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

	// Read all raw parquets (union_by_name handles mixed schemas from old/new data)
	rawGlob := filepath.Join(rawDir, "*.parquet")
	_, err = ddb.ExecContext(ctx, fmt.Sprintf(
		`CREATE TABLE raw_events AS SELECT * FROM read_parquet('%s', union_by_name=true)`, rawGlob))
	if err != nil {
		return fmt.Errorf("read raw parquets: %w", err)
	}

	// Ensure session_id column exists for legacy parquet files without it
	_, _ = ddb.ExecContext(ctx, `ALTER TABLE raw_events ADD COLUMN session_id VARCHAR DEFAULT ''`)

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

	// Aggregate visits: session-based with legacy fallback
	visitsPath := filepath.Join(tmpDir, "visits.parquet")
	_, err = ddb.ExecContext(ctx, fmt.Sprintf(`
		COPY (
			SELECT * FROM (
				-- Pass 1: session_id present (new SDK data)
				SELECT
					session_id,
					tenant_id,
					min(event_time) AS session_start,
					max(event_time) AS session_end,
					epoch(max(event_time) - min(event_time)) AS duration_seconds,
					count(*) AS event_count,
					count(DISTINCT
						CASE WHEN context IS NOT NULL AND context != '{}'
						     THEN json_extract_string(context, '$.page_url')
						     ELSE NULL
						END
					) AS page_count,
					first(
						CASE WHEN context IS NOT NULL AND context != '{}'
						     THEN json_extract_string(context, '$.page_url')
						     ELSE NULL
						END
						ORDER BY event_time ASC
					) AS landing_page,
					last(
						CASE WHEN context IS NOT NULL AND context != '{}'
						     THEN json_extract_string(context, '$.page_url')
						     ELSE NULL
						END
						ORDER BY event_time ASC
					) AS exit_page
				FROM raw_events
				WHERE session_id IS NOT NULL AND session_id != ''
				GROUP BY session_id, tenant_id

				UNION ALL

				-- Pass 2: no session_id (legacy data) — hourly pseudo-sessions
				SELECT
					concat('legacy-', tenant_id, '-', strftime(date_trunc('hour', event_time), '%%Y%%m%%dT%%H')) AS session_id,
					tenant_id,
					min(event_time) AS session_start,
					max(event_time) AS session_end,
					epoch(max(event_time) - min(event_time)) AS duration_seconds,
					count(*) AS event_count,
					count(DISTINCT
						CASE WHEN context IS NOT NULL AND context != '{}'
						     THEN json_extract_string(context, '$.page_url')
						     ELSE event_id
						END
					) AS page_count,
					first(
						CASE WHEN context IS NOT NULL AND context != '{}'
						     THEN json_extract_string(context, '$.page_url')
						     ELSE NULL
						END
						ORDER BY event_time ASC
					) AS landing_page,
					last(
						CASE WHEN context IS NOT NULL AND context != '{}'
						     THEN json_extract_string(context, '$.page_url')
						     ELSE NULL
						END
						ORDER BY event_time ASC
					) AS exit_page
				FROM raw_events
				WHERE session_id IS NULL OR session_id = ''
				GROUP BY tenant_id, date_trunc('hour', event_time)
			)
			ORDER BY session_start
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
