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
	return "module_types/happy_path/crud_and_schemas"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register new user
	email := fmt.Sprintf("e2e_module_types_%d@example.com", time.Now().UnixNano())
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

	// 3. POST /api/v1/module_types -> 201 (#71)
	createReq := openapi.CreateModuleTypeRequest{
		Name:     "test-source",
		Category: openapi.CreateModuleTypeRequestCategorySource,
	}
	var createResp openapi.ModuleType
	code, body, err = client.PostJSON(ctx, "/api/v1/module_types", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create module type: status=%d body=%s", code, string(body))
	}
	if createResp.Id == "" {
		return fmt.Errorf("create module type: missing id in response")
	}
	if createResp.Name != "test-source" {
		return fmt.Errorf("create module type: name mismatch: got=%s want=test-source", createResp.Name)
	}
	if createResp.Category != openapi.Source {
		return fmt.Errorf("create module type: category mismatch: got=%s want=source", createResp.Category)
	}
	moduleTypeID := createResp.Id

	// 4. GET /api/v1/module_types -> 200 (#72)
	var listResp openapi.ListResponse[openapi.ModuleType]
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
		if item.Id == moduleTypeID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("list module types: created module type %s not found in list", moduleTypeID)
	}

	// 5. GET /api/v1/module_types/{id} -> 200 (#73)
	var getResp openapi.ModuleType
	code, body, err = client.GetJSON(ctx, "/api/v1/module_types/"+moduleTypeID, &getResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get module type: status=%d body=%s", code, string(body))
	}
	if getResp.Id != moduleTypeID {
		return fmt.Errorf("get module type: id mismatch: got=%s want=%s", getResp.Id, moduleTypeID)
	}
	if getResp.Name != "test-source" {
		return fmt.Errorf("get module type: name mismatch: got=%s want=test-source", getResp.Name)
	}
	if getResp.Category != openapi.Source {
		return fmt.Errorf("get module type: category mismatch: got=%s want=source", getResp.Category)
	}

	// 6. POST /api/v1/module_types/{id}/schemas -> 201 (#74)
	schemaReq := openapi.CreateModuleTypeSchemaRequest{
		JsonSchema: `{"type":"object"}`,
	}
	var schemaResp openapi.ModuleTypeSchema
	code, body, err = client.PostJSON(ctx, "/api/v1/module_types/"+moduleTypeID+"/schemas", schemaReq, &schemaResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create module type schema: status=%d body=%s", code, string(body))
	}
	if schemaResp.Id == "" {
		return fmt.Errorf("create module type schema: missing id in response")
	}
	if schemaResp.ModuleTypeId != moduleTypeID {
		return fmt.Errorf("create module type schema: module_type_id mismatch: got=%s want=%s", schemaResp.ModuleTypeId, moduleTypeID)
	}
	if schemaResp.Version < 1 {
		return fmt.Errorf("create module type schema: version should be >= 1, got=%d", schemaResp.Version)
	}

	// 7. GET /api/v1/module_types/{id}/schemas -> 200 (#75)
	var listSchemasResp openapi.ListResponse[openapi.ModuleTypeSchema]
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
