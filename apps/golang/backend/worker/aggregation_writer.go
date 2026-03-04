package worker

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/user/micro-dp/storage"
)

type AggregationWriter struct {
	minio *storage.MinIOClient
}

func NewAggregationWriter(minio *storage.MinIOClient) *AggregationWriter {
	return &AggregationWriter{minio: minio}
}

// AggregateEvents reads raw event parquet files for a tenant+date and writes
// aggregated events and visits parquet files.
func (w *AggregationWriter) AggregateEvents(ctx context.Context, tenantID, datePart string) error {
	// List raw parquet files for this tenant+date
	prefix := fmt.Sprintf("events/%s/dt=%s/", tenantID, datePart)
	keys, err := w.minio.ListObjectKeys(ctx, prefix)
	if err != nil {
		return fmt.Errorf("list raw files: %w", err)
	}
	if len(keys) == 0 {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", "micro-dp-agg-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download raw parquet files
	rawDir := filepath.Join(tmpDir, "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		return fmt.Errorf("create raw dir: %w", err)
	}

	for i, key := range keys {
		destPath := filepath.Join(rawDir, fmt.Sprintf("raw_%d.parquet", i))
		if err := w.minio.DownloadToFile(ctx, key, destPath); err != nil {
			return fmt.Errorf("download %s: %w", key, err)
		}
	}

	// Open in-memory DuckDB and aggregate
	ddb, err := sql.Open("duckdb", "")
	if err != nil {
		return fmt.Errorf("open duckdb: %w", err)
	}
	defer ddb.Close()

	// Read all raw parquets
	rawGlob := filepath.Join(rawDir, "*.parquet")
	_, err = ddb.ExecContext(ctx, fmt.Sprintf(
		`CREATE TABLE raw_events AS SELECT * FROM read_parquet('%s')`, rawGlob))
	if err != nil {
		return fmt.Errorf("read raw parquets: %w", err)
	}

	// Aggregate events: count by event_name per hour
	eventsPath := filepath.Join(tmpDir, "events.parquet")
	_, err = ddb.ExecContext(ctx, fmt.Sprintf(`
		COPY (
			SELECT
				tenant_id,
				event_name,
				date_trunc('hour', event_time) AS hour,
				count(*) AS event_count
			FROM raw_events
			GROUP BY tenant_id, event_name, date_trunc('hour', event_time)
			ORDER BY hour, event_name
		) TO '%s' (FORMAT PARQUET)`, eventsPath))
	if err != nil {
		return fmt.Errorf("aggregate events: %w", err)
	}

	// Aggregate visits: unique sessions per hour
	visitsPath := filepath.Join(tmpDir, "visits.parquet")
	_, err = ddb.ExecContext(ctx, fmt.Sprintf(`
		COPY (
			SELECT
				tenant_id,
				date_trunc('hour', event_time) AS hour,
				count(DISTINCT
					CASE WHEN context IS NOT NULL AND context != '{}'
					     THEN json_extract_string(context, '$.page_url')
					     ELSE event_id
					END
				) AS unique_pages,
				count(*) AS total_events
			FROM raw_events
			GROUP BY tenant_id, date_trunc('hour', event_time)
			ORDER BY hour
		) TO '%s' (FORMAT PARQUET)`, visitsPath))
	if err != nil {
		return fmt.Errorf("aggregate visits: %w", err)
	}

	// Upload aggregated files
	now := time.Now().UTC()
	timestamp := now.Format("20060102T150405")

	eventsData, err := os.ReadFile(eventsPath)
	if err != nil {
		return fmt.Errorf("read events parquet: %w", err)
	}
	eventsKey := fmt.Sprintf("aggregated/events/%s/dt=%s/%s.parquet", tenantID, datePart, timestamp)
	if err := w.minio.PutParquet(ctx, eventsKey, eventsData); err != nil {
		return fmt.Errorf("upload events parquet: %w", err)
	}

	visitsData, err := os.ReadFile(visitsPath)
	if err != nil {
		return fmt.Errorf("read visits parquet: %w", err)
	}
	visitsKey := fmt.Sprintf("aggregated/visits/%s/dt=%s/%s.parquet", tenantID, datePart, timestamp)
	if err := w.minio.PutParquet(ctx, visitsKey, visitsData); err != nil {
		return fmt.Errorf("upload visits parquet: %w", err)
	}

	return nil
}
