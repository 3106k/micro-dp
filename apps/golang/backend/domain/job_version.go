package domain

import (
	"context"
	"errors"
	"time"
)

const (
	JobVersionStatusDraft     = "draft"
	JobVersionStatusPublished = "published"
)

var (
	ErrJobVersionNotFound  = errors.New("job version not found")
	ErrJobVersionImmutable = errors.New("published version cannot be modified")
)

type JobVersion struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	JobID       string     `json:"job_id"`
	Version     int        `json:"version"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type JobVersionRepository interface {
	Create(ctx context.Context, v *JobVersion) error
	FindByID(ctx context.Context, tenantID, id string) (*JobVersion, error)
	ListByJobID(ctx context.Context, tenantID, jobID string) ([]JobVersion, error)
	Publish(ctx context.Context, tenantID, id string) error
	NextVersion(ctx context.Context, jobID string) (int, error)
}
