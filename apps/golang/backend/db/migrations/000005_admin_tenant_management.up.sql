ALTER TABLE users ADD COLUMN is_superadmin INTEGER NOT NULL DEFAULT 0;

ALTER TABLE tenants ADD COLUMN is_active INTEGER NOT NULL DEFAULT 1;

CREATE TABLE admin_audit_logs (
    id TEXT PRIMARY KEY,
    actor_user_id TEXT NOT NULL REFERENCES users(id),
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    metadata_json TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_admin_audit_logs_actor_user_id ON admin_audit_logs(actor_user_id);
CREATE INDEX idx_admin_audit_logs_created_at ON admin_audit_logs(created_at);
