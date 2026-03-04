package worker

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/storage"
)

type TransformResult struct {
	RowCount  int64
	OutputKey string
}

type TransformWriter struct {
	minio    *storage.MinIOClient
	datasets domain.DatasetRepository
}

func NewTransformWriter(minio *storage.MinIOClient, datasets domain.DatasetRepository) *TransformWriter {
	return &TransformWriter{minio: minio, datasets: datasets}
}

func (w *TransformWriter) Execute(ctx context.Context, msg *domain.TransformJobMessage) (*TransformResult, error) {
	// tmpDir is only needed for output Parquet (COPY TO requires a local path)
	tmpDir, err := os.MkdirTemp("", "micro-dp-transform-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Fetch input datasets
	datasets := make([]*domain.Dataset, 0, len(msg.DatasetIDs))
	for _, id := range msg.DatasetIDs {
		ds, err := w.datasets.FindByID(ctx, msg.TenantID, id)
		if err != nil {
			return nil, fmt.Errorf("find dataset %s: %w", id, err)
		}
		datasets = append(datasets, ds)
	}

	// Open DuckDB in-memory
	duckDB, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer duckDB.Close()

	// Configure httpfs for direct S3/MinIO reads
	s3Cfg := w.minio.S3Config()
	if err := storage.ConfigureDuckDBHTTPFS(ctx, duckDB, s3Cfg); err != nil {
		return nil, fmt.Errorf("configure httpfs: %w", err)
	}

	// Register each dataset as a VIEW reading directly from S3
	for _, ds := range datasets {
		uri := storage.S3ParquetURI(s3Cfg.Bucket, ds.StoragePath)
		viewSQL := fmt.Sprintf(`CREATE VIEW "%s" AS SELECT * FROM read_parquet('%s')`, ds.Name, uri)
		if _, err := duckDB.ExecContext(ctx, viewSQL); err != nil {
			return nil, fmt.Errorf("create view %s: %w", ds.Name, err)
		}
	}

	// Execute user SQL
	createResult := fmt.Sprintf("CREATE TABLE _result AS SELECT * FROM (%s) AS _q", msg.SQL)
	if _, err := duckDB.ExecContext(ctx, createResult); err != nil {
		return nil, fmt.Errorf("execute sql: %w", err)
	}

	// Get schema
	schemaJSON, err := ExtractEnrichedSchema(ctx, duckDB, "_result")
	if err != nil {
		return nil, fmt.Errorf("extract schema: %w", err)
	}

	// Count rows
	var rowCount int64
	if err := duckDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM _result").Scan(&rowCount); err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}

	// Export to Parquet
	parquetPath := filepath.Join(tmpDir, "output.parquet")
	if _, err := duckDB.ExecContext(ctx, fmt.Sprintf("COPY _result TO '%s' (FORMAT PARQUET)", parquetPath)); err != nil {
		return nil, fmt.Errorf("copy to parquet: %w", err)
	}

	// Upload to MinIO
	data, err := os.ReadFile(parquetPath)
	if err != nil {
		return nil, fmt.Errorf("read parquet: %w", err)
	}

	now := time.Now().UTC()
	outputKey := fmt.Sprintf("transforms/%s/dt=%s/%s.parquet",
		msg.TenantID,
		now.Format("2006-01-02"),
		msg.JobRunID,
	)
	if err := w.minio.PutParquet(ctx, outputKey, data); err != nil {
		return nil, fmt.Errorf("upload parquet: %w", err)
	}

	// Upsert dataset
	datasetName := fmt.Sprintf("transform_%s", msg.JobRunID[:8])
	lastUpdated := now
	dataset := &domain.Dataset{
		ID:            uuid.New().String(),
		TenantID:      msg.TenantID,
		Name:          datasetName,
		SourceType:    domain.SourceTypeTransform,
		SchemaJSON:    &schemaJSON,
		RowCount:      &rowCount,
		StoragePath:   outputKey,
		LastUpdatedAt: &lastUpdated,
	}
	if err := w.datasets.Upsert(ctx, dataset); err != nil {
		return nil, fmt.Errorf("upsert dataset: %w", err)
	}

	return &TransformResult{
		RowCount:  rowCount,
		OutputKey: outputKey,
	}, nil
}

