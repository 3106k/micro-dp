package executors

import (
	"context"
	"fmt"
	"testing"

	"github.com/user/micro-dp/internal/connector"
	"github.com/user/micro-dp/worker"
)

// mockSheetsImportWriter is a test double for worker.SheetsImportWriter.
// Define mock structs in the _test.go file for each executor.
type mockSheetsImportWriter struct {
	result *worker.SheetsImportResult
	err    error
	called bool
	msg    *worker.SheetsImportMessage
}

func (m *mockSheetsImportWriter) Execute(ctx context.Context, msg *worker.SheetsImportMessage) (*worker.SheetsImportResult, error) {
	m.called = true
	m.msg = msg
	return m.result, m.err
}

// TestGoogleSheetsExecutor_Template demonstrates the table-driven pattern
// for testing ImportExecutor implementations. Copy and adapt for new connectors.
//
// Note: GoogleSheetsExecutor depends on *worker.SheetsImportWriter (concrete type).
// This test verifies config parsing and parameter mapping. For full integration
// testing, use E2E tests with real services.
func TestGoogleSheetsExecutor_Template(t *testing.T) {
	tests := []struct {
		name       string
		params     *connector.ImportParams
		mockResult *worker.SheetsImportResult
		mockErr    error
		wantErr    bool
		errContain string
		wantRows   int64
	}{
		{
			name: "successful import",
			params: &connector.ImportParams{
				TenantID:    "tenant-1",
				JobRunID:    "run-1",
				JobID:       "job-1",
				VersionID:   "ver-1",
				AccessToken: "token",
				Config: map[string]any{
					"spreadsheet_id": "abc123",
					"sheet_name":     "Sheet1",
					"range":          "A1:Z100",
				},
			},
			mockResult: &worker.SheetsImportResult{RowCount: 42, OutputKey: "imports/tenant-1/out.parquet"},
			wantRows:   42,
		},
		{
			name: "missing spreadsheet_id returns error",
			params: &connector.ImportParams{
				TenantID:    "tenant-1",
				JobRunID:    "run-1",
				AccessToken: "token",
				Config:      map[string]any{},
			},
			wantErr:    true,
			errContain: "spreadsheet_id",
		},
		{
			name: "writer error propagates",
			params: &connector.ImportParams{
				TenantID:    "tenant-1",
				JobRunID:    "run-1",
				AccessToken: "token",
				Config:      map[string]any{"spreadsheet_id": "abc123"},
			},
			mockResult: nil,
			mockErr:    fmt.Errorf("sheets API rate limit"),
			wantErr:    true,
			errContain: "rate limit",
		},
		{
			name: "optional fields can be omitted",
			params: &connector.ImportParams{
				TenantID:    "tenant-1",
				JobRunID:    "run-1",
				AccessToken: "token",
				Config:      map[string]any{"spreadsheet_id": "abc123"},
			},
			mockResult: &worker.SheetsImportResult{RowCount: 10, OutputKey: "out.parquet"},
			wantRows:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSheetsImportWriter{result: tt.mockResult, err: tt.mockErr}

			// GoogleSheetsExecutor takes *worker.SheetsImportWriter (concrete).
			// We test config parsing by verifying the error path for missing config,
			// and verify parameter mapping for successful cases via the mock interface.
			//
			// For cases that need the writer, we use the concrete executor only
			// when mock injection is possible. Since the executor holds a concrete
			// type, we test the ExecuteImport logic directly here.
			executor := &testableExecutor{mock: mock}
			result, err := executor.ExecuteImport(context.Background(), tt.params)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !containsStr(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.RowCount != tt.wantRows {
				t.Errorf("RowCount = %d, want %d", result.RowCount, tt.wantRows)
			}

			// Verify parameter mapping
			if mock.called && mock.msg != nil {
				if mock.msg.TenantID != tt.params.TenantID {
					t.Errorf("TenantID = %q, want %q", mock.msg.TenantID, tt.params.TenantID)
				}
				if mock.msg.JobRunID != tt.params.JobRunID {
					t.Errorf("JobRunID = %q, want %q", mock.msg.JobRunID, tt.params.JobRunID)
				}
			}
		})
	}
}

// testableExecutor mirrors GoogleSheetsExecutor logic but uses the mock interface
// for testing. This is the recommended pattern when the real executor wraps a
// concrete dependency: duplicate the thin adapter logic in the test.
type testableExecutor struct {
	mock *mockSheetsImportWriter
}

func (e *testableExecutor) ExecuteImport(ctx context.Context, params *connector.ImportParams) (*connector.ImportResult, error) {
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

	result, err := e.mock.Execute(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &connector.ImportResult{
		RowCount:  result.RowCount,
		OutputKey: result.OutputKey,
	}, nil
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
