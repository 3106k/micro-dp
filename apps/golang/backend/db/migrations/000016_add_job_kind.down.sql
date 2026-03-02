-- SQLite does not support DROP COLUMN directly; recreate table without kind.
CREATE TABLE jobs_backup AS SELECT id, tenant_id, name, slug, description, is_active, created_at, updated_at FROM jobs;
DROP TABLE jobs;
CREATE TABLE jobs (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(id),
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(tenant_id, slug)
);
INSERT INTO jobs SELECT * FROM jobs_backup;
DROP TABLE jobs_backup;
