package connector

import "context"

// TestResult holds the result of a real connection test.
type TestResult struct {
	OK      bool
	Code    string // e.g. "ok", "unauthorized", "forbidden", "not_found", "invalid_config"
	Message string
}

// ConnectionTester performs real connectivity tests for a specific connector type.
type ConnectionTester interface {
	Test(ctx context.Context, configJSON string, accessToken string) *TestResult
}
