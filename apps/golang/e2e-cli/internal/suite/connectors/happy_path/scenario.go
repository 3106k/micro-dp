package happy_path

import (
	"context"
	"encoding/json"
	"fmt"
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
	return "connectors/happy_path/crud_and_test"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_connectors_%d@example.com", time.Now().UnixNano())
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

	// 3. GET /api/v1/connectors → 200, items non-empty
	var listResp struct {
		Items []json.RawMessage `json:"items"`
	}
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
	var sourceListResp struct {
		Items []struct {
			Kind string `json:"kind"`
		} `json:"items"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/connectors?kind=source", &sourceListResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list source connectors: status=%d body=%s", code, string(body))
	}
	for i, item := range sourceListResp.Items {
		if item.Kind != "source" {
			return fmt.Errorf("list source connectors: item[%d] kind=%s, expected source", i, item.Kind)
		}
	}

	// 5. GET /api/v1/connectors/source-postgres → 200, spec present
	var detailResp struct {
		ID   string         `json:"id"`
		Spec map[string]any `json:"spec"`
	}
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
	createReq := map[string]string{
		"name":        "test-pg",
		"type":        "source-postgres",
		"config_json": validConfig,
	}
	var createResp struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create connection (valid): status=%d body=%s", code, string(body))
	}

	// 8. POST /api/v1/connections (unknown type) → 400
	unknownReq := map[string]string{
		"name":        "bad-type",
		"type":        "source-unknown-db",
		"config_json": "{}",
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", unknownReq, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("create connection (unknown type): expected 400 got=%d body=%s", code, string(body))
	}

	// 9. POST /api/v1/connections (valid type + invalid config) → 422
	invalidReq := map[string]string{
		"name":        "bad-config",
		"type":        "source-postgres",
		"config_json": "{}",
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", invalidReq, nil)
	if err != nil {
		return err
	}
	if code != 422 {
		return fmt.Errorf("create connection (invalid config): expected 422 got=%d body=%s", code, string(body))
	}

	// 10. POST /api/v1/connections/test (valid) → 200, status=ok
	testReq := map[string]string{
		"type":        "source-postgres",
		"config_json": validConfig,
	}
	var testResp struct {
		Status string `json:"status"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections/test", testReq, &testResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("test connection (valid): status=%d body=%s", code, string(body))
	}
	if testResp.Status != "ok" {
		return fmt.Errorf("test connection: expected status=ok got=%s", testResp.Status)
	}

	// 11. POST /api/v1/connections/test (invalid config) → 422
	testInvalidReq := map[string]string{
		"type":        "source-postgres",
		"config_json": "{}",
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections/test", testInvalidReq, nil)
	if err != nil {
		return err
	}
	if code != 422 {
		return fmt.Errorf("test connection (invalid): expected 422 got=%d body=%s", code, string(body))
	}

	return nil
}
