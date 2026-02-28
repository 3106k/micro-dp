package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrJobNotFound      = errors.New("job not found")
	ErrJobSlugDuplicate = errors.New("job slug already exists")
)

type Job struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	FindByID(ctx context.Context, tenantID, id string) (*Job, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Job, error)
	Update(ctx context.Context, job *Job) error
}
