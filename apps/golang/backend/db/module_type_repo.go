package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/user/micro-dp/domain"
)

// ---- ModuleTypeRepo ----

type ModuleTypeRepo struct {
	db DBTX
}

func NewModuleTypeRepo(db DBTX) *ModuleTypeRepo {
	return &ModuleTypeRepo{db: db}
}

func (r *ModuleTypeRepo) Create(ctx context.Context, mt *domain.ModuleType) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO module_types (id, tenant_id, name, category, created_at, updated_at)
		 VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`,
		mt.ID, mt.TenantID, mt.Name, mt.Category,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrModuleTypeNameDuplicate
		}
		return err
	}
	return nil
}

func (r *ModuleTypeRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.ModuleType, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, category, created_at, updated_at
		 FROM module_types WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	var mt domain.ModuleType
	if err := row.Scan(&mt.ID, &mt.TenantID, &mt.Name, &mt.Category, &mt.CreatedAt, &mt.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrModuleTypeNotFound
		}
		return nil, err
	}
	return &mt, nil
}

func (r *ModuleTypeRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.ModuleType, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, name, category, created_at, updated_at
		 FROM module_types WHERE tenant_id = ?
		 ORDER BY name`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []domain.ModuleType
	for rows.Next() {
		var mt domain.ModuleType
		if err := rows.Scan(&mt.ID, &mt.TenantID, &mt.Name, &mt.Category, &mt.CreatedAt, &mt.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, mt)
	}
	return types, rows.Err()
}

// ---- ModuleTypeSchemaRepo ----

type ModuleTypeSchemaRepo struct {
	db DBTX
}

func NewModuleTypeSchemaRepo(db DBTX) *ModuleTypeSchemaRepo {
	return &ModuleTypeSchemaRepo{db: db}
}

func (r *ModuleTypeSchemaRepo) Create(ctx context.Context, s *domain.ModuleTypeSchema) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO module_type_schemas (id, tenant_id, module_type_id, version, json_schema, created_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'))`,
		s.ID, s.TenantID, s.ModuleTypeID, s.Version, s.JSONSchema,
	)
	return err
}

func (r *ModuleTypeSchemaRepo) ListByModuleTypeID(ctx context.Context, tenantID, moduleTypeID string) ([]domain.ModuleTypeSchema, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, module_type_id, version, json_schema, created_at
		 FROM module_type_schemas WHERE tenant_id = ? AND module_type_id = ?
		 ORDER BY version DESC`, tenantID, moduleTypeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []domain.ModuleTypeSchema
	for rows.Next() {
		var s domain.ModuleTypeSchema
		if err := rows.Scan(&s.ID, &s.TenantID, &s.ModuleTypeID, &s.Version, &s.JSONSchema, &s.CreatedAt); err != nil {
			return nil, err
		}
		schemas = append(schemas, s)
	}
	return schemas, rows.Err()
}

func (r *ModuleTypeSchemaRepo) NextVersion(ctx context.Context, moduleTypeID string) (int, error) {
	var maxVersion sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT MAX(version) FROM module_type_schemas WHERE module_type_id = ?`, moduleTypeID,
	).Scan(&maxVersion)
	if err != nil {
		return 0, err
	}
	if !maxVersion.Valid {
		return 1, nil
	}
	return int(maxVersion.Int64) + 1, nil
}
