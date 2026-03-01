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
	return "events/happy_path/ingest_event"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_events_%d@example.com", time.Now().UnixNano())
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

	// 3. POST /api/v1/events (valid payload) → 202
	eventReq := map[string]any{
		"event_id":   "test-event-1",
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
	if eventResp.EventID != "test-event-1" {
		return fmt.Errorf("ingest event: event_id mismatch got=%s", eventResp.EventID)
	}
	if eventResp.Status != "accepted" {
		return fmt.Errorf("ingest event: status mismatch got=%s", eventResp.Status)
	}

	// 4. POST same event_id → 409 (duplicate)
	code, body, err = client.PostJSON(ctx, "/api/v1/events", eventReq, nil)
	if err != nil {
		return err
	}
	if code != 409 {
		return fmt.Errorf("duplicate event: expected 409 got=%d body=%s", code, string(body))
	}

	// 5. POST missing event_name → 400
	badReq := map[string]any{
		"event_id":   "test-event-2",
		"event_time": time.Now().UTC().Format(time.RFC3339),
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/events", badReq, nil)
	if err != nil {
		return err
	}
	if code != 400 {
		return fmt.Errorf("missing event_name: expected 400 got=%d body=%s", code, string(body))
	}

	return nil
}
