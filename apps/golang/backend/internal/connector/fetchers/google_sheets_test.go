package fetchers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

// mockFetcherRT intercepts HTTP requests and returns a preconfigured status + body.
// Use this pattern for any SchemaFetcher that calls external APIs via http.DefaultClient.
type mockFetcherRT struct {
	statusCode int
	body       string
}

func (m *mockFetcherRT) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.body)),
		Header:     make(http.Header),
	}, nil
}

// withMockTransport temporarily replaces http.DefaultTransport for the duration of the test.
func withMockTransport(t *testing.T, statusCode int, body string) {
	t.Helper()
	original := http.DefaultTransport
	http.DefaultTransport = &mockFetcherRT{statusCode: statusCode, body: body}
	t.Cleanup(func() { http.DefaultTransport = original })
}

const sheetsAPIResponseJSON = `{
  "properties": {"title": "My Spreadsheet"},
  "sheets": [
    {
      "properties": {
        "title": "Sheet1",
        "sheetId": 0,
        "index": 0,
        "gridProperties": {"rowCount": 100, "columnCount": 10}
      }
    },
    {
      "properties": {
        "title": "Sheet2",
        "sheetId": 1,
        "index": 1,
        "gridProperties": {"rowCount": 50, "columnCount": 5}
      }
    }
  ]
}`

// TestGoogleSheetsFetcher_Template demonstrates the table-driven pattern
// for testing SchemaFetcher implementations. Copy and adapt for new connectors.
func TestGoogleSheetsFetcher_Template(t *testing.T) {
	tests := []struct {
		name       string
		config     string
		token      string
		statusCode int
		body       string
		wantErr    bool
		errContain string
		wantItems  int
	}{
		{
			name:       "successful fetch returns sheets",
			config:     `{"spreadsheet_id":"abc123"}`,
			token:      "valid-token",
			statusCode: http.StatusOK,
			body:       sheetsAPIResponseJSON,
			wantItems:  2,
		},
		{
			name:       "unauthorized returns credential_expired error",
			config:     `{"spreadsheet_id":"abc123"}`,
			token:      "expired-token",
			statusCode: http.StatusUnauthorized,
			body:       `{}`,
			wantErr:    true,
			errContain: "credential_expired",
		},
		{
			name:       "forbidden returns permissions error",
			config:     `{"spreadsheet_id":"abc123"}`,
			token:      "no-scope-token",
			statusCode: http.StatusForbidden,
			body:       `{}`,
			wantErr:    true,
			errContain: "insufficient permissions",
		},
		{
			name:       "not found returns spreadsheet not found",
			config:     `{"spreadsheet_id":"nonexistent"}`,
			token:      "valid-token",
			statusCode: http.StatusNotFound,
			body:       `{}`,
			wantErr:    true,
			errContain: "not found",
		},
		{
			name:       "missing spreadsheet_id in config",
			config:     `{}`,
			token:      "valid-token",
			wantErr:    true,
			errContain: "spreadsheet_id is required",
		},
		{
			name:       "invalid config JSON",
			config:     `{invalid}`,
			token:      "valid-token",
			wantErr:    true,
			errContain: "invalid config JSON",
		},
		{
			name:       "empty access token",
			config:     `{"spreadsheet_id":"abc123"}`,
			token:      "",
			wantErr:    true,
			errContain: "no access token",
		},
	}

	fetcher := NewGoogleSheetsFetcher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.statusCode != 0 {
				withMockTransport(t, tt.statusCode, tt.body)
			}

			result, err := fetcher.FetchSchema(context.Background(), tt.config, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !contains(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Items) != tt.wantItems {
				t.Errorf("items count = %d, want %d", len(result.Items), tt.wantItems)
			}

			// Verify schema item structure
			if tt.wantItems > 0 {
				item := result.Items[0]
				if item.Name == "" {
					t.Error("first item name is empty")
				}
				if item.Type != "sheet" {
					t.Errorf("item type = %q, want %q", item.Type, "sheet")
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
