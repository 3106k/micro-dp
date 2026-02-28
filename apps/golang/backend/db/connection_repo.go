package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/user/micro-dp/domain"
)

type ConnectionRepo struct {
	db DBTX
}

func NewConnectionRepo(db DBTX) *ConnectionRepo {
	return &ConnectionRepo{db: db}
}

func (r *ConnectionRepo) Create(ctx context.Context, c *domain.Connection) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO connections (id, tenant_id, name, type, config_json, secret_ref, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		c.ID, c.TenantID, c.Name, c.Type, c.ConfigJSON, c.SecretRef,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrConnectionNameDuplicate
		}
		return err
	}
	return nil
}

func (r *ConnectionRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Connection, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, type, config_json, secret_ref, created_at, updated_at
		 FROM connections WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	var c domain.Connection
	if err := row.Scan(&c.ID, &c.TenantID, &c.Name, &c.Type, &c.ConfigJSON, &c.SecretRef, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrConnectionNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *ConnectionRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Connection, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, type, config_json, secret_ref, created_at, updated_at
		 FROM connections WHERE tenant_id = ?
		 ORDER BY name`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []domain.Connection
	for rows.Next() {
		var c domain.Connection
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Type, &c.ConfigJSON, &c.SecretRef, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		connections = append(connections, c)
	}
	return connections, rows.Err()
}

func (r *ConnectionRepo) Update(ctx context.Context, c *domain.Connection) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE connections SET name = ?, type = ?, config_json = ?, secret_ref = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		c.Name, c.Type, c.ConfigJSON, c.SecretRef, c.TenantID, c.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrConnectionNameDuplicate
		}
		return err
	}
	return nil
}

func (r *ConnectionRepo) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM connections WHERE tenant_id = ? AND id = ?`,
		tenantID, id,
	)
	return err
}
