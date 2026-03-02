package happy_path

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/openapi"
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
	registerReq := openapi.RegisterRequest{
		Email:       openapi.Email(email),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}
	var registerResp openapi.RegisterResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: status=%d body=%s", code, string(body))
	}

	// 2. Login
	loginReq := openapi.LoginRequest{
		Email:    openapi.Email(email),
		Password: s.password,
	}
	var loginResp openapi.LoginResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login: status=%d body=%s", code, string(body))
	}
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantId)

	// 3. POST /api/v1/events (valid payload) → 202
	eventReq := openapi.IngestEventRequest{
		EventId:    "test-event-1",
		EventName:  "page_view",
		Properties: openapi.Ptr(map[string]interface{}{"page": "/home"}),
		EventTime:  time.Now().UTC(),
	}
	var eventResp openapi.IngestEventResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/events", eventReq, &eventResp)
	if err != nil {
		return err
	}
	if code != 202 {
		return fmt.Errorf("ingest event: status=%d body=%s", code, string(body))
	}
	if eventResp.EventId != "test-event-1" {
		return fmt.Errorf("ingest event: event_id mismatch got=%s", eventResp.EventId)
	}
	if eventResp.Status != openapi.IngestEventResponseStatusAccepted {
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

	// 5. POST missing event_name → 400 (intentionally incomplete — cannot use generated type)
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
