package domain

import (
	"context"
	"errors"
	"time"
)

var ErrDatasetNotFound = errors.New("dataset not found")

const (
	SourceTypeTracker = "tracker"
	SourceTypeParquet = "parquet"
	SourceTypeImport  = "import"
)

type Dataset struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Name          string     `json:"name"`
	SourceType    string     `json:"source_type"`
	SchemaJSON    *string    `json:"schema_json,omitempty"`
	RowCount      *int64     `json:"row_count,omitempty"`
	StoragePath   string     `json:"storage_path"`
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type DatasetListFilter struct {
	Query      string
	SourceType string
	Limit      int
	Offset     int
}

type DatasetRepository interface {
	FindByID(ctx context.Context, tenantID, id string) (*Dataset, error)
	ListByTenant(ctx context.Context, tenantID string, filter DatasetListFilter) ([]Dataset, error)
	Create(ctx context.Context, d *Dataset) error
	Update(ctx context.Context, d *Dataset) error
	Upsert(ctx context.Context, d *Dataset) error
}
