package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type JobRunArtifactRepo struct {
	db DBTX
}

func NewJobRunArtifactRepo(db DBTX) *JobRunArtifactRepo {
	return &JobRunArtifactRepo{db: db}
}

func (r *JobRunArtifactRepo) Create(ctx context.Context, a *domain.JobRunArtifact) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO job_run_artifacts (id, tenant_id, job_run_id, job_run_module_id, name, artifact_type, storage_type, storage_path, uri, size_bytes, content_type, checksum, metadata_json, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		a.ID, a.TenantID, a.JobRunID, a.JobRunModuleID, a.Name, a.ArtifactType, a.StorageType, a.StoragePath, a.URI, a.SizeBytes, a.ContentType, a.Checksum, a.MetadataJSON,
	)
	return err
}

func (r *JobRunArtifactRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.JobRunArtifact, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, job_run_id, job_run_module_id, name, artifact_type, storage_type, storage_path, uri, size_bytes, content_type, checksum, metadata_json, created_at
		 FROM job_run_artifacts WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	var a domain.JobRunArtifact
	if err := row.Scan(
		&a.ID, &a.TenantID, &a.JobRunID, &a.JobRunModuleID, &a.Name, &a.ArtifactType, &a.StorageType, &a.StoragePath, &a.URI, &a.SizeBytes, &a.ContentType, &a.Checksum, &a.MetadataJSON, &a.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrJobRunArtifactNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *JobRunArtifactRepo) ListByJobRunID(ctx context.Context, tenantID, jobRunID string) ([]domain.JobRunArtifact, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, job_run_id, job_run_module_id, name, artifact_type, storage_type, storage_path, uri, size_bytes, content_type, checksum, metadata_json, created_at
		 FROM job_run_artifacts WHERE tenant_id = ? AND job_run_id = ?
		 ORDER BY created_at`, tenantID, jobRunID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []domain.JobRunArtifact
	for rows.Next() {
		var a domain.JobRunArtifact
		if err := rows.Scan(
			&a.ID, &a.TenantID, &a.JobRunID, &a.JobRunModuleID, &a.Name, &a.ArtifactType, &a.StorageType, &a.StoragePath, &a.URI, &a.SizeBytes, &a.ContentType, &a.Checksum, &a.MetadataJSON, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}
