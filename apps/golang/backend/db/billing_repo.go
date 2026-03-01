package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type BillingSubscriptionRepo struct {
	db DBTX
}

func NewBillingSubscriptionRepo(db DBTX) *BillingSubscriptionRepo {
	return &BillingSubscriptionRepo{db: db}
}

func (r *BillingSubscriptionRepo) FindByTenantID(ctx context.Context, tenantID string) (*domain.BillingSubscription, error) {
	return r.findOne(ctx,
		`SELECT id, stripe_customer_id, stripe_subscription_id, stripe_price_id, subscription_status, current_period_end, updated_at
		 FROM tenant_billing_subscriptions
		 WHERE id = ?`,
		tenantID,
	)
}

func (r *BillingSubscriptionRepo) FindByStripeCustomerID(ctx context.Context, customerID string) (*domain.BillingSubscription, error) {
	return r.findOne(ctx,
		`SELECT id, stripe_customer_id, stripe_subscription_id, stripe_price_id, subscription_status, current_period_end, updated_at
		 FROM tenant_billing_subscriptions
		 WHERE stripe_customer_id = ?`,
		customerID,
	)
}

func (r *BillingSubscriptionRepo) findOne(ctx context.Context, q string, arg any) (*domain.BillingSubscription, error) {
	var sub domain.BillingSubscription
	var stripeCustomerID sql.NullString
	var stripeSubscriptionID sql.NullString
	var stripePriceID sql.NullString
	err := r.db.QueryRowContext(ctx, q, arg).Scan(
		&sub.TenantID,
		&stripeCustomerID,
		&stripeSubscriptionID,
		&stripePriceID,
		&sub.SubscriptionStatus,
		&sub.CurrentPeriodEnd,
		&sub.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrBillingNotFound
		}
		return nil, err
	}
	if stripeCustomerID.Valid {
		sub.StripeCustomerID = stripeCustomerID.String
	}
	if stripeSubscriptionID.Valid {
		sub.StripeSubscriptionID = stripeSubscriptionID.String
	}
	if stripePriceID.Valid {
		sub.StripePriceID = stripePriceID.String
	}
	return &sub, nil
}

func (r *BillingSubscriptionRepo) UpsertByTenantID(ctx context.Context, sub *domain.BillingSubscription) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tenant_billing_subscriptions
		 (id, stripe_customer_id, stripe_subscription_id, stripe_price_id, subscription_status, current_period_end, created_at, updated_at)
		 VALUES (?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(id) DO UPDATE SET
		   stripe_customer_id = excluded.stripe_customer_id,
		   stripe_subscription_id = excluded.stripe_subscription_id,
		   stripe_price_id = excluded.stripe_price_id,
		   subscription_status = excluded.subscription_status,
		   current_period_end = excluded.current_period_end,
		   updated_at = datetime('now')`,
		sub.TenantID,
		sub.StripeCustomerID,
		sub.StripeSubscriptionID,
		sub.StripePriceID,
		sub.SubscriptionStatus,
		sub.CurrentPeriodEnd,
	)
	return err
}

type StripeWebhookEventRepo struct {
	db DBTX
}

func NewStripeWebhookEventRepo(db DBTX) *StripeWebhookEventRepo {
	return &StripeWebhookEventRepo{db: db}
}

func (r *StripeWebhookEventRepo) MarkProcessed(ctx context.Context, eventID, eventType string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO stripe_webhook_events (id, event_type, created_at)
		 VALUES (?, ?, datetime('now'))
		 ON CONFLICT(id) DO NOTHING`,
		eventID, eventType,
	)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

type BillingAuditLogRepo struct {
	db DBTX
}

func NewBillingAuditLogRepo(db DBTX) *BillingAuditLogRepo {
	return &BillingAuditLogRepo{db: db}
}

func (r *BillingAuditLogRepo) Create(ctx context.Context, log *domain.BillingAuditLog) error {
	id := log.ID
	if id == "" {
		id = uuid.New().String()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO billing_audit_logs
		 (id, tenant_id, event_type, stripe_event_id, payload_json, created_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'))`,
		id, log.TenantID, log.EventType, log.StripeEventID, log.PayloadJSON,
	)
	return err
}
