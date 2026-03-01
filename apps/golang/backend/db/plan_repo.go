package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

// --- PlanRepo ---

type PlanRepo struct {
	db DBTX
}

func NewPlanRepo(db DBTX) *PlanRepo {
	return &PlanRepo{db: db}
}

func scanPlan(s interface{ Scan(...any) error }) (*domain.Plan, error) {
	var p domain.Plan
	if err := s.Scan(
		&p.ID, &p.Name, &p.DisplayName,
		&p.MaxEventsPerDay, &p.MaxStorageBytes, &p.MaxRowsPerDay, &p.MaxUploadsPerDay,
		&p.IsDefault, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &p, nil
}

const planColumns = `id, name, display_name, max_events_per_day, max_storage_bytes, max_rows_per_day, max_uploads_per_day, is_default, created_at, updated_at`

func (r *PlanRepo) FindByID(ctx context.Context, id string) (*domain.Plan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+planColumns+` FROM plans WHERE id = ?`, id,
	)
	p, err := scanPlan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PlanRepo) FindByName(ctx context.Context, name string) (*domain.Plan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+planColumns+` FROM plans WHERE name = ?`, name,
	)
	p, err := scanPlan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PlanRepo) FindDefault(ctx context.Context) (*domain.Plan, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+planColumns+` FROM plans WHERE is_default = TRUE LIMIT 1`,
	)
	p, err := scanPlan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PlanRepo) ListAll(ctx context.Context) ([]domain.Plan, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+planColumns+` FROM plans ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []domain.Plan
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, *p)
	}
	return plans, rows.Err()
}

func (r *PlanRepo) Create(ctx context.Context, p *domain.Plan) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO plans (id, name, display_name, max_events_per_day, max_storage_bytes, max_rows_per_day, max_uploads_per_day, is_default, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		p.ID, p.Name, p.DisplayName, p.MaxEventsPerDay, p.MaxStorageBytes, p.MaxRowsPerDay, p.MaxUploadsPerDay, p.IsDefault,
	)
	return err
}

func (r *PlanRepo) Update(ctx context.Context, p *domain.Plan) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE plans SET name = ?, display_name = ?, max_events_per_day = ?, max_storage_bytes = ?, max_rows_per_day = ?, max_uploads_per_day = ?, is_default = ?, updated_at = datetime('now')
		 WHERE id = ?`,
		p.Name, p.DisplayName, p.MaxEventsPerDay, p.MaxStorageBytes, p.MaxRowsPerDay, p.MaxUploadsPerDay, p.IsDefault, p.ID,
	)
	return err
}

// --- TenantPlanRepo ---

type TenantPlanRepo struct {
	db DBTX
}

func NewTenantPlanRepo(db DBTX) *TenantPlanRepo {
	return &TenantPlanRepo{db: db}
}

func (r *TenantPlanRepo) FindByTenantID(ctx context.Context, tenantID string) (*domain.TenantPlan, error) {
	var tp domain.TenantPlan
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, plan_id, started_at, expires_at, created_at, updated_at
		 FROM tenant_plans WHERE tenant_id = ?`, tenantID,
	)
	if err := row.Scan(&tp.ID, &tp.TenantID, &tp.PlanID, &tp.StartedAt, &tp.ExpiresAt, &tp.CreatedAt, &tp.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTenantPlanNotFound
		}
		return nil, err
	}
	return &tp, nil
}

func (r *TenantPlanRepo) Upsert(ctx context.Context, tp *domain.TenantPlan) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tenant_plans (id, tenant_id, plan_id, started_at, expires_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(tenant_id) DO UPDATE SET
		   plan_id = excluded.plan_id,
		   started_at = excluded.started_at,
		   expires_at = excluded.expires_at,
		   updated_at = datetime('now')`,
		tp.ID, tp.TenantID, tp.PlanID, tp.StartedAt, tp.ExpiresAt,
	)
	return err
}
