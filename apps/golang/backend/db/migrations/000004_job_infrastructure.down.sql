-- ============================================================
-- 000004_job_infrastructure.down.sql
-- Reverse the job infrastructure migration.
-- ============================================================

DROP TABLE IF EXISTS job_run_artifacts;
DROP TABLE IF EXISTS job_run_modules;

-- Restore job_runs with project_id
CREATE TABLE job_runs_old (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT REFERENCES tenants(id),
    project_id      TEXT NOT NULL DEFAULT '',
    job_id          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'queued',
    checkpoint_json TEXT,
    progress_json   TEXT,
    attempt         INTEGER NOT NULL DEFAULT 0,
    next_run_at     DATETIME,
    last_error      TEXT,
    started_at      DATETIME,
    finished_at     DATETIME,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at      DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO job_runs_old (id, tenant_id, project_id, job_id, status, checkpoint_json, progress_json, attempt, next_run_at, last_error, started_at, finished_at, created_at, updated_at)
SELECT id, tenant_id, '', job_id, status, checkpoint_json, progress_json, attempt, next_run_at, last_error, started_at, finished_at, created_at, updated_at
FROM job_runs;

DROP TABLE job_runs;
ALTER TABLE job_runs_old RENAME TO job_runs;
CREATE INDEX idx_job_runs_tenant_id ON job_runs(tenant_id);

DROP TABLE IF EXISTS job_module_edges;
DROP TABLE IF EXISTS job_modules;
DROP TABLE IF EXISTS connections;
DROP TABLE IF EXISTS module_type_schemas;
DROP TABLE IF EXISTS module_types;
DROP TABLE IF EXISTS job_versions;
DROP TABLE IF EXISTS jobs;
