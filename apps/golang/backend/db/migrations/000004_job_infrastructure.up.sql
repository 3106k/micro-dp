-- ============================================================
-- 000004_job_infrastructure.up.sql
-- Job definitions, versions, modules, DAG edges, run modules,
-- artifacts, module types, schemas, and connections.
-- Also redefines job_runs to drop project_id and add
-- job_version_id + run_snapshot_json.
-- ============================================================

-- ----------------------------------------------------------
-- jobs
-- ----------------------------------------------------------
CREATE TABLE jobs (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL REFERENCES tenants(id),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_active  INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(tenant_id, slug)
);
CREATE INDEX idx_jobs_tenant_id ON jobs(tenant_id);

-- ----------------------------------------------------------
-- job_versions
-- ----------------------------------------------------------
CREATE TABLE job_versions (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL REFERENCES tenants(id),
    job_id       TEXT NOT NULL REFERENCES jobs(id),
    version      INTEGER NOT NULL,
    status       TEXT NOT NULL DEFAULT 'draft',  -- draft | published
    published_at DATETIME,
    created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(job_id, version)
);
CREATE INDEX idx_job_versions_tenant_id ON job_versions(tenant_id);
CREATE INDEX idx_job_versions_job_id ON job_versions(job_id);

-- ----------------------------------------------------------
-- module_types
-- ----------------------------------------------------------
CREATE TABLE module_types (
    id        TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    name      TEXT NOT NULL,
    category  TEXT NOT NULL,  -- source | transform | destination
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(tenant_id, name)
);
CREATE INDEX idx_module_types_tenant_id ON module_types(tenant_id);

-- ----------------------------------------------------------
-- module_type_schemas
-- ----------------------------------------------------------
CREATE TABLE module_type_schemas (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL REFERENCES tenants(id),
    module_type_id  TEXT NOT NULL REFERENCES module_types(id),
    version         INTEGER NOT NULL,
    json_schema     TEXT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(module_type_id, version)
);
CREATE INDEX idx_module_type_schemas_tenant_id ON module_type_schemas(tenant_id);
CREATE INDEX idx_module_type_schemas_module_type_id ON module_type_schemas(module_type_id);

-- ----------------------------------------------------------
-- connections
-- ----------------------------------------------------------
CREATE TABLE connections (
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
CREATE INDEX idx_connections_tenant_id ON connections(tenant_id);

-- ----------------------------------------------------------
-- job_modules
-- ----------------------------------------------------------
CREATE TABLE job_modules (
    id                     TEXT PRIMARY KEY,
    tenant_id              TEXT NOT NULL REFERENCES tenants(id),
    job_version_id         TEXT NOT NULL REFERENCES job_versions(id),
    module_type_id         TEXT NOT NULL REFERENCES module_types(id),
    module_type_schema_id  TEXT REFERENCES module_type_schemas(id),
    connection_id          TEXT REFERENCES connections(id),
    name                   TEXT NOT NULL,
    config_json            TEXT NOT NULL DEFAULT '{}',
    position_x             REAL NOT NULL DEFAULT 0,
    position_y             REAL NOT NULL DEFAULT 0,
    created_at             DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at             DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_job_modules_tenant_id ON job_modules(tenant_id);
CREATE INDEX idx_job_modules_job_version_id ON job_modules(job_version_id);

-- ----------------------------------------------------------
-- job_module_edges
-- ----------------------------------------------------------
CREATE TABLE job_module_edges (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL REFERENCES tenants(id),
    job_version_id   TEXT NOT NULL REFERENCES job_versions(id),
    source_module_id TEXT NOT NULL REFERENCES job_modules(id),
    target_module_id TEXT NOT NULL REFERENCES job_modules(id),
    created_at       DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(job_version_id, source_module_id, target_module_id),
    CHECK(source_module_id <> target_module_id)
);
CREATE INDEX idx_job_module_edges_tenant_id ON job_module_edges(tenant_id);
CREATE INDEX idx_job_module_edges_job_version_id ON job_module_edges(job_version_id);

-- ----------------------------------------------------------
-- Redefine job_runs: drop project_id, add job_version_id + run_snapshot_json
-- SQLite does not support DROP COLUMN before 3.35.0 so we use
-- CREATE → INSERT SELECT → DROP → RENAME pattern.
-- ----------------------------------------------------------
CREATE TABLE job_runs_new (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT REFERENCES tenants(id),
    job_id           TEXT NOT NULL,
    job_version_id   TEXT REFERENCES job_versions(id),
    status           TEXT NOT NULL DEFAULT 'queued',
    run_snapshot_json TEXT,
    checkpoint_json  TEXT,
    progress_json    TEXT,
    attempt          INTEGER NOT NULL DEFAULT 0,
    next_run_at      DATETIME,
    last_error       TEXT,
    started_at       DATETIME,
    finished_at      DATETIME,
    created_at       DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at       DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO job_runs_new (id, tenant_id, job_id, status, checkpoint_json, progress_json, attempt, next_run_at, last_error, started_at, finished_at, created_at, updated_at)
SELECT id, tenant_id, job_id, status, checkpoint_json, progress_json, attempt, next_run_at, last_error, started_at, finished_at, created_at, updated_at
FROM job_runs;

DROP TABLE job_runs;
ALTER TABLE job_runs_new RENAME TO job_runs;

CREATE INDEX idx_job_runs_tenant_id ON job_runs(tenant_id);
CREATE INDEX idx_job_runs_job_version_id ON job_runs(job_version_id);

-- ----------------------------------------------------------
-- job_run_modules
-- ----------------------------------------------------------
CREATE TABLE job_run_modules (
    id            TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL REFERENCES tenants(id),
    job_run_id    TEXT NOT NULL REFERENCES job_runs(id),
    job_module_id TEXT NOT NULL REFERENCES job_modules(id),
    status        TEXT NOT NULL DEFAULT 'queued',
    input_json    TEXT,
    output_json   TEXT,
    error_message TEXT,
    started_at    DATETIME,
    finished_at   DATETIME,
    created_at    DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at    DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_job_run_modules_tenant_id ON job_run_modules(tenant_id);
CREATE INDEX idx_job_run_modules_job_run_id ON job_run_modules(job_run_id);

-- ----------------------------------------------------------
-- job_run_artifacts
-- ----------------------------------------------------------
CREATE TABLE job_run_artifacts (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL REFERENCES tenants(id),
    job_run_id        TEXT NOT NULL REFERENCES job_runs(id),
    job_run_module_id TEXT REFERENCES job_run_modules(id),
    name              TEXT NOT NULL,
    artifact_type     TEXT NOT NULL,
    storage_path      TEXT NOT NULL,
    size_bytes        INTEGER NOT NULL DEFAULT 0,
    content_type      TEXT NOT NULL DEFAULT 'application/octet-stream',
    created_at        DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_job_run_artifacts_tenant_id ON job_run_artifacts(tenant_id);
CREATE INDEX idx_job_run_artifacts_job_run_id ON job_run_artifacts(job_run_id);
