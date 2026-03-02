package google_start

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/openapi"
)

type Scenario struct{}

func NewScenario() *Scenario {
	return &Scenario{}
}

func (s *Scenario) ID() string {
	return "auth/google_start/without_oauth_config"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// GET /api/v1/auth/google/start when OAuth is NOT configured (typical E2E env)
	// Expected: 500 with error "google oauth is not configured"
	var errResp openapi.ErrorResponse
	code, body, err := client.GetJSON(ctx, "/api/v1/auth/google/start", &errResp)
	if err != nil {
		return err
	}
	if code != 500 {
		return fmt.Errorf("expected 500, got %d body=%s", code, string(body))
	}
	if !strings.Contains(errResp.Error, "google oauth is not configured") {
		return fmt.Errorf("expected error containing 'google oauth is not configured', got %q", errResp.Error)
	}
	return nil
}
