package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPlanNotFound       = errors.New("plan not found")
	ErrTenantPlanNotFound = errors.New("tenant plan not found")
)

type Plan struct {
	ID               string
	Name             string
	DisplayName      string
	MaxEventsPerDay  int
	MaxStorageBytes  int64
	MaxRowsPerDay    int
	MaxUploadsPerDay int
	IsDefault        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type TenantPlan struct {
	ID        string
	TenantID  string
	PlanID    string
	StartedAt time.Time
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PlanRepository interface {
	FindByID(ctx context.Context, id string) (*Plan, error)
	FindByName(ctx context.Context, name string) (*Plan, error)
	FindDefault(ctx context.Context) (*Plan, error)
	ListAll(ctx context.Context) ([]Plan, error)
	Create(ctx context.Context, p *Plan) error
	Update(ctx context.Context, p *Plan) error
}

type TenantPlanRepository interface {
	FindByTenantID(ctx context.Context, tenantID string) (*TenantPlan, error)
	Upsert(ctx context.Context, tp *TenantPlan) error
}
