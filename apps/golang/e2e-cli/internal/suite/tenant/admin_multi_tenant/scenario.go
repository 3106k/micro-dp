package admin_multi_tenant

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
}

func NewScenario(adminEmail, adminPassword string) *Scenario {
	return &Scenario{
		adminEmail:    adminEmail,
		adminPassword: adminPassword,
	}
}

func (s *Scenario) ID() string {
	return "tenant/admin_multi_tenant/create_and_membership"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	if s.adminEmail == "" || s.adminPassword == "" {
		return runner.Skip("admin credentials are not configured")
	}

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
	client.SetToken(loginResp.Token)
	client.SetTenantID("")

	ts := time.Now().UnixNano()
	tenantNames := []string{
		fmt.Sprintf("E2E Tenant A %d", ts),
		fmt.Sprintf("E2E Tenant B %d", ts),
	}
	createdIDs := make([]string, 0, len(tenantNames))
	for _, name := range tenantNames {
		var tr openapi.Tenant
		code, body, err = client.PostJSON(ctx, "/api/v1/admin/tenants", openapi.AdminCreateTenantRequest{
			Name: name,
		}, &tr)
		if err != nil {
			return err
		}
		if code != 201 {
			return fmt.Errorf("create admin tenant: expected 201, got %d body=%s", code, string(body))
		}
		if tr.Id == "" || !tr.IsActive {
			return fmt.Errorf("create admin tenant: invalid response body=%s", string(body))
		}
		createdIDs = append(createdIDs, tr.Id)
	}

	var meResp openapi.MeResponse
	code, body, err = client.GetJSON(ctx, "/api/v1/auth/me", &meResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("me: expected 200, got %d body=%s", code, string(body))
	}
	if meResp.PlatformRole != openapi.Superadmin {
		return fmt.Errorf("me: expected platform_role=superadmin body=%s", string(body))
	}

	tenantSet := map[string]bool{}
	for _, t := range meResp.Tenants {
		tenantSet[t.Id] = true
	}
	for _, id := range createdIDs {
		if !tenantSet[id] {
			return fmt.Errorf("me: expected tenant %s in memberships", id)
		}
	}

	return nil
}
