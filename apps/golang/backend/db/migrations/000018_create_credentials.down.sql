-- SQLite does not support DROP COLUMN, so recreate connections without credential_id.
CREATE TABLE connections_new (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL REFERENCES tenants(id),
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,
    config_json TEXT NOT NULL DEFAULT '{}',
    secret_ref  TEXT,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(tenant_id, name)
);

INSERT INTO connections_new (id, tenant_id, name, type, config_json, secret_ref, created_at, updated_at)
SELECT id, tenant_id, name, type, config_json, secret_ref, created_at, updated_at FROM connections;

DROP TABLE connections;
ALTER TABLE connections_new RENAME TO connections;

DROP TABLE credentials;
