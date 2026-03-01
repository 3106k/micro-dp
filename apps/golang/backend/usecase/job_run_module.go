package usecase

import (
	"context"
	"fmt"

	"github.com/user/micro-dp/domain"
)

type JobRunModuleService struct {
	runModules domain.JobRunModuleRepository
}

func NewJobRunModuleService(runModules domain.JobRunModuleRepository) *JobRunModuleService {
	return &JobRunModuleService{runModules: runModules}
}

func (s *JobRunModuleService) Get(ctx context.Context, id string) (*domain.JobRunModule, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.runModules.FindByID(ctx, tenantID, id)
}

func (s *JobRunModuleService) ListByJobRun(ctx context.Context, jobRunID string) ([]domain.JobRunModule, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.runModules.ListByJobRunID(ctx, tenantID, jobRunID)
}
