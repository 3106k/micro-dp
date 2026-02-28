CREATE TABLE job_runs (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    checkpoint_json TEXT,
    progress_json TEXT,
    attempt INTEGER NOT NULL DEFAULT 0,
    next_run_at DATETIME,
    last_error TEXT,
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
