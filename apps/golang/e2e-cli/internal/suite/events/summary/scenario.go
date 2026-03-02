package summary

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
	return "events/summary/get_summary"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_events_summary_%d@example.com", time.Now().UnixNano())
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
		return fmt.Errorf("register: status=%d body=%s", code, string(body))
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
		return fmt.Errorf("login: status=%d body=%s", code, string(body))
	}
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantID)

	// 3. POST /api/v1/events with a valid event -> 202
	eventReq := map[string]any{
		"event_id":   fmt.Sprintf("summary-test-%d", time.Now().UnixNano()),
		"event_name": "page_view",
		"properties": map[string]any{"page": "/home"},
		"event_time": time.Now().UTC().Format(time.RFC3339),
	}
	var eventResp struct {
		EventID string `json:"event_id"`
		Status  string `json:"status"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/events", eventReq, &eventResp)
	if err != nil {
		return err
	}
	if code != 202 {
		return fmt.Errorf("ingest event: status=%d body=%s", code, string(body))
	}

	// 4. GET /api/v1/events/summary -> 200
	var summaryResp struct {
		Counts []map[string]any `json:"counts"`
		Total  int64            `json:"total"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/events/summary", &summaryResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("get summary: status=%d body=%s", code, string(body))
	}

	// 5. Verify response has counts array
	if summaryResp.Counts == nil {
		return fmt.Errorf("get summary: counts is nil, body=%s", string(body))
	}

	// 6. GET /api/v1/events/summary without auth -> 401
	savedToken := loginResp.Token
	client.SetToken("")
	code, body, err = client.GetJSON(ctx, "/api/v1/events/summary", nil)
	if err != nil {
		return err
	}
	if code != 401 {
		return fmt.Errorf("get summary without auth: expected 401 got=%d body=%s", code, string(body))
	}
	client.SetToken(savedToken)

	return nil
}
