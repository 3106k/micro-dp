package tenants

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
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
	var loginResp struct {
		Token string `json:"token"`
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/login", map[string]string{
		"email":    s.adminEmail,
		"password": s.adminPassword,
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
	type tenantItem struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		IsActive bool   `json:"is_active"`
	}
	var listResp struct {
		Items []tenantItem `json:"items"`
	}
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
	firstTenantID := listResp.Items[0].ID

	// 3. PATCH /api/v1/admin/tenants/{id} -> 200 (#87)
	var patchResp tenantItem
	code, body, err = client.PatchJSON(ctx, "/api/v1/admin/tenants/"+firstTenantID, map[string]any{
		"is_active": true,
	}, &patchResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("patch tenant: expected 200, got %d body=%s", code, string(body))
	}
	if patchResp.ID != firstTenantID {
		return fmt.Errorf("patch tenant: expected id=%s, got id=%s body=%s", firstTenantID, patchResp.ID, string(body))
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
	var regResp struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	client.SetToken("")
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/register", map[string]string{
		"email":        regEmail,
		"password":     s.password,
		"display_name": s.displayName,
	}, &regResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register regular user: expected 201, got %d body=%s", code, string(body))
	}

	// Login as regular user
	var regLoginResp struct {
		Token string `json:"token"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", map[string]string{
		"email":    regEmail,
		"password": s.password,
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
