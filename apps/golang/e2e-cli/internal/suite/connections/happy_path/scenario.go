package happy_path

import (
	"context"
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
	return "connections/happy_path/get_update_delete"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	ts := time.Now().UnixNano()

	// 1. Register new user
	email := fmt.Sprintf("e2e_connections_%d@example.com", ts)
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

	// 3. POST /api/v1/connections → 201 (create)
	validConfig := `{"host":"localhost","port":5432,"database":"mydb","username":"user","password":"pass"}`
	createReq := map[string]string{
		"name":        "e2e-pg",
		"type":        "source-postgres",
		"config_json": validConfig,
	}
	var createResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/connections", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create connection: status=%d body=%s", code, string(body))
	}
	if createResp.ID == "" {
		return fmt.Errorf("create connection: id is empty")
	}
	connID := createResp.ID

	// 4. GET /api/v1/connections/{id} → 200 (#76)
	var getResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/connections/"+connID, &getResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get connection: status=%d body=%s", code, string(body))
	}
	if getResp.Name != "e2e-pg" {
		return fmt.Errorf("get connection: expected name=e2e-pg got=%s", getResp.Name)
	}
	if getResp.ID != connID {
		return fmt.Errorf("get connection: expected id=%s got=%s", connID, getResp.ID)
	}

	// 5. PUT /api/v1/connections/{id} → 200 (#77)
	updatedConfig := `{"host":"localhost","port":5432,"database":"mydb2","username":"user","password":"pass"}`
	updateReq := map[string]string{
		"name":        "e2e-pg-updated",
		"type":        "source-postgres",
		"config_json": updatedConfig,
	}
	var updateResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	code, body, err = client.PutJSON(ctx, "/api/v1/connections/"+connID, updateReq, &updateResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("update connection: status=%d body=%s", code, string(body))
	}
	if updateResp.Name != "e2e-pg-updated" {
		return fmt.Errorf("update connection: expected name=e2e-pg-updated got=%s", updateResp.Name)
	}

	// 6. GET /api/v1/connections/{id} → 200 (verify update persisted)
	var getAfterUpdateResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/connections/"+connID, &getAfterUpdateResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get after update: status=%d body=%s", code, string(body))
	}
	if getAfterUpdateResp.Name != "e2e-pg-updated" {
		return fmt.Errorf("get after update: expected name=e2e-pg-updated got=%s", getAfterUpdateResp.Name)
	}

	// 7. DELETE /api/v1/connections/{id} → 204 (#78)
	code, body, err = client.Delete(ctx, "/api/v1/connections/"+connID)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("delete connection: expected 204 got=%d body=%s", code, string(body))
	}

	// 8. GET /api/v1/connections/{id} → 404 (verify deletion)
	code, body, err = client.GetJSON(ctx, "/api/v1/connections/"+connID, nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("get after delete: expected 404 got=%d body=%s", code, string(body))
	}

	// 9. GET /api/v1/connections/nonexistent → 404
	code, body, err = client.GetJSON(ctx, "/api/v1/connections/nonexistent", nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("get nonexistent: expected 404 got=%d body=%s", code, string(body))
	}

	return nil
}
