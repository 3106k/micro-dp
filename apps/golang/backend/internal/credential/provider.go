package credential

import (
	"context"

	"golang.org/x/oauth2"
)

// OAuthProvider abstracts an OAuth 2.0 credential provider (e.g., Google).
type OAuthProvider interface {
	// ProviderName returns the canonical name (e.g., "google").
	ProviderName() string
	// Scopes returns the OAuth scopes.
	Scopes() []string
	// AuthCodeURL builds the authorization URL with PKCE.
	AuthCodeURL(state, codeChallenge string) string
	// Exchange trades the authorization code for tokens.
	Exchange(ctx context.Context, code, codeVerifier string) (*oauth2.Token, error)
	// FetchLabel fetches a human-readable label (e.g., email).
	FetchLabel(ctx context.Context, token *oauth2.Token) (string, error)
	// RefreshToken refreshes an expired access token.
	RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error)
	// OAuthEnabled returns true if this provider is configured.
	OAuthEnabled() bool
	// PostRedirectURL returns the URL to redirect after OAuth completion.
	PostRedirectURL() string
}
