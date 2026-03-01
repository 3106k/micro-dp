package usecase

import (
	"context"
	"fmt"

	"github.com/user/micro-dp/domain"
)

type JobRunArtifactService struct {
	runArtifacts domain.JobRunArtifactRepository
}

func NewJobRunArtifactService(runArtifacts domain.JobRunArtifactRepository) *JobRunArtifactService {
	return &JobRunArtifactService{runArtifacts: runArtifacts}
}

func (s *JobRunArtifactService) Get(ctx context.Context, id string) (*domain.JobRunArtifact, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.runArtifacts.FindByID(ctx, tenantID, id)
}

func (s *JobRunArtifactService) ListByJobRun(ctx context.Context, jobRunID string) ([]domain.JobRunArtifact, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.runArtifacts.ListByJobRunID(ctx, tenantID, jobRunID)
}
