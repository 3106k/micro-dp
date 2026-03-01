package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

const (
	tokenCookieName         = "micro-dp-token"
	tenantCookieName        = "micro-dp-tenant-id"
	oauthStateCookieName    = "micro-dp-oauth-state"
	oauthNonceCookieName    = "micro-dp-oauth-nonce"
	oauthVerifierCookieName = "micro-dp-oauth-verifier"
)

type AuthHandler struct {
	auth *usecase.AuthService
}

func NewAuthHandler(auth *usecase.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req openapi.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if string(req.Email) == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	displayName := ""
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}

	userID, tenantID, err := h.auth.Register(r.Context(), string(req.Email), req.Password, displayName)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, openapi.RegisterResponse{
		UserId:   userID,
		TenantId: tenantID,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req openapi.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if string(req.Email) == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	token, err := h.auth.Login(r.Context(), string(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, openapi.LoginResponse{Token: token})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := domain.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, tenants, err := h.auth.Me(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	oaTenants := make([]tenantResponse, len(tenants))
	for i, t := range tenants {
		oaTenants[i] = toOpenAPITenant(t)
	}

	writeJSON(w, http.StatusOK, struct {
		UserID       string           `json:"user_id"`
		Email        string           `json:"email"`
		DisplayName  string           `json:"display_name"`
		PlatformRole string           `json:"platform_role"`
		Tenants      []tenantResponse `json:"tenants"`
	}{
		UserID:       user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		PlatformRole: user.PlatformRole,
		Tenants:      oaTenants,
	})
}

func (h *AuthHandler) GoogleStart(w http.ResponseWriter, r *http.Request) {
	if !h.auth.OAuthEnabled() {
		log.Printf("auth_google_start failed: oauth not configured")
		writeError(w, http.StatusInternalServerError, "google oauth is not configured")
		return
	}

	state, err := h.auth.GenerateOAuthState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initialize oauth state")
		return
	}
	nonce, err := h.auth.GenerateOAuthNonce()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initialize oauth nonce")
		return
	}
	codeVerifier, codeChallenge, err := h.auth.GeneratePKCEVerifier()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initialize oauth pkce")
		return
	}

	authURL, err := h.auth.BuildGoogleAuthURL(state, nonce, codeChallenge)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build oauth url")
		return
	}

	secure := requestSecure(r)
	http.SetCookie(w, &http.Cookie{Name: oauthStateCookieName, Value: state, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: 600})
	http.SetCookie(w, &http.Cookie{Name: oauthNonceCookieName, Value: nonce, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: 600})
	http.SetCookie(w, &http.Cookie{Name: oauthVerifierCookieName, Value: codeVerifier, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: 600})

	log.Printf("auth_google_start success")
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.auth.OAuthEnabled() {
		writeError(w, http.StatusInternalServerError, "google oauth is not configured")
		return
	}

	if !redirectURIMatchesRequest(r, h.auth.GoogleRedirectURL()) {
		log.Printf("auth_google_callback failed: redirect uri mismatch")
		writeError(w, http.StatusBadRequest, "invalid redirect uri")
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		log.Printf("auth_google_callback failed: missing code/state")
		writeError(w, http.StatusBadRequest, "missing oauth code or state")
		return
	}

	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != state {
		log.Printf("auth_google_callback failed: state mismatch")
		h.redirectOAuthFailure(w, r, "invalid oauth state")
		return
	}

	nonceCookie, err := r.Cookie(oauthNonceCookieName)
	if err != nil || nonceCookie.Value == "" {
		log.Printf("auth_google_callback failed: nonce missing")
		h.redirectOAuthFailure(w, r, "invalid oauth nonce")
		return
	}
	verifierCookie, err := r.Cookie(oauthVerifierCookieName)
	if err != nil || verifierCookie.Value == "" {
		log.Printf("auth_google_callback failed: verifier missing")
		h.redirectOAuthFailure(w, r, "invalid oauth verifier")
		return
	}

	token, tenantID, err := h.auth.CompleteGoogleOAuth(r.Context(), code, verifierCookie.Value, nonceCookie.Value)
	if err != nil {
		log.Printf("auth_google_callback failed: %v", err)
		h.redirectOAuthFailure(w, r, "google oauth failed")
		return
	}

	secure := requestSecure(r)
	http.SetCookie(w, &http.Cookie{Name: tokenCookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: int((24 * time.Hour).Seconds())})
	if tenantID != "" {
		http.SetCookie(w, &http.Cookie{Name: tenantCookieName, Value: tenantID, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: int((24 * time.Hour).Seconds())})
	}
	clearOAuthCookies(w, secure)

	log.Printf("auth_google_callback success")
	http.Redirect(w, r, h.auth.PostLoginRedirectURL(), http.StatusFound)
}

func (h *AuthHandler) redirectOAuthFailure(w http.ResponseWriter, r *http.Request, reason string) {
	secure := requestSecure(r)
	clearOAuthCookies(w, secure)

	failureURL := h.auth.PostFailureRedirectURL()
	if parsed, err := url.Parse(failureURL); err == nil {
		q := parsed.Query()
		q.Set("reason", reason)
		parsed.RawQuery = q.Encode()
		failureURL = parsed.String()
	}

	http.Redirect(w, r, failureURL, http.StatusFound)
}

func clearOAuthCookies(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: oauthStateCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: oauthNonceCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: oauthVerifierCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: -1})
}

func requestSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func redirectURIMatchesRequest(r *http.Request, expected string) bool {
	expectedURL, err := url.Parse(expected)
	if err != nil {
		return false
	}
	scheme := "http"
	if requestSecure(r) {
		scheme = "https"
	}
	actual := &url.URL{Scheme: scheme, Host: r.Host, Path: r.URL.Path}
	return strings.EqualFold(actual.Scheme, expectedURL.Scheme) && strings.EqualFold(actual.Host, expectedURL.Host) && actual.Path == expectedURL.Path
}
