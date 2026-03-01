package usecase

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type MeteringService struct {
	usage domain.UsageRepository
}

func NewMeteringService(usage domain.UsageRepository) *MeteringService {
	return &MeteringService{usage: usage}
}

func today() string {
	return time.Now().UTC().Format("2006-01-02")
}

// RecordEvents records event ingest usage (called by Worker, best-effort).
func (s *MeteringService) RecordEvents(ctx context.Context, tenantID string, count int) error {
	date := today()
	if err := s.usage.IncrementEvents(ctx, tenantID, date, count); err != nil {
		return err
	}
	return s.usage.RecordEvent(ctx, &domain.UsageEvent{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		EventType: domain.UsageTypeEventsIngest,
		Delta:     count,
	})
}

// RecordUpload records upload rows + storage usage (called by Worker, best-effort).
func (s *MeteringService) RecordUpload(ctx context.Context, tenantID string, rowCount int, storageBytes int64) error {
	date := today()
	if err := s.usage.IncrementRows(ctx, tenantID, date, rowCount); err != nil {
		return err
	}
	if err := s.usage.IncrementStorage(ctx, tenantID, date, storageBytes); err != nil {
		return err
	}
	return s.usage.RecordEvent(ctx, &domain.UsageEvent{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		EventType: domain.UsageTypeStorageWrite,
		Delta:     rowCount,
	})
}

// RecordUploadCount increments the daily upload counter by 1.
func (s *MeteringService) RecordUploadCount(ctx context.Context, tenantID string) error {
	date := today()
	if err := s.usage.IncrementUploads(ctx, tenantID, date, 1); err != nil {
		return err
	}
	return s.usage.RecordEvent(ctx, &domain.UsageEvent{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		EventType: domain.UsageTypeUploadComplete,
		Delta:     1,
	})
}

// RecordEventsBestEffort logs errors but does not return them.
func (s *MeteringService) RecordEventsBestEffort(ctx context.Context, tenantID string, count int) {
	if err := s.RecordEvents(ctx, tenantID, count); err != nil {
		log.Printf("metering record events error tenant=%s: %v", tenantID, err)
	}
}

// RecordUploadBestEffort logs errors but does not return them.
func (s *MeteringService) RecordUploadBestEffort(ctx context.Context, tenantID string, rowCount int, storageBytes int64) {
	if err := s.RecordUpload(ctx, tenantID, rowCount, storageBytes); err != nil {
		log.Printf("metering record upload error tenant=%s: %v", tenantID, err)
	}
}
