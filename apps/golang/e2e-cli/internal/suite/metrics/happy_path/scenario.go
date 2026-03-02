package happy_path

import (
	"context"
	"fmt"
	"strings"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
)

type Scenario struct{}

func NewScenario() *Scenario {
	return &Scenario{}
}

func (s *Scenario) ID() string {
	return "metrics/happy_path/basic"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// /metrics returns text/plain (Prometheus format), not JSON.
	// When out is nil, doJSON still returns the body bytes without attempting JSON unmarshal.
	code, body, err := client.GetJSON(ctx, "/metrics", nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("unexpected status code: %d body=%s", code, string(body))
	}
	if len(body) == 0 {
		return fmt.Errorf("response body is empty")
	}
	text := string(body)
	if !strings.Contains(text, "go_") && !strings.Contains(text, "process_") {
		return fmt.Errorf("response does not contain expected Prometheus metrics (go_ or process_)")
	}
	return nil
}
