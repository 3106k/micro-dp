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
	return "plan/happy_path/get_plan_and_usage"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register new user
	email := fmt.Sprintf("e2e_plan_%d@example.com", time.Now().UnixNano())
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

	// 3. GET /api/v1/plan -> 200
	var planResp openapi.TenantPlanResponse
	code, body, err = client.GetJSON(ctx, "/api/v1/plan", &planResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get plan: status=%d body=%s", code, string(body))
	}
	if planResp.Plan.Id == "" {
		return fmt.Errorf("get plan: plan.id is empty, body=%s", string(body))
	}
	if planResp.Plan.Name == "" {
		return fmt.Errorf("get plan: plan.name is empty, body=%s", string(body))
	}

	// 4. GET /api/v1/usage/summary -> 200
	var usageResp openapi.UsageSummaryResponse
	code, body, err = client.GetJSON(ctx, "/api/v1/usage/summary", &usageResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get usage summary: status=%d body=%s", code, string(body))
	}
	if usageResp.Date == "" {
		return fmt.Errorf("get usage summary: date is empty, body=%s", string(body))
	}

	// 5. GET /api/v1/plan without auth -> 401
	savedToken := loginResp.Token
	client.SetToken("")
	code, body, err = client.GetJSON(ctx, "/api/v1/plan", nil)
	if err != nil {
		return err
	}
	if code != 401 {
		return fmt.Errorf("get plan without auth: expected 401 got=%d body=%s", code, string(body))
	}
	client.SetToken(savedToken)

	return nil
}
