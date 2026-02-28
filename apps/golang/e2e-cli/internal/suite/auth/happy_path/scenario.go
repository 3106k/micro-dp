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
	registerReq := map[string]string{
		"email":        s.email,
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
		return fmt.Errorf("register unexpected status code: %d body=%s", code, string(body))
	}
	if registerResp.UserID == "" || registerResp.TenantID == "" {
		return fmt.Errorf("register response missing user_id/tenant_id")
	}

	loginReq := map[string]string{
		"email":    s.email,
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
		return fmt.Errorf("login unexpected status code: %d body=%s", code, string(body))
	}
	if loginResp.Token == "" {
		return fmt.Errorf("login response missing token")
	}
	client.SetToken(loginResp.Token)

	var meResp struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}
	code, body, err = client.GetJSON(ctx, "/api/v1/auth/me", &meResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("me unexpected status code: %d body=%s", code, string(body))
	}
	if meResp.UserID != registerResp.UserID {
		return fmt.Errorf("me user_id mismatch: got=%s want=%s", meResp.UserID, registerResp.UserID)
	}
	if meResp.Email != s.email {
		return fmt.Errorf("me email mismatch: got=%s want=%s", meResp.Email, s.email)
	}
	return nil
}
