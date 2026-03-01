CREATE TABLE IF NOT EXISTS tenant_billing_subscriptions_old (
    id TEXT PRIMARY KEY REFERENCES tenants(id),
    stripe_customer_id TEXT NOT NULL DEFAULT '' UNIQUE,
    stripe_subscription_id TEXT NOT NULL DEFAULT '' UNIQUE,
    stripe_price_id TEXT NOT NULL DEFAULT '',
    subscription_status TEXT NOT NULL DEFAULT 'inactive',
    current_period_end DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO tenant_billing_subscriptions_old (
    id,
    stripe_customer_id,
    stripe_subscription_id,
    stripe_price_id,
    subscription_status,
    current_period_end,
    created_at,
    updated_at
)
SELECT
    id,
    IFNULL(stripe_customer_id, ''),
    IFNULL(stripe_subscription_id, ''),
    stripe_price_id,
    subscription_status,
    current_period_end,
    created_at,
    updated_at
FROM tenant_billing_subscriptions;

DROP INDEX IF EXISTS idx_tenant_billing_subscriptions_status;
DROP TABLE tenant_billing_subscriptions;
ALTER TABLE tenant_billing_subscriptions_old RENAME TO tenant_billing_subscriptions;

CREATE INDEX IF NOT EXISTS idx_tenant_billing_subscriptions_status
    ON tenant_billing_subscriptions(subscription_status);
