package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserIdentityNotFound = errors.New("user identity not found")
	ErrUserIdentityExists   = errors.New("user identity already exists")
)

type UserIdentity struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Subject   string    `json:"subject"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserIdentityRepository interface {
	Create(ctx context.Context, identity *UserIdentity) error
	FindByProviderSubject(ctx context.Context, provider, subject string) (*UserIdentity, error)
}
