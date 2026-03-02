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
	return "module_types/happy_path/crud_and_schemas"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register new user
	email := fmt.Sprintf("e2e_module_types_%d@example.com", time.Now().UnixNano())
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

	// 3. POST /api/v1/module_types -> 201 (#71)
	createReq := map[string]string{
		"name":     "test-source",
		"category": "source",
	}
	var createResp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/module_types", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create module type: status=%d body=%s", code, string(body))
	}
	if createResp.ID == "" {
		return fmt.Errorf("create module type: missing id in response")
	}
	if createResp.Name != "test-source" {
		return fmt.Errorf("create module type: name mismatch: got=%s want=test-source", createResp.Name)
	}
	if createResp.Category != "source" {
		return fmt.Errorf("create module type: category mismatch: got=%s want=source", createResp.Category)
	}
	moduleTypeID := createResp.ID

	// 4. GET /api/v1/module_types -> 200 (#72)
	var listResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/module_types", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list module types: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) < 1 {
		return fmt.Errorf("list module types: expected at least 1 item, got %d", len(listResp.Items))
	}
	found := false
	for _, item := range listResp.Items {
		if item.ID == moduleTypeID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("list module types: created module type %s not found in list", moduleTypeID)
	}

	// 5. GET /api/v1/module_types/{id} -> 200 (#73)
	var getResp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/module_types/"+moduleTypeID, &getResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get module type: status=%d body=%s", code, string(body))
	}
	if getResp.ID != moduleTypeID {
		return fmt.Errorf("get module type: id mismatch: got=%s want=%s", getResp.ID, moduleTypeID)
	}
	if getResp.Name != "test-source" {
		return fmt.Errorf("get module type: name mismatch: got=%s want=test-source", getResp.Name)
	}
	if getResp.Category != "source" {
		return fmt.Errorf("get module type: category mismatch: got=%s want=source", getResp.Category)
	}

	// 6. POST /api/v1/module_types/{id}/schemas -> 201 (#74)
	schemaReq := map[string]string{
		"version":     "1.0.0",
		"json_schema": `{"type":"object"}`,
	}
	var schemaResp struct {
		ID           string `json:"id"`
		ModuleTypeID string `json:"module_type_id"`
		Version      string `json:"version"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/module_types/"+moduleTypeID+"/schemas", schemaReq, &schemaResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create module type schema: status=%d body=%s", code, string(body))
	}
	if schemaResp.ID == "" {
		return fmt.Errorf("create module type schema: missing id in response")
	}
	if schemaResp.ModuleTypeID != moduleTypeID {
		return fmt.Errorf("create module type schema: module_type_id mismatch: got=%s want=%s", schemaResp.ModuleTypeID, moduleTypeID)
	}
	if schemaResp.Version != "1.0.0" {
		return fmt.Errorf("create module type schema: version mismatch: got=%s want=1.0.0", schemaResp.Version)
	}

	// 7. GET /api/v1/module_types/{id}/schemas -> 200 (#75)
	var listSchemasResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/module_types/"+moduleTypeID+"/schemas", &listSchemasResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list module type schemas: status=%d body=%s", code, string(body))
	}
	if len(listSchemasResp.Items) < 1 {
		return fmt.Errorf("list module type schemas: expected at least 1 item, got %d", len(listSchemasResp.Items))
	}

	// 8. GET /api/v1/module_types/nonexistent -> 404 (error case)
	code, body, err = client.GetJSON(ctx, "/api/v1/module_types/nonexistent", nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("get nonexistent module type: expected 404 got=%d body=%s", code, string(body))
	}

	return nil
}
