package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type TemplateRunRepo struct {
	db DBTX
}

func NewTemplateRunRepo(db DBTX) *TemplateRunRepo {
	return &TemplateRunRepo{db: db}
}

func (r *TemplateRunRepo) Create(ctx context.Context, tr *domain.TemplateRun) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO template_runs (id, tenant_id, template_type, status, skip_reason, dashboard_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		tr.ID, tr.TenantID, tr.TemplateType, tr.Status, tr.SkipReason, tr.DashboardID,
	)
	return err
}

func (r *TemplateRunRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.TemplateRun, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, template_type, status, skip_reason, dashboard_id, created_at
		 FROM template_runs WHERE tenant_id = ? AND id = ?`,
		tenantID, id,
	)
	var tr domain.TemplateRun
	if err := row.Scan(&tr.ID, &tr.TenantID, &tr.TemplateType, &tr.Status, &tr.SkipReason, &tr.DashboardID, &tr.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTemplateRunNotFound
		}
		return nil, err
	}
	return &tr, nil
}

func (r *TemplateRunRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.TemplateRun, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, template_type, status, skip_reason, dashboard_id, created_at
		 FROM template_runs WHERE tenant_id = ? ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []domain.TemplateRun
	for rows.Next() {
		var tr domain.TemplateRun
		if err := rows.Scan(&tr.ID, &tr.TenantID, &tr.TemplateType, &tr.Status, &tr.SkipReason, &tr.DashboardID, &tr.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, tr)
	}
	return runs, rows.Err()
}
