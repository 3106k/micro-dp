package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

type UploadRepo struct {
	db DBTX
}

func NewUploadRepo(db DBTX) *UploadRepo {
	return &UploadRepo{db: db}
}

func scanUpload(s interface{ Scan(...any) error }) (*domain.Upload, error) {
	var u domain.Upload
	if err := s.Scan(&u.ID, &u.TenantID, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func scanUploadFile(s interface{ Scan(...any) error }) (*domain.UploadFile, error) {
	var f domain.UploadFile
	if err := s.Scan(&f.ID, &f.TenantID, &f.UploadID, &f.FileName, &f.ObjectKey, &f.ContentType, &f.SizeBytes, &f.CreatedAt); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *UploadRepo) CreateUpload(ctx context.Context, u *domain.Upload) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO uploads (id, tenant_id, status, created_at, updated_at)
		 VALUES (?, ?, ?, datetime('now'), datetime('now'))`,
		u.ID, u.TenantID, u.Status,
	)
	return err
}

func (r *UploadRepo) CreateUploadFile(ctx context.Context, f *domain.UploadFile) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO upload_files (id, tenant_id, upload_id, file_name, object_key, content_type, size_bytes, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		f.ID, f.TenantID, f.UploadID, f.FileName, f.ObjectKey, f.ContentType, f.SizeBytes,
	)
	return err
}

func (r *UploadRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Upload, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, status, created_at, updated_at
		 FROM uploads WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	u, err := scanUpload(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUploadNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *UploadRepo) FindFilesByUploadID(ctx context.Context, tenantID, uploadID string) ([]domain.UploadFile, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, upload_id, file_name, object_key, content_type, size_bytes, created_at
		 FROM upload_files WHERE tenant_id = ? AND upload_id = ?`, tenantID, uploadID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []domain.UploadFile
	for rows.Next() {
		f, err := scanUploadFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, *f)
	}
	return files, rows.Err()
}

func (r *UploadRepo) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE uploads SET status = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		status, tenantID, id,
	)
	return err
}
