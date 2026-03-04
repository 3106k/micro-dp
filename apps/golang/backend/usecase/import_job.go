package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type ImportJobService struct {
	jobs        *JobService
	moduleTypes domain.ModuleTypeRepository
	versions    domain.JobVersionRepository
	modules     domain.JobModuleRepository
}

func NewImportJobService(
	jobs *JobService,
	moduleTypes domain.ModuleTypeRepository,
	versions domain.JobVersionRepository,
	modules domain.JobModuleRepository,
) *ImportJobService {
	return &ImportJobService{
		jobs:        jobs,
		moduleTypes: moduleTypes,
		versions:    versions,
		modules:     modules,
	}
}

type CreateImportJobInput struct {
	Name          string
	Slug          string
	Description   string
	ConnectionID  string
	SpreadsheetID string
	SheetName     string
	Range         string
}

type CreateImportJobResult struct {
	Job     *domain.Job
	Version *domain.JobVersion
}

func (s *ImportJobService) CreateImportJob(ctx context.Context, input CreateImportJobInput) (*CreateImportJobResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Create Job
	job, err := s.jobs.CreateJob(ctx, input.Name, input.Slug, input.Description, domain.JobKindImport)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	// Ensure "Google Sheets Import" module type exists
	mtName := "Google Sheets Import"
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

	// Build config_json
	configMap := map[string]string{
		"spreadsheet_id": input.SpreadsheetID,
	}
	if input.SheetName != "" {
		configMap["sheet_name"] = input.SheetName
	}
	if input.Range != "" {
		configMap["range"] = input.Range
	}
	configBytes, _ := json.Marshal(configMap)

	connID := input.ConnectionID
	mod := &domain.JobModule{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		JobVersionID: version.ID,
		ModuleTypeID: mt.ID,
		ConnectionID: &connID,
		Name:         "Google Sheets Import",
		ConfigJSON:   string(configBytes),
	}
	if err := s.modules.Create(ctx, mod); err != nil {
		return nil, fmt.Errorf("create module: %w", err)
	}

	return &CreateImportJobResult{
		Job:     job,
		Version: version,
	}, nil
}
