package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type DashboardRepo struct {
	db DBTX
}

func NewDashboardRepo(db DBTX) *DashboardRepo {
	return &DashboardRepo{db: db}
}

func (r *DashboardRepo) Create(ctx context.Context, d *domain.Dashboard) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO dashboards (id, tenant_id, name, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`,
		d.ID, d.TenantID, d.Name, d.Description,
	)
	return err
}

func (r *DashboardRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Dashboard, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, description, created_at, updated_at
		 FROM dashboards WHERE tenant_id = ? AND id = ?`,
		tenantID, id,
	)
	var d domain.Dashboard
	if err := row.Scan(&d.ID, &d.TenantID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDashboardNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *DashboardRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Dashboard, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, description, created_at, updated_at
		 FROM dashboards WHERE tenant_id = ? ORDER BY name`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dashboards []domain.Dashboard
	for rows.Next() {
		var d domain.Dashboard
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Name, &d.Description, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		dashboards = append(dashboards, d)
	}
	return dashboards, rows.Err()
}

func (r *DashboardRepo) Update(ctx context.Context, d *domain.Dashboard) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE dashboards SET name = ?, description = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		d.Name, d.Description, d.TenantID, d.ID,
	)
	return err
}

func (r *DashboardRepo) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM dashboards WHERE tenant_id = ? AND id = ?`,
		tenantID, id,
	)
	return err
}
