package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrDatasetNotFound  = errors.New("dataset not found")
	ErrColumnNotFound   = errors.New("column not found")
)

const (
	SourceTypeTracker   = "tracker"
	SourceTypeParquet   = "parquet"
	SourceTypeImport    = "import"
	SourceTypeTransform = "transform"
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

type DatasetColumnMeta struct {
	Name         string           `json:"column_name"`
	Type         string           `json:"column_type"`
	Nullable     bool             `json:"nullable,omitempty"`
	Description  string           `json:"description,omitempty"`
	SemanticType string           `json:"semantic_type,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
	SampleValues []interface{}    `json:"sample_values,omitempty"`
	Statistics   *ColumnStatistics `json:"statistics,omitempty"`
}

type ColumnStatistics struct {
	Min           *string `json:"min,omitempty"`
	Max           *string `json:"max,omitempty"`
	NullRate      float64 `json:"null_rate"`
	DistinctCount int64   `json:"distinct_count"`
}

// ParseColumns parses schema_json into []DatasetColumnMeta.
func (d *Dataset) ParseColumns() ([]DatasetColumnMeta, error) {
	if d.SchemaJSON == nil || *d.SchemaJSON == "" {
		return nil, nil
	}
	var cols []DatasetColumnMeta
	if err := json.Unmarshal([]byte(*d.SchemaJSON), &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

// SetColumns serializes []DatasetColumnMeta into schema_json.
func (d *Dataset) SetColumns(cols []DatasetColumnMeta) error {
	data, err := json.Marshal(cols)
	if err != nil {
		return err
	}
	s := string(data)
	d.SchemaJSON = &s
	return nil
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
