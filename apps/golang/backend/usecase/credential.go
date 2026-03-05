package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/credential"
)

// ErrUnsupportedProvider is returned when the requested provider is not registered.
var ErrUnsupportedProvider = fmt.Errorf("unsupported or unconfigured provider")

type CredentialService struct {
	credentials domain.CredentialRepository
	providers   map[string]credential.OAuthProvider
	hmacSecret  []byte
}

func NewCredentialService(
	credentials domain.CredentialRepository,
	providers []credential.OAuthProvider,
	jwtSecret string,
) *CredentialService {
	m := make(map[string]credential.OAuthProvider, len(providers))
	for _, p := range providers {
		m[p.ProviderName()] = p
	}
	return &CredentialService{
		credentials: credentials,
		providers:   m,
		hmacSecret:  []byte(jwtSecret),
	}
}

// GetProvider returns the OAuthProvider for the given name.
func (s *CredentialService) GetProvider(name string) (credential.OAuthProvider, error) {
	p, ok := s.providers[name]
	if !ok {
		return nil, ErrUnsupportedProvider
	}
	return p, nil
}

// ProviderEnabled returns true if the named provider is registered and configured.
func (s *CredentialService) ProviderEnabled(name string) bool {
	p, ok := s.providers[name]
	return ok && p.OAuthEnabled()
}

// ProviderPostRedirectURL returns the post-redirect URL for the named provider.
func (s *CredentialService) ProviderPostRedirectURL(name string) string {
	p, ok := s.providers[name]
	if !ok {
		return "http://localhost:3000/integrations"
	}
	return p.PostRedirectURL()
}

func (s *CredentialService) List(ctx context.Context) ([]domain.Credential, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.credentials.ListByTenant(ctx, tenantID)
}

func (s *CredentialService) Delete(ctx context.Context, id string) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	if _, err := s.credentials.FindByID(ctx, tenantID, id); err != nil {
		return err
	}

	return s.credentials.Delete(ctx, tenantID, id)
}

func (s *CredentialService) BuildAuthURL(providerName, userID, tenantID, codeChallenge string) (string, error) {
	p, ok := s.providers[providerName]
	if !ok || !p.OAuthEnabled() {
		return "", ErrOAuthNotConfigured
	}

	state := s.signState(userID, tenantID)
	url := p.AuthCodeURL(state, codeChallenge)
	return url, nil
}

func (s *CredentialService) CompleteOAuth(ctx context.Context, providerName, code, codeVerifier, state string) error {
	p, ok := s.providers[providerName]
	if !ok || !p.OAuthEnabled() {
		return ErrOAuthNotConfigured
	}

	userID, tenantID, err := s.verifyState(state)
	if err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	oauthToken, err := p.Exchange(ctx, code, codeVerifier)
	if err != nil {
		return err
	}

	// Fetch label (e.g., email) — best effort
	label, _ := p.FetchLabel(ctx, oauthToken)

	scopes := strings.Join(p.Scopes(), " ")

	// Upsert: update if exists, create otherwise
	existing, err := s.credentials.FindByUserAndProvider(ctx, userID, tenantID, providerName)
	if err == nil {
		existing.AccessToken = oauthToken.AccessToken
		existing.RefreshToken = oauthToken.RefreshToken
		if !oauthToken.Expiry.IsZero() {
			existing.TokenExpiry = &oauthToken.Expiry
		}
		existing.Scopes = scopes
		if label != "" {
			existing.ProviderLabel = label
		}
		return s.credentials.Update(ctx, existing)
	}

	cred := &domain.Credential{
		ID:            uuid.New().String(),
		UserID:        userID,
		TenantID:      tenantID,
		Provider:      providerName,
		ProviderLabel: label,
		AccessToken:   oauthToken.AccessToken,
		RefreshToken:  oauthToken.RefreshToken,
		Scopes:        scopes,
	}
	if !oauthToken.Expiry.IsZero() {
		cred.TokenExpiry = &oauthToken.Expiry
	}

	return s.credentials.Create(ctx, cred)
}

// GetValidAccessToken returns a valid access token for a credential, refreshing if needed.
func (s *CredentialService) GetValidAccessToken(ctx context.Context, tenantID, credentialID string) (string, error) {
	cred, err := s.credentials.FindByID(ctx, tenantID, credentialID)
	if err != nil {
		return "", err
	}

	if cred.TokenExpiry != nil && time.Now().Before(*cred.TokenExpiry) {
		return cred.AccessToken, nil
	}

	if cred.RefreshToken == "" {
		return cred.AccessToken, nil
	}

	p, ok := s.providers[cred.Provider]
	if !ok {
		return "", fmt.Errorf("no provider registered for %q", cred.Provider)
	}

	newToken, err := p.RefreshToken(ctx, cred.RefreshToken)
	if err != nil {
		log.Printf("credential_refresh failed cred=%s: %v", credentialID, err)
		return "", fmt.Errorf("token refresh failed: %w", err)
	}

	cred.AccessToken = newToken.AccessToken
	if newToken.RefreshToken != "" {
		cred.RefreshToken = newToken.RefreshToken
	}
	if !newToken.Expiry.IsZero() {
		cred.TokenExpiry = &newToken.Expiry
	}
	if err := s.credentials.Update(ctx, cred); err != nil {
		log.Printf("credential_update failed cred=%s: %v", credentialID, err)
	}

	return newToken.AccessToken, nil
}

// GeneratePKCE generates a PKCE verifier and challenge pair.
func GeneratePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return verifier, challenge, nil
}

// signState encodes user_id:tenant_id with HMAC signature.
func (s *CredentialService) signState(userID, tenantID string) string {
	payload := userID + ":" + tenantID
	mac := hmac.New(sha256.New, s.hmacSecret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
}

// verifyState verifies the HMAC signature and returns user_id, tenant_id.
func (s *CredentialService) verifyState(state string) (string, string, error) {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed state")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", fmt.Errorf("decode payload: %w", err)
	}
	payload := string(payloadBytes)

	mac := hmac.New(sha256.New, s.hmacSecret)
	mac.Write(payloadBytes)
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return "", "", fmt.Errorf("invalid signature")
	}

	p := strings.SplitN(payload, ":", 2)
	if len(p) != 2 || p[0] == "" || p[1] == "" {
		return "", "", fmt.Errorf("invalid payload")
	}

	return p[0], p[1], nil
}
