package handler

import (
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
)

type tenantResponse struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func toOpenAPIJobRun(jr *domain.JobRun) openapi.JobRun {
	out := openapi.JobRun{
		Id:           jr.ID,
		TenantId:     jr.TenantID,
		JobId:        jr.JobID,
		JobVersionId: jr.JobVersionID,
		Status:       openapi.JobRunStatus(jr.Status),
		StartedAt:    jr.StartedAt,
	}
	if jr.FinishedAt != nil {
		out.FinishedAt = jr.FinishedAt
	}
	return out
}

func toOpenAPITenant(t domain.Tenant) tenantResponse {
	return tenantResponse{
		Id:       t.ID,
		Name:     t.Name,
		IsActive: t.IsActive,
	}
}

func toOpenAPIJob(j *domain.Job) openapi.Job {
	out := openapi.Job{
		Id:       j.ID,
		TenantId: j.TenantID,
		Name:     j.Name,
		Slug:     j.Slug,
		IsActive: j.IsActive,
	}
	if j.Description != "" {
		out.Description = &j.Description
	}
	out.CreatedAt = &j.CreatedAt
	out.UpdatedAt = &j.UpdatedAt
	return out
}

func toOpenAPIJobVersion(v *domain.JobVersion) openapi.JobVersion {
	return openapi.JobVersion{
		Id:          v.ID,
		TenantId:    v.TenantID,
		JobId:       v.JobID,
		Version:     v.Version,
		Status:      openapi.JobVersionStatus(v.Status),
		PublishedAt: v.PublishedAt,
		CreatedAt:   &v.CreatedAt,
		UpdatedAt:   &v.UpdatedAt,
	}
}

func toOpenAPIJobModule(m *domain.JobModule) openapi.JobModule {
	px := float32(m.PositionX)
	py := float32(m.PositionY)
	out := openapi.JobModule{
		Id:                 m.ID,
		TenantId:           m.TenantID,
		JobVersionId:       m.JobVersionID,
		ModuleTypeId:       m.ModuleTypeID,
		ModuleTypeSchemaId: m.ModuleTypeSchemaID,
		ConnectionId:       m.ConnectionID,
		Name:               m.Name,
		PositionX:          &px,
		PositionY:          &py,
	}
	if m.ConfigJSON != "" && m.ConfigJSON != "{}" {
		out.ConfigJson = &m.ConfigJSON
	}
	return out
}

func toOpenAPIJobModuleEdge(e *domain.JobModuleEdge) openapi.JobModuleEdge {
	return openapi.JobModuleEdge{
		Id:             e.ID,
		TenantId:       e.TenantID,
		JobVersionId:   e.JobVersionID,
		SourceModuleId: e.SourceModuleID,
		TargetModuleId: e.TargetModuleID,
	}
}

func toOpenAPIModuleType(mt *domain.ModuleType) openapi.ModuleType {
	return openapi.ModuleType{
		Id:        mt.ID,
		TenantId:  mt.TenantID,
		Name:      mt.Name,
		Category:  openapi.ModuleTypeCategory(mt.Category),
		CreatedAt: &mt.CreatedAt,
		UpdatedAt: &mt.UpdatedAt,
	}
}

func toOpenAPIModuleTypeSchema(s *domain.ModuleTypeSchema) openapi.ModuleTypeSchema {
	return openapi.ModuleTypeSchema{
		Id:           s.ID,
		TenantId:     s.TenantID,
		ModuleTypeId: s.ModuleTypeID,
		Version:      s.Version,
		JsonSchema:   s.JSONSchema,
		CreatedAt:    &s.CreatedAt,
	}
}

func toOpenAPIDataset(d *domain.Dataset) openapi.Dataset {
	out := openapi.Dataset{
		Id:          d.ID,
		TenantId:    d.TenantID,
		Name:        d.Name,
		SourceType:  openapi.DatasetSourceType(d.SourceType),
		StoragePath: d.StoragePath,
	}
	out.SchemaJson = d.SchemaJSON
	out.RowCount = d.RowCount
	out.LastUpdatedAt = d.LastUpdatedAt
	out.CreatedAt = &d.CreatedAt
	out.UpdatedAt = &d.UpdatedAt
	return out
}

func toOpenAPIUpload(u *domain.Upload, files []domain.UploadFile) openapi.Upload {
	apiFiles := make([]openapi.UploadFile, len(files))
	for i, f := range files {
		apiFiles[i] = toOpenAPIUploadFile(&f)
	}
	return openapi.Upload{
		Id:        u.ID,
		TenantId:  u.TenantID,
		Status:    openapi.UploadStatus(u.Status),
		Files:     apiFiles,
		CreatedAt: &u.CreatedAt,
		UpdatedAt: &u.UpdatedAt,
	}
}

func toOpenAPIUploadFile(f *domain.UploadFile) openapi.UploadFile {
	return openapi.UploadFile{
		Id:          f.ID,
		UploadId:    f.UploadID,
		FileName:    f.FileName,
		ObjectKey:   f.ObjectKey,
		ContentType: f.ContentType,
		SizeBytes:   f.SizeBytes,
		CreatedAt:   &f.CreatedAt,
	}
}

func toOpenAPIJobRunModule(m *domain.JobRunModule) openapi.JobRunModule {
	out := openapi.JobRunModule{
		Id:          m.ID,
		TenantId:    m.TenantID,
		JobRunId:    m.JobRunID,
		JobModuleId: m.JobModuleID,
		Status:      openapi.JobRunModuleStatus(m.Status),
		Attempt:     m.Attempt,
		InputJson:   m.InputJSON,
		OutputJson:  m.OutputJSON,
		MetricsJson: m.MetricsJSON,
		ErrorCode:   m.ErrorCode,
		ErrorMessage: m.ErrorMessage,
		StartedAt:   m.StartedAt,
		FinishedAt:  m.FinishedAt,
		CreatedAt:   &m.CreatedAt,
		UpdatedAt:   &m.UpdatedAt,
	}
	return out
}

func toOpenAPIJobRunArtifact(a *domain.JobRunArtifact) openapi.JobRunArtifact {
	out := openapi.JobRunArtifact{
		Id:             a.ID,
		TenantId:       a.TenantID,
		JobRunId:       a.JobRunID,
		JobRunModuleId: a.JobRunModuleID,
		Name:           a.Name,
		ArtifactType:   a.ArtifactType,
		StorageType:    a.StorageType,
		StoragePath:    a.StoragePath,
		SizeBytes:      a.SizeBytes,
		ContentType:    a.ContentType,
		Checksum:       a.Checksum,
		MetadataJson:   a.MetadataJSON,
		CreatedAt:      &a.CreatedAt,
	}
	if a.URI != "" {
		out.Uri = &a.URI
	}
	return out
}

func toOpenAPIPlan(p *domain.Plan) openapi.Plan {
	return openapi.Plan{
		Id:               p.ID,
		Name:             p.Name,
		DisplayName:      p.DisplayName,
		MaxEventsPerDay:  p.MaxEventsPerDay,
		MaxStorageBytes:  p.MaxStorageBytes,
		MaxRowsPerDay:    p.MaxRowsPerDay,
		MaxUploadsPerDay: p.MaxUploadsPerDay,
		IsDefault:        p.IsDefault,
	}
}

func toOpenAPITenantPlanResponse(p *domain.Plan, tp *domain.TenantPlan) openapi.TenantPlanResponse {
	resp := openapi.TenantPlanResponse{
		Plan:      toOpenAPIPlan(p),
		StartedAt: tp.StartedAt,
	}
	if tp.ExpiresAt != nil {
		resp.ExpiresAt = tp.ExpiresAt
	}
	return resp
}

func toOpenAPIUsageSummary(s *domain.UsageSummary) openapi.UsageSummaryResponse {
	resp := openapi.UsageSummaryResponse{
		Date:         s.Date,
		EventsCount:  s.EventsCount,
		StorageBytes: s.StorageBytes,
		RowsCount:    s.RowsCount,
		UploadsCount: s.UploadsCount,
	}
	if s.Plan != nil {
		p := toOpenAPIPlan(s.Plan)
		resp.Plan = &p
	}
	return resp
}

func toOpenAPIConnection(c *domain.Connection) openapi.Connection {
	out := openapi.Connection{
		Id:       c.ID,
		TenantId: c.TenantID,
		Name:     c.Name,
		Type:     c.Type,
	}
	if c.ConfigJSON != "" && c.ConfigJSON != "{}" {
		out.ConfigJson = &c.ConfigJSON
	}
	out.SecretRef = c.SecretRef
	out.CreatedAt = &c.CreatedAt
	out.UpdatedAt = &c.UpdatedAt
	return out
}
