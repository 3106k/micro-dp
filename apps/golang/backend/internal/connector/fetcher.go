package connector

import "context"

// SchemaColumn describes a single column within a schema item.
type SchemaColumn struct {
	Name            string `json:"name"`
	Type            string `json:"type"` // e.g. "integer", "varchar", "timestamp"
	Nullable        bool   `json:"nullable,omitempty"`
	PrimaryKey      bool   `json:"primary_key,omitempty"`
	CursorCandidate bool   `json:"cursor_candidate,omitempty"`
}

// SchemaItem represents a single schema item (sheet, table, view, etc.).
type SchemaItem struct {
	Name                 string         `json:"name"`
	Type                 string         `json:"type"` // "sheet", "table", "view"
	Columns              []SchemaColumn `json:"columns,omitempty"`
	PrimaryKey           []string       `json:"primary_key,omitempty"`
	CursorField          string         `json:"cursor_field,omitempty"`
	SupportsIncremental  bool           `json:"supports_incremental,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
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
