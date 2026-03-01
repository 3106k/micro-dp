package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/user/micro-dp/domain"
)

type JobRunModuleRepo struct {
	db DBTX
}

func NewJobRunModuleRepo(db DBTX) *JobRunModuleRepo {
	return &JobRunModuleRepo{db: db}
}

func (r *JobRunModuleRepo) Create(ctx context.Context, m *domain.JobRunModule) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO job_run_modules (id, tenant_id, job_run_id, job_module_id, status, attempt, input_json, output_json, metrics_json, error_code, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		m.ID, m.TenantID, m.JobRunID, m.JobModuleID, m.Status, m.Attempt, m.InputJSON, m.OutputJSON, m.MetricsJSON, m.ErrorCode,
	)
	return err
}

func (r *JobRunModuleRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.JobRunModule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, job_run_id, job_module_id, status, attempt,
		        input_json, output_json, metrics_json, error_code, error_message,
		        started_at, finished_at, created_at, updated_at
		 FROM job_run_modules WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	var m domain.JobRunModule
	if err := row.Scan(
		&m.ID, &m.TenantID, &m.JobRunID, &m.JobModuleID, &m.Status, &m.Attempt,
		&m.InputJSON, &m.OutputJSON, &m.MetricsJSON, &m.ErrorCode, &m.ErrorMessage,
		&m.StartedAt, &m.FinishedAt, &m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrJobRunModuleNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *JobRunModuleRepo) ListByJobRunID(ctx context.Context, tenantID, jobRunID string) ([]domain.JobRunModule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, job_run_id, job_module_id, status, attempt,
		        input_json, output_json, metrics_json, error_code, error_message,
		        started_at, finished_at, created_at, updated_at
		 FROM job_run_modules WHERE tenant_id = ? AND job_run_id = ?
		 ORDER BY created_at`, tenantID, jobRunID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []domain.JobRunModule
	for rows.Next() {
		var m domain.JobRunModule
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.JobRunID, &m.JobModuleID, &m.Status, &m.Attempt,
			&m.InputJSON, &m.OutputJSON, &m.MetricsJSON, &m.ErrorCode, &m.ErrorMessage,
			&m.StartedAt, &m.FinishedAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

func (r *JobRunModuleRepo) UpdateStatus(ctx context.Context, tenantID, id, status string, errorCode, errorMessage *string, startedAt, finishedAt *time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE job_run_modules
		 SET status = ?, error_code = ?, error_message = ?, started_at = ?, finished_at = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		status, errorCode, errorMessage, startedAt, finishedAt, tenantID, id,
	)
	return err
}
