package fetchers

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

// GoogleSheetsFetcher fetches sheet information from a Google Spreadsheet.
type GoogleSheetsFetcher struct{}

func NewGoogleSheetsFetcher() *GoogleSheetsFetcher {
	return &GoogleSheetsFetcher{}
}

type sheetsAPIResponse struct {
	Properties struct {
		Title string `json:"title"`
	} `json:"properties"`
	Sheets []struct {
		Properties struct {
			Title      string `json:"title"`
			SheetID    int    `json:"sheetId"`
			Index      int    `json:"index"`
			GridProps  *struct {
				RowCount    int `json:"rowCount"`
				ColumnCount int `json:"columnCount"`
			} `json:"gridProperties"`
		} `json:"properties"`
	} `json:"sheets"`
}

func (f *GoogleSheetsFetcher) FetchSchema(ctx context.Context, configJSON string, accessToken string) (*connector.SchemaResult, error) {
	var cfg googleSheetsConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	if cfg.SpreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required")
	}
	if accessToken == "" {
		return nil, fmt.Errorf("no access token available")
	}

	url := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s?fields=properties.title,sheets.properties",
		cfg.SpreadsheetID,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// parse below
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("credential_expired")
	case http.StatusForbidden:
		return nil, fmt.Errorf("insufficient permissions to access this spreadsheet")
	case http.StatusNotFound:
		return nil, fmt.Errorf("spreadsheet not found")
	default:
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp sheetsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	items := make([]connector.SchemaItem, 0, len(apiResp.Sheets))
	for _, s := range apiResp.Sheets {
		meta := map[string]any{}
		if s.Properties.GridProps != nil {
			meta["row_count"] = s.Properties.GridProps.RowCount
			meta["column_count"] = s.Properties.GridProps.ColumnCount
		}
		items = append(items, connector.SchemaItem{
			Name:     s.Properties.Title,
			Type:     "sheet",
			Metadata: meta,
		})
	}

	return &connector.SchemaResult{
		Title: apiResp.Properties.Title,
		Items: items,
	}, nil
}
