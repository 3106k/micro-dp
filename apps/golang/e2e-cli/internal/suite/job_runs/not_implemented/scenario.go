package not_implemented

import (
	"context"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/runner"
)

type Scenario struct{}

func NewScenario() *Scenario {
	return &Scenario{}
}

func (s *Scenario) ID() string {
	return "job_runs/happy_path/create_and_get"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	return runner.Skip("job_runs API is not implemented yet")
}
