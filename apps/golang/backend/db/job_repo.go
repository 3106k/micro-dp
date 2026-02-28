package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/user/micro-dp/domain"
)

type JobRepo struct {
	db DBTX
}

func NewJobRepo(db DBTX) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(ctx context.Context, job *domain.Job) error {
	isActive := 0
	if job.IsActive {
		isActive = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO jobs (id, tenant_id, name, slug, description, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		job.ID, job.TenantID, job.Name, job.Slug, job.Description, isActive,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrJobSlugDuplicate
		}
		return err
	}
	return nil
}

func (r *JobRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Job, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, slug, description, is_active, created_at, updated_at
		 FROM jobs WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	return scanJob(row)
}

func (r *JobRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Job, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, slug, description, is_active, created_at, updated_at
		 FROM jobs WHERE tenant_id = ?
		 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		var isActive int
		if err := rows.Scan(&j.ID, &j.TenantID, &j.Name, &j.Slug, &j.Description, &isActive, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		j.IsActive = isActive != 0
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *JobRepo) Update(ctx context.Context, job *domain.Job) error {
	isActive := 0
	if job.IsActive {
		isActive = 1
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE jobs SET name = ?, slug = ?, description = ?, is_active = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		job.Name, job.Slug, job.Description, isActive, job.TenantID, job.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrJobSlugDuplicate
		}
		return err
	}
	return nil
}

func scanJob(row *sql.Row) (*domain.Job, error) {
	var j domain.Job
	var isActive int
	if err := row.Scan(&j.ID, &j.TenantID, &j.Name, &j.Slug, &j.Description, &isActive, &j.CreatedAt, &j.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrJobNotFound
		}
		return nil, err
	}
	j.IsActive = isActive != 0
	return &j, nil
}
