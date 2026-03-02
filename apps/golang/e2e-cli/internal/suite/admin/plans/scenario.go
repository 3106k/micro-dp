package plans

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
	return "admin/plans/crud_and_assign"
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
	client.SetToken(loginResp.Token)
	client.SetTenantID("")

	// 2. GET /api/v1/admin/plans -> 200 (#88)
	type planItem struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		DisplayName     string `json:"display_name"`
		MaxEventsPerDay int    `json:"max_events_per_day"`
		MaxStorageBytes int64  `json:"max_storage_bytes"`
		MaxRowsPerDay   int    `json:"max_rows_per_day"`
		MaxUploadsPerDay int   `json:"max_uploads_per_day"`
		IsDefault       bool   `json:"is_default"`
	}
	var listPlansResp struct {
		Items []planItem `json:"items"`
	}
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
	createReq := map[string]any{
		"name":                planName,
		"display_name":        "E2E Test Plan",
		"max_events_per_day":  1000,
		"max_storage_bytes":   1073741824,
		"max_rows_per_day":    5000,
		"max_uploads_per_day": 10,
	}
	var createdPlan planItem
	code, body, err = client.PostJSON(ctx, "/api/v1/admin/plans", createReq, &createdPlan)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create plan: expected 201, got %d body=%s", code, string(body))
	}
	if createdPlan.ID == "" {
		return fmt.Errorf("create plan: id is empty, body=%s", string(body))
	}
	if createdPlan.Name != planName {
		return fmt.Errorf("create plan: expected name=%s, got name=%s body=%s", planName, createdPlan.Name, string(body))
	}
	if createdPlan.DisplayName != "E2E Test Plan" {
		return fmt.Errorf("create plan: expected display_name='E2E Test Plan', got '%s' body=%s", createdPlan.DisplayName, string(body))
	}

	createdPlanID := createdPlan.ID

	// 4. PUT /api/v1/admin/plans/{id} -> 200 (#90)
	updateReq := map[string]any{
		"display_name": "E2E Test Plan Updated",
	}
	var updatedPlan planItem
	code, body, err = client.PutJSON(ctx, "/api/v1/admin/plans/"+createdPlanID, updateReq, &updatedPlan)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("update plan: expected 200, got %d body=%s", code, string(body))
	}
	if updatedPlan.ID != createdPlanID {
		return fmt.Errorf("update plan: expected id=%s, got id=%s body=%s", createdPlanID, updatedPlan.ID, string(body))
	}
	if updatedPlan.DisplayName != "E2E Test Plan Updated" {
		return fmt.Errorf("update plan: expected display_name='E2E Test Plan Updated', got '%s' body=%s", updatedPlan.DisplayName, string(body))
	}

	// 5. GET /api/v1/admin/tenants -> 200 to get a tenant_id
	type tenantItem struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		IsActive bool   `json:"is_active"`
	}
	var listTenantsResp struct {
		Items []tenantItem `json:"items"`
	}
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

	tenantID := listTenantsResp.Items[0].ID

	// 6. POST /api/v1/admin/tenants/{tenant_id}/plan -> 200 (#91)
	assignReq := map[string]string{
		"plan_id": createdPlanID,
	}
	var assignResp struct {
		Plan struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"plan"`
		StartedAt string `json:"started_at"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/admin/tenants/"+tenantID+"/plan", assignReq, &assignResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("assign plan: expected 200, got %d body=%s", code, string(body))
	}
	if assignResp.Plan.ID != createdPlanID {
		return fmt.Errorf("assign plan: expected plan.id=%s, got plan.id=%s body=%s", createdPlanID, assignResp.Plan.ID, string(body))
	}
	if assignResp.StartedAt == "" {
		return fmt.Errorf("assign plan: started_at is empty, body=%s", string(body))
	}

	return nil
}
