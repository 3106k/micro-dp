package happy_path

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
	return "jobs/happy_path/crud_and_versions"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register new user
	email := fmt.Sprintf("e2e_jobs_%d@example.com", time.Now().UnixNano())
	registerReq := map[string]string{
		"email":        email,
		"password":     s.password,
		"display_name": s.displayName,
	}
	var registerResp struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: expected 201, got %d body=%s", code, string(body))
	}

	// 2. Login
	loginReq := map[string]string{
		"email":    email,
		"password": s.password,
	}
	var loginResp struct {
		Token string `json:"token"`
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

	// 3. POST /api/v1/jobs -> 201 (create job)
	jobReq := map[string]interface{}{
		"name":        "E2E Test Job",
		"slug":        fmt.Sprintf("e2e-test-job-%d", time.Now().UnixNano()),
		"description": "E2E test job for jobs scenario",
	}
	var createJobResp struct {
		ID       string `json:"id"`
		TenantID string `json:"tenant_id"`
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		IsActive bool   `json:"is_active"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/jobs", jobReq, &createJobResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create job: expected 201, got %d body=%s", code, string(body))
	}
	if createJobResp.ID == "" {
		return fmt.Errorf("create job: missing id in response")
	}
	if createJobResp.Name != "E2E Test Job" {
		return fmt.Errorf("create job: name mismatch: got=%s want=E2E Test Job", createJobResp.Name)
	}
	if createJobResp.Slug == "" {
		return fmt.Errorf("create job: missing slug in response")
	}
	jobID := createJobResp.ID

	// 4. GET /api/v1/jobs -> 200 (list jobs)
	var listJobsResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/jobs", &listJobsResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list jobs: expected 200, got %d body=%s", code, string(body))
	}
	if len(listJobsResp.Items) < 1 {
		return fmt.Errorf("list jobs: expected at least 1 item, got %d", len(listJobsResp.Items))
	}
	found := false
	for _, item := range listJobsResp.Items {
		if item.ID == jobID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("list jobs: created job %s not found in list", jobID)
	}

	// 5. GET /api/v1/jobs/{id} -> 200 (#65)
	var getJobResp struct {
		ID       string `json:"id"`
		TenantID string `json:"tenant_id"`
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		IsActive bool   `json:"is_active"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/jobs/"+jobID, &getJobResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get job: expected 200, got %d body=%s", code, string(body))
	}
	if getJobResp.ID != jobID {
		return fmt.Errorf("get job: id mismatch: got=%s want=%s", getJobResp.ID, jobID)
	}
	if getJobResp.Name != "E2E Test Job" {
		return fmt.Errorf("get job: name mismatch: got=%s want=E2E Test Job", getJobResp.Name)
	}

	// 6. PUT /api/v1/jobs/{id} -> 200 (#66)
	updateReq := map[string]interface{}{
		"name":        "Updated E2E Job",
		"slug":        getJobResp.Slug,
		"is_active":   true,
		"description": "Updated description",
	}
	var updateJobResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	code, body, err = client.PutJSON(ctx, "/api/v1/jobs/"+jobID, updateReq, &updateJobResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("update job: expected 200, got %d body=%s", code, string(body))
	}
	if updateJobResp.Name != "Updated E2E Job" {
		return fmt.Errorf("update job: name mismatch: got=%s want=Updated E2E Job", updateJobResp.Name)
	}

	// Prerequisite for versions: create a module type
	moduleTypeReq := map[string]string{
		"name":     fmt.Sprintf("e2e-source-%d", time.Now().UnixNano()),
		"category": "source",
	}
	var moduleTypeResp struct {
		ID string `json:"id"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/module_types", moduleTypeReq, &moduleTypeResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create module type: expected 201, got %d body=%s", code, string(body))
	}
	if moduleTypeResp.ID == "" {
		return fmt.Errorf("create module type: missing id in response")
	}

	// 7. POST /api/v1/jobs/{job_id}/versions -> 201 (#67)
	versionReq := map[string]interface{}{
		"modules": []map[string]interface{}{
			{
				"module_type_id": moduleTypeResp.ID,
				"name":           "source-module",
			},
		},
	}
	var createVersionResp struct {
		ID      string `json:"id"`
		JobID   string `json:"job_id"`
		Version int    `json:"version"`
		Status  string `json:"status"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/jobs/"+jobID+"/versions", versionReq, &createVersionResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create version: expected 201, got %d body=%s", code, string(body))
	}
	if createVersionResp.ID == "" {
		return fmt.Errorf("create version: missing id in response")
	}
	if createVersionResp.Version < 1 {
		return fmt.Errorf("create version: expected version >= 1, got %d", createVersionResp.Version)
	}
	if createVersionResp.Status != "draft" {
		return fmt.Errorf("create version: expected status 'draft', got '%s'", createVersionResp.Status)
	}
	versionID := createVersionResp.ID

	// 8. GET /api/v1/jobs/{job_id}/versions -> 200 (list versions)
	var listVersionsResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/jobs/"+jobID+"/versions", &listVersionsResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list versions: expected 200, got %d body=%s", code, string(body))
	}
	if len(listVersionsResp.Items) < 1 {
		return fmt.Errorf("list versions: expected at least 1 item, got %d", len(listVersionsResp.Items))
	}
	found = false
	for _, item := range listVersionsResp.Items {
		if item.ID == versionID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("list versions: created version %s not found in list", versionID)
	}

	// 9. GET /api/v1/jobs/{job_id}/versions/{version_id} -> 200 (#69)
	var versionDetailResp struct {
		Version struct {
			ID      string `json:"id"`
			JobID   string `json:"job_id"`
			Version int    `json:"version"`
			Status  string `json:"status"`
		} `json:"version"`
		Modules []struct {
			ID           string `json:"id"`
			ModuleTypeID string `json:"module_type_id"`
			Name         string `json:"name"`
		} `json:"modules"`
		Edges []struct {
			ID string `json:"id"`
		} `json:"edges"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/jobs/"+jobID+"/versions/"+versionID, &versionDetailResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get version detail: expected 200, got %d body=%s", code, string(body))
	}
	if versionDetailResp.Version.ID != versionID {
		return fmt.Errorf("get version detail: id mismatch: got=%s want=%s", versionDetailResp.Version.ID, versionID)
	}
	if len(versionDetailResp.Modules) < 1 {
		return fmt.Errorf("get version detail: expected at least 1 module, got %d", len(versionDetailResp.Modules))
	}

	// 10. POST /api/v1/jobs/{job_id}/versions/{version_id}/publish -> 200 (#70)
	var publishResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/jobs/"+jobID+"/versions/"+versionID+"/publish", nil, &publishResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("publish version: expected 200, got %d body=%s", code, string(body))
	}
	if publishResp.Status != "published" {
		return fmt.Errorf("publish version: expected status 'published', got '%s'", publishResp.Status)
	}

	// 11. GET /api/v1/jobs/nonexistent -> 404 (error case)
	code, body, err = client.GetJSON(ctx, "/api/v1/jobs/nonexistent", nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("get nonexistent job: expected 404, got %d body=%s", code, string(body))
	}

	return nil
}
