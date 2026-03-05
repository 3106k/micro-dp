package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrChartNotFound = errors.New("chart not found")
)

type Chart struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	ChartType  string    `json:"chart_type"`
	DatasetID  string    `json:"dataset_id"`
	Measure    string    `json:"measure"`
	Dimension  string    `json:"dimension"`
	ConfigJSON *string   `json:"config_json,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ChartRepository interface {
	Create(ctx context.Context, c *Chart) error
	FindByID(ctx context.Context, tenantID, id string) (*Chart, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Chart, error)
	Update(ctx context.Context, c *Chart) error
	Delete(ctx context.Context, tenantID, id string) error
}
