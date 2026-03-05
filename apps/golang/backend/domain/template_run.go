package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTemplateRunNotFound = errors.New("template run not found")
)

type TemplateRun struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	TemplateType string    `json:"template_type"`
	Status       string    `json:"status"`
	SkipReason   *string   `json:"skip_reason,omitempty"`
	DashboardID  *string   `json:"dashboard_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type TemplateRunRepository interface {
	Create(ctx context.Context, tr *TemplateRun) error
	FindByID(ctx context.Context, tenantID, id string) (*TemplateRun, error)
	ListByTenant(ctx context.Context, tenantID string) ([]TemplateRun, error)
}
