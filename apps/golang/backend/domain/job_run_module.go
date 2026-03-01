package domain

import (
	"context"
	"errors"
	"time"
)

const (
	ModuleStatusQueued   = "queued"
	ModuleStatusRunning  = "running"
	ModuleStatusSuccess  = "success"
	ModuleStatusFailed   = "failed"
	ModuleStatusCanceled = "canceled"
)

var ValidModuleStatuses = []string{
	ModuleStatusQueued,
	ModuleStatusRunning,
	ModuleStatusSuccess,
	ModuleStatusFailed,
	ModuleStatusCanceled,
}

var (
	ErrJobRunModuleNotFound = errors.New("job run module not found")
	ErrInvalidModuleStatus  = errors.New("invalid module status")
)

type JobRunModule struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	JobRunID     string     `json:"job_run_id"`
	JobModuleID  string     `json:"job_module_id"`
	Status       string     `json:"status"`
	Attempt      int        `json:"attempt"`
	InputJSON    *string    `json:"input_json,omitempty"`
	OutputJSON   *string    `json:"output_json,omitempty"`
	MetricsJSON  *string    `json:"metrics_json,omitempty"`
	ErrorCode    *string    `json:"error_code,omitempty"`
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
	UpdateStatus(ctx context.Context, tenantID, id, status string, errorCode, errorMessage *string, startedAt, finishedAt *time.Time) error
}
