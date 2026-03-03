package domain

import (
	"context"
	"errors"
	"time"
)

const (
	StatusQueued   = "queued"
	StatusRunning  = "running"
	StatusSuccess  = "success"
	StatusFailed   = "failed"
	StatusCanceled = "canceled"
)

var (
	ErrJobRunNotFound         = errors.New("job run not found")
	ErrJobRunAlreadyProcessed = errors.New("job run already processed")
	ErrNoPublishedVersion     = errors.New("no published version available")
)

// RunSnapshot captures all information needed to execute a job run.
type RunSnapshot struct {
	JobKind   string              `json:"job_kind"`
	JobID     string              `json:"job_id"`
	VersionID string              `json:"version_id"`
	Modules   []RunSnapshotModule `json:"modules"`
	Edges     []RunSnapshotEdge   `json:"edges"`
}

type RunSnapshotModule struct {
	ID           string  `json:"id"`
	ModuleTypeID string  `json:"module_type_id"`
	Category     string  `json:"category"`
	Name         string  `json:"name"`
	ConfigJSON   string  `json:"config_json"`
	ConnectionID *string `json:"connection_id,omitempty"`
}

type RunSnapshotEdge struct {
	SourceModuleID string `json:"source_module_id"`
	TargetModuleID string `json:"target_module_id"`
}

// JobRunMessage is the message sent through the job run queue.
type JobRunMessage struct {
	JobRunID string `json:"job_run_id"`
	TenantID string `json:"tenant_id"`
}

// JobRunQueue defines the interface for the job run execution queue.
type JobRunQueue interface {
	Enqueue(ctx context.Context, msg *JobRunMessage) error
	Dequeue(ctx context.Context) (*JobRunMessage, error)
	MarkProcessed(ctx context.Context, jobRunID string) error
	EnqueueDLQ(ctx context.Context, msg *JobRunMessage, reason string) error
}

type JobRun struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	JobID           string     `json:"job_id"`
	JobVersionID    *string    `json:"job_version_id,omitempty"`
	Status          string     `json:"status"`
	RunSnapshotJSON *string    `json:"run_snapshot_json,omitempty"`
	CheckpointJSON  *string    `json:"checkpoint_json,omitempty"`
	ProgressJSON    *string    `json:"progress_json,omitempty"`
	Attempt         int        `json:"attempt"`
	NextRunAt       *time.Time `json:"next_run_at,omitempty"`
	LastError       *string    `json:"last_error,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type JobRunRepository interface {
	Create(ctx context.Context, jr *JobRun) error
	FindByID(ctx context.Context, tenantID, id string) (*JobRun, error)
	ListByTenant(ctx context.Context, tenantID string) ([]JobRun, error)
	ListReady(ctx context.Context) ([]JobRun, error)
	UpdateStatus(ctx context.Context, tenantID, id, status string) error
	UpdateStarted(ctx context.Context, id string) error
	UpdateFailed(ctx context.Context, id, lastError string) error
}
