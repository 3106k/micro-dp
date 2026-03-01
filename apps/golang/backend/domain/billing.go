package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrBillingNotFound           = errors.New("billing not found")
	ErrStripeEventAlreadyHandled = errors.New("stripe event already handled")
)

type BillingSubscription struct {
	TenantID             string
	StripeCustomerID     string
	StripeSubscriptionID string
	StripePriceID        string
	SubscriptionStatus   string
	CurrentPeriodEnd     *time.Time
	UpdatedAt            time.Time
}

type BillingSubscriptionRepository interface {
	FindByTenantID(ctx context.Context, tenantID string) (*BillingSubscription, error)
	FindByStripeCustomerID(ctx context.Context, customerID string) (*BillingSubscription, error)
	UpsertByTenantID(ctx context.Context, sub *BillingSubscription) error
}

type StripeWebhookEventRepository interface {
	MarkProcessed(ctx context.Context, eventID, eventType string) (bool, error)
}

type BillingAuditLog struct {
	ID            string
	TenantID      string
	EventType     string
	StripeEventID string
	PayloadJSON   string
	CreatedAt     time.Time
}

type BillingAuditLogRepository interface {
	Create(ctx context.Context, log *BillingAuditLog) error
}
