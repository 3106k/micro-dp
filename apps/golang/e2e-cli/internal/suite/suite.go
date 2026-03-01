package suite

import (
	"fmt"

	"github.com/user/micro-dp/e2e-cli/internal/config"
	"github.com/user/micro-dp/e2e-cli/internal/runner"
	authfailure "github.com/user/micro-dp/e2e-cli/internal/suite/auth/failure"
	authcase "github.com/user/micro-dp/e2e-cli/internal/suite/auth/happy_path"
	datasetscase "github.com/user/micro-dp/e2e-cli/internal/suite/datasets/happy_path"
	eventscase "github.com/user/micro-dp/e2e-cli/internal/suite/events/happy_path"
	healthcase "github.com/user/micro-dp/e2e-cli/internal/suite/health/healthz"
	jobrunscase "github.com/user/micro-dp/e2e-cli/internal/suite/job_runs/happy_path"
	adminmultitenant "github.com/user/micro-dp/e2e-cli/internal/suite/tenant/admin_multi_tenant"
	tenantisolation "github.com/user/micro-dp/e2e-cli/internal/suite/tenant/isolation"
	membersauthorization "github.com/user/micro-dp/e2e-cli/internal/suite/members/authorization"
	membershappypath "github.com/user/micro-dp/e2e-cli/internal/suite/members/happy_path"
	uploadscase "github.com/user/micro-dp/e2e-cli/internal/suite/uploads/happy_path"
)

func Build(cfg *config.Config) ([]runner.Scenario, error) {
	scenarios := make([]runner.Scenario, 0)
	for _, suiteName := range cfg.Suites {
		switch suiteName {
		case "health":
			scenarios = append(scenarios, healthcase.NewScenario())
		case "auth":
			scenarios = append(scenarios,
				authcase.NewScenario(cfg.AuthEmail, cfg.AuthPassword, cfg.DisplayName),
				authfailure.NewScenario(cfg.AuthPassword),
			)
		case "job_runs":
			scenarios = append(scenarios, jobrunscase.NewScenario("", cfg.AuthPassword, cfg.DisplayName))
		case "datasets":
			scenarios = append(scenarios, datasetscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "events":
			scenarios = append(scenarios, eventscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "uploads":
			scenarios = append(scenarios, uploadscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "tenant":
			scenarios = append(scenarios,
				tenantisolation.NewScenario(cfg.AuthPassword, cfg.DisplayName),
				adminmultitenant.NewScenario(cfg.AdminEmail, cfg.AdminPassword),
			)
		case "members":
			scenarios = append(scenarios,
				membershappypath.NewScenario(cfg.AuthPassword, cfg.DisplayName),
				membersauthorization.NewScenario(cfg.AuthPassword, cfg.DisplayName),
			)
		default:
			return nil, fmt.Errorf("unknown suite: %s", suiteName)
		}
	}
	return scenarios, nil
}
