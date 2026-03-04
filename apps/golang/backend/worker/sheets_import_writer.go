package worker

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/storage"
)

type SheetsImportMessage struct {
	JobRunID      string
	TenantID      string
	SpreadsheetID string
	SheetName     string
	Range         string
	AccessToken   string
	JobID         string
	VersionID     string
}

type SheetsImportResult struct {
	RowCount  int64
	OutputKey string
}

type SheetsImportWriter struct {
	minio    *storage.MinIOClient
	datasets domain.DatasetRepository
}

func NewSheetsImportWriter(minio *storage.MinIOClient, datasets domain.DatasetRepository) *SheetsImportWriter {
	return &SheetsImportWriter{minio: minio, datasets: datasets}
}

func (w *SheetsImportWriter) Execute(ctx context.Context, msg *SheetsImportMessage) (*SheetsImportResult, error) {
	tmpDir, err := os.MkdirTemp("", "micro-dp-sheets-import-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Resolve sheet name if empty
	sheetName := msg.SheetName
	spreadsheetTitle := ""
	if sheetName == "" {
		title, firstSheet, err := w.getSpreadsheetInfo(ctx, msg.AccessToken, msg.SpreadsheetID)
		if err != nil {
			return nil, fmt.Errorf("get spreadsheet info: %w", err)
		}
		sheetName = firstSheet
		spreadsheetTitle = title
	} else {
		title, _, err := w.getSpreadsheetInfo(ctx, msg.AccessToken, msg.SpreadsheetID)
		if err != nil {
			spreadsheetTitle = msg.SpreadsheetID
		} else {
			spreadsheetTitle = title
		}
	}

	// Fetch data from Sheets API
	values, err := w.getSheetValues(ctx, msg.AccessToken, msg.SpreadsheetID, sheetName, msg.Range)
	if err != nil {
		return nil, fmt.Errorf("get sheet values: %w", err)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("sheet returned no data")
	}

	// Write to CSV temp file
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := w.writeCSV(csvPath, values); err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}

	// DuckDB: CSV → Parquet
	duckDB, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer duckDB.Close()

	_, err = duckDB.ExecContext(ctx, fmt.Sprintf("CREATE TABLE imported AS SELECT * FROM read_csv('%s', header=true, auto_detect=true, delim=',', null_padding=true, ignore_errors=true)", csvPath))
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	schemaJSON, err := extractSheetsSchema(ctx, duckDB)
	if err != nil {
		return nil, fmt.Errorf("extract schema: %w", err)
	}

	var rowCount int64
	if err := duckDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM imported").Scan(&rowCount); err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}

	parquetPath := filepath.Join(tmpDir, "output.parquet")
	_, err = duckDB.ExecContext(ctx, fmt.Sprintf("COPY imported TO '%s' (FORMAT PARQUET)", parquetPath))
	if err != nil {
		return nil, fmt.Errorf("copy to parquet: %w", err)
	}

	// Upload to MinIO
	data, err := os.ReadFile(parquetPath)
	if err != nil {
		return nil, fmt.Errorf("read parquet: %w", err)
	}

	now := time.Now().UTC()
	outputKey := fmt.Sprintf("sheets_imports/%s/dt=%s/%s.parquet",
		msg.TenantID,
		now.Format("2006-01-02"),
		msg.JobRunID,
	)
	if err := w.minio.PutParquet(ctx, outputKey, data); err != nil {
		return nil, fmt.Errorf("upload parquet: %w", err)
	}

	// Upsert dataset
	datasetName := fmt.Sprintf("%s - %s", spreadsheetTitle, sheetName)
	lastUpdated := now
	dataset := &domain.Dataset{
		ID:            uuid.New().String(),
		TenantID:      msg.TenantID,
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

	return &SheetsImportResult{
		RowCount:  rowCount,
		OutputKey: outputKey,
	}, nil
}

// getSpreadsheetInfo fetches the spreadsheet title and first sheet name.
func (w *SheetsImportWriter) getSpreadsheetInfo(ctx context.Context, accessToken, spreadsheetID string) (title string, firstSheet string, err error) {
	apiURL := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=properties.title,sheets.properties.title", url.PathEscape(spreadsheetID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("sheets api returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Properties struct {
			Title string `json:"title"`
		} `json:"properties"`
		Sheets []struct {
			Properties struct {
				Title string `json:"title"`
			} `json:"properties"`
		} `json:"sheets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Sheets) == 0 {
		return "", "", fmt.Errorf("spreadsheet has no sheets")
	}

	return result.Properties.Title, result.Sheets[0].Properties.Title, nil
}

// getSheetValues fetches cell values from the Sheets API.
func (w *SheetsImportWriter) getSheetValues(ctx context.Context, accessToken, spreadsheetID, sheetName, cellRange string) ([][]interface{}, error) {
	rangeStr := sheetName
	if cellRange != "" {
		rangeStr = sheetName + "!" + cellRange
	}

	apiURL := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s",
		url.PathEscape(spreadsheetID),
		url.PathEscape(rangeStr),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sheets api returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Values [][]interface{} `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Values, nil
}

// writeCSV writes the Sheets values to a CSV file. First row is treated as header.
// Sheets API omits trailing empty cells, so rows may have fewer columns than the header.
// This function pads all rows to match the header length.
func (w *SheetsImportWriter) writeCSV(path string, values [][]interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Determine max column count from header (first row)
	numCols := 0
	if len(values) > 0 {
		numCols = len(values[0])
	}

	for _, row := range values {
		record := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			if i < len(row) {
				record[i] = fmt.Sprintf("%v", row[i])
			}
			// else remains "" (empty string)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func extractSheetsSchema(ctx context.Context, db *sql.DB) (string, error) {
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
