package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type JobRunService struct {
	jobRuns domain.JobRunRepository
}

func NewJobRunService(jobRuns domain.JobRunRepository) *JobRunService {
	return &JobRunService{jobRuns: jobRuns}
}

func (s *JobRunService) Create(ctx context.Context, projectID, jobID string) (*domain.JobRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	jr := &domain.JobRun{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		ProjectID: projectID,
		JobID:     jobID,
		Status:    "queued",
	}

	if err := s.jobRuns.Create(ctx, jr); err != nil {
		return nil, err
	}

	return s.jobRuns.FindByID(ctx, tenantID, jr.ID)
}

func (s *JobRunService) Get(ctx context.Context, id string) (*domain.JobRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	return s.jobRuns.FindByID(ctx, tenantID, id)
}

func (s *JobRunService) List(ctx context.Context) ([]domain.JobRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	return s.jobRuns.ListByTenant(ctx, tenantID)
}
