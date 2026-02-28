package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type JobRunRepo struct {
	db DBTX
}

func NewJobRunRepo(db DBTX) *JobRunRepo {
	return &JobRunRepo{db: db}
}

func (r *JobRunRepo) Create(ctx context.Context, jr *domain.JobRun) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO job_runs (id, tenant_id, job_id, job_version_id, status, run_snapshot_json, attempt, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 0, datetime('now'), datetime('now'))`,
		jr.ID, jr.TenantID, jr.JobID, jr.JobVersionID, jr.Status, jr.RunSnapshotJSON,
	)
	return err
}

func (r *JobRunRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.JobRun, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, job_id, job_version_id, status,
		        run_snapshot_json, checkpoint_json, progress_json, attempt,
		        next_run_at, last_error, started_at, finished_at,
		        created_at, updated_at
		 FROM job_runs WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	return scanJobRun(row)
}

func (r *JobRunRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.JobRun, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, job_id, job_version_id, status,
		        run_snapshot_json, checkpoint_json, progress_json, attempt,
		        next_run_at, last_error, started_at, finished_at,
		        created_at, updated_at
		 FROM job_runs WHERE tenant_id = ?
		 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobRuns []domain.JobRun
	for rows.Next() {
		var jr domain.JobRun
		if err := rows.Scan(
			&jr.ID, &jr.TenantID, &jr.JobID, &jr.JobVersionID, &jr.Status,
			&jr.RunSnapshotJSON, &jr.CheckpointJSON, &jr.ProgressJSON, &jr.Attempt,
			&jr.NextRunAt, &jr.LastError, &jr.StartedAt, &jr.FinishedAt,
			&jr.CreatedAt, &jr.UpdatedAt,
		); err != nil {
			return nil, err
		}
		jobRuns = append(jobRuns, jr)
	}
	return jobRuns, rows.Err()
}

func (r *JobRunRepo) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE job_runs SET status = ?, updated_at = datetime('now') WHERE tenant_id = ? AND id = ?`,
		status, tenantID, id,
	)
	return err
}

func scanJobRun(row *sql.Row) (*domain.JobRun, error) {
	var jr domain.JobRun
	if err := row.Scan(
		&jr.ID, &jr.TenantID, &jr.JobID, &jr.JobVersionID, &jr.Status,
		&jr.RunSnapshotJSON, &jr.CheckpointJSON, &jr.ProgressJSON, &jr.Attempt,
		&jr.NextRunAt, &jr.LastError, &jr.StartedAt, &jr.FinishedAt,
		&jr.CreatedAt, &jr.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrJobRunNotFound
		}
		return nil, err
	}
	return &jr, nil
}
