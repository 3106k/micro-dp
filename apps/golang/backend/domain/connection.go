package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrConnectionNotFound      = errors.New("connection not found")
	ErrConnectionNameDuplicate = errors.New("connection name already exists")
)

type Connection struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	ConfigJSON string    `json:"config_json"`
	SecretRef  *string   `json:"secret_ref,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ConnectionRepository interface {
	Create(ctx context.Context, c *Connection) error
	FindByID(ctx context.Context, tenantID, id string) (*Connection, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Connection, error)
	Update(ctx context.Context, c *Connection) error
	Delete(ctx context.Context, tenantID, id string) error
}
