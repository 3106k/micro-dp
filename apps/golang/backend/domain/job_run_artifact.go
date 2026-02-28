package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrJobRunArtifactNotFound = errors.New("job run artifact not found")
)

type JobRunArtifact struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	JobRunID        string    `json:"job_run_id"`
	JobRunModuleID  *string   `json:"job_run_module_id,omitempty"`
	Name            string    `json:"name"`
	ArtifactType    string    `json:"artifact_type"`
	StoragePath     string    `json:"storage_path"`
	SizeBytes       int64     `json:"size_bytes"`
	ContentType     string    `json:"content_type"`
	CreatedAt       time.Time `json:"created_at"`
}

type JobRunArtifactRepository interface {
	Create(ctx context.Context, a *JobRunArtifact) error
	FindByID(ctx context.Context, tenantID, id string) (*JobRunArtifact, error)
	ListByJobRunID(ctx context.Context, tenantID, jobRunID string) ([]JobRunArtifact, error)
}
