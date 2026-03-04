package testers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/user/micro-dp/internal/connector"
)

type googleSheetsConfig struct {
	SpreadsheetID string `json:"spreadsheet_id"`
}

// GoogleSheetsTester tests connectivity to a Google Spreadsheet.
type GoogleSheetsTester struct{}

func NewGoogleSheetsTester() *GoogleSheetsTester {
	return &GoogleSheetsTester{}
}

func (t *GoogleSheetsTester) Test(ctx context.Context, configJSON string, accessToken string) *connector.TestResult {
	var cfg googleSheetsConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return &connector.TestResult{OK: false, Code: "invalid_config", Message: "invalid config JSON"}
	}
	if cfg.SpreadsheetID == "" {
		return &connector.TestResult{OK: false, Code: "invalid_config", Message: "spreadsheet_id is required"}
	}
	if accessToken == "" {
		return &connector.TestResult{OK: false, Code: "unauthorized", Message: "no access token available"}
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=properties.title", cfg.SpreadsheetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return &connector.TestResult{OK: false, Code: "invalid_config", Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &connector.TestResult{OK: false, Code: "invalid_config", Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode == http.StatusOK:
		return &connector.TestResult{OK: true, Code: "ok", Message: "connected successfully"}
	case resp.StatusCode == http.StatusUnauthorized:
		return &connector.TestResult{OK: false, Code: "unauthorized", Message: "access token is invalid or expired"}
	case resp.StatusCode == http.StatusForbidden:
		return &connector.TestResult{OK: false, Code: "forbidden", Message: "insufficient permissions to access this spreadsheet"}
	case resp.StatusCode == http.StatusNotFound:
		return &connector.TestResult{OK: false, Code: "not_found", Message: "spreadsheet not found"}
	default:
		return &connector.TestResult{OK: false, Code: "invalid_config", Message: fmt.Sprintf("unexpected status %d", resp.StatusCode)}
	}
}
