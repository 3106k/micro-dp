package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type TxRunner interface {
	RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error
}

type JobModuleRepoFactory interface {
	WithTx(tx *sql.Tx) domain.JobModuleRepository
}

type JobModuleEdgeRepoFactory interface {
	WithTx(tx *sql.Tx) domain.JobModuleEdgeRepository
}

type JobService struct {
	jobs              domain.JobRepository
	versions          domain.JobVersionRepository
	modules           domain.JobModuleRepository
	edges             domain.JobModuleEdgeRepository
	moduleTypeSchemas domain.ModuleTypeSchemaRepository
	txRunner          TxRunner
}

func NewJobService(
	jobs domain.JobRepository,
	versions domain.JobVersionRepository,
	modules domain.JobModuleRepository,
	edges domain.JobModuleEdgeRepository,
	moduleTypeSchemas domain.ModuleTypeSchemaRepository,
	txRunner TxRunner,
) *JobService {
	return &JobService{
		jobs:              jobs,
		versions:          versions,
		modules:           modules,
		edges:             edges,
		moduleTypeSchemas: moduleTypeSchemas,
		txRunner:          txRunner,
	}
}

func (s *JobService) CreateJob(ctx context.Context, name, slug, description string) (*domain.Job, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	job := &domain.Job{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        name,
		Slug:        slug,
		Description: description,
		IsActive:    true,
	}

	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, err
	}

	return s.jobs.FindByID(ctx, tenantID, job.ID)
}

func (s *JobService) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.jobs.FindByID(ctx, tenantID, id)
}

func (s *JobService) ListJobs(ctx context.Context) ([]domain.Job, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.jobs.ListByTenant(ctx, tenantID)
}

func (s *JobService) UpdateJob(ctx context.Context, id, name, slug, description string, isActive bool) (*domain.Job, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	job, err := s.jobs.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	job.Name = name
	job.Slug = slug
	job.Description = description
	job.IsActive = isActive

	if err := s.jobs.Update(ctx, job); err != nil {
		return nil, err
	}

	return s.jobs.FindByID(ctx, tenantID, id)
}

type CreateModuleInput struct {
	ModuleTypeID       string
	ModuleTypeSchemaID *string
	ConnectionID       *string
	Name               string
	ConfigJSON         string
	PositionX          float64
	PositionY          float64
}

type CreateEdgeInput struct {
	SourceModuleIndex int
	TargetModuleIndex int
}

func (s *JobService) CreateVersion(ctx context.Context, jobID string, modules []CreateModuleInput, edges []CreateEdgeInput) (*domain.JobVersion, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify job exists
	if _, err := s.jobs.FindByID(ctx, tenantID, jobID); err != nil {
		return nil, err
	}

	// Validate module_type_schema_id belongs to the specified module_type_id
	for _, m := range modules {
		if m.ModuleTypeSchemaID == nil {
			continue
		}
		schema, err := s.moduleTypeSchemas.FindByID(ctx, tenantID, *m.ModuleTypeSchemaID)
		if err != nil {
			return nil, fmt.Errorf("module_type_schema %s not found: %w", *m.ModuleTypeSchemaID, err)
		}
		if schema.ModuleTypeID != m.ModuleTypeID {
			return nil, fmt.Errorf("schema %s does not belong to module_type %s", *m.ModuleTypeSchemaID, m.ModuleTypeID)
		}
	}

	nextVer, err := s.versions.NextVersion(ctx, jobID)
	if err != nil {
		return nil, err
	}

	version := &domain.JobVersion{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		JobID:    jobID,
		Version:  nextVer,
		Status:   domain.JobVersionStatusDraft,
	}

	if err := s.versions.Create(ctx, version); err != nil {
		return nil, err
	}

	// Create modules
	moduleIDs := make([]string, len(modules))
	for i, m := range modules {
		mod := &domain.JobModule{
			ID:                 uuid.New().String(),
			TenantID:           tenantID,
			JobVersionID:       version.ID,
			ModuleTypeID:       m.ModuleTypeID,
			ModuleTypeSchemaID: m.ModuleTypeSchemaID,
			ConnectionID:       m.ConnectionID,
			Name:               m.Name,
			ConfigJSON:         m.ConfigJSON,
			PositionX:          m.PositionX,
			PositionY:          m.PositionY,
		}
		if err := s.modules.Create(ctx, mod); err != nil {
			return nil, err
		}
		moduleIDs[i] = mod.ID
	}

	// Create edges
	for _, e := range edges {
		if e.SourceModuleIndex < 0 || e.SourceModuleIndex >= len(moduleIDs) ||
			e.TargetModuleIndex < 0 || e.TargetModuleIndex >= len(moduleIDs) {
			return nil, fmt.Errorf("invalid module index in edge")
		}
		edge := &domain.JobModuleEdge{
			ID:             uuid.New().String(),
			TenantID:       tenantID,
			JobVersionID:   version.ID,
			SourceModuleID: moduleIDs[e.SourceModuleIndex],
			TargetModuleID: moduleIDs[e.TargetModuleIndex],
		}
		if err := s.edges.Create(ctx, edge); err != nil {
			return nil, err
		}
	}

	return s.versions.FindByID(ctx, tenantID, version.ID)
}

func (s *JobService) ListVersions(ctx context.Context, jobID string) ([]domain.JobVersion, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.versions.ListByJobID(ctx, tenantID, jobID)
}

func (s *JobService) GetVersionDetail(ctx context.Context, jobID, versionID string) (*domain.JobVersion, []domain.JobModule, []domain.JobModuleEdge, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, nil, nil, fmt.Errorf("tenant id not found in context")
	}

	v, err := s.versions.FindByID(ctx, tenantID, versionID)
	if err != nil {
		return nil, nil, nil, err
	}
	if v.JobID != jobID {
		return nil, nil, nil, domain.ErrJobVersionNotFound
	}

	mods, err := s.modules.ListByJobVersionID(ctx, tenantID, versionID)
	if err != nil {
		return nil, nil, nil, err
	}

	edgeList, err := s.edges.ListByJobVersionID(ctx, tenantID, versionID)
	if err != nil {
		return nil, nil, nil, err
	}

	return v, mods, edgeList, nil
}

func (s *JobService) PublishVersion(ctx context.Context, jobID, versionID string) (*domain.JobVersion, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	v, err := s.versions.FindByID(ctx, tenantID, versionID)
	if err != nil {
		return nil, err
	}
	if v.JobID != jobID {
		return nil, domain.ErrJobVersionNotFound
	}
	if v.Status == domain.JobVersionStatusPublished {
		return nil, domain.ErrJobVersionImmutable
	}

	if err := s.versions.Publish(ctx, tenantID, versionID); err != nil {
		return nil, err
	}

	return s.versions.FindByID(ctx, tenantID, versionID)
}
