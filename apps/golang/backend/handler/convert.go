package handler

import (
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
)

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

func toOpenAPITenant(t domain.Tenant) openapi.Tenant {
	return openapi.Tenant{
		Id:   t.ID,
		Name: t.Name,
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
