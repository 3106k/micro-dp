package domain

import (
	"context"
	"errors"
)

var ErrUploadAlreadyProcessed = errors.New("upload already processed")

type UploadJobFile struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type UploadJobMessage struct {
	UploadID string          `json:"upload_id"`
	TenantID string          `json:"tenant_id"`
	Files    []UploadJobFile `json:"files"`
}

type UploadJobQueue interface {
	Enqueue(ctx context.Context, msg *UploadJobMessage) error
	Dequeue(ctx context.Context) (*UploadJobMessage, error)
	MarkProcessed(ctx context.Context, uploadID string) error
	EnqueueDLQ(ctx context.Context, msg *UploadJobMessage, reason string) error
}
