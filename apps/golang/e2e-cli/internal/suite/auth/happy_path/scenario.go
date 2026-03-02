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
		email = fmt.Sprintf("e2e_%d@example.com", time.Now().UnixNano())
	}
	return &Scenario{
		email:       email,
		password:    password,
		displayName: displayName,
	}
}

func (s *Scenario) ID() string {
	return "auth/happy_path/register_login_me"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	registerReq := openapi.RegisterRequest{
		Email:       openapi.Email(s.email),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}
	var registerResp openapi.RegisterResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register unexpected status code: %d body=%s", code, string(body))
	}
	if registerResp.UserId == "" || registerResp.TenantId == "" {
		return fmt.Errorf("register response missing user_id/tenant_id")
	}

	loginReq := openapi.LoginRequest{
		Email:    openapi.Email(s.email),
		Password: s.password,
	}
	var loginResp openapi.LoginResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login unexpected status code: %d body=%s", code, string(body))
	}
	if loginResp.Token == "" {
		return fmt.Errorf("login response missing token")
	}
	client.SetToken(loginResp.Token)

	var meResp openapi.MeResponse
	code, body, err = client.GetJSON(ctx, "/api/v1/auth/me", &meResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("me unexpected status code: %d body=%s", code, string(body))
	}
	if meResp.UserId != registerResp.UserId {
		return fmt.Errorf("me user_id mismatch: got=%s want=%s", meResp.UserId, registerResp.UserId)
	}
	if string(meResp.Email) != s.email {
		return fmt.Errorf("me email mismatch: got=%s want=%s", string(meResp.Email), s.email)
	}
	return nil
}
