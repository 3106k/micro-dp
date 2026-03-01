package connector

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed definitions/sources/*.json definitions/destinations/*.json
var definitionsFS embed.FS

// Definition represents a connector definition loaded from embedded JSON.
type Definition struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Kind        string          `json:"kind"`
	Icon        string          `json:"icon"`
	Description string          `json:"description"`
	Spec        json.RawMessage `json:"spec"`
}

// Registry holds all connector definitions and their compiled JSON Schemas.
type Registry struct {
	defs    map[string]*Definition
	schemas map[string]*jsonschema.Schema
}

var (
	globalRegistry *Registry
	globalOnce     sync.Once
)

// Global returns the singleton Registry, loading definitions on first call.
// Panics (via log.Fatalf) if any definition is invalid.
func Global() *Registry {
	globalOnce.Do(func() {
		r, err := load()
		if err != nil {
			log.Fatalf("connector registry: %v", err)
		}
		globalRegistry = r
	})
	return globalRegistry
}

func load() (*Registry, error) {
	r := &Registry{
		defs:    make(map[string]*Definition),
		schemas: make(map[string]*jsonschema.Schema),
	}

	dirs := []string{"definitions/sources", "definitions/destinations"}
	for _, dir := range dirs {
		entries, err := definitionsFS.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("read dir %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			path := dir + "/" + e.Name()
			data, err := definitionsFS.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", path, err)
			}

			var def Definition
			if err := json.Unmarshal(data, &def); err != nil {
				return nil, fmt.Errorf("parse %s: %w", path, err)
			}
			if def.ID == "" || def.Kind == "" {
				return nil, fmt.Errorf("definition %s: id and kind are required", path)
			}
			if _, exists := r.defs[def.ID]; exists {
				return nil, fmt.Errorf("duplicate definition id: %s", def.ID)
			}

			// Compile JSON Schema for the spec.
			schema, err := compileSchema(def.ID, def.Spec)
			if err != nil {
				return nil, fmt.Errorf("compile schema for %s: %w", def.ID, err)
			}

			r.defs[def.ID] = &def
			r.schemas[def.ID] = schema
		}
	}

	log.Printf("connector registry: loaded %d definitions", len(r.defs))
	return r, nil
}

func compileSchema(id string, raw json.RawMessage) (*jsonschema.Schema, error) {
	var specObj any
	if err := json.Unmarshal(raw, &specObj); err != nil {
		return nil, fmt.Errorf("unmarshal spec: %w", err)
	}

	c := jsonschema.NewCompiler()
	url := "connector://" + id + "/spec.json"
	if err := c.AddResource(url, specObj); err != nil {
		return nil, fmt.Errorf("add resource: %w", err)
	}
	return c.Compile(url)
}

// Get returns a definition by ID or nil if not found.
func (r *Registry) Get(id string) *Definition {
	return r.defs[id]
}

// Exists returns true if a definition with the given ID exists.
func (r *Registry) Exists(id string) bool {
	_, ok := r.defs[id]
	return ok
}

// List returns definitions filtered by kind. Pass "" to return all.
func (r *Registry) List(kind string) []*Definition {
	out := make([]*Definition, 0, len(r.defs))
	for _, d := range r.defs {
		if kind == "" || d.Kind == kind {
			out = append(out, d)
		}
	}
	return out
}

// ValidateConfig validates a JSON config string against the connector's spec.
// Returns nil if valid. Returns an error describing validation failures.
func (r *Registry) ValidateConfig(connectorID, configJSON string) error {
	schema, ok := r.schemas[connectorID]
	if !ok {
		return fmt.Errorf("unknown connector: %s", connectorID)
	}

	var configObj any
	if err := json.Unmarshal([]byte(configJSON), &configObj); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return schema.Validate(configObj)
}
