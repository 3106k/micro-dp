package domain

import (
	"context"
	"errors"
	"time"
)

const (
	ModuleTypeCategorySource      = "source"
	ModuleTypeCategoryTransform   = "transform"
	ModuleTypeCategoryDestination = "destination"
)

var (
	ErrModuleTypeNotFound      = errors.New("module type not found")
	ErrModuleTypeNameDuplicate = errors.New("module type name already exists")
)

type ModuleType struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ModuleTypeSchema struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	ModuleTypeID string    `json:"module_type_id"`
	Version      int       `json:"version"`
	JSONSchema   string    `json:"json_schema"`
	CreatedAt    time.Time `json:"created_at"`
}

type ModuleTypeRepository interface {
	Create(ctx context.Context, mt *ModuleType) error
	FindByID(ctx context.Context, tenantID, id string) (*ModuleType, error)
	ListByTenant(ctx context.Context, tenantID string) ([]ModuleType, error)
}

type ModuleTypeSchemaRepository interface {
	Create(ctx context.Context, s *ModuleTypeSchema) error
	FindByID(ctx context.Context, tenantID, id string) (*ModuleTypeSchema, error)
	ListByModuleTypeID(ctx context.Context, tenantID, moduleTypeID string) ([]ModuleTypeSchema, error)
	NextVersion(ctx context.Context, moduleTypeID string) (int, error)
}
