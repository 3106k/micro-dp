package failure

import (
	"context"
	"fmt"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
)

type Scenario struct {
	password string
}

func NewScenario(password string) *Scenario {
	return &Scenario{password: password}
}

func (s *Scenario) ID() string {
	return "auth/failure/unauthenticated_and_wrong_password"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// Save current token and clear it for unauthenticated request
	client.SetToken("")

	var errResp struct {
		Error string `json:"error"`
	}

	// GET /api/v1/auth/me without token → 401
	code, body, err := client.GetJSON(ctx, "/api/v1/auth/me", &errResp)
	if err != nil {
		return err
	}
	if code != 401 {
		return fmt.Errorf("me without token: expected 401, got %d body=%s", code, string(body))
	}

	// POST /api/v1/auth/login with wrong password → 401
	loginReq := map[string]string{
		"email":    "nonexistent@example.com",
		"password": s.password,
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &errResp)
	if err != nil {
		return err
	}
	if code != 401 {
		return fmt.Errorf("login with wrong creds: expected 401, got %d body=%s", code, string(body))
	}

	return nil
}
