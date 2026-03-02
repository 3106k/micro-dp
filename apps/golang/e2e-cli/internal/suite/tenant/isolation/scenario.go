package isolation

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
	return "tenant/isolation/cross_tenant_forbidden"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	ts := time.Now().UnixNano()

	// Register User A → gets tenant_a
	emailA := fmt.Sprintf("e2e_tenantA_%d@example.com", ts)
	var registerA openapi.RegisterResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", openapi.RegisterRequest{
		Email:       openapi.Email(emailA),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}, &registerA)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register A: expected 201, got %d body=%s", code, string(body))
	}

	// Register User B → gets tenant_b
	emailB := fmt.Sprintf("e2e_tenantB_%d@example.com", ts)
	var registerB openapi.RegisterResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/register", openapi.RegisterRequest{
		Email:       openapi.Email(emailB),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}, &registerB)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register B: expected 201, got %d body=%s", code, string(body))
	}

	// Login as User A
	var loginResp openapi.LoginResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", openapi.LoginRequest{
		Email:    openapi.Email(emailA),
		Password: s.password,
	}, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login A: expected 200, got %d body=%s", code, string(body))
	}

	// Use User A's token but User B's tenant → 403
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerB.TenantId)

	var errResp openapi.ErrorResponse
	code, body, err = client.GetJSON(ctx, "/api/v1/job_runs", &errResp)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("cross-tenant access: expected 403, got %d body=%s", code, string(body))
	}

	return nil
}
