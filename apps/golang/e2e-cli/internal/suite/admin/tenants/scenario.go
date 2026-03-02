package tenants

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/openapi"
	"github.com/user/micro-dp/e2e-cli/internal/runner"
)

type Scenario struct {
	adminEmail    string
	adminPassword string
	password      string
	displayName   string
}

func NewScenario(adminEmail, adminPassword, password, displayName string) *Scenario {
	return &Scenario{
		adminEmail:    adminEmail,
		adminPassword: adminPassword,
		password:      password,
		displayName:   displayName,
	}
}

func (s *Scenario) ID() string {
	return "admin/tenants/list_and_patch"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	if s.adminEmail == "" || s.adminPassword == "" {
		return runner.Skip("admin credentials are not configured")
	}

	// 1. Login with admin credentials
	var loginResp openapi.LoginResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/login", openapi.LoginRequest{
		Email:    openapi.Email(s.adminEmail),
		Password: s.adminPassword,
	}, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("admin login: expected 200, got %d body=%s", code, string(body))
	}
	adminToken := loginResp.Token
	client.SetToken(adminToken)
	client.SetTenantID("")

	// 2. GET /api/v1/admin/tenants -> 200 (#86)
	var listResp openapi.ListResponse[openapi.Tenant]
	code, body, err = client.GetJSON(ctx, "/api/v1/admin/tenants", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list tenants: expected 200, got %d body=%s", code, string(body))
	}
	if listResp.Items == nil {
		return fmt.Errorf("list tenants: items is nil, body=%s", string(body))
	}
	if len(listResp.Items) == 0 {
		return fmt.Errorf("list tenants: items is empty, body=%s", string(body))
	}

	// Capture first tenant id for PATCH
	firstTenantID := listResp.Items[0].Id

	// 3. PATCH /api/v1/admin/tenants/{id} -> 200 (#87)
	var patchResp openapi.Tenant
	code, body, err = client.PatchJSON(ctx, "/api/v1/admin/tenants/"+firstTenantID, openapi.AdminUpdateTenantRequest{
		IsActive: openapi.Ptr(true),
	}, &patchResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("patch tenant: expected 200, got %d body=%s", code, string(body))
	}
	if patchResp.Id != firstTenantID {
		return fmt.Errorf("patch tenant: expected id=%s, got id=%s body=%s", firstTenantID, patchResp.Id, string(body))
	}
	if !patchResp.IsActive {
		return fmt.Errorf("patch tenant: expected is_active=true, body=%s", string(body))
	}

	// 4. GET /api/v1/admin/tenants without auth -> 401
	client.SetToken("")
	code, body, err = client.GetJSON(ctx, "/api/v1/admin/tenants", nil)
	if err != nil {
		return err
	}
	if code != 401 {
		return fmt.Errorf("list tenants without auth: expected 401, got %d body=%s", code, string(body))
	}

	// 5. GET /api/v1/admin/tenants with non-admin token -> 403
	// Register a regular user
	ts := time.Now().UnixNano()
	regEmail := fmt.Sprintf("e2e_admin_tenants_%d@example.com", ts)
	var regResp openapi.RegisterResponse
	client.SetToken("")
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/register", openapi.RegisterRequest{
		Email:       openapi.Email(regEmail),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}, &regResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register regular user: expected 201, got %d body=%s", code, string(body))
	}

	// Login as regular user
	var regLoginResp openapi.LoginResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", openapi.LoginRequest{
		Email:    openapi.Email(regEmail),
		Password: s.password,
	}, &regLoginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login regular user: expected 200, got %d body=%s", code, string(body))
	}

	// Try admin endpoint with regular user token -> 403
	client.SetToken(regLoginResp.Token)
	client.SetTenantID("")
	code, body, err = client.GetJSON(ctx, "/api/v1/admin/tenants", nil)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("list tenants as non-admin: expected 403, got %d body=%s", code, string(body))
	}

	return nil
}
