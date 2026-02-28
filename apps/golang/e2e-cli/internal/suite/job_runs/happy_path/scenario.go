package happy_path

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
)

type Scenario struct {
	email       string
	password    string
	displayName string
}

func NewScenario(email, password, displayName string) *Scenario {
	if email == "" {
		email = fmt.Sprintf("e2e_jr_%d@example.com", time.Now().UnixNano())
	}
	return &Scenario{
		email:       email,
		password:    password,
		displayName: displayName,
	}
}

func (s *Scenario) ID() string {
	return "job_runs/happy_path/create_list_get"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// Register a new user
	var registerResp struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	registerReq := map[string]string{
		"email":        s.email,
		"password":     s.password,
		"display_name": s.displayName,
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: expected 201, got %d body=%s", code, string(body))
	}

	// Login
	var loginResp struct {
		Token string `json:"token"`
	}
	loginReq := map[string]string{
		"email":    s.email,
		"password": s.password,
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login: expected 200, got %d body=%s", code, string(body))
	}

	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantID)

	// POST /api/v1/job-runs → 201
	var createResp struct {
		ID        string `json:"id"`
		TenantID  string `json:"tenant_id"`
		ProjectID string `json:"project_id"`
		JobID     string `json:"job_id"`
		Status    string `json:"status"`
	}
	createReq := map[string]string{
		"project_id": "proj-001",
		"job_id":     "job-001",
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/job-runs", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create job run: expected 201, got %d body=%s", code, string(body))
	}
	if createResp.ID == "" {
		return fmt.Errorf("create job run: missing id in response")
	}
	if createResp.TenantID != registerResp.TenantID {
		return fmt.Errorf("create job run: tenant_id mismatch: got=%s want=%s", createResp.TenantID, registerResp.TenantID)
	}
	if createResp.Status != "queued" {
		return fmt.Errorf("create job run: expected status 'queued', got '%s'", createResp.Status)
	}

	// GET /api/v1/job-runs → 200
	var listResp []struct {
		ID string `json:"id"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/job-runs", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list job runs: expected 200, got %d body=%s", code, string(body))
	}
	found := false
	for _, jr := range listResp {
		if jr.ID == createResp.ID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("list job runs: created job run %s not found in list", createResp.ID)
	}

	// GET /api/v1/job-runs/{id} → 200
	var getResp struct {
		ID        string `json:"id"`
		TenantID  string `json:"tenant_id"`
		ProjectID string `json:"project_id"`
		JobID     string `json:"job_id"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/job-runs/"+createResp.ID, &getResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get job run: expected 200, got %d body=%s", code, string(body))
	}
	if getResp.ID != createResp.ID {
		return fmt.Errorf("get job run: id mismatch: got=%s want=%s", getResp.ID, createResp.ID)
	}

	return nil
}
