package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type JobVersionRepo struct {
	db DBTX
}

func NewJobVersionRepo(db DBTX) *JobVersionRepo {
	return &JobVersionRepo{db: db}
}

func (r *JobVersionRepo) Create(ctx context.Context, v *domain.JobVersion) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO job_versions (id, tenant_id, job_id, version, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		v.ID, v.TenantID, v.JobID, v.Version, v.Status,
	)
	return err
}

func (r *JobVersionRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.JobVersion, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, job_id, version, status, published_at, created_at, updated_at
		 FROM job_versions WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	return scanJobVersion(row)
}

func (r *JobVersionRepo) ListByJobID(ctx context.Context, tenantID, jobID string) ([]domain.JobVersion, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, job_id, version, status, published_at, created_at, updated_at
		 FROM job_versions WHERE tenant_id = ? AND job_id = ?
		 ORDER BY version DESC`, tenantID, jobID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []domain.JobVersion
	for rows.Next() {
		var v domain.JobVersion
		if err := rows.Scan(&v.ID, &v.TenantID, &v.JobID, &v.Version, &v.Status, &v.PublishedAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *JobVersionRepo) Publish(ctx context.Context, tenantID, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE job_versions SET status = 'published', published_at = datetime('now'), updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	return err
}

func (r *JobVersionRepo) NextVersion(ctx context.Context, jobID string) (int, error) {
	var maxVersion sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT MAX(version) FROM job_versions WHERE job_id = ?`, jobID,
	).Scan(&maxVersion)
	if err != nil {
		return 0, err
	}
	if !maxVersion.Valid {
		return 1, nil
	}
	return int(maxVersion.Int64) + 1, nil
}

func scanJobVersion(row *sql.Row) (*domain.JobVersion, error) {
	var v domain.JobVersion
	if err := row.Scan(&v.ID, &v.TenantID, &v.JobID, &v.Version, &v.Status, &v.PublishedAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrJobVersionNotFound
		}
		return nil, err
	}
	return &v, nil
}
