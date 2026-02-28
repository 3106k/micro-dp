package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type JobRunService struct {
	jobRuns domain.JobRunRepository
	jobs    domain.JobRepository
}

func NewJobRunService(jobRuns domain.JobRunRepository, jobs domain.JobRepository) *JobRunService {
	return &JobRunService{jobRuns: jobRuns, jobs: jobs}
}

func (s *JobRunService) Create(ctx context.Context, jobID string, jobVersionID *string) (*domain.JobRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify job exists
	if _, err := s.jobs.FindByID(ctx, tenantID, jobID); err != nil {
		return nil, err
	}

	jr := &domain.JobRun{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		JobID:        jobID,
		JobVersionID: jobVersionID,
		Status:       domain.StatusQueued,
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
