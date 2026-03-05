package credential

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var defaultGoogleScopes = []string{
	"https://www.googleapis.com/auth/spreadsheets.readonly",
}

// GoogleConfig holds configuration for the Google OAuth provider.
type GoogleConfig struct {
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	PostRedirectURI string
	Scopes          []string // defaults to spreadsheets.readonly if empty
}

// GoogleProvider implements OAuthProvider for Google.
type GoogleProvider struct {
	cfg GoogleConfig
}

// NewGoogleProvider creates a new GoogleProvider.
func NewGoogleProvider(cfg GoogleConfig) *GoogleProvider {
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = defaultGoogleScopes
	}
	return &GoogleProvider{cfg: cfg}
}

func (p *GoogleProvider) ProviderName() string {
	return "google"
}

func (p *GoogleProvider) Scopes() []string {
	return p.cfg.Scopes
}

func (p *GoogleProvider) OAuthEnabled() bool {
	return p.cfg.ClientID != "" && p.cfg.ClientSecret != "" && p.cfg.RedirectURL != ""
}

func (p *GoogleProvider) PostRedirectURL() string {
	if p.cfg.PostRedirectURI != "" {
		return p.cfg.PostRedirectURI
	}
	return "http://localhost:3000/integrations"
}

func (p *GoogleProvider) oauth2Config() oauth2.Config {
	return oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		RedirectURL:  p.cfg.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       p.cfg.Scopes,
	}
}

func (p *GoogleProvider) AuthCodeURL(state, codeChallenge string) string {
	cfg := p.oauth2Config()
	return cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code, codeVerifier string) (*oauth2.Token, error) {
	cfg := p.oauth2Config()
	token, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("oauth exchange: %w", err)
	}
	return token, nil
}

func (p *GoogleProvider) FetchLabel(ctx context.Context, token *oauth2.Token) (string, error) {
	cfg := p.oauth2Config()
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return "", fmt.Errorf("fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("decode userinfo: %w", err)
	}
	return userInfo.Email, nil
}

func (p *GoogleProvider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	cfg := p.oauth2Config()
	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Time{}, // force refresh
	}
	newToken, err := cfg.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh: %w", err)
	}
	return newToken, nil
}
