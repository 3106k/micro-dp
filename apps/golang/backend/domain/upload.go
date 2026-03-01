package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUploadNotFound       = errors.New("upload not found")
	ErrUploadAlreadyComplete = errors.New("upload already complete")
)

const (
	UploadStatusPresigned = "presigned"
	UploadStatusUploaded  = "uploaded"
)

type Upload struct {
	ID        string
	TenantID  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UploadFile struct {
	ID          string
	TenantID    string
	UploadID    string
	FileName    string
	ObjectKey   string
	ContentType string
	SizeBytes   int64
	CreatedAt   time.Time
}

type UploadRepository interface {
	CreateUpload(ctx context.Context, u *Upload) error
	CreateUploadFile(ctx context.Context, f *UploadFile) error
	FindByID(ctx context.Context, tenantID, id string) (*Upload, error)
	FindFilesByUploadID(ctx context.Context, tenantID, uploadID string) ([]UploadFile, error)
	UpdateStatus(ctx context.Context, tenantID, id, status string) error
}

type PresignedURLGenerator interface {
	GeneratePresignedPutURL(ctx context.Context, objectKey, contentType string, expiry time.Duration) (string, time.Time, error)
}
