DROP INDEX IF EXISTS idx_admin_audit_logs_created_at;
DROP INDEX IF EXISTS idx_admin_audit_logs_actor_user_id;
DROP TABLE IF EXISTS admin_audit_logs;

PRAGMA foreign_keys=OFF;

CREATE TABLE tenants_old (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO tenants_old (id, name, created_at, updated_at)
SELECT id, name, created_at, updated_at
FROM tenants;

DROP TABLE tenants;
ALTER TABLE tenants_old RENAME TO tenants;

CREATE TABLE users_old (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO users_old (id, email, password_hash, display_name, created_at, updated_at)
SELECT id, email, password_hash, display_name, created_at, updated_at
FROM users;

DROP TABLE users;
ALTER TABLE users_old RENAME TO users;

PRAGMA foreign_keys=ON;
