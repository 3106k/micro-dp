package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BaseURL      string
	Token        string
	TenantID     string
	Suites       []string
	JSONOut      string
	AuthEmail    string
	AuthPassword string
	DisplayName  string
}

func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("e2e-cli", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	baseURL := fs.String("base-url", envOr("E2E_BASE_URL", "http://localhost:8080"), "API base URL")
	token := fs.String("token", envOr("E2E_TOKEN", ""), "Bearer token for authenticated requests")
	tenantID := fs.String("tenant-id", envOr("E2E_TENANT_ID", ""), "Tenant header value")
	suitesRaw := fs.String("suites", envOr("E2E_SUITES", "health,auth,job_runs"), "Comma separated suite names")
	jsonOut := fs.String("json-out", envOr("E2E_JSON_OUT", "e2e-report.json"), "JSON report output path")
	authEmail := fs.String("auth-email", envOr("E2E_AUTH_EMAIL", ""), "Email for auth suite (optional)")
	authPassword := fs.String("auth-password", envOr("E2E_AUTH_PASSWORD", "Passw0rd!123"), "Password for auth suite")
	displayName := fs.String("display-name", envOr("E2E_DISPLAY_NAME", "E2E User"), "Display name for auth suite")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	suites := splitCSV(*suitesRaw)
	if len(suites) == 0 {
		return nil, fmt.Errorf("at least one suite is required")
	}

	return &Config{
		BaseURL:      strings.TrimRight(*baseURL, "/"),
		Token:        *token,
		TenantID:     *tenantID,
		Suites:       suites,
		JSONOut:      *jsonOut,
		AuthEmail:    *authEmail,
		AuthPassword: *authPassword,
		DisplayName:  *displayName,
	}, nil
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
