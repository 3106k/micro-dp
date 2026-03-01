package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
)

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserTenant struct {
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	FindByID(ctx context.Context, id string) (*Tenant, error)
	ListAll(ctx context.Context) ([]Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	AddUserToTenant(ctx context.Context, ut *UserTenant) error
	ListByUserID(ctx context.Context, userID string) ([]Tenant, error)
	IsUserInTenant(ctx context.Context, userID, tenantID string) (bool, error)
}
