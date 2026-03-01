package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type UsageRepo struct {
	db DBTX
}

func NewUsageRepo(db DBTX) *UsageRepo {
	return &UsageRepo{db: db}
}

func (r *UsageRepo) IncrementEvents(ctx context.Context, tenantID, date string, delta int) error {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO usage_daily (id, tenant_id, date, events_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(tenant_id, date) DO UPDATE SET
		   events_count = events_count + ?,
		   updated_at = datetime('now')`,
		id, tenantID, date, delta, delta,
	)
	return err
}

func (r *UsageRepo) IncrementStorage(ctx context.Context, tenantID, date string, deltaBytes int64) error {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO usage_daily (id, tenant_id, date, storage_bytes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(tenant_id, date) DO UPDATE SET
		   storage_bytes = storage_bytes + ?,
		   updated_at = datetime('now')`,
		id, tenantID, date, deltaBytes, deltaBytes,
	)
	return err
}

func (r *UsageRepo) IncrementRows(ctx context.Context, tenantID, date string, delta int) error {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO usage_daily (id, tenant_id, date, rows_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(tenant_id, date) DO UPDATE SET
		   rows_count = rows_count + ?,
		   updated_at = datetime('now')`,
		id, tenantID, date, delta, delta,
	)
	return err
}

func (r *UsageRepo) IncrementUploads(ctx context.Context, tenantID, date string, delta int) error {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO usage_daily (id, tenant_id, date, uploads_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(tenant_id, date) DO UPDATE SET
		   uploads_count = uploads_count + ?,
		   updated_at = datetime('now')`,
		id, tenantID, date, delta, delta,
	)
	return err
}

func (r *UsageRepo) FindDailyByTenantAndDate(ctx context.Context, tenantID, date string) (*domain.UsageDaily, error) {
	var u domain.UsageDaily
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, date, events_count, storage_bytes, rows_count, uploads_count, created_at, updated_at
		 FROM usage_daily WHERE tenant_id = ? AND date = ?`, tenantID, date,
	)
	if err := row.Scan(&u.ID, &u.TenantID, &u.Date, &u.EventsCount, &u.StorageBytes, &u.RowsCount, &u.UploadsCount, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UsageRepo) RecordEvent(ctx context.Context, e *domain.UsageEvent) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO usage_events (id, tenant_id, event_type, delta, recorded_at)
		 VALUES (?, ?, ?, ?, datetime('now'))`,
		e.ID, e.TenantID, e.EventType, e.Delta,
	)
	return err
}
