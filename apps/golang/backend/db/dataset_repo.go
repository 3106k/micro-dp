package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/micro-dp/domain"
)

type DatasetRepo struct {
	db DBTX
}

func NewDatasetRepo(db DBTX) *DatasetRepo {
	return &DatasetRepo{db: db}
}

func scanDataset(s interface{ Scan(...any) error }) (*domain.Dataset, error) {
	var d domain.Dataset
	if err := s.Scan(
		&d.ID, &d.TenantID, &d.Name, &d.SourceType,
		&d.SchemaJSON, &d.RowCount, &d.StoragePath,
		&d.LastUpdatedAt, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DatasetRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Dataset, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, source_type, schema_json, row_count, storage_path, last_updated_at, created_at, updated_at
		 FROM datasets WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	d, err := scanDataset(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDatasetNotFound
		}
		return nil, err
	}
	return d, nil
}

func (r *DatasetRepo) ListByTenant(ctx context.Context, tenantID string, filter domain.DatasetListFilter) ([]domain.Dataset, error) {
	query := `SELECT id, tenant_id, name, source_type, schema_json, row_count, storage_path, last_updated_at, created_at, updated_at
		 FROM datasets WHERE tenant_id = ?`
	args := []any{tenantID}

	if filter.Query != "" {
		query += ` AND name LIKE ?`
		args = append(args, fmt.Sprintf("%%%s%%", filter.Query))
	}
	if filter.SourceType != "" {
		query += ` AND source_type = ?`
		args = append(args, filter.SourceType)
	}

	query += ` ORDER BY name`

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasets []domain.Dataset
	for rows.Next() {
		d, err := scanDataset(rows)
		if err != nil {
			return nil, err
		}
		datasets = append(datasets, *d)
	}
	return datasets, rows.Err()
}

func (r *DatasetRepo) Create(ctx context.Context, d *domain.Dataset) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO datasets (id, tenant_id, name, source_type, schema_json, row_count, storage_path, last_updated_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		d.ID, d.TenantID, d.Name, d.SourceType, d.SchemaJSON, d.RowCount, d.StoragePath, d.LastUpdatedAt,
	)
	return err
}

func (r *DatasetRepo) Update(ctx context.Context, d *domain.Dataset) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE datasets SET name = ?, source_type = ?, schema_json = ?, row_count = ?, storage_path = ?, last_updated_at = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		d.Name, d.SourceType, d.SchemaJSON, d.RowCount, d.StoragePath, d.LastUpdatedAt, d.TenantID, d.ID,
	)
	return err
}
