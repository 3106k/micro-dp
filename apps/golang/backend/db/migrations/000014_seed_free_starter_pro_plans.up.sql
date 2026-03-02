-- Seed: web plans (free/starter/pro)
INSERT OR IGNORE INTO plans (
    id,
    name,
    display_name,
    max_events_per_day,
    max_storage_bytes,
    max_rows_per_day,
    max_uploads_per_day,
    is_default
) VALUES
    ('plan-free-default', 'free', 'Free', 10000, 1073741824, 100000, 10, FALSE),
    ('plan-starter-default', 'starter', 'Starter', 100000, 10737418240, 1000000, 100, FALSE),
    ('plan-pro-default', 'pro', 'Pro', 1000000, 107374182400, 10000000, 1000, FALSE);
