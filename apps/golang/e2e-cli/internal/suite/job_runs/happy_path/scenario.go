package happy_path

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/openapi"
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
	var registerResp openapi.RegisterResponse
	registerReq := openapi.RegisterRequest{
		Email:       openapi.Email(s.email),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: expected 201, got %d body=%s", code, string(body))
	}

	// Login
	var loginResp openapi.LoginResponse
	loginReq := openapi.LoginRequest{
		Email:    openapi.Email(s.email),
		Password: s.password,
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login: expected 200, got %d body=%s", code, string(body))
	}

	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantId)

	// POST /api/v1/jobs → 201 (create a job first)
	var jobResp openapi.Job
	jobReq := openapi.CreateJobRequest{
		Name: "Test Job",
		Slug: "test-job",
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/jobs", jobReq, &jobResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create job: expected 201, got %d body=%s", code, string(body))
	}
	if jobResp.Id == "" {
		return fmt.Errorf("create job: missing id in response")
	}

	// POST /api/v1/job_runs → 201
	var createResp openapi.JobRun
	createReq := openapi.CreateJobRunRequest{
		JobId: jobResp.Id,
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/job_runs", createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create job run: expected 201, got %d body=%s", code, string(body))
	}
	if createResp.Id == "" {
		return fmt.Errorf("create job run: missing id in response")
	}
	if createResp.TenantId != registerResp.TenantId {
		return fmt.Errorf("create job run: tenant_id mismatch: got=%s want=%s", createResp.TenantId, registerResp.TenantId)
	}
	if createResp.Status != openapi.JobRunStatusQueued {
		return fmt.Errorf("create job run: expected status 'queued', got '%s'", createResp.Status)
	}

	// GET /api/v1/job_runs → 200
	var listResp openapi.ListResponse[openapi.JobRun]
	code, body, err = client.GetJSON(ctx, "/api/v1/job_runs", &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list job runs: expected 200, got %d body=%s", code, string(body))
	}
	found := false
	for _, jr := range listResp.Items {
		if jr.Id == createResp.Id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("list job runs: created job run %s not found in list", createResp.Id)
	}

	// GET /api/v1/job_runs/{id} → 200
	var getResp openapi.JobRun
	code, body, err = client.GetJSON(ctx, "/api/v1/job_runs/"+createResp.Id, &getResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get job run: expected 200, got %d body=%s", code, string(body))
	}
	if getResp.Id != createResp.Id {
		return fmt.Errorf("get job run: id mismatch: got=%s want=%s", getResp.Id, createResp.Id)
	}

	return nil
}
