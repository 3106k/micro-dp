-- Plans: quota template
CREATE TABLE IF NOT EXISTS plans (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    max_events_per_day INTEGER NOT NULL DEFAULT -1,
    max_storage_bytes BIGINT NOT NULL DEFAULT -1,
    max_rows_per_day INTEGER NOT NULL DEFAULT -1,
    max_uploads_per_day INTEGER NOT NULL DEFAULT -1,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tenant ↔ Plan mapping
CREATE TABLE IF NOT EXISTS tenant_plans (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    plan_id TEXT NOT NULL REFERENCES plans(id),
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id)
);

-- Daily usage counters (upsert pattern)
CREATE TABLE IF NOT EXISTS usage_daily (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    date TEXT NOT NULL,
    events_count INTEGER NOT NULL DEFAULT 0,
    storage_bytes BIGINT NOT NULL DEFAULT 0,
    rows_count INTEGER NOT NULL DEFAULT 0,
    uploads_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, date)
);

-- Usage event log (append-only audit)
CREATE TABLE IF NOT EXISTS usage_events (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    event_type TEXT NOT NULL,
    delta INTEGER NOT NULL,
    recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_usage_daily_tenant_date ON usage_daily(tenant_id, date);
CREATE INDEX IF NOT EXISTS idx_usage_events_tenant ON usage_events(tenant_id, recorded_at);

-- Seed: default OSS plan (unlimited)
INSERT OR IGNORE INTO plans (id, name, display_name, max_events_per_day, max_storage_bytes, max_rows_per_day, max_uploads_per_day, is_default)
VALUES ('plan-oss-default', 'oss', 'OSS (Unlimited)', -1, -1, -1, -1, TRUE);
