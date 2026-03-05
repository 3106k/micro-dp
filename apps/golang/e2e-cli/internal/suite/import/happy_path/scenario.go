package happy_path

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/openapi"
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
	return "import/happy_path/create_import_job"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	ts := time.Now().UnixNano()

	// 1. Register new user
	email := fmt.Sprintf("e2e_import_%d@example.com", ts)
	registerReq := openapi.RegisterRequest{
		Email:       openapi.Email(email),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}
	var registerResp openapi.RegisterResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: status=%d body=%s", code, string(body))
	}

	// 2. Login
	loginReq := openapi.LoginRequest{
		Email:    openapi.Email(email),
		Password: s.password,
	}
	var loginResp openapi.LoginResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login: status=%d body=%s", code, string(body))
	}
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantId)

	// 3. Create a source-google-sheets connection
	connConfig := `{"credentials_json":"{}"}`
	createConnReq := openapi.CreateConnectionRequest{
		Name:       "e2e-gsheets",
		Type:       "source-google-sheets",
		ConfigJson: openapi.Ptr(connConfig),
	}
	var connResp openapi.Connection
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", createConnReq, &connResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create connection: status=%d body=%s", code, string(body))
	}
	connID := connResp.Id

	// 4. Normal: POST /api/v1/import/jobs with source_config → 201
	srcCfg := map[string]interface{}{
		"spreadsheet_id": "abc123",
		"sheet_name":     "Sheet1",
		"range":          "A1:Z1000",
	}
	exec := openapi.ImportExecutionSaveOnly
	createReq := openapi.CreateImportJobRequest{
		Name:         "e2e-import",
		Slug:         fmt.Sprintf("e2e-import-%d", ts),
		ConnectionId: connID,
		Description:  openapi.Ptr("E2E import test"),
		Execution:    &exec,
		SourceConfig: &srcCfg,
	}
	var createResp openapi.CreateImportJobResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/import/jobs", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create import job: status=%d body=%s", code, string(body))
	}
	if createResp.Job.Id == "" {
		return fmt.Errorf("create import job: job.id is empty")
	}
	if createResp.Version.Id == "" {
		return fmt.Errorf("create import job: version.id is empty")
	}

	// 5. source_config omitted → 201 (optional field)
	createReqNoSrc := openapi.CreateImportJobRequest{
		Name:         "e2e-import-no-src",
		Slug:         fmt.Sprintf("e2e-import-no-src-%d", ts),
		ConnectionId: connID,
		Execution:    &exec,
	}
	var createRespNoSrc openapi.CreateImportJobResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/import/jobs", createReqNoSrc, &createRespNoSrc)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create import job (no source_config): status=%d body=%s", code, string(body))
	}

	// 6. Validation: name missing → 400
	createReqNoName := openapi.CreateImportJobRequest{
		Name:         "",
		Slug:         fmt.Sprintf("e2e-import-noname-%d", ts),
		ConnectionId: connID,
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/import/jobs", createReqNoName, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("create import job (no name): expected 400 got=%d body=%s", code, string(body))
	}

	// 7. Validation: connection_id missing → 400
	createReqNoConn := openapi.CreateImportJobRequest{
		Name: "e2e-import-noconn",
		Slug: fmt.Sprintf("e2e-import-noconn-%d", ts),
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/import/jobs", createReqNoConn, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("create import job (no connection_id): expected 400 got=%d body=%s", code, string(body))
	}

	// 8. Invalid connection_id → 400
	createReqBadConn := openapi.CreateImportJobRequest{
		Name:         "e2e-import-badconn",
		Slug:         fmt.Sprintf("e2e-import-badconn-%d", ts),
		ConnectionId: "nonexistent-connection-id",
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/import/jobs", createReqBadConn, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("create import job (bad connection_id): expected 400 got=%d body=%s", code, string(body))
	}

	return nil
}
