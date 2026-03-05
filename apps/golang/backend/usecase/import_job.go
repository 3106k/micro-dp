package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/connector"
)

type ImportJobService struct {
	jobs        *JobService
	jobRuns     *JobRunService
	moduleTypes domain.ModuleTypeRepository
	versions    domain.JobVersionRepository
	modules     domain.JobModuleRepository
	connections domain.ConnectionRepository
}

func NewImportJobService(
	jobs *JobService,
	jobRuns *JobRunService,
	moduleTypes domain.ModuleTypeRepository,
	versions domain.JobVersionRepository,
	modules domain.JobModuleRepository,
	connections domain.ConnectionRepository,
) *ImportJobService {
	return &ImportJobService{
		jobs:        jobs,
		jobRuns:     jobRuns,
		moduleTypes: moduleTypes,
		versions:    versions,
		modules:     modules,
		connections: connections,
	}
}

type CreateImportJobInput struct {
	Name         string
	Slug         string
	Description  string
	ConnectionID string
	SourceConfig map[string]interface{}
	Execution    string
}

type CreateImportJobResult struct {
	Job     *domain.Job
	Version *domain.JobVersion
	JobRun  *domain.JobRun
}

func (s *ImportJobService) CreateImportJob(ctx context.Context, input CreateImportJobInput) (*CreateImportJobResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Look up connection to determine connector type
	conn, err := s.connections.FindByID(ctx, tenantID, input.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	def := connector.Global().Get(conn.Type)
	if def == nil {
		return nil, fmt.Errorf("unknown connector type: %s", conn.Type)
	}

	// Create Job
	job, err := s.jobs.CreateJob(ctx, input.Name, input.Slug, input.Description, domain.JobKindImport)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	// Ensure module type exists for this connector
	mtName := def.Name
	mt, err := s.moduleTypes.FindByTenantAndName(ctx, tenantID, mtName)
	if err != nil {
		mt = &domain.ModuleType{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			Name:     mtName,
			Category: domain.ModuleTypeCategorySource,
		}
		if err := s.moduleTypes.Create(ctx, mt); err != nil {
			mt2, err2 := s.moduleTypes.FindByTenantAndName(ctx, tenantID, mtName)
			if err2 != nil {
				return nil, fmt.Errorf("create module type: %w", err)
			}
			mt = mt2
		}
	}

	// Create Version
	nextVer, err := s.versions.NextVersion(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("next version: %w", err)
	}

	version := &domain.JobVersion{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		JobID:    job.ID,
		Version:  nextVer,
		Status:   domain.JobVersionStatusDraft,
	}
	if err := s.versions.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create version: %w", err)
	}

	// Build config_json from source_config
	configBytes, _ := json.Marshal(input.SourceConfig)

	connID := input.ConnectionID
	mod := &domain.JobModule{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		JobVersionID: version.ID,
		ModuleTypeID: mt.ID,
		ConnectionID: &connID,
		Name:         def.Name,
		ConfigJSON:   string(configBytes),
	}
	if err := s.modules.Create(ctx, mod); err != nil {
		return nil, fmt.Errorf("create module: %w", err)
	}

	out := &CreateImportJobResult{
		Job:     job,
		Version: version,
	}

	execution := input.Execution
	if execution == "" {
		execution = "save_only"
	}

	if execution == "immediate" {
		// Auto-publish the version
		if _, err := s.jobs.PublishVersion(ctx, job.ID, version.ID); err != nil {
			return nil, fmt.Errorf("publish version: %w", err)
		}

		// Create job run (JobRunService builds RunSnapshot and sets status=queued)
		jr, err := s.jobRuns.Create(ctx, job.ID, &version.ID)
		if err != nil {
			return nil, fmt.Errorf("create job run: %w", err)
		}
		out.JobRun = jr
	}

	return out, nil
}
