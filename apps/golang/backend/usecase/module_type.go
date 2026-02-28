package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type ModuleTypeService struct {
	moduleTypes domain.ModuleTypeRepository
	schemas     domain.ModuleTypeSchemaRepository
}

func NewModuleTypeService(
	moduleTypes domain.ModuleTypeRepository,
	schemas domain.ModuleTypeSchemaRepository,
) *ModuleTypeService {
	return &ModuleTypeService{
		moduleTypes: moduleTypes,
		schemas:     schemas,
	}
}

func (s *ModuleTypeService) Create(ctx context.Context, name, category string) (*domain.ModuleType, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	mt := &domain.ModuleType{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		Name:     name,
		Category: category,
	}

	if err := s.moduleTypes.Create(ctx, mt); err != nil {
		return nil, err
	}

	return s.moduleTypes.FindByID(ctx, tenantID, mt.ID)
}

func (s *ModuleTypeService) Get(ctx context.Context, id string) (*domain.ModuleType, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.moduleTypes.FindByID(ctx, tenantID, id)
}

func (s *ModuleTypeService) List(ctx context.Context) ([]domain.ModuleType, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.moduleTypes.ListByTenant(ctx, tenantID)
}

func (s *ModuleTypeService) CreateSchema(ctx context.Context, moduleTypeID, jsonSchema string) (*domain.ModuleTypeSchema, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify module type exists
	if _, err := s.moduleTypes.FindByID(ctx, tenantID, moduleTypeID); err != nil {
		return nil, err
	}

	nextVer, err := s.schemas.NextVersion(ctx, moduleTypeID)
	if err != nil {
		return nil, err
	}

	schema := &domain.ModuleTypeSchema{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		ModuleTypeID: moduleTypeID,
		Version:      nextVer,
		JSONSchema:   jsonSchema,
	}

	if err := s.schemas.Create(ctx, schema); err != nil {
		return nil, err
	}

	return schema, nil
}

func (s *ModuleTypeService) ListSchemas(ctx context.Context, moduleTypeID string) ([]domain.ModuleTypeSchema, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.schemas.ListByModuleTypeID(ctx, tenantID, moduleTypeID)
}
