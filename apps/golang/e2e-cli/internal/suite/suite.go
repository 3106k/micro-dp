package suite

import (
	"fmt"

	"github.com/user/micro-dp/e2e-cli/internal/config"
	"github.com/user/micro-dp/e2e-cli/internal/runner"
	authcase "github.com/user/micro-dp/e2e-cli/internal/suite/auth/happy_path"
	healthcase "github.com/user/micro-dp/e2e-cli/internal/suite/health/healthz"
	jobrunscase "github.com/user/micro-dp/e2e-cli/internal/suite/job_runs/not_implemented"
)

func Build(cfg *config.Config) ([]runner.Scenario, error) {
	scenarios := make([]runner.Scenario, 0)
	for _, suiteName := range cfg.Suites {
		switch suiteName {
		case "health":
			scenarios = append(scenarios, healthcase.NewScenario())
		case "auth":
			scenarios = append(scenarios, authcase.NewScenario(cfg.AuthEmail, cfg.AuthPassword, cfg.DisplayName))
		case "job_runs":
			scenarios = append(scenarios, jobrunscase.NewScenario())
		default:
			return nil, fmt.Errorf("unknown suite: %s", suiteName)
		}
	}
	return scenarios, nil
}
