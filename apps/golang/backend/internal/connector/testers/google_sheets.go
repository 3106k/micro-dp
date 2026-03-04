package testers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/user/micro-dp/internal/connector"
)

// GoogleSheetsTester tests connectivity by verifying the OAuth credential
// against the Google Sheets API.
type GoogleSheetsTester struct{}

func NewGoogleSheetsTester() *GoogleSheetsTester {
	return &GoogleSheetsTester{}
}

func (t *GoogleSheetsTester) Test(ctx context.Context, configJSON string, accessToken string) *connector.TestResult {
	if accessToken == "" {
		return &connector.TestResult{OK: false, Code: "unauthorized", Message: "no access token available"}
	}

	// Verify the token by calling the Sheets API with a minimal request.
	// Getting spreadsheet "" returns 404 if the token is valid, 401 if not.
	// We use a lightweight endpoint that only requires spreadsheets.readonly scope.
	url := "https://sheets.googleapis.com/v4/spreadsheets/__test_connectivity__?fields=properties.title"
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

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNotFound:
		// 404 = token is valid but spreadsheet doesn't exist (expected)
		return &connector.TestResult{OK: true, Code: "ok", Message: "connected successfully"}
	case http.StatusUnauthorized:
		return &connector.TestResult{OK: false, Code: "unauthorized", Message: "access token is invalid or expired"}
	case http.StatusForbidden:
		return &connector.TestResult{OK: false, Code: "forbidden", Message: "insufficient permissions"}
	default:
		return &connector.TestResult{OK: false, Code: "invalid_config", Message: fmt.Sprintf("unexpected status %d", resp.StatusCode)}
	}
}
