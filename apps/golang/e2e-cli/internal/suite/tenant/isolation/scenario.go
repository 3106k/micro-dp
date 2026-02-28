package isolation

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
	return "tenant/isolation/cross_tenant_forbidden"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	ts := time.Now().UnixNano()

	// Register User A → gets tenant_a
	emailA := fmt.Sprintf("e2e_tenantA_%d@example.com", ts)
	var registerA struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", map[string]string{
		"email":        emailA,
		"password":     s.password,
		"display_name": s.displayName,
	}, &registerA)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register A: expected 201, got %d body=%s", code, string(body))
	}

	// Register User B → gets tenant_b
	emailB := fmt.Sprintf("e2e_tenantB_%d@example.com", ts)
	var registerB struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/register", map[string]string{
		"email":        emailB,
		"password":     s.password,
		"display_name": s.displayName,
	}, &registerB)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register B: expected 201, got %d body=%s", code, string(body))
	}

	// Login as User A
	var loginResp struct {
		Token string `json:"token"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", map[string]string{
		"email":    emailA,
		"password": s.password,
	}, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login A: expected 200, got %d body=%s", code, string(body))
	}

	// Use User A's token but User B's tenant → 403
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerB.TenantID)

	var errResp struct {
		Error string `json:"error"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/job-runs", &errResp)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("cross-tenant access: expected 403, got %d body=%s", code, string(body))
	}

	return nil
}
