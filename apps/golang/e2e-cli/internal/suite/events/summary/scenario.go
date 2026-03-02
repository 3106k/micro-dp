package summary

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
	return "events/summary/get_summary"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register
	email := fmt.Sprintf("e2e_events_summary_%d@example.com", time.Now().UnixNano())
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

	// 3. POST /api/v1/events with a valid event -> 202
	eventReq := openapi.IngestEventRequest{
		EventId:    fmt.Sprintf("summary-test-%d", time.Now().UnixNano()),
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

	// 4. GET /api/v1/events/summary -> 200
	var summaryResp openapi.EventsSummaryResponse
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
