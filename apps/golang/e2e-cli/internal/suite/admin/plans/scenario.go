package plans

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
	return "admin/plans/crud_and_assign"
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
	client.SetToken(loginResp.Token)
	client.SetTenantID("")

	// 2. GET /api/v1/admin/plans -> 200 (#88)
	var listPlansResp openapi.ListResponse[openapi.Plan]
	code, body, err = client.GetJSON(ctx, "/api/v1/admin/plans", &listPlansResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list plans: expected 200, got %d body=%s", code, string(body))
	}
	if listPlansResp.Items == nil {
		return fmt.Errorf("list plans: items is nil, body=%s", string(body))
	}

	// 3. POST /api/v1/admin/plans -> 201 (#89)
	ts := time.Now().UnixNano()
	planName := fmt.Sprintf("e2e-test-plan-%d", ts)
	var createdPlan openapi.Plan
	code, body, err = client.PostJSON(ctx, "/api/v1/admin/plans", openapi.CreatePlanRequest{
		Name:             planName,
		DisplayName:      "E2E Test Plan",
		MaxEventsPerDay:  openapi.Ptr(1000),
		MaxStorageBytes:  openapi.Ptr(int64(1073741824)),
		MaxRowsPerDay:    openapi.Ptr(5000),
		MaxUploadsPerDay: openapi.Ptr(10),
	}, &createdPlan)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create plan: expected 201, got %d body=%s", code, string(body))
	}
	if createdPlan.Id == "" {
		return fmt.Errorf("create plan: id is empty, body=%s", string(body))
	}
	if createdPlan.Name != planName {
		return fmt.Errorf("create plan: expected name=%s, got name=%s body=%s", planName, createdPlan.Name, string(body))
	}
	if createdPlan.DisplayName != "E2E Test Plan" {
		return fmt.Errorf("create plan: expected display_name='E2E Test Plan', got '%s' body=%s", createdPlan.DisplayName, string(body))
	}

	createdPlanID := createdPlan.Id

	// 4. PUT /api/v1/admin/plans/{id} -> 200 (#90)
	var updatedPlan openapi.Plan
	code, body, err = client.PutJSON(ctx, "/api/v1/admin/plans/"+createdPlanID, openapi.UpdatePlanRequest{
		DisplayName: openapi.Ptr("E2E Test Plan Updated"),
	}, &updatedPlan)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("update plan: expected 200, got %d body=%s", code, string(body))
	}
	if updatedPlan.Id != createdPlanID {
		return fmt.Errorf("update plan: expected id=%s, got id=%s body=%s", createdPlanID, updatedPlan.Id, string(body))
	}
	if updatedPlan.DisplayName != "E2E Test Plan Updated" {
		return fmt.Errorf("update plan: expected display_name='E2E Test Plan Updated', got '%s' body=%s", updatedPlan.DisplayName, string(body))
	}

	// 5. GET /api/v1/admin/tenants -> 200 to get a tenant_id
	var listTenantsResp openapi.ListResponse[openapi.Tenant]
	code, body, err = client.GetJSON(ctx, "/api/v1/admin/tenants", &listTenantsResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list tenants: expected 200, got %d body=%s", code, string(body))
	}
	if len(listTenantsResp.Items) == 0 {
		return fmt.Errorf("list tenants: no tenants found, body=%s", string(body))
	}

	tenantID := listTenantsResp.Items[0].Id

	// 6. POST /api/v1/admin/tenants/{tenant_id}/plan -> 200 (#91)
	var assignResp openapi.TenantPlanResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/admin/tenants/"+tenantID+"/plan", openapi.AssignPlanRequest{
		PlanId: createdPlanID,
	}, &assignResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("assign plan: expected 200, got %d body=%s", code, string(body))
	}
	if assignResp.Plan.Id != createdPlanID {
		return fmt.Errorf("assign plan: expected plan.id=%s, got plan.id=%s body=%s", createdPlanID, assignResp.Plan.Id, string(body))
	}
	if assignResp.StartedAt.IsZero() {
		return fmt.Errorf("assign plan: started_at is empty, body=%s", string(body))
	}

	return nil
}
