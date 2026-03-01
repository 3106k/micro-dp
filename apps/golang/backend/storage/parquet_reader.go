package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

type ParquetColumn struct {
	Name string
	Type string
}

type ParquetRowsResult struct {
	Columns   []ParquetColumn
	Rows      []map[string]any
	TotalRows int64
}

func ReadParquetRows(ctx context.Context, filePath string, limit, offset int) (*ParquetRowsResult, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer db.Close()

	// Total row count
	var totalRows int64
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM read_parquet('%s')", filePath)
	if err := db.QueryRowContext(ctx, countQ).Scan(&totalRows); err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}

	// Column info via DESCRIBE
	descQ := fmt.Sprintf(
		"CREATE TEMP TABLE _preview AS SELECT * FROM read_parquet('%s') LIMIT 0",
		filePath,
	)
	if _, err := db.ExecContext(ctx, descQ); err != nil {
		return nil, fmt.Errorf("create preview table: %w", err)
	}
	descRows, err := db.QueryContext(ctx, "DESCRIBE _preview")
	if err != nil {
		return nil, fmt.Errorf("describe: %w", err)
	}
	var columns []ParquetColumn
	for descRows.Next() {
		var name, colType string
		var null, key, def, extra sql.NullString
		if err := descRows.Scan(&name, &colType, &null, &key, &def, &extra); err != nil {
			descRows.Close()
			return nil, fmt.Errorf("scan describe: %w", err)
		}
		columns = append(columns, ParquetColumn{Name: name, Type: colType})
	}
	descRows.Close()

	// Read rows with LIMIT/OFFSET
	dataQ := fmt.Sprintf(
		"SELECT * FROM read_parquet('%s') LIMIT %d OFFSET %d",
		filePath, limit, offset,
	)
	dataRows, err := db.QueryContext(ctx, dataQ)
	if err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}
	defer dataRows.Close()

	cols, _ := dataRows.Columns()
	var rows []map[string]any
	for dataRows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := dataRows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			v := values[i]
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			row[col] = v
		}
		rows = append(rows, row)
	}
	if rows == nil {
		rows = []map[string]any{}
	}

	return &ParquetRowsResult{
		Columns:   columns,
		Rows:      rows,
		TotalRows: totalRows,
	}, nil
}
