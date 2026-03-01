package happy_path

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
	return "uploads/happy_path/presign_and_complete"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_uploads_%d@example.com", time.Now().UnixNano())
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

	// 3. POST /api/v1/uploads/presign (single file) → 201
	presignReq := map[string]any{
		"files": []map[string]any{
			{
				"filename":     "test-data.csv",
				"content_type": "text/csv",
				"size_bytes":   1024,
			},
		},
	}
	var presignResp struct {
		UploadID string `json:"upload_id"`
		Files    []struct {
			FileID       string `json:"file_id"`
			Filename     string `json:"filename"`
			PresignedURL string `json:"presigned_url"`
			ObjectKey    string `json:"object_key"`
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
	if presignResp.Files[0].ObjectKey == "" {
		return fmt.Errorf("presign: object_key is empty")
	}

	// 3b. PUT CSV data to presigned URL
	csvData := []byte("id,name,age\n1,Alice,30\n2,Bob,25\n3,Charlie,35\n")
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

	// 4. POST /api/v1/uploads/{id}/complete → 200, status=uploaded
	var completeResp struct {
		ID       string `json:"id"`
		TenantID string `json:"tenant_id"`
		Status   string `json:"status"`
		Files    []struct {
			ID        string `json:"id"`
			UploadID  string `json:"upload_id"`
			FileName  string `json:"file_name"`
			ObjectKey string `json:"object_key"`
		} `json:"files"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/"+presignResp.UploadID+"/complete", nil, &completeResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("complete: status=%d body=%s", code, string(body))
	}
	if completeResp.Status != "uploaded" {
		return fmt.Errorf("complete: expected status=uploaded, got=%s", completeResp.Status)
	}
	if len(completeResp.Files) != 1 {
		return fmt.Errorf("complete: expected 1 file, got=%d", len(completeResp.Files))
	}

	// 4b. Poll for Worker to process CSV→Parquet (up to 15s)
	var lastBody []byte
	found := false
	for attempt := 0; attempt < 15; attempt++ {
		time.Sleep(1 * time.Second)
		var datasetsResp struct {
			Items []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				SourceType  string `json:"source_type"`
				StoragePath string `json:"storage_path"`
				RowCount    *int64 `json:"row_count"`
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
			if ds.Name == "test-data" && ds.SourceType == "import" {
				found = true
				if ds.StoragePath == "" {
					return fmt.Errorf("dataset storage_path is empty")
				}
				if ds.RowCount == nil || *ds.RowCount != 3 {
					return fmt.Errorf("dataset row_count: expected 3, got=%v", ds.RowCount)
				}
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return fmt.Errorf("dataset 'test-data' with source_type=import not found after 15s in %s", string(lastBody))
	}

	// 5. POST /api/v1/uploads/{id}/complete again → 409
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/"+presignResp.UploadID+"/complete", nil, nil)
	if err != nil {
		return err
	}
	if code != 409 {
		return fmt.Errorf("complete again: expected 409 got=%d body=%s", code, string(body))
	}

	// 6. POST /api/v1/uploads/presign (invalid extension .exe) → 400
	badExtReq := map[string]any{
		"files": []map[string]any{
			{
				"filename":     "malware.exe",
				"content_type": "application/octet-stream",
				"size_bytes":   1024,
			},
		},
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/presign", badExtReq, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("bad extension: expected 400 got=%d body=%s", code, string(body))
	}

	// 7. POST /api/v1/uploads/presign (size exceeds 100MB) → 400
	bigFileReq := map[string]any{
		"files": []map[string]any{
			{
				"filename":     "huge.csv",
				"content_type": "text/csv",
				"size_bytes":   200 * 1024 * 1024, // 200MB
			},
		},
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/presign", bigFileReq, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("size exceed: expected 400 got=%d body=%s", code, string(body))
	}

	// 8. POST /api/v1/uploads/nonexistent/complete → 404
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/nonexistent/complete", nil, nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("nonexistent complete: expected 404 got=%d body=%s", code, string(body))
	}

	return nil
}
