package admin_multi_tenant

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
	client.SetToken(loginResp.Token)
	client.SetTenantID("")

	type tenantResp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		IsActive bool   `json:"is_active"`
	}

	ts := time.Now().UnixNano()
	tenantNames := []string{
		fmt.Sprintf("E2E Tenant A %d", ts),
		fmt.Sprintf("E2E Tenant B %d", ts),
	}
	createdIDs := make([]string, 0, len(tenantNames))
	for _, name := range tenantNames {
		var tr tenantResp
		code, body, err = client.PostJSON(ctx, "/api/v1/admin/tenants", map[string]string{
			"name": name,
		}, &tr)
		if err != nil {
			return err
		}
		if code != 201 {
			return fmt.Errorf("create admin tenant: expected 201, got %d body=%s", code, string(body))
		}
		if tr.ID == "" || !tr.IsActive {
			return fmt.Errorf("create admin tenant: invalid response body=%s", string(body))
		}
		createdIDs = append(createdIDs, tr.ID)
	}

	var meResp struct {
		UserID       string `json:"user_id"`
		IsSuperadmin bool   `json:"is_superadmin"`
		Tenants      []struct {
			ID string `json:"id"`
		} `json:"tenants"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/auth/me", &meResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("me: expected 200, got %d body=%s", code, string(body))
	}
	if !meResp.IsSuperadmin {
		return fmt.Errorf("me: expected is_superadmin=true body=%s", string(body))
	}

	tenantSet := map[string]bool{}
	for _, t := range meResp.Tenants {
		tenantSet[t.ID] = true
	}
	for _, id := range createdIDs {
		if !tenantSet[id] {
			return fmt.Errorf("me: expected tenant %s in memberships", id)
		}
	}

	return nil
}
