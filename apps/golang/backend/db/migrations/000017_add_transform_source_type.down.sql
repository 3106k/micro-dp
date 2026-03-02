-- Revert: remove 'transform' from datasets.source_type CHECK constraint.

DELETE FROM datasets WHERE source_type = 'transform';

CREATE TABLE datasets_new (
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

INSERT INTO datasets_new SELECT * FROM datasets;
DROP TABLE datasets;
ALTER TABLE datasets_new RENAME TO datasets;

CREATE INDEX idx_datasets_tenant_id ON datasets(tenant_id);
CREATE INDEX idx_datasets_source_type ON datasets(tenant_id, source_type);
