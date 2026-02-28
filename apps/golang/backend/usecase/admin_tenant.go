package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
)

type AdminTenantService struct {
	tenants   domain.TenantRepository
	auditLogs domain.AdminAuditLogRepository
}

func NewAdminTenantService(tenants domain.TenantRepository, auditLogs domain.AdminAuditLogRepository) *AdminTenantService {
	return &AdminTenantService{
		tenants:   tenants,
		auditLogs: auditLogs,
	}
}

func (s *AdminTenantService) CreateTenant(ctx context.Context, actorUserID, name string) (*domain.Tenant, error) {
	tenant := &domain.Tenant{
		ID:       uuid.New().String(),
		Name:     name,
		IsActive: true,
	}
	if err := s.tenants.Create(ctx, tenant); err != nil {
		return nil, err
	}

	if err := s.tenants.AddUserToTenant(ctx, &domain.UserTenant{
		UserID:   actorUserID,
		TenantID: tenant.ID,
		Role:     "owner",
	}); err != nil {
		return nil, err
	}

	meta, err := json.Marshal(map[string]any{
		"name":      tenant.Name,
		"is_active": tenant.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal audit metadata: %w", err)
	}
	if err := s.auditLogs.Create(ctx, &domain.AdminAuditLog{
		ID:           uuid.New().String(),
		ActorUserID:  actorUserID,
		Action:       "tenant.create",
		TargetType:   "tenant",
		TargetID:     tenant.ID,
		MetadataJSON: string(meta),
	}); err != nil {
		return nil, err
	}

	return s.tenants.FindByID(ctx, tenant.ID)
}

func (s *AdminTenantService) ListTenants(ctx context.Context) ([]domain.Tenant, error) {
	return s.tenants.ListAll(ctx)
}

func (s *AdminTenantService) UpdateTenant(ctx context.Context, actorUserID, tenantID string, name *string, isActive *bool) (*domain.Tenant, error) {
	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if name != nil {
		tenant.Name = *name
	}
	if isActive != nil {
		tenant.IsActive = *isActive
	}

	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}

	meta, err := json.Marshal(map[string]any{
		"name":      tenant.Name,
		"is_active": tenant.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal audit metadata: %w", err)
	}
	if err := s.auditLogs.Create(ctx, &domain.AdminAuditLog{
		ID:           uuid.New().String(),
		ActorUserID:  actorUserID,
		Action:       "tenant.update",
		TargetType:   "tenant",
		TargetID:     tenant.ID,
		MetadataJSON: string(meta),
	}); err != nil {
		return nil, err
	}

	return s.tenants.FindByID(ctx, tenant.ID)
}
