package worker

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/storage"
)

type ParquetWriter struct {
	minio *storage.MinIOClient
}

func NewParquetWriter(minio *storage.MinIOClient) *ParquetWriter {
	return &ParquetWriter{minio: minio}
}

func (w *ParquetWriter) WriteBatch(ctx context.Context, events []*domain.EventQueueMessage) error {
	if len(events) == 0 {
		return nil
	}

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return fmt.Errorf("open duckdb: %w", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `CREATE TABLE events (
		event_id VARCHAR,
		tenant_id VARCHAR,
		event_name VARCHAR,
		properties VARCHAR,
		event_time TIMESTAMP,
		received_at TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	stmt, err := db.PrepareContext(ctx, `INSERT INTO events VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, e := range events {
		_, err := stmt.ExecContext(ctx, e.EventID, e.TenantID, e.EventName, e.Properties, e.EventTime, e.ReceivedAt)
		if err != nil {
			return fmt.Errorf("insert event %s: %w", e.EventID, err)
		}
	}

	tmpDir, err := os.MkdirTemp("", "micro-dp-parquet-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	parquetPath := filepath.Join(tmpDir, "batch.parquet")
	_, err = db.ExecContext(ctx, fmt.Sprintf("COPY events TO '%s' (FORMAT PARQUET)", parquetPath))
	if err != nil {
		return fmt.Errorf("copy to parquet: %w", err)
	}

	data, err := os.ReadFile(parquetPath)
	if err != nil {
		return fmt.Errorf("read parquet: %w", err)
	}

	tenantID := events[0].TenantID
	now := time.Now().UTC()
	objectKey := fmt.Sprintf("events/%s/dt=%s/%d_%s.parquet",
		tenantID,
		now.Format("2006-01-02"),
		now.UnixMilli(),
		events[0].EventID[:8],
	)

	if err := w.minio.PutParquet(ctx, objectKey, data); err != nil {
		return fmt.Errorf("upload parquet: %w", err)
	}

	return nil
}
