package handler

import (
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/openapi"
)

func toOpenAPIJobRun(jr *domain.JobRun) openapi.JobRun {
	out := openapi.JobRun{
		Id:        jr.ID,
		TenantId:  jr.TenantID,
		ProjectId: jr.ProjectID,
		JobId:     jr.JobID,
		Status:    openapi.JobRunStatus(jr.Status),
		StartedAt: jr.StartedAt,
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
