package suite

import (
	"fmt"

	"github.com/user/micro-dp/e2e-cli/internal/config"
	"github.com/user/micro-dp/e2e-cli/internal/runner"
	adminplans "github.com/user/micro-dp/e2e-cli/internal/suite/admin/plans"
	admintenants "github.com/user/micro-dp/e2e-cli/internal/suite/admin/tenants"
	authfailure "github.com/user/micro-dp/e2e-cli/internal/suite/auth/failure"
	googlecallback "github.com/user/micro-dp/e2e-cli/internal/suite/auth/google_callback"
	googlestart "github.com/user/micro-dp/e2e-cli/internal/suite/auth/google_start"
	authcase "github.com/user/micro-dp/e2e-cli/internal/suite/auth/happy_path"
	billingcase "github.com/user/micro-dp/e2e-cli/internal/suite/billing/happy_path"
	connectionscase "github.com/user/micro-dp/e2e-cli/internal/suite/connections/happy_path"
	connectorscase "github.com/user/micro-dp/e2e-cli/internal/suite/connectors/happy_path"
	datasetscase "github.com/user/micro-dp/e2e-cli/internal/suite/datasets/happy_path"
	datasetsrowspreview "github.com/user/micro-dp/e2e-cli/internal/suite/datasets/rows_preview"
	eventscase "github.com/user/micro-dp/e2e-cli/internal/suite/events/happy_path"
	eventssummary "github.com/user/micro-dp/e2e-cli/internal/suite/events/summary"
	healthcase "github.com/user/micro-dp/e2e-cli/internal/suite/health/healthz"
	jobrunscase "github.com/user/micro-dp/e2e-cli/internal/suite/job_runs/happy_path"
	jobscase "github.com/user/micro-dp/e2e-cli/internal/suite/jobs/happy_path"
	membersauthorization "github.com/user/micro-dp/e2e-cli/internal/suite/members/authorization"
	membershappypath "github.com/user/micro-dp/e2e-cli/internal/suite/members/happy_path"
	metricscase "github.com/user/micro-dp/e2e-cli/internal/suite/metrics/happy_path"
	moduletypescase "github.com/user/micro-dp/e2e-cli/internal/suite/module_types/happy_path"
	plancase "github.com/user/micro-dp/e2e-cli/internal/suite/plan/happy_path"
	adminmultitenant "github.com/user/micro-dp/e2e-cli/internal/suite/tenant/admin_multi_tenant"
	tenantisolation "github.com/user/micro-dp/e2e-cli/internal/suite/tenant/isolation"
	dashboardscase "github.com/user/micro-dp/e2e-cli/internal/suite/dashboards/happy_path"
	importcase "github.com/user/micro-dp/e2e-cli/internal/suite/import/happy_path"
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
				googlestart.NewScenario(),
				googlecallback.NewScenario(),
			)
		case "metrics":
			scenarios = append(scenarios, metricscase.NewScenario())
		case "connectors":
			scenarios = append(scenarios, connectorscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "connections":
			scenarios = append(scenarios, connectionscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "datasets":
			scenarios = append(scenarios,
				datasetscase.NewScenario(cfg.AuthPassword, cfg.DisplayName),
				datasetsrowspreview.NewScenario(cfg.AuthPassword, cfg.DisplayName),
			)
		case "events":
			scenarios = append(scenarios,
				eventscase.NewScenario(cfg.AuthPassword, cfg.DisplayName),
				eventssummary.NewScenario(cfg.AuthPassword, cfg.DisplayName),
			)
		case "jobs":
			scenarios = append(scenarios, jobscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "job_runs":
			scenarios = append(scenarios, jobrunscase.NewScenario("", cfg.AuthPassword, cfg.DisplayName))
		case "module_types":
			scenarios = append(scenarios, moduletypescase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "import":
			scenarios = append(scenarios, importcase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "uploads":
			scenarios = append(scenarios, uploadscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "plan":
			scenarios = append(scenarios, plancase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "billing":
			scenarios = append(scenarios, billingcase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
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
		case "dashboards":
			scenarios = append(scenarios, dashboardscase.NewScenario(cfg.AuthPassword, cfg.DisplayName))
		case "admin":
			scenarios = append(scenarios,
				admintenants.NewScenario(cfg.AdminEmail, cfg.AdminPassword, cfg.AuthPassword, cfg.DisplayName),
				adminplans.NewScenario(cfg.AdminEmail, cfg.AdminPassword),
			)
		default:
			return nil, fmt.Errorf("unknown suite: %s", suiteName)
		}
	}
	return scenarios, nil
}
