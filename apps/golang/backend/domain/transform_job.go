package domain

import (
	"context"
	"errors"
)

var ErrTransformAlreadyProcessed = errors.New("transform already processed")

type TransformJobMessage struct {
	JobRunID   string   `json:"job_run_id"`
	TenantID   string   `json:"tenant_id"`
	SQL        string   `json:"sql"`
	DatasetIDs []string `json:"dataset_ids"`
	JobID      string   `json:"job_id"`
	VersionID  string   `json:"version_id"`
}

type TransformJobQueue interface {
	Enqueue(ctx context.Context, msg *TransformJobMessage) error
	Dequeue(ctx context.Context) (*TransformJobMessage, error)
	MarkProcessed(ctx context.Context, jobRunID string) error
	EnqueueDLQ(ctx context.Context, msg *TransformJobMessage, reason string) error
}
