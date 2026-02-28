package healthz

import (
	"context"
	"fmt"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
)

type Scenario struct{}

func NewScenario() *Scenario {
	return &Scenario{}
}

func (s *Scenario) ID() string {
	return "health/healthz/basic"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	var resp struct {
		Status string `json:"status"`
	}
	code, body, err := client.GetJSON(ctx, "/healthz", &resp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("unexpected status code: %d body=%s", code, string(body))
	}
	if resp.Status != "ok" && resp.Status != "degraded" {
		return fmt.Errorf("unexpected status field: %q", resp.Status)
	}
	return nil
}
