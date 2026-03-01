package domain

import (
	"context"
	"errors"
	"time"
)

var ErrQuotaExceeded = errors.New("quota exceeded")

const (
	UsageTypeEventsIngest  = "events_ingest"
	UsageTypeUploadComplete = "upload_complete"
	UsageTypeStorageWrite  = "storage_write"
)

type UsageDaily struct {
	ID           string
	TenantID     string
	Date         string
	EventsCount  int
	StorageBytes int64
	RowsCount    int
	UploadsCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UsageEvent struct {
	ID         string
	TenantID   string
	EventType  string
	Delta      int
	RecordedAt time.Time
}

type UsageSummary struct {
	TenantID     string
	Date         string
	EventsCount  int
	StorageBytes int64
	RowsCount    int
	UploadsCount int
	Plan         *Plan
}

type UsageRepository interface {
	IncrementEvents(ctx context.Context, tenantID, date string, delta int) error
	IncrementStorage(ctx context.Context, tenantID, date string, deltaBytes int64) error
	IncrementRows(ctx context.Context, tenantID, date string, delta int) error
	IncrementUploads(ctx context.Context, tenantID, date string, delta int) error
	FindDailyByTenantAndDate(ctx context.Context, tenantID, date string) (*UsageDaily, error)
	RecordEvent(ctx context.Context, e *UsageEvent) error
}
