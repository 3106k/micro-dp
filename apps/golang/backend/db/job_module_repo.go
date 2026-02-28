package db

import (
	"context"
	"database/sql"

	"github.com/user/micro-dp/domain"
)

// ---- JobModuleRepo ----

type JobModuleRepo struct {
	db DBTX
}

func NewJobModuleRepo(db DBTX) *JobModuleRepo {
	return &JobModuleRepo{db: db}
}

func (r *JobModuleRepo) Create(ctx context.Context, m *domain.JobModule) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO job_modules (id, tenant_id, job_version_id, module_type_id, module_type_schema_id, connection_id, name, config_json, position_x, position_y, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		m.ID, m.TenantID, m.JobVersionID, m.ModuleTypeID, m.ModuleTypeSchemaID, m.ConnectionID,
		m.Name, m.ConfigJSON, m.PositionX, m.PositionY,
	)
	return err
}

func (r *JobModuleRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.JobModule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, job_version_id, module_type_id, module_type_schema_id, connection_id,
		        name, config_json, position_x, position_y, created_at, updated_at
		 FROM job_modules WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	var m domain.JobModule
	if err := row.Scan(
		&m.ID, &m.TenantID, &m.JobVersionID, &m.ModuleTypeID, &m.ModuleTypeSchemaID, &m.ConnectionID,
		&m.Name, &m.ConfigJSON, &m.PositionX, &m.PositionY, &m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrJobModuleNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *JobModuleRepo) ListByJobVersionID(ctx context.Context, tenantID, jobVersionID string) ([]domain.JobModule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, job_version_id, module_type_id, module_type_schema_id, connection_id,
		        name, config_json, position_x, position_y, created_at, updated_at
		 FROM job_modules WHERE tenant_id = ? AND job_version_id = ?
		 ORDER BY created_at`, tenantID, jobVersionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []domain.JobModule
	for rows.Next() {
		var m domain.JobModule
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.JobVersionID, &m.ModuleTypeID, &m.ModuleTypeSchemaID, &m.ConnectionID,
			&m.Name, &m.ConfigJSON, &m.PositionX, &m.PositionY, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

// ---- JobModuleEdgeRepo ----

type JobModuleEdgeRepo struct {
	db DBTX
}

func NewJobModuleEdgeRepo(db DBTX) *JobModuleEdgeRepo {
	return &JobModuleEdgeRepo{db: db}
}

func (r *JobModuleEdgeRepo) Create(ctx context.Context, e *domain.JobModuleEdge) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO job_module_edges (id, tenant_id, job_version_id, source_module_id, target_module_id, created_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'))`,
		e.ID, e.TenantID, e.JobVersionID, e.SourceModuleID, e.TargetModuleID,
	)
	return err
}

func (r *JobModuleEdgeRepo) ListByJobVersionID(ctx context.Context, tenantID, jobVersionID string) ([]domain.JobModuleEdge, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, job_version_id, source_module_id, target_module_id, created_at
		 FROM job_module_edges WHERE tenant_id = ? AND job_version_id = ?
		 ORDER BY created_at`, tenantID, jobVersionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []domain.JobModuleEdge
	for rows.Next() {
		var e domain.JobModuleEdge
		if err := rows.Scan(&e.ID, &e.TenantID, &e.JobVersionID, &e.SourceModuleID, &e.TargetModuleID, &e.CreatedAt); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}
