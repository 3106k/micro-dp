package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type WriteKeyRepo struct {
	db DBTX
}

func NewWriteKeyRepo(db DBTX) *WriteKeyRepo {
	return &WriteKeyRepo{db: db}
}

func scanWriteKey(s interface{ Scan(...any) error }) (*domain.WriteKey, error) {
	var wk domain.WriteKey
	if err := s.Scan(&wk.ID, &wk.TenantID, &wk.Name, &wk.KeyHash, &wk.KeyPrefix, &wk.IsActive, &wk.CreatedAt, &wk.UpdatedAt); err != nil {
		return nil, err
	}
	return &wk, nil
}

func (r *WriteKeyRepo) Create(ctx context.Context, wk *domain.WriteKey) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO write_keys (id, tenant_id, name, key_hash, key_prefix, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		wk.ID, wk.TenantID, wk.Name, wk.KeyHash, wk.KeyPrefix, wk.IsActive,
	)
	return err
}

func (r *WriteKeyRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.WriteKey, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, key_hash, key_prefix, is_active, created_at, updated_at
		 FROM write_keys WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	wk, err := scanWriteKey(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrWriteKeyNotFound
		}
		return nil, err
	}
	return wk, nil
}

func (r *WriteKeyRepo) FindByKeyHash(ctx context.Context, keyHash string) (*domain.WriteKey, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, key_hash, key_prefix, is_active, created_at, updated_at
		 FROM write_keys WHERE key_hash = ? AND is_active = 1`, keyHash,
	)
	wk, err := scanWriteKey(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrWriteKeyNotFound
		}
		return nil, err
	}
	return wk, nil
}

func (r *WriteKeyRepo) FindByTenantID(ctx context.Context, tenantID string) ([]domain.WriteKey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, key_hash, key_prefix, is_active, created_at, updated_at
		 FROM write_keys WHERE tenant_id = ? ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []domain.WriteKey
	for rows.Next() {
		wk, err := scanWriteKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, *wk)
	}
	return keys, rows.Err()
}

func (r *WriteKeyRepo) Delete(ctx context.Context, tenantID, id string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM write_keys WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrWriteKeyNotFound
	}
	return nil
}

func (r *WriteKeyRepo) UpdateKeyHash(ctx context.Context, tenantID, id, keyHash, keyPrefix string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE write_keys SET key_hash = ?, key_prefix = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		keyHash, keyPrefix, tenantID, id,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrWriteKeyNotFound
	}
	return nil
}
