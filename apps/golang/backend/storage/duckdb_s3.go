package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// ConfigureDuckDBHTTPFS loads the httpfs extension and configures S3 credentials for DuckDB.
func ConfigureDuckDBHTTPFS(ctx context.Context, db *sql.DB, cfg S3Config) error {
	stmts := []string{
		"INSTALL httpfs",
		"LOAD httpfs",
		fmt.Sprintf("SET s3_endpoint = '%s'", cfg.Endpoint),
		fmt.Sprintf("SET s3_access_key_id = '%s'", cfg.AccessKey),
		fmt.Sprintf("SET s3_secret_access_key = '%s'", cfg.SecretKey),
		fmt.Sprintf("SET s3_use_ssl = %v", cfg.Secure),
		"SET s3_url_style = 'path'",
		fmt.Sprintf("SET s3_region = '%s'", cfg.Region),
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("duckdb httpfs setup (%s): %w", stmt, err)
		}
	}
	return nil
}

// S3ParquetURI builds an s3:// URI from a bucket name and object key.
func S3ParquetURI(bucket, objectKey string) string {
	return fmt.Sprintf("s3://%s/%s", bucket, objectKey)
}
