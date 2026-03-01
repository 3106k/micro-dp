CREATE TABLE IF NOT EXISTS tenant_billing_subscriptions (
    id TEXT PRIMARY KEY REFERENCES tenants(id),
    stripe_customer_id TEXT NOT NULL DEFAULT '' UNIQUE,
    stripe_subscription_id TEXT NOT NULL DEFAULT '' UNIQUE,
    stripe_price_id TEXT NOT NULL DEFAULT '',
    subscription_status TEXT NOT NULL DEFAULT 'inactive',
    current_period_end DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tenant_billing_subscriptions_status
    ON tenant_billing_subscriptions(subscription_status);

CREATE TABLE IF NOT EXISTS stripe_webhook_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS billing_audit_logs (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    event_type TEXT NOT NULL,
    stripe_event_id TEXT NOT NULL,
    payload_json TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_billing_audit_logs_tenant_id_created_at
    ON billing_audit_logs(tenant_id, created_at);
