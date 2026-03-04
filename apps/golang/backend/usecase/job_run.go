package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type JobRunService struct {
	jobRuns     domain.JobRunRepository
	jobs        domain.JobRepository
	versions    domain.JobVersionRepository
	modules     domain.JobModuleRepository
	edges       domain.JobModuleEdgeRepository
	moduleTypes domain.ModuleTypeRepository
}

func NewJobRunService(
	jobRuns domain.JobRunRepository,
	jobs domain.JobRepository,
	versions domain.JobVersionRepository,
	modules domain.JobModuleRepository,
	edges domain.JobModuleEdgeRepository,
	moduleTypes domain.ModuleTypeRepository,
) *JobRunService {
	return &JobRunService{
		jobRuns:     jobRuns,
		jobs:        jobs,
		versions:    versions,
		modules:     modules,
		edges:       edges,
		moduleTypes: moduleTypes,
	}
}

func (s *JobRunService) Create(ctx context.Context, jobID string, jobVersionID *string) (*domain.JobRun, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify job exists and get kind
	job, err := s.jobs.FindByID(ctx, tenantID, jobID)
	if err != nil {
		return nil, err
	}

	// Resolve version: use provided or auto-select latest published
	var versionID string
	if jobVersionID != nil && *jobVersionID != "" {
		versionID = *jobVersionID
		if _, err := s.versions.FindByID(ctx, tenantID, versionID); err != nil {
			return nil, err
		}
	} else {
		versions, err := s.versions.ListByJobID(ctx, tenantID, jobID)
		if err != nil {
			return nil, fmt.Errorf("list versions: %w", err)
		}
		var latest *domain.JobVersion
		for i := range versions {
			if versions[i].Status == domain.JobVersionStatusPublished {
				if latest == nil || versions[i].Version > latest.Version {
					latest = &versions[i]
				}
			}
		}
		if latest == nil {
			return nil, domain.ErrNoPublishedVersion
		}
		versionID = latest.ID
	}

	// Build RunSnapshot
	modules, err := s.modules.ListByJobVersionID(ctx, tenantID, versionID)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}

	edges, err := s.edges.ListByJobVersionID(ctx, tenantID, versionID)
	if err != nil {
		return nil, fmt.Errorf("list edges: %w", err)
	}

	snapshotModules := make([]domain.RunSnapshotModule, len(modules))
	for i, m := range modules {
		mt, err := s.moduleTypes.FindByID(ctx, tenantID, m.ModuleTypeID)
		if err != nil {
			return nil, fmt.Errorf("find module type %s: %w", m.ModuleTypeID, err)
		}
		snapshotModules[i] = domain.RunSnapshotModule{
			ID:           m.ID,
			ModuleTypeID: m.ModuleTypeID,
			Category:     mt.Category,
			Name:         m.Name,
			ConfigJSON:   m.ConfigJSON,
			ConnectionID: m.ConnectionID,
		}
	}

	snapshotEdges := make([]domain.RunSnapshotEdge, len(edges))
	for i, e := range edges {
		snapshotEdges[i] = domain.RunSnapshotEdge{
			SourceModuleID: e.SourceModuleID,
			TargetModuleID: e.TargetModuleID,
		}
	}

	snapshot := domain.RunSnapshot{
		JobKind:   job.Kind,
		JobID:     job.ID,
		VersionID: versionID,
		Modules:   snapshotModules,
		Edges:     snapshotEdges,
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}
	snapshotStr := string(snapshotJSON)

	jr := &domain.JobRun{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		JobID:           jobID,
		JobVersionID:    &versionID,
		Status:          domain.StatusQueued,
		RunSnapshotJSON: &snapshotStr,
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
