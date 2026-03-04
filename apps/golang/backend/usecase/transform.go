package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/storage"
)

type TransformService struct {
	datasets    domain.DatasetRepository
	minio       *storage.MinIOClient
	jobs        *JobService
	moduleTypes domain.ModuleTypeRepository
	jobRuns     domain.JobRunRepository
	versions    domain.JobVersionRepository
	modules     domain.JobModuleRepository
	queue       domain.TransformJobQueue
}

func NewTransformService(
	datasets domain.DatasetRepository,
	minio *storage.MinIOClient,
	jobs *JobService,
	moduleTypes domain.ModuleTypeRepository,
	jobRuns domain.JobRunRepository,
	versions domain.JobVersionRepository,
	modules domain.JobModuleRepository,
	queue domain.TransformJobQueue,
) *TransformService {
	return &TransformService{
		datasets:    datasets,
		minio:       minio,
		jobs:        jobs,
		moduleTypes: moduleTypes,
		jobRuns:     jobRuns,
		versions:    versions,
		modules:     modules,
		queue:       queue,
	}
}

type ValidateResult struct {
	Valid   bool
	Error   string
	Columns []ColumnInfo
}

type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type PreviewResult struct {
	Columns  []ColumnInfo
	Rows     []map[string]any
	RowCount int
}

type CreateTransformJobInput struct {
	Name        string
	Slug        string
	Description string
	SQL         string
	DatasetIDs  []string
	Execution   string // "save_only", "immediate", "scheduled"
	ScheduledAt *time.Time
}

type CreateTransformJobResult struct {
	Job     *domain.Job
	Version *domain.JobVersion
	JobRun  *domain.JobRun
}

func (s *TransformService) setupDuckDB(ctx context.Context, tenantID string, datasetIDs []string) (*sql.DB, error) {
	datasets := make([]*domain.Dataset, 0, len(datasetIDs))
	for _, id := range datasetIDs {
		ds, err := s.datasets.FindByID(ctx, tenantID, id)
		if err != nil {
			return nil, fmt.Errorf("dataset %s: %w", id, err)
		}
		datasets = append(datasets, ds)
	}

	duckDB, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}

	s3Cfg := s.minio.S3Config()
	if err := storage.ConfigureDuckDBHTTPFS(ctx, duckDB, s3Cfg); err != nil {
		duckDB.Close()
		return nil, fmt.Errorf("configure httpfs: %w", err)
	}

	// Register each dataset as a VIEW reading directly from S3
	for _, ds := range datasets {
		uri := storage.S3ParquetURI(s3Cfg.Bucket, ds.StoragePath)
		viewSQL := fmt.Sprintf("CREATE VIEW %s AS SELECT * FROM read_parquet('%s')", quoteIdentifier(ds.Name), uri)
		if _, err := duckDB.ExecContext(ctx, viewSQL); err != nil {
			duckDB.Close()
			return nil, fmt.Errorf("create view %s: %w", ds.Name, err)
		}
	}

	return duckDB, nil
}

func (s *TransformService) ValidateSQL(ctx context.Context, sqlStr string, datasetIDs []string) (*ValidateResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	if sqlStr == "" {
		return &ValidateResult{Valid: false, Error: "SQL is required"}, nil
	}
	if len(datasetIDs) == 0 {
		return &ValidateResult{Valid: false, Error: "at least one dataset is required"}, nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	duckDB, err := s.setupDuckDB(timeoutCtx, tenantID, datasetIDs)
	if err != nil {
		return &ValidateResult{Valid: false, Error: err.Error()}, nil
	}
	defer duckDB.Close()

	// EXPLAIN to check syntax
	if _, err := duckDB.ExecContext(timeoutCtx, fmt.Sprintf("EXPLAIN %s", sqlStr)); err != nil {
		return &ValidateResult{Valid: false, Error: err.Error()}, nil
	}

	// Get columns from the query
	createTmp := fmt.Sprintf("CREATE TEMP TABLE _r AS SELECT * FROM (%s) AS _q LIMIT 0", sqlStr)
	if _, err := duckDB.ExecContext(timeoutCtx, createTmp); err != nil {
		return &ValidateResult{Valid: false, Error: err.Error()}, nil
	}

	columns, err := describeTable(timeoutCtx, duckDB, "_r")
	if err != nil {
		return &ValidateResult{Valid: false, Error: err.Error()}, nil
	}

	return &ValidateResult{Valid: true, Columns: columns}, nil
}

func (s *TransformService) PreviewSQL(ctx context.Context, sqlStr string, datasetIDs []string, limit int) (*PreviewResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	duckDB, err := s.setupDuckDB(timeoutCtx, tenantID, datasetIDs)
	if err != nil {
		return nil, fmt.Errorf("setup duckdb: %w", err)
	}
	defer duckDB.Close()

	query := fmt.Sprintf("SELECT * FROM (%s) AS _q LIMIT %d", sqlStr, limit)
	rows, err := duckDB.QueryContext(timeoutCtx, query)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("get column types: %w", err)
	}

	columns := make([]ColumnInfo, len(colNames))
	for i, name := range colNames {
		columns[i] = ColumnInfo{Name: name, Type: colTypes[i].DatabaseTypeName()}
	}

	var resultRows []map[string]any
	for rows.Next() {
		values := make([]any, len(colNames))
		valuePtrs := make([]any, len(colNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		row := make(map[string]any, len(colNames))
		for i, name := range colNames {
			row[name] = values[i]
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return &PreviewResult{
		Columns:  columns,
		Rows:     resultRows,
		RowCount: len(resultRows),
	}, nil
}

func (s *TransformService) CreateTransformJob(ctx context.Context, input CreateTransformJobInput) (*CreateTransformJobResult, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Validate SQL first
	result, err := s.ValidateSQL(ctx, input.SQL, input.DatasetIDs)
	if err != nil {
		return nil, err
	}
	if !result.Valid {
		return nil, fmt.Errorf("invalid SQL: %s", result.Error)
	}

	// Create Job
	job, err := s.jobs.CreateJob(ctx, input.Name, input.Slug, input.Description, domain.JobKindTransform)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	// Ensure "SQL Transform" module type exists
	mt, err := s.moduleTypes.FindByTenantAndName(ctx, tenantID, "SQL Transform")
	if err != nil {
		mt = &domain.ModuleType{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			Name:     "SQL Transform",
			Category: domain.ModuleTypeCategoryTransform,
		}
		if err := s.moduleTypes.Create(ctx, mt); err != nil {
			// If already created by race, try to find again
			mt2, err2 := s.moduleTypes.FindByTenantAndName(ctx, tenantID, "SQL Transform")
			if err2 != nil {
				return nil, fmt.Errorf("create module type: %w", err)
			}
			mt = mt2
		}
	}

	// Create Version
	nextVer, err := s.versions.NextVersion(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("next version: %w", err)
	}

	version := &domain.JobVersion{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		JobID:    job.ID,
		Version:  nextVer,
		Status:   domain.JobVersionStatusDraft,
	}
	if err := s.versions.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create version: %w", err)
	}

	// Create Module with SQL config
	configJSON := fmt.Sprintf(`{"sql":%q,"dataset_ids":%s}`, input.SQL, toJSONArray(input.DatasetIDs))
	mod := &domain.JobModule{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		JobVersionID: version.ID,
		ModuleTypeID: mt.ID,
		Name:         "SQL Transform",
		ConfigJSON:   configJSON,
	}
	if err := s.modules.Create(ctx, mod); err != nil {
		return nil, fmt.Errorf("create module: %w", err)
	}

	out := &CreateTransformJobResult{
		Job:     job,
		Version: version,
	}

	execution := input.Execution
	if execution == "" {
		execution = "save_only"
	}

	if execution == "immediate" || execution == "scheduled" {
		status := domain.StatusQueued
		var startedAt *time.Time
		if execution == "scheduled" && input.ScheduledAt != nil {
			status = "scheduled"
		}

		jr := &domain.JobRun{
			ID:           uuid.New().String(),
			TenantID:     tenantID,
			JobID:        job.ID,
			JobVersionID: &version.ID,
			Status:       status,
			StartedAt:    startedAt,
		}
		if err := s.jobRuns.Create(ctx, jr); err != nil {
			return nil, fmt.Errorf("create job run: %w", err)
		}

		// Re-read to get timestamps
		jr, err = s.jobRuns.FindByID(ctx, tenantID, jr.ID)
		if err != nil {
			return nil, fmt.Errorf("find job run: %w", err)
		}
		out.JobRun = jr

		// Enqueue for immediate execution
		if execution == "immediate" && s.queue != nil {
			msg := &domain.TransformJobMessage{
				JobRunID:   jr.ID,
				TenantID:   tenantID,
				SQL:        input.SQL,
				DatasetIDs: input.DatasetIDs,
				JobID:      job.ID,
				VersionID:  version.ID,
			}
			if err := s.queue.Enqueue(ctx, msg); err != nil {
				return nil, fmt.Errorf("enqueue transform: %w", err)
			}
		}
	}

	return out, nil
}

func quoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, name)
}

func describeTable(ctx context.Context, db *sql.DB, tableName string) ([]ColumnInfo, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("DESCRIBE %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var name, colType string
		var null, key, def, extra sql.NullString
		if err := rows.Scan(&name, &colType, &null, &key, &def, &extra); err != nil {
			return nil, err
		}
		columns = append(columns, ColumnInfo{Name: name, Type: colType})
	}
	return columns, rows.Err()
}

func toJSONArray(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	result := "["
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%q", s)
	}
	result += "]"
	return result
}
