CREATE TABLE dashboards (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL REFERENCES tenants(id),
    name        TEXT NOT NULL,
    description TEXT,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_dashboards_tenant_id ON dashboards(tenant_id);

CREATE TABLE charts (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL REFERENCES tenants(id),
    name        TEXT NOT NULL,
    chart_type  TEXT NOT NULL CHECK(chart_type IN ('line', 'bar', 'pie')),
    dataset_id  TEXT NOT NULL REFERENCES datasets(id),
    measure     TEXT NOT NULL,
    dimension   TEXT NOT NULL,
    config_json TEXT,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_charts_tenant_id ON charts(tenant_id);
CREATE INDEX idx_charts_dataset_id ON charts(dataset_id);

CREATE TABLE dashboard_widgets (
    id           TEXT PRIMARY KEY,
    dashboard_id TEXT NOT NULL REFERENCES dashboards(id) ON DELETE CASCADE,
    chart_id     TEXT NOT NULL REFERENCES charts(id) ON DELETE CASCADE,
    position     INTEGER NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(dashboard_id, chart_id)
);
CREATE INDEX idx_dashboard_widgets_dashboard_id ON dashboard_widgets(dashboard_id);
CREATE INDEX idx_dashboard_widgets_chart_id ON dashboard_widgets(chart_id);

CREATE TABLE template_runs (
    id             TEXT PRIMARY KEY,
    tenant_id      TEXT NOT NULL REFERENCES tenants(id),
    template_type  TEXT NOT NULL,
    status         TEXT NOT NULL CHECK(status IN ('success', 'failed', 'skipped')),
    skip_reason    TEXT,
    dashboard_id   TEXT REFERENCES dashboards(id),
    created_at     DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_template_runs_tenant_id ON template_runs(tenant_id);
