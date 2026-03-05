package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrWriteKeyNotFound = errors.New("write key not found")
	ErrWriteKeyInvalid  = errors.New("invalid write key")
)

type WriteKey struct {
	ID        string
	TenantID  string
	Name      string
	KeyHash   string
	KeyPrefix string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WriteKeyRepository interface {
	Create(ctx context.Context, wk *WriteKey) error
	FindByID(ctx context.Context, tenantID, id string) (*WriteKey, error)
	FindByKeyHash(ctx context.Context, keyHash string) (*WriteKey, error)
	FindByTenantID(ctx context.Context, tenantID string) ([]WriteKey, error)
	Delete(ctx context.Context, tenantID, id string) error
	UpdateKeyHash(ctx context.Context, tenantID, id, keyHash, keyPrefix string) error
}
