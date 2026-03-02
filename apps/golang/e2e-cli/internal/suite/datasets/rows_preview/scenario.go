package rows_preview

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
)

type Scenario struct {
	password    string
	displayName string
}

func NewScenario(password, displayName string) *Scenario {
	return &Scenario{
		password:    password,
		displayName: displayName,
	}
}

func (s *Scenario) ID() string {
	return "datasets/rows_preview/get_rows"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_dataset_rows_%d@example.com", time.Now().UnixNano())
	registerReq := map[string]string{
		"email":        email,
		"password":     s.password,
		"display_name": s.displayName,
	}
	var registerResp struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: status=%d body=%s", code, string(body))
	}

	// 2. Login
	loginReq := map[string]string{
		"email":    email,
		"password": s.password,
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login: status=%d body=%s", code, string(body))
	}
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantID)

	// 3. POST /api/v1/uploads/presign (single file) -> 201
	presignReq := map[string]any{
		"files": []map[string]any{
			{
				"filename":     "test-rows.csv",
				"content_type": "text/csv",
				"size_bytes":   50,
			},
		},
	}
	var presignResp struct {
		UploadID string `json:"upload_id"`
		Files    []struct {
			FileID       string `json:"file_id"`
			PresignedURL string `json:"presigned_url"`
		} `json:"files"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/presign", presignReq, &presignResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("presign: status=%d body=%s", code, string(body))
	}
	if presignResp.UploadID == "" {
		return fmt.Errorf("presign: upload_id is empty")
	}
	if len(presignResp.Files) != 1 {
		return fmt.Errorf("presign: expected 1 file, got=%d", len(presignResp.Files))
	}
	if presignResp.Files[0].PresignedURL == "" {
		return fmt.Errorf("presign: presigned_url is empty")
	}

	// 4. PUT CSV data to presigned URL
	csvData := []byte("id,name,age\n1,Alice,30\n2,Bob,25\n")
	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, presignResp.Files[0].PresignedURL, bytes.NewReader(csvData))
	if err != nil {
		return fmt.Errorf("create put request: %w", err)
	}
	putReq.Header.Set("Content-Type", "text/csv")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return fmt.Errorf("put csv to presigned url: %w", err)
	}
	putResp.Body.Close()
	if putResp.StatusCode != 200 {
		return fmt.Errorf("put csv: expected 200 got=%d", putResp.StatusCode)
	}

	// 5. POST /api/v1/uploads/{upload_id}/complete -> 200
	var completeResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/"+presignResp.UploadID+"/complete", nil, &completeResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("complete: status=%d body=%s", code, string(body))
	}

	// 6. Poll GET /api/v1/datasets?source_type=import for up to 15 seconds
	var datasetID string
	var lastBody []byte
	found := false
	for attempt := 0; attempt < 15; attempt++ {
		time.Sleep(1 * time.Second)
		var datasetsResp struct {
			Items []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				SourceType string `json:"source_type"`
			} `json:"items"`
		}
		code, lastBody, err = client.GetJSON(ctx, "/api/v1/datasets?source_type=import", &datasetsResp)
		if err != nil {
			return err
		}
		if code != 200 {
			return fmt.Errorf("datasets list: status=%d body=%s", code, string(lastBody))
		}
		for _, ds := range datasetsResp.Items {
			if ds.SourceType == "import" {
				datasetID = ds.ID
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return fmt.Errorf("dataset with source_type=import not found after 15s in %s", string(lastBody))
	}

	// 7. GET /api/v1/datasets/{id}/rows -> 200
	var rowsResp struct {
		Columns   []map[string]any `json:"columns"`
		Rows      []map[string]any `json:"rows"`
		TotalRows int64            `json:"total_rows"`
		Limit     int              `json:"limit"`
		Offset    int              `json:"offset"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/datasets/"+datasetID+"/rows", &rowsResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get rows: status=%d body=%s", code, string(body))
	}
	if len(rowsResp.Columns) == 0 {
		return fmt.Errorf("get rows: columns array is empty")
	}
	if len(rowsResp.Rows) == 0 {
		return fmt.Errorf("get rows: rows array is empty")
	}
	if rowsResp.TotalRows == 0 {
		return fmt.Errorf("get rows: total_rows is 0")
	}

	// 8. GET /api/v1/datasets/{id}/rows?limit=1&offset=0 -> 200
	var paginatedResp struct {
		Columns   []map[string]any `json:"columns"`
		Rows      []map[string]any `json:"rows"`
		TotalRows int64            `json:"total_rows"`
		Limit     int              `json:"limit"`
		Offset    int              `json:"offset"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/datasets/"+datasetID+"/rows?limit=1&offset=0", &paginatedResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get rows paginated: status=%d body=%s", code, string(body))
	}
	if len(paginatedResp.Rows) != 1 {
		return fmt.Errorf("get rows paginated: expected 1 row, got=%d", len(paginatedResp.Rows))
	}

	// 9. GET /api/v1/datasets/nonexistent/rows -> 404
	code, body, err = client.GetJSON(ctx, "/api/v1/datasets/nonexistent/rows", nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("get nonexistent rows: expected 404 got=%d body=%s", code, string(body))
	}

	return nil
}
