package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/storage"
)

type ImportResult struct {
	RowCount   int64
	SchemaJSON string
	OutputKey  string
}

type CSVImportWriter struct {
	minio    *storage.MinIOClient
	datasets domain.DatasetRepository
}

func NewCSVImportWriter(minio *storage.MinIOClient, datasets domain.DatasetRepository) *CSVImportWriter {
	return &CSVImportWriter{minio: minio, datasets: datasets}
}

func (w *CSVImportWriter) ProcessFile(ctx context.Context, tenantID string, file domain.UploadJobFile) (*ImportResult, error) {
	tmpDir, err := os.MkdirTemp("", "micro-dp-csv-import-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download CSV from MinIO
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := w.minio.DownloadToFile(ctx, file.ObjectKey, csvPath); err != nil {
		return nil, fmt.Errorf("download csv: %w", err)
	}

	// Open DuckDB in-memory
	duckDB, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer duckDB.Close()

	// Import CSV
	_, err = duckDB.ExecContext(ctx, fmt.Sprintf("CREATE TABLE imported AS SELECT * FROM read_csv_auto('%s')", csvPath))
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	// Extract schema
	schemaJSON, err := extractSchema(ctx, duckDB)
	if err != nil {
		return nil, fmt.Errorf("extract schema: %w", err)
	}

	// Count rows
	var rowCount int64
	if err := duckDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM imported").Scan(&rowCount); err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}

	// Export to Parquet
	parquetPath := filepath.Join(tmpDir, "output.parquet")
	_, err = duckDB.ExecContext(ctx, fmt.Sprintf("COPY imported TO '%s' (FORMAT PARQUET)", parquetPath))
	if err != nil {
		return nil, fmt.Errorf("copy to parquet: %w", err)
	}

	// Upload Parquet to MinIO
	data, err := os.ReadFile(parquetPath)
	if err != nil {
		return nil, fmt.Errorf("read parquet: %w", err)
	}

	now := time.Now().UTC()
	outputKey := fmt.Sprintf("imports/%s/dt=%s/%s.parquet",
		tenantID,
		now.Format("2006-01-02"),
		file.FileID,
	)
	if err := w.minio.PutParquet(ctx, outputKey, data); err != nil {
		return nil, fmt.Errorf("upload parquet: %w", err)
	}

	// Upsert dataset
	datasetName := strings.TrimSuffix(file.FileName, filepath.Ext(file.FileName))
	lastUpdated := now
	dataset := &domain.Dataset{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Name:          datasetName,
		SourceType:    domain.SourceTypeImport,
		SchemaJSON:    &schemaJSON,
		RowCount:      &rowCount,
		StoragePath:   outputKey,
		LastUpdatedAt: &lastUpdated,
	}
	if err := w.datasets.Upsert(ctx, dataset); err != nil {
		return nil, fmt.Errorf("upsert dataset: %w", err)
	}

	return &ImportResult{
		RowCount:   rowCount,
		SchemaJSON: schemaJSON,
		OutputKey:  outputKey,
	}, nil
}

func extractSchema(ctx context.Context, db *sql.DB) (string, error) {
	rows, err := db.QueryContext(ctx, "DESCRIBE imported")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type column struct {
		Name string `json:"column_name"`
		Type string `json:"column_type"`
	}
	var columns []column
	for rows.Next() {
		var name, colType string
		var null, key, def, extra sql.NullString
		if err := rows.Scan(&name, &colType, &null, &key, &def, &extra); err != nil {
			return "", err
		}
		columns = append(columns, column{Name: name, Type: colType})
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	data, err := json.Marshal(columns)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
