package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/user/micro-dp/domain"
)

// ExtractEnrichedSchema extracts column metadata with statistics and sample values from a DuckDB table.
func ExtractEnrichedSchema(ctx context.Context, db *sql.DB, tableName string) (string, error) {
	cols, err := describeTable(ctx, db, tableName)
	if err != nil {
		return "", fmt.Errorf("describe table: %w", err)
	}

	for i := range cols {
		stats, err := computeColumnStats(ctx, db, tableName, cols[i].Name)
		if err != nil {
			log.Printf("schema_extractor: stats for %s.%s failed: %v", tableName, cols[i].Name, err)
			continue
		}
		cols[i].Statistics = stats

		samples, err := sampleColumnValues(ctx, db, tableName, cols[i].Name)
		if err != nil {
			log.Printf("schema_extractor: samples for %s.%s failed: %v", tableName, cols[i].Name, err)
			continue
		}
		cols[i].SampleValues = samples
	}

	data, err := json.Marshal(cols)
	if err != nil {
		return "", fmt.Errorf("marshal schema: %w", err)
	}
	return string(data), nil
}

func describeTable(ctx context.Context, db *sql.DB, tableName string) ([]domain.DatasetColumnMeta, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("DESCRIBE %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []domain.DatasetColumnMeta
	for rows.Next() {
		var name, colType string
		var null, key, def, extra sql.NullString
		if err := rows.Scan(&name, &colType, &null, &key, &def, &extra); err != nil {
			return nil, err
		}
		nullable := null.Valid && null.String == "YES"
		cols = append(cols, domain.DatasetColumnMeta{
			Name:     name,
			Type:     colType,
			Nullable: nullable,
		})
	}
	return cols, rows.Err()
}

func computeColumnStats(ctx context.Context, db *sql.DB, tableName, colName string) (*domain.ColumnStatistics, error) {
	query := fmt.Sprintf(
		`SELECT
			CAST(MIN("%s") AS VARCHAR),
			CAST(MAX("%s") AS VARCHAR),
			COUNT(DISTINCT "%s"),
			SUM(CASE WHEN "%s" IS NULL THEN 1 ELSE 0 END),
			COUNT(*)
		FROM %s`,
		colName, colName, colName, colName, tableName,
	)

	var minVal, maxVal sql.NullString
	var distinctCount, nullCount, totalCount int64
	if err := db.QueryRowContext(ctx, query).Scan(&minVal, &maxVal, &distinctCount, &nullCount, &totalCount); err != nil {
		return nil, err
	}

	stats := &domain.ColumnStatistics{
		DistinctCount: distinctCount,
	}
	if minVal.Valid {
		stats.Min = &minVal.String
	}
	if maxVal.Valid {
		stats.Max = &maxVal.String
	}
	if totalCount > 0 {
		stats.NullRate = float64(nullCount) / float64(totalCount)
	}
	return stats, nil
}

func sampleColumnValues(ctx context.Context, db *sql.DB, tableName, colName string) ([]interface{}, error) {
	query := fmt.Sprintf(
		`SELECT DISTINCT CAST("%s" AS VARCHAR) FROM %s WHERE "%s" IS NOT NULL LIMIT 5`,
		colName, tableName, colName,
	)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []interface{}
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		samples = append(samples, val)
	}
	return samples, rows.Err()
}
