CREATE TABLE uploads (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL REFERENCES tenants(id),
    status     TEXT NOT NULL DEFAULT 'presigned' CHECK(status IN ('presigned', 'uploaded')),
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_uploads_tenant_id ON uploads(tenant_id);

CREATE TABLE upload_files (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL REFERENCES tenants(id),
    upload_id    TEXT NOT NULL REFERENCES uploads(id),
    file_name    TEXT NOT NULL,
    object_key   TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes   INTEGER NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_upload_files_upload_id ON upload_files(upload_id);
