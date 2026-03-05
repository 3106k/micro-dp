package executors

import (
	"context"
	"fmt"

	"github.com/user/micro-dp/internal/connector"
	"github.com/user/micro-dp/worker"
)

// GoogleSheetsExecutor adapts the existing SheetsImportWriter to the ImportExecutor interface.
type GoogleSheetsExecutor struct {
	writer *worker.SheetsImportWriter
}

// NewGoogleSheetsExecutor creates a new GoogleSheetsExecutor wrapping the given SheetsImportWriter.
func NewGoogleSheetsExecutor(writer *worker.SheetsImportWriter) *GoogleSheetsExecutor {
	return &GoogleSheetsExecutor{writer: writer}
}

func (e *GoogleSheetsExecutor) ExecuteImport(ctx context.Context, params *connector.ImportParams) (*connector.ImportResult, error) {
	spreadsheetID, _ := params.Config["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return nil, fmt.Errorf("import config missing spreadsheet_id")
	}
	sheetName, _ := params.Config["sheet_name"].(string)
	cellRange, _ := params.Config["range"].(string)

	msg := &worker.SheetsImportMessage{
		JobRunID:      params.JobRunID,
		TenantID:      params.TenantID,
		SpreadsheetID: spreadsheetID,
		SheetName:     sheetName,
		Range:         cellRange,
		AccessToken:   params.AccessToken,
		JobID:         params.JobID,
		VersionID:     params.VersionID,
	}

	result, err := e.writer.Execute(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &connector.ImportResult{
		RowCount:  result.RowCount,
		OutputKey: result.OutputKey,
	}, nil
}
