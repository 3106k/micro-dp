DROP INDEX IF EXISTS idx_billing_audit_logs_tenant_id_created_at;
DROP TABLE IF EXISTS billing_audit_logs;

DROP TABLE IF EXISTS stripe_webhook_events;

DROP INDEX IF EXISTS idx_tenant_billing_subscriptions_status;
DROP TABLE IF EXISTS tenant_billing_subscriptions;
