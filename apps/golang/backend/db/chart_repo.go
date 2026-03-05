package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type ChartRepo struct {
	db DBTX
}

func NewChartRepo(db DBTX) *ChartRepo {
	return &ChartRepo{db: db}
}

func (r *ChartRepo) Create(ctx context.Context, c *domain.Chart) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO charts (id, tenant_id, name, chart_type, dataset_id, measure, dimension, config_json, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		c.ID, c.TenantID, c.Name, c.ChartType, c.DatasetID, c.Measure, c.Dimension, c.ConfigJSON,
	)
	return err
}

func (r *ChartRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Chart, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, chart_type, dataset_id, measure, dimension, config_json, created_at, updated_at
		 FROM charts WHERE tenant_id = ? AND id = ?`,
		tenantID, id,
	)
	var c domain.Chart
	if err := row.Scan(&c.ID, &c.TenantID, &c.Name, &c.ChartType, &c.DatasetID, &c.Measure, &c.Dimension, &c.ConfigJSON, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrChartNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *ChartRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Chart, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, chart_type, dataset_id, measure, dimension, config_json, created_at, updated_at
		 FROM charts WHERE tenant_id = ? ORDER BY name`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charts []domain.Chart
	for rows.Next() {
		var c domain.Chart
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.ChartType, &c.DatasetID, &c.Measure, &c.Dimension, &c.ConfigJSON, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		charts = append(charts, c)
	}
	return charts, rows.Err()
}

func (r *ChartRepo) Update(ctx context.Context, c *domain.Chart) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE charts SET name = ?, chart_type = ?, dataset_id = ?, measure = ?, dimension = ?, config_json = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		c.Name, c.ChartType, c.DatasetID, c.Measure, c.Dimension, c.ConfigJSON, c.TenantID, c.ID,
	)
	return err
}

func (r *ChartRepo) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM charts WHERE tenant_id = ? AND id = ?`,
		tenantID, id,
	)
	return err
}
