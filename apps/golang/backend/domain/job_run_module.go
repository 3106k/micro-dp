package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrJobRunModuleNotFound = errors.New("job run module not found")
)

type JobRunModule struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	JobRunID     string     `json:"job_run_id"`
	JobModuleID  string     `json:"job_module_id"`
	Status       string     `json:"status"`
	InputJSON    *string    `json:"input_json,omitempty"`
	OutputJSON   *string    `json:"output_json,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type JobRunModuleRepository interface {
	Create(ctx context.Context, m *JobRunModule) error
	FindByID(ctx context.Context, tenantID, id string) (*JobRunModule, error)
	ListByJobRunID(ctx context.Context, tenantID, jobRunID string) ([]JobRunModule, error)
	UpdateStatus(ctx context.Context, tenantID, id, status string) error
}
