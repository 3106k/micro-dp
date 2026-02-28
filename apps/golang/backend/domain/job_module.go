package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrJobModuleNotFound = errors.New("job module not found")
)

type JobModule struct {
	ID                  string    `json:"id"`
	TenantID            string    `json:"tenant_id"`
	JobVersionID        string    `json:"job_version_id"`
	ModuleTypeID        string    `json:"module_type_id"`
	ModuleTypeSchemaID  *string   `json:"module_type_schema_id,omitempty"`
	ConnectionID        *string   `json:"connection_id,omitempty"`
	Name                string    `json:"name"`
	ConfigJSON          string    `json:"config_json"`
	PositionX           float64   `json:"position_x"`
	PositionY           float64   `json:"position_y"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type JobModuleEdge struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	JobVersionID   string    `json:"job_version_id"`
	SourceModuleID string    `json:"source_module_id"`
	TargetModuleID string    `json:"target_module_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type JobModuleRepository interface {
	Create(ctx context.Context, m *JobModule) error
	FindByID(ctx context.Context, tenantID, id string) (*JobModule, error)
	ListByJobVersionID(ctx context.Context, tenantID, jobVersionID string) ([]JobModule, error)
}

type JobModuleEdgeRepository interface {
	Create(ctx context.Context, e *JobModuleEdge) error
	ListByJobVersionID(ctx context.Context, tenantID, jobVersionID string) ([]JobModuleEdge, error)
}
