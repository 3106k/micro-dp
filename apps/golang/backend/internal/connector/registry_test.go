package connector

import (
	"fmt"
	"strings"
	"testing"
)

// TestDefinitionsLoad verifies that all embedded JSON definitions load
// and compile without errors. This is the first gate for any new connector.
func TestDefinitionsLoad(t *testing.T) {
	r, err := load()
	if err != nil {
		t.Fatalf("load definitions: %v", err)
	}
	if len(r.defs) == 0 {
		t.Fatal("expected at least one definition")
	}
}

// TestDefinitionsLint checks structural invariants for every definition.
// New connectors automatically inherit these checks.
func TestDefinitionsLint(t *testing.T) {
	r, err := load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	for id, def := range r.defs {
		t.Run(id, func(t *testing.T) {
			// Required fields
			if def.ID == "" {
				t.Error("id is empty")
			}
			if def.Name == "" {
				t.Error("name is empty")
			}
			if def.Kind == "" {
				t.Error("kind is empty")
			}
			if def.Kind != "source" && def.Kind != "destination" {
				t.Errorf("kind must be source or destination, got %q", def.Kind)
			}

			// ID format: {kind}-{rest}
			if !strings.HasPrefix(def.ID, def.Kind+"-") {
				t.Errorf("id %q should start with %q", def.ID, def.Kind+"-")
			}

			// Capabilities must be known values
			validCaps := map[string]bool{"testable": true, "fetchable": true, "importable": true}
			for _, cap := range def.Capabilities {
				if !validCaps[cap] {
					t.Errorf("unknown capability %q", cap)
				}
			}

			// Spec must be valid JSON (already parsed, but check non-nil)
			if len(def.Spec) == 0 {
				t.Error("spec is empty")
			}

			// Schema must have compiled successfully
			if _, ok := r.schemas[def.ID]; !ok {
				t.Error("schema not compiled")
			}
		})
	}
}

// TestDefinitionsNoDuplicateIDs checks that no two definitions share the same id.
// load() already enforces this, but this test makes the invariant explicit.
func TestDefinitionsNoDuplicateIDs(t *testing.T) {
	r, err := load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	seen := map[string]bool{}
	for id := range r.defs {
		if seen[id] {
			t.Errorf("duplicate id: %s", id)
		}
		seen[id] = true
	}
}

// TestValidateConfig_Template demonstrates the pattern for testing
// config validation against a connector's spec. Copy and adapt for new connectors.
func TestValidateConfig_Template(t *testing.T) {
	r, err := load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Find a connector with required properties for a meaningful test.
	// source-postgres has required fields: host, port, database, username, password.
	def := r.Get("source-postgres")
	if def == nil {
		t.Skip("source-postgres not found")
	}

	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  `{"host":"localhost","port":5432,"database":"test","username":"user","password":"pass"}`,
			wantErr: false,
		},
		{
			name:    "missing required field",
			config:  `{"host":"localhost","port":5432}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			config:  `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty object",
			config:  `{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", def.ID, tt.name), func(t *testing.T) {
			err := r.ValidateConfig(def.ID, tt.config)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
