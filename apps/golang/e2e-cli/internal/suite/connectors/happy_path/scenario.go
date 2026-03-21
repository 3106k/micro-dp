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
	return "connectors/happy_path/crud_and_test"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_connectors_%d@example.com", time.Now().UnixNano())
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

	// 3. GET /api/v1/connectors → 200, items non-empty
	var listResp openapi.ListResponse[openapi.ConnectorDefinition]
	code, body, err = client.GetJSON(ctx, "/api/v1/connectors", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list connectors: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) == 0 {
		return fmt.Errorf("list connectors: expected non-empty items")
	}

	// 4. GET /api/v1/connectors?kind=source → 200, all items have kind=source
	var sourceListResp openapi.ListResponse[openapi.ConnectorDefinition]
	code, body, err = client.GetJSON(ctx, "/api/v1/connectors?kind=source", &sourceListResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list source connectors: status=%d body=%s", code, string(body))
	}
	for i, item := range sourceListResp.Items {
		if item.Kind != openapi.ConnectorKindSource {
			return fmt.Errorf("list source connectors: item[%d] kind=%s, expected source", i, item.Kind)
		}
	}

	// 5. GET /api/v1/connectors/source-postgres → 200, spec present
	var detailResp openapi.ConnectorDefinitionDetail
	code, body, err = client.GetJSON(ctx, "/api/v1/connectors/source-postgres", &detailResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get connector detail: status=%d body=%s", code, string(body))
	}
	if detailResp.Spec == nil {
		return fmt.Errorf("get connector detail: spec is nil")
	}

	// 6. GET /api/v1/connectors/nonexistent → 404
	code, body, err = client.GetJSON(ctx, "/api/v1/connectors/nonexistent", nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("get nonexistent connector: expected 404 got=%d body=%s", code, string(body))
	}

	// 7. POST /api/v1/connections (valid type + valid config) → 201
	validConfig := `{"host":"localhost","port":5432,"database":"mydb","username":"user","password":"pass"}`
	createReq := openapi.CreateConnectionRequest{
		Name:       "test-pg",
		Type:       "source-postgres",
		ConfigJson: openapi.Ptr(validConfig),
	}
	var createResp openapi.Connection
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create connection (valid): status=%d body=%s", code, string(body))
	}

	// 8. POST /api/v1/connections (unknown type) → 400
	unknownReq := openapi.CreateConnectionRequest{
		Name:       "bad-type",
		Type:       "source-unknown-db",
		ConfigJson: openapi.Ptr("{}"),
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", unknownReq, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("create connection (unknown type): expected 400 got=%d body=%s", code, string(body))
	}

	// 9. POST /api/v1/connections (valid type + invalid config) → 422
	invalidReq := openapi.CreateConnectionRequest{
		Name:       "bad-config",
		Type:       "source-postgres",
		ConfigJson: openapi.Ptr("{}"),
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", invalidReq, nil)
	if err != nil {
		return err
	}
	if code != 422 {
		return fmt.Errorf("create connection (invalid config): expected 422 got=%d body=%s", code, string(body))
	}

	// 10. POST /api/v1/connections/test (valid) → 200, status=ok
	testReq := openapi.TestConnectionRequest{
		Type:       "source-postgres",
		ConfigJson: validConfig,
	}
	var testResp openapi.TestConnectionResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/connections/test", testReq, &testResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("test connection (valid): status=%d body=%s", code, string(body))
	}
	if testResp.Validation.Status != openapi.ValidationResultStatusOk {
		return fmt.Errorf("test connection: expected validation.status=ok got=%s", testResp.Validation.Status)
	}
	// source-postgres has no tester registered → connectivity should be skipped
	if testResp.Connectivity.Status != openapi.ConnectivityResultStatusSkipped {
		return fmt.Errorf("test connection: expected connectivity.status=skipped got=%s", testResp.Connectivity.Status)
	}

	// 11. POST /api/v1/connections/test (invalid config) → 200 with validation.status=failed
	testInvalidReq := openapi.TestConnectionRequest{
		Type:       "source-postgres",
		ConfigJson: "{}",
	}
	var testInvalidResp openapi.TestConnectionResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/connections/test", testInvalidReq, &testInvalidResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("test connection (invalid): expected 200 got=%d body=%s", code, string(body))
	}
	if testInvalidResp.Validation.Status != openapi.ValidationResultStatusFailed {
		return fmt.Errorf("test connection (invalid): expected validation.status=failed got=%s", testInvalidResp.Validation.Status)
	}

	return nil
}
