package connector

import "context"

// SchemaItem represents a single schema item (sheet, table, view, etc.).
type SchemaItem struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"` // "sheet", "table", "view"
	Metadata map[string]any `json:"metadata,omitempty"`
}

// SchemaResult holds the result of a schema fetch.
type SchemaResult struct {
	Title string
	Items []SchemaItem
}

// SchemaFetcher retrieves schema information from an external data source.
type SchemaFetcher interface {
	FetchSchema(ctx context.Context, configJSON string, accessToken string) (*SchemaResult, error)
}
