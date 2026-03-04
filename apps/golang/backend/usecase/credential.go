package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/user/micro-dp/domain"
)

type GoogleCredentialOAuthConfig struct {
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	PostRedirectURL string
}

type CredentialService struct {
	credentials domain.CredentialRepository
	oauthCfg    GoogleCredentialOAuthConfig
	hmacSecret  []byte
}

func NewCredentialService(
	credentials domain.CredentialRepository,
	oauthCfg GoogleCredentialOAuthConfig,
	jwtSecret string,
) *CredentialService {
	return &CredentialService{
		credentials: credentials,
		oauthCfg:    oauthCfg,
		hmacSecret:  []byte(jwtSecret),
	}
}

func (s *CredentialService) OAuthEnabled() bool {
	return s.oauthCfg.ClientID != "" &&
		s.oauthCfg.ClientSecret != "" &&
		s.oauthCfg.RedirectURL != ""
}

func (s *CredentialService) RedirectURL() string {
	return s.oauthCfg.RedirectURL
}

func (s *CredentialService) PostRedirectURL() string {
	if s.oauthCfg.PostRedirectURL != "" {
		return s.oauthCfg.PostRedirectURL
	}
	return "http://localhost:3000/integrations"
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

func (s *CredentialService) BuildGoogleCredentialAuthURL(userID, tenantID, codeChallenge string) (string, error) {
	if !s.OAuthEnabled() {
		return "", ErrOAuthNotConfigured
	}

	state := s.signState(userID, tenantID)

	cfg := s.googleOAuth2Config()
	url := cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
	return url, nil
}

func (s *CredentialService) CompleteGoogleCredentialOAuth(ctx context.Context, code, codeVerifier, state string) error {
	if !s.OAuthEnabled() {
		return ErrOAuthNotConfigured
	}

	userID, tenantID, err := s.verifyState(state)
	if err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	cfg := s.googleOAuth2Config()
	oauthToken, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return fmt.Errorf("oauth exchange: %w", err)
	}

	// Extract email from userinfo endpoint for the label
	label := ""
	client := cfg.Client(ctx, oauthToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err == nil {
		defer resp.Body.Close()
		var userInfo struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err == nil && userInfo.Email != "" {
			label = userInfo.Email
		}
	}

	// Upsert: update if exists, create otherwise
	existing, err := s.credentials.FindByUserAndProvider(ctx, userID, tenantID, googleProvider)
	if err == nil {
		existing.AccessToken = oauthToken.AccessToken
		existing.RefreshToken = oauthToken.RefreshToken
		if !oauthToken.Expiry.IsZero() {
			existing.TokenExpiry = &oauthToken.Expiry
		}
		existing.Scopes = strings.Join(cfg.Scopes, " ")
		if label != "" {
			existing.ProviderLabel = label
		}
		return s.credentials.Update(ctx, existing)
	}

	cred := &domain.Credential{
		ID:           uuid.New().String(),
		UserID:       userID,
		TenantID:     tenantID,
		Provider:     googleProvider,
		ProviderLabel: label,
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
		Scopes:       strings.Join(cfg.Scopes, " "),
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

	cfg := s.googleOAuth2Config()
	token := &oauth2.Token{
		AccessToken:  cred.AccessToken,
		RefreshToken: cred.RefreshToken,
		Expiry:       time.Time{},
	}
	newToken, err := cfg.TokenSource(ctx, token).Token()
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

func (s *CredentialService) googleOAuth2Config() oauth2.Config {
	return oauth2.Config{
		ClientID:     s.oauthCfg.ClientID,
		ClientSecret: s.oauthCfg.ClientSecret,
		RedirectURL:  s.oauthCfg.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/spreadsheets.readonly"},
	}
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
