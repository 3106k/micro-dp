package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCredentialNotFound      = errors.New("credential not found")
	ErrCredentialAlreadyExists = errors.New("credential already exists")
)

type Credential struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	TenantID      string     `json:"tenant_id"`
	Provider      string     `json:"provider"`
	ProviderLabel string     `json:"provider_label"`
	AccessToken   string     `json:"access_token"`
	RefreshToken  string     `json:"refresh_token"`
	TokenExpiry   *time.Time `json:"token_expiry,omitempty"`
	Scopes        string     `json:"scopes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CredentialRepository interface {
	Create(ctx context.Context, c *Credential) error
	FindByID(ctx context.Context, tenantID, id string) (*Credential, error)
	FindByUserAndProvider(ctx context.Context, userID, tenantID, provider string) (*Credential, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Credential, error)
	Update(ctx context.Context, c *Credential) error
	Delete(ctx context.Context, tenantID, id string) error
}
