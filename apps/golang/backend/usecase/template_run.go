package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
)

type TemplateRunService struct {
	templateRuns domain.TemplateRunRepository
}

func NewTemplateRunService(templateRuns domain.TemplateRunRepository) *TemplateRunService {
	return &TemplateRunService{templateRuns: templateRuns}
}

func (s *TemplateRunService) Create(ctx context.Context, templateType string) (*domain.TemplateRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	skipReason := "not yet implemented"
	tr := &domain.TemplateRun{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		TemplateType: templateType,
		Status:       "skipped",
		SkipReason:   &skipReason,
	}
	if err := s.templateRuns.Create(ctx, tr); err != nil {
		return nil, err
	}
	return s.templateRuns.FindByID(ctx, tenantID, tr.ID)
}

func (s *TemplateRunService) Get(ctx context.Context, id string) (*domain.TemplateRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.templateRuns.FindByID(ctx, tenantID, id)
}

func (s *TemplateRunService) List(ctx context.Context) ([]domain.TemplateRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.templateRuns.ListByTenant(ctx, tenantID)
}
