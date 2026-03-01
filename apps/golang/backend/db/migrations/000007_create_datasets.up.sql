CREATE TABLE datasets (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL REFERENCES tenants(id),
    name            TEXT NOT NULL,
    source_type     TEXT NOT NULL CHECK(source_type IN ('tracker', 'parquet', 'import')),
    schema_json     TEXT,
    row_count       INTEGER,
    storage_path    TEXT NOT NULL DEFAULT '',
    last_updated_at DATETIME,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(tenant_id, name)
);
CREATE INDEX idx_datasets_tenant_id ON datasets(tenant_id);
CREATE INDEX idx_datasets_source_type ON datasets(tenant_id, source_type);
