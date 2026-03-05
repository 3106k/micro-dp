package connector

import "context"

// ImportParams holds the generic parameters for an import executor.
type ImportParams struct {
	TenantID    string
	JobRunID    string
	JobID       string
	VersionID   string
	Config      map[string]any // module config_json parsed as generic map
	AccessToken string         // empty for connectors that don't require credentials
}

// ImportResult holds the result of an import execution.
type ImportResult struct {
	RowCount  int64
	OutputKey string
}

// ImportExecutor performs data import for a specific connector type.
type ImportExecutor interface {
	ExecuteImport(ctx context.Context, params *ImportParams) (*ImportResult, error)
}
