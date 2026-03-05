package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
	"github.com/user/micro-dp/usecase"
)

const (
	credOAuthStateCookieName    = "micro-dp-cred-oauth-state"
	credOAuthVerifierCookieName = "micro-dp-cred-oauth-verifier"
)

type CredentialHandler struct {
	credentials *usecase.CredentialService
}

func NewCredentialHandler(credentials *usecase.CredentialService) *CredentialHandler {
	return &CredentialHandler{credentials: credentials}
}

func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	credentials, err := h.credentials.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if credentials == nil {
		credentials = []domain.Credential{}
	}

	items := make([]openapi.Credential, len(credentials))
	for i := range credentials {
		items[i] = toOpenAPICredential(&credentials[i])
	}

	writeJSON(w, http.StatusOK, struct {
		Items []openapi.Credential `json:"items"`
	}{Items: items})
}

func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := h.credentials.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrCredentialNotFound) {
			writeError(w, http.StatusNotFound, "credential not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CredentialHandler) OAuthStart(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	if !h.credentials.ProviderEnabled(providerName) {
		log.Printf("credential_oauth_start failed: unsupported or unconfigured provider %q", providerName)
		writeError(w, http.StatusBadRequest, "unsupported or unconfigured provider")
		return
	}

	userID, ok := domain.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID, ok := domain.TenantIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing tenant")
		return
	}

	codeVerifier, codeChallenge, err := usecase.GeneratePKCE()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to initialize pkce")
		return
	}

	authURL, err := h.credentials.BuildAuthURL(providerName, userID, tenantID, codeChallenge)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build oauth url")
		return
	}

	secure := requestSecure(r)
	http.SetCookie(w, &http.Cookie{Name: credOAuthVerifierCookieName, Value: codeVerifier, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: 600})

	log.Printf("credential_oauth_start success provider=%s", providerName)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *CredentialHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	if !h.credentials.ProviderEnabled(providerName) {
		writeError(w, http.StatusBadRequest, "unsupported or unconfigured provider")
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		log.Printf("credential_oauth_callback failed: missing code/state provider=%s", providerName)
		h.redirectCredentialFailure(w, r, providerName, "missing oauth code or state")
		return
	}

	verifierCookie, err := r.Cookie(credOAuthVerifierCookieName)
	if err != nil || verifierCookie.Value == "" {
		log.Printf("credential_oauth_callback failed: verifier missing provider=%s", providerName)
		h.redirectCredentialFailure(w, r, providerName, "invalid oauth verifier")
		return
	}

	if err := h.credentials.CompleteOAuth(r.Context(), providerName, code, verifierCookie.Value, state); err != nil {
		log.Printf("credential_oauth_callback failed: %v provider=%s", err, providerName)
		h.redirectCredentialFailure(w, r, providerName, "credential oauth failed")
		return
	}

	secure := requestSecure(r)
	clearCredentialOAuthCookies(w, secure)

	log.Printf("credential_oauth_callback success provider=%s", providerName)
	redirectURL := h.credentials.ProviderPostRedirectURL(providerName) + "?linked=true"
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *CredentialHandler) redirectCredentialFailure(w http.ResponseWriter, r *http.Request, providerName, reason string) {
	secure := requestSecure(r)
	clearCredentialOAuthCookies(w, secure)

	failureURL := h.credentials.ProviderPostRedirectURL(providerName)
	if parsed, err := url.Parse(failureURL); err == nil {
		q := parsed.Query()
		q.Set("error", reason)
		parsed.RawQuery = q.Encode()
		failureURL = parsed.String()
	}

	http.Redirect(w, r, failureURL, http.StatusFound)
}

func clearCredentialOAuthCookies(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: credOAuthStateCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: credOAuthVerifierCookieName, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: secure, MaxAge: -1})
}
