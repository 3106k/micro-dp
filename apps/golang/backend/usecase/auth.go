package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/notification"
)

const (
	googleOIDCIssuer = "https://accounts.google.com"
	googleProvider   = "google"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrOAuthNotConfigured = errors.New("oauth not configured")
	ErrInvalidOAuthState  = errors.New("invalid oauth state")
	ErrInvalidOAuthToken  = errors.New("invalid oauth token")
)

type GoogleOAuthConfig struct {
	ClientID            string
	ClientSecret        string
	RedirectURL         string
	PostLoginRedirect   string
	PostFailureRedirect string
}

type AuthService struct {
	users       domain.UserRepository
	identities  domain.UserIdentityRepository
	tenants     domain.TenantRepository
	jwtSecret   []byte
	emailSender notification.EmailSender
	googleOAuth GoogleOAuthConfig
}

func NewAuthService(
	users domain.UserRepository,
	identities domain.UserIdentityRepository,
	tenants domain.TenantRepository,
	jwtSecret string,
	emailSender notification.EmailSender,
	googleOAuth GoogleOAuthConfig,
) *AuthService {
	return &AuthService{
		users:       users,
		identities:  identities,
		tenants:     tenants,
		jwtSecret:   []byte(jwtSecret),
		emailSender: emailSender,
		googleOAuth: googleOAuth,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (userID, tenantID string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	userID = uuid.New().String()
	user := &domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
		PlatformRole: domain.PlatformRoleUser,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return "", "", err
	}

	tenantID = uuid.New().String()
	tenant := &domain.Tenant{
		ID:       tenantID,
		Name:     displayName + "'s Workspace",
		IsActive: true,
	}
	if err := s.tenants.Create(ctx, tenant); err != nil {
		return "", "", err
	}

	ut := &domain.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     "owner",
	}
	if err := s.tenants.AddUserToTenant(ctx, ut); err != nil {
		return "", "", err
	}

	if s.emailSender != nil {
		subject, html, text, err := notification.RenderWelcome(displayName)
		if err == nil {
			if sendErr := s.emailSender.Send(ctx, &notification.EmailMessage{
				To: email, Subject: subject, HTML: html, Text: text,
			}); sendErr != nil {
				log.Printf("welcome email failed to=%s: %v", email, sendErr)
			}
		}
	}

	return userID, tenantID, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (token string, err error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.issueToken(user.ID)
}

func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, []domain.Tenant, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	tenants, err := s.tenants.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return user, tenants, nil
}

func (s *AuthService) OAuthEnabled() bool {
	return s.googleOAuth.ClientID != "" &&
		s.googleOAuth.ClientSecret != "" &&
		s.googleOAuth.RedirectURL != ""
}

func (s *AuthService) GoogleRedirectURL() string {
	return s.googleOAuth.RedirectURL
}

func (s *AuthService) PostLoginRedirectURL() string {
	if s.googleOAuth.PostLoginRedirect != "" {
		return s.googleOAuth.PostLoginRedirect
	}
	return "http://localhost:3000/dashboard"
}

func (s *AuthService) PostFailureRedirectURL() string {
	if s.googleOAuth.PostFailureRedirect != "" {
		return s.googleOAuth.PostFailureRedirect
	}
	return "http://localhost:3000/signin?error=oauth_failed"
}

func (s *AuthService) BuildGoogleAuthURL(state, nonce, codeChallenge string) (string, error) {
	if !s.OAuthEnabled() {
		return "", ErrOAuthNotConfigured
	}

	cfg := oauth2.Config{
		ClientID:     s.googleOAuth.ClientID,
		ClientSecret: s.googleOAuth.ClientSecret,
		RedirectURL:  s.googleOAuth.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile"},
	}

	url := cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	return url, nil
}

func (s *AuthService) CompleteGoogleOAuth(
	ctx context.Context,
	code string,
	codeVerifier string,
	expectedNonce string,
) (token string, defaultTenantID string, err error) {
	if !s.OAuthEnabled() {
		return "", "", ErrOAuthNotConfigured
	}

	cfg := oauth2.Config{
		ClientID:     s.googleOAuth.ClientID,
		ClientSecret: s.googleOAuth.ClientSecret,
		RedirectURL:  s.googleOAuth.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile"},
	}

	oauthToken, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return "", "", fmt.Errorf("oauth exchange: %w", err)
	}

	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", "", ErrInvalidOAuthToken
	}

	verifier := oidc.NewVerifier(
		googleOIDCIssuer,
		oidc.NewRemoteKeySet(ctx, "https://www.googleapis.com/oauth2/v3/certs"),
		&oidc.Config{ClientID: s.googleOAuth.ClientID},
	)
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", "", fmt.Errorf("verify id token: %w", err)
	}

	var claims struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Nonce         string `json:"nonce"`
		Issuer        string `json:"iss"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", "", fmt.Errorf("parse id token claims: %w", err)
	}

	if claims.Nonce == "" || claims.Nonce != expectedNonce {
		return "", "", ErrInvalidOAuthState
	}
	if claims.Issuer != googleOIDCIssuer && claims.Issuer != "accounts.google.com" {
		return "", "", ErrInvalidOAuthToken
	}
	if claims.Subject == "" || claims.Email == "" || !claims.EmailVerified {
		return "", "", ErrInvalidOAuthToken
	}

	userID, err := s.findOrCreateGoogleUser(ctx, claims.Subject, claims.Email, claims.Name)
	if err != nil {
		return "", "", err
	}

	token, err = s.issueToken(userID)
	if err != nil {
		return "", "", err
	}

	tenants, err := s.tenants.ListByUserID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if len(tenants) > 0 {
		defaultTenantID = tenants[0].ID
	}

	return token, defaultTenantID, nil
}

func (s *AuthService) GenerateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *AuthService) GenerateOAuthNonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *AuthService) GeneratePKCEVerifier() (string, string, error) {
	verifier, err := s.GenerateOAuthState()
	if err != nil {
		return "", "", err
	}
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return verifier, challenge, nil
}

func (s *AuthService) findOrCreateGoogleUser(ctx context.Context, subject, email, name string) (string, error) {
	identity, err := s.identities.FindByProviderSubject(ctx, googleProvider, subject)
	if err == nil {
		return identity.UserID, nil
	}
	if !errors.Is(err, domain.ErrUserIdentityNotFound) {
		return "", err
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			return "", err
		}

		displayName := strings.TrimSpace(name)
		if displayName == "" {
			displayName = strings.Split(email, "@")[0]
		}

		newUserID := uuid.New().String()
		if err := s.users.Create(ctx, &domain.User{
			ID:           newUserID,
			Email:        email,
			PasswordHash: "",
			DisplayName:  displayName,
			PlatformRole: domain.PlatformRoleUser,
		}); err != nil {
			return "", err
		}

		tenantID := uuid.New().String()
		if err := s.tenants.Create(ctx, &domain.Tenant{
			ID:       tenantID,
			Name:     displayName + "'s Workspace",
			IsActive: true,
		}); err != nil {
			return "", err
		}
		if err := s.tenants.AddUserToTenant(ctx, &domain.UserTenant{
			UserID:   newUserID,
			TenantID: tenantID,
			Role:     "owner",
		}); err != nil {
			return "", err
		}

		user = &domain.User{ID: newUserID, Email: email}
	}

	if err := s.identities.Create(ctx, &domain.UserIdentity{
		ID:       uuid.New().String(),
		Provider: googleProvider,
		Subject:  subject,
		UserID:   user.ID,
		Email:    email,
	}); err != nil {
		if errors.Is(err, domain.ErrUserIdentityExists) {
			existing, findErr := s.identities.FindByProviderSubject(ctx, googleProvider, subject)
			if findErr == nil {
				return existing.UserID, nil
			}
		}
		return "", err
	}

	return user.ID, nil
}

func (s *AuthService) issueToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}
	return signed, nil
}
