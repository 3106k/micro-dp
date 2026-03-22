package testers

import (
	"context"
	"net/http"
	"testing"
)

// mockRoundTripper intercepts HTTP requests and returns a preconfigured response.
// Use this pattern for any connector tester that calls external APIs via http.DefaultClient.
type mockRoundTripper struct {
	statusCode int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}, nil
}

// withMockTransport temporarily replaces http.DefaultTransport for the duration of the test.
// Restores the original transport via t.Cleanup.
func withMockTransport(t *testing.T, statusCode int) {
	t.Helper()
	original := http.DefaultTransport
	http.DefaultTransport = &mockRoundTripper{statusCode: statusCode}
	t.Cleanup(func() { http.DefaultTransport = original })
}

// TestGoogleSheetsTester_Template demonstrates the table-driven pattern
// for testing ConnectionTester implementations. Copy and adapt for new connectors.
func TestGoogleSheetsTester_Template(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		statusCode int
		wantOK     bool
		wantCode   string
	}{
		{
			name:       "valid token returns ok on 200",
			token:      "valid-token",
			statusCode: http.StatusOK,
			wantOK:     true,
			wantCode:   "ok",
		},
		{
			name:       "valid token returns ok on 404 (spreadsheet not found is expected)",
			token:      "valid-token",
			statusCode: http.StatusNotFound,
			wantOK:     true,
			wantCode:   "ok",
		},
		{
			name:       "expired token returns unauthorized",
			token:      "expired-token",
			statusCode: http.StatusUnauthorized,
			wantOK:     false,
			wantCode:   "unauthorized",
		},
		{
			name:       "insufficient scope returns forbidden",
			token:      "no-scope-token",
			statusCode: http.StatusForbidden,
			wantOK:     false,
			wantCode:   "forbidden",
		},
		{
			name:       "server error returns invalid_config",
			token:      "valid-token",
			statusCode: http.StatusInternalServerError,
			wantOK:     false,
			wantCode:   "invalid_config",
		},
		{
			name:     "empty token returns unauthorized without HTTP call",
			token:    "",
			wantOK:   false,
			wantCode: "unauthorized",
		},
	}

	tester := NewGoogleSheetsTester()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.token != "" {
				withMockTransport(t, tt.statusCode)
			}

			result := tester.Test(context.Background(), `{}`, tt.token)

			if result.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v", result.OK, tt.wantOK)
			}
			if result.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", result.Code, tt.wantCode)
			}
		})
	}
}
