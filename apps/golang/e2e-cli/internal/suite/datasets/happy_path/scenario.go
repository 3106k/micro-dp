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
	return "datasets/happy_path/list_and_get"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_datasets_%d@example.com", time.Now().UnixNano())
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

	// 3. GET /api/v1/datasets → 200, empty items
	var listResp openapi.ListResponse[openapi.Dataset]
	code, body, err = client.GetJSON(ctx, "/api/v1/datasets", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list datasets: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 0 {
		return fmt.Errorf("list datasets: expected 0 items, got=%d", len(listResp.Items))
	}

	// 4. GET /api/v1/datasets/nonexistent → 404
	code, body, err = client.GetJSON(ctx, "/api/v1/datasets/nonexistent", nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("get nonexistent dataset: expected 404 got=%d body=%s", code, string(body))
	}

	// 5. GET /api/v1/datasets?source_type=tracker → 200
	code, body, err = client.GetJSON(ctx, "/api/v1/datasets?source_type=tracker", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list datasets by source_type: status=%d body=%s", code, string(body))
	}

	// 6. GET /api/v1/datasets?q=test → 200
	code, body, err = client.GetJSON(ctx, "/api/v1/datasets?q=test", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list datasets by query: status=%d body=%s", code, string(body))
	}

	return nil
}
