package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrJobRunNotFound = errors.New("job run not found")
)

type JobRun struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	ProjectID      string     `json:"project_id"`
	JobID          string     `json:"job_id"`
	Status         string     `json:"status"`
	CheckpointJSON *string    `json:"checkpoint_json,omitempty"`
	ProgressJSON   *string    `json:"progress_json,omitempty"`
	Attempt        int        `json:"attempt"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	LastError      *string    `json:"last_error,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type JobRunRepository interface {
	Create(ctx context.Context, jr *JobRun) error
	FindByID(ctx context.Context, tenantID, id string) (*JobRun, error)
	ListByTenant(ctx context.Context, tenantID string) ([]JobRun, error)
}
