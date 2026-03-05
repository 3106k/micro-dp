package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDashboardNotFound = errors.New("dashboard not found")
	ErrWidgetNotFound    = errors.New("widget not found")
	ErrWidgetDuplicate   = errors.New("widget already exists for this chart on this dashboard")
)

type Dashboard struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DashboardWidget struct {
	ID          string    `json:"id"`
	DashboardID string    `json:"dashboard_id"`
	ChartID     string    `json:"chart_id"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
}

type DashboardRepository interface {
	Create(ctx context.Context, d *Dashboard) error
	FindByID(ctx context.Context, tenantID, id string) (*Dashboard, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Dashboard, error)
	Update(ctx context.Context, d *Dashboard) error
	Delete(ctx context.Context, tenantID, id string) error
}

type DashboardWidgetRepository interface {
	Create(ctx context.Context, w *DashboardWidget) error
	FindByID(ctx context.Context, dashboardID, id string) (*DashboardWidget, error)
	ListByDashboard(ctx context.Context, dashboardID string) ([]DashboardWidget, error)
	Delete(ctx context.Context, dashboardID, id string) error
}
