package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type TenantRepo struct {
	db DBTX
}

func NewTenantRepo(db DBTX) *TenantRepo {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tenants (id, name, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, datetime('now'), datetime('now'))`,
		tenant.ID, tenant.Name, tenant.IsActive,
	)
	return err
}

func (r *TenantRepo) FindByID(ctx context.Context, id string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, is_active, created_at, updated_at FROM tenants WHERE id = ?`, id,
	).Scan(&t.ID, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) ListAll(ctx context.Context) ([]domain.Tenant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, is_active, created_at, updated_at
		 FROM tenants
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func (r *TenantRepo) Update(ctx context.Context, tenant *domain.Tenant) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tenants
		 SET name = ?, is_active = ?, updated_at = datetime('now')
		 WHERE id = ?`,
		tenant.Name, tenant.IsActive, tenant.ID,
	)
	return err
}

func (r *TenantRepo) AddUserToTenant(ctx context.Context, ut *domain.UserTenant) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_tenants (user_id, tenant_id, role, created_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		ut.UserID, ut.TenantID, ut.Role,
	)
	return err
}

func (r *TenantRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Tenant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.id, t.name, t.is_active, t.created_at, t.updated_at
		 FROM tenants t
		 INNER JOIN user_tenants ut ON ut.tenant_id = t.id
		 WHERE ut.user_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func (r *TenantRepo) IsUserInTenant(ctx context.Context, userID, tenantID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_tenants WHERE user_id = ? AND tenant_id = ?`,
		userID, tenantID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
