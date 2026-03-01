package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

const (
	maxFileSize     = 100 * 1024 * 1024 // 100 MB
	maxFilesPerReq  = 10
	presignExpiry   = 15 * time.Minute
)

var allowedExtensions = map[string]bool{
	".csv":     true,
	".json":    true,
	".parquet": true,
	".xlsx":    true,
	".txt":     true,
	".tsv":     true,
	".gz":      true,
	".zip":     true,
}

type UploadFileInput struct {
	Filename    string
	ContentType string
	SizeBytes   int64
}

type UploadPresignedFile struct {
	FileID       string
	Filename     string
	PresignedURL string
	ObjectKey    string
	ExpiresAt    time.Time
}

type UploadPresignResult struct {
	UploadID string
	Files    []UploadPresignedFile
}

type UploadService struct {
	uploads   domain.UploadRepository
	presigner domain.PresignedURLGenerator
	queue     domain.UploadJobQueue
}

func NewUploadService(uploads domain.UploadRepository, presigner domain.PresignedURLGenerator, queue domain.UploadJobQueue) *UploadService {
	return &UploadService{uploads: uploads, presigner: presigner, queue: queue}
}

func (s *UploadService) CreatePresign(ctx context.Context, files []UploadFileInput) (*UploadPresignResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("at least one file is required")
	}
	if len(files) > maxFilesPerReq {
		return nil, fmt.Errorf("too many files: max %d", maxFilesPerReq)
	}

	for _, f := range files {
		if f.SizeBytes <= 0 {
			return nil, fmt.Errorf("invalid size for file %q", f.Filename)
		}
		if f.SizeBytes > maxFileSize {
			return nil, fmt.Errorf("file %q exceeds max size %d bytes", f.Filename, maxFileSize)
		}
		ext := strings.ToLower(filepath.Ext(f.Filename))
		if !allowedExtensions[ext] {
			return nil, fmt.Errorf("file extension %q is not allowed", ext)
		}
	}

	uploadID := uuid.New().String()
	upload := &domain.Upload{
		ID:       uploadID,
		TenantID: tenantID,
		Status:   domain.UploadStatusPresigned,
	}
	if err := s.uploads.CreateUpload(ctx, upload); err != nil {
		return nil, fmt.Errorf("create upload: %w", err)
	}

	datePart := time.Now().UTC().Format("2006-01-02")
	result := &UploadPresignResult{UploadID: uploadID}

	for _, f := range files {
		fileID := uuid.New().String()
		ext := strings.ToLower(filepath.Ext(f.Filename))
		objectKey := fmt.Sprintf("uploads/%s/%s/%s%s", tenantID, datePart, fileID, ext)

		presignedURL, expiresAt, err := s.presigner.GeneratePresignedPutURL(ctx, objectKey, f.ContentType, presignExpiry)
		if err != nil {
			return nil, fmt.Errorf("generate presigned url: %w", err)
		}

		uf := &domain.UploadFile{
			ID:          fileID,
			TenantID:    tenantID,
			UploadID:    uploadID,
			FileName:    f.Filename,
			ObjectKey:   objectKey,
			ContentType: f.ContentType,
			SizeBytes:   f.SizeBytes,
		}
		if err := s.uploads.CreateUploadFile(ctx, uf); err != nil {
			return nil, fmt.Errorf("create upload file: %w", err)
		}

		result.Files = append(result.Files, UploadPresignedFile{
			FileID:       fileID,
			Filename:     f.Filename,
			PresignedURL: presignedURL,
			ObjectKey:    objectKey,
			ExpiresAt:    expiresAt,
		})
	}

	return result, nil
}

func (s *UploadService) Complete(ctx context.Context, uploadID string) (*domain.Upload, []domain.UploadFile, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("tenant id not found in context")
	}

	upload, err := s.uploads.FindByID(ctx, tenantID, uploadID)
	if err != nil {
		return nil, nil, err
	}

	if upload.Status == domain.UploadStatusUploaded {
		return nil, nil, domain.ErrUploadAlreadyComplete
	}

	if err := s.uploads.UpdateStatus(ctx, tenantID, uploadID, domain.UploadStatusUploaded); err != nil {
		return nil, nil, fmt.Errorf("update status: %w", err)
	}

	upload, err = s.uploads.FindByID(ctx, tenantID, uploadID)
	if err != nil {
		return nil, nil, err
	}

	files, err := s.uploads.FindFilesByUploadID(ctx, tenantID, uploadID)
	if err != nil {
		return nil, nil, fmt.Errorf("find files: %w", err)
	}

	// Enqueue upload job for CSV→Parquet conversion
	jobMsg := &domain.UploadJobMessage{
		UploadID: uploadID,
		TenantID: tenantID,
		Files:    toJobFiles(files),
	}
	if err := s.queue.Enqueue(ctx, jobMsg); err != nil {
		return nil, nil, fmt.Errorf("enqueue upload job: %w", err)
	}

	return upload, files, nil
}

func toJobFiles(files []domain.UploadFile) []domain.UploadJobFile {
	result := make([]domain.UploadJobFile, len(files))
	for i, f := range files {
		result[i] = domain.UploadJobFile{
			FileID:      f.ID,
			FileName:    f.FileName,
			ObjectKey:   f.ObjectKey,
			ContentType: f.ContentType,
			SizeBytes:   f.SizeBytes,
		}
	}
	return result
}
