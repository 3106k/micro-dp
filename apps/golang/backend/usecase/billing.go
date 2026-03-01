package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v82"
	billingportalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/webhook"
	"github.com/user/micro-dp/domain"
)

var (
	ErrBillingNotConfigured = errors.New("billing is not configured")
)

type StripeConfig struct {
	SecretKey         string
	WebhookSecret     string
	DefaultSuccessURL string
	DefaultCancelURL  string
	DefaultReturnURL  string
	PriceIDToPlanName map[string]string
}

type BillingService struct {
	subscriptions domain.BillingSubscriptionRepository
	webhookEvents domain.StripeWebhookEventRepository
	auditLogs     domain.BillingAuditLogRepository
	plans         domain.PlanRepository
	planService   *PlanService
	stripeCfg     StripeConfig
}

func NewBillingService(
	subscriptions domain.BillingSubscriptionRepository,
	webhookEvents domain.StripeWebhookEventRepository,
	auditLogs domain.BillingAuditLogRepository,
	plans domain.PlanRepository,
	planService *PlanService,
	stripeCfg StripeConfig,
) *BillingService {
	return &BillingService{
		subscriptions: subscriptions,
		webhookEvents: webhookEvents,
		auditLogs:     auditLogs,
		plans:         plans,
		planService:   planService,
		stripeCfg:     stripeCfg,
	}
}

func ParsePriceIDToPlanMap(raw string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		p := strings.SplitN(part, ":", 2)
		if len(p) != 2 {
			continue
		}
		priceID := strings.TrimSpace(p[0])
		planName := strings.TrimSpace(p[1])
		if priceID == "" || planName == "" {
			continue
		}
		out[priceID] = planName
	}
	return out
}

func (s *BillingService) Enabled() bool {
	return strings.TrimSpace(s.stripeCfg.SecretKey) != ""
}

func (s *BillingService) CreateCheckoutSession(ctx context.Context, priceID, successURL, cancelURL string) (string, error) {
	if !s.Enabled() {
		return "", ErrBillingNotConfigured
	}
	if strings.TrimSpace(priceID) == "" {
		return "", fmt.Errorf("price_id is required")
	}
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant id not found in context")
	}

	customerID, err := s.ensureCustomer(ctx, tenantID)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(successURL) == "" {
		successURL = s.stripeCfg.DefaultSuccessURL
	}
	if strings.TrimSpace(cancelURL) == "" {
		cancelURL = s.stripeCfg.DefaultCancelURL
	}
	if successURL == "" || cancelURL == "" {
		return "", fmt.Errorf("success_url and cancel_url are required")
	}

	stripe.Key = s.stripeCfg.SecretKey
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		Customer:   stripe.String(customerID),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
	}
	params.AddMetadata("tenant_id", tenantID)

	session, err := checkoutsession.New(params)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}
	if session.URL == "" {
		return "", fmt.Errorf("checkout session url is empty")
	}
	return session.URL, nil
}

func (s *BillingService) CreatePortalSession(ctx context.Context, returnURL string) (string, error) {
	if !s.Enabled() {
		return "", ErrBillingNotConfigured
	}
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant id not found in context")
	}

	customerID, err := s.ensureCustomer(ctx, tenantID)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(returnURL) == "" {
		returnURL = s.stripeCfg.DefaultReturnURL
	}
	if returnURL == "" {
		return "", fmt.Errorf("return_url is required")
	}

	stripe.Key = s.stripeCfg.SecretKey
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}
	session, err := billingportalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("create portal session: %w", err)
	}
	if session.URL == "" {
		return "", fmt.Errorf("portal session url is empty")
	}
	return session.URL, nil
}

func (s *BillingService) GetSubscription(ctx context.Context) (*domain.BillingSubscription, *domain.Plan, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("tenant id not found in context")
	}
	sub, err := s.subscriptions.FindByTenantID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrBillingNotFound) {
			return &domain.BillingSubscription{
				TenantID:           tenantID,
				SubscriptionStatus: "inactive",
			}, nil, nil
		}
		return nil, nil, err
	}

	plan, _, err := s.planService.GetTenantPlan(ctx)
	if err != nil && !errors.Is(err, domain.ErrPlanNotFound) {
		return nil, nil, err
	}
	return sub, plan, nil
}

func (s *BillingService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.Enabled() || strings.TrimSpace(s.stripeCfg.WebhookSecret) == "" {
		return ErrBillingNotConfigured
	}
	event, err := webhook.ConstructEvent(payload, signature, s.stripeCfg.WebhookSecret)
	if err != nil {
		return fmt.Errorf("verify webhook signature: %w", err)
	}

	created, err := s.webhookEvents.MarkProcessed(ctx, event.ID, string(event.Type))
	if err != nil {
		return err
	}
	if !created {
		return nil
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return err
		}
		tenantID := session.Metadata["tenant_id"]
		customerID := ""
		if session.Customer != nil {
			customerID = session.Customer.ID
		}
		subscriptionID := ""
		if session.Subscription != nil {
			subscriptionID = session.Subscription.ID
		}
		if tenantID == "" || customerID == "" {
			return fmt.Errorf("invalid checkout.session.completed payload")
		}
		_ = s.audit(ctx, tenantID, string(event.Type), event.ID, payload)
		return s.syncTenantSubscription(ctx, tenantID, customerID, subscriptionID, "", "active", nil)

	case "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return err
		}
		customerID := ""
		if sub.Customer != nil {
			customerID = sub.Customer.ID
		}
		if customerID == "" {
			return fmt.Errorf("invalid subscription payload")
		}
		billingSub, err := s.subscriptions.FindByStripeCustomerID(ctx, customerID)
		if err != nil {
			return err
		}
		priceID := ""
		if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
			priceID = sub.Items.Data[0].Price.ID
		}
		var periodEnd *time.Time
		if len(sub.Items.Data) > 0 && sub.Items.Data[0].CurrentPeriodEnd > 0 {
			t := time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0).UTC()
			periodEnd = &t
		}
		status := string(sub.Status)
		if event.Type == "customer.subscription.deleted" {
			status = "canceled"
		}
		_ = s.audit(ctx, billingSub.TenantID, string(event.Type), event.ID, payload)
		return s.syncTenantSubscription(ctx, billingSub.TenantID, customerID, sub.ID, priceID, status, periodEnd)

	case "invoice.paid", "invoice.payment_failed":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return err
		}
		customerID := ""
		if invoice.Customer != nil {
			customerID = invoice.Customer.ID
		}
		if customerID == "" {
			return fmt.Errorf("invalid invoice payload")
		}
		billingSub, err := s.subscriptions.FindByStripeCustomerID(ctx, customerID)
		if err != nil {
			return err
		}
		status := billingSub.SubscriptionStatus
		if event.Type == "invoice.payment_failed" {
			status = "past_due"
		}
		_ = s.audit(ctx, billingSub.TenantID, string(event.Type), event.ID, payload)
		return s.syncTenantSubscription(
			ctx,
			billingSub.TenantID,
			customerID,
			billingSub.StripeSubscriptionID,
			billingSub.StripePriceID,
			status,
			billingSub.CurrentPeriodEnd,
		)
	}

	return nil
}

func (s *BillingService) audit(ctx context.Context, tenantID, eventType, eventID string, payload []byte) error {
	if s.auditLogs == nil {
		return nil
	}
	if err := s.auditLogs.Create(ctx, &domain.BillingAuditLog{
		TenantID:      tenantID,
		EventType:     eventType,
		StripeEventID: eventID,
		PayloadJSON:   string(payload),
	}); err != nil {
		log.Printf("billing audit log failed tenant=%s event=%s: %v", tenantID, eventType, err)
		return err
	}
	return nil
}

func (s *BillingService) syncTenantSubscription(
	ctx context.Context,
	tenantID, customerID, subscriptionID, priceID, status string,
	currentPeriodEnd *time.Time,
) error {
	sub := &domain.BillingSubscription{
		TenantID:             tenantID,
		StripeCustomerID:     customerID,
		StripeSubscriptionID: subscriptionID,
		StripePriceID:        priceID,
		SubscriptionStatus:   status,
		CurrentPeriodEnd:     currentPeriodEnd,
	}
	if err := s.subscriptions.UpsertByTenantID(ctx, sub); err != nil {
		return err
	}

	plan, err := s.resolvePlan(ctx, priceID, status)
	if err != nil {
		return err
	}
	if plan == nil {
		return nil
	}
	_, _, err = s.planService.AssignPlan(ctx, tenantID, plan.ID)
	return err
}

func (s *BillingService) resolvePlan(ctx context.Context, priceID, status string) (*domain.Plan, error) {
	paid := status == "active" || status == "trialing" || status == "past_due"
	if paid {
		if planName, ok := s.stripeCfg.PriceIDToPlanName[priceID]; ok && planName != "" {
			return s.plans.FindByName(ctx, planName)
		}
	}
	return s.plans.FindDefault(ctx)
}

func (s *BillingService) ensureCustomer(ctx context.Context, tenantID string) (string, error) {
	sub, err := s.subscriptions.FindByTenantID(ctx, tenantID)
	if err != nil && !errors.Is(err, domain.ErrBillingNotFound) {
		return "", err
	}
	if sub != nil && strings.TrimSpace(sub.StripeCustomerID) != "" {
		return sub.StripeCustomerID, nil
	}

	stripe.Key = s.stripeCfg.SecretKey
	params := &stripe.CustomerParams{}
	params.AddMetadata("tenant_id", tenantID)
	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("create stripe customer: %w", err)
	}
	if err := s.subscriptions.UpsertByTenantID(ctx, &domain.BillingSubscription{
		TenantID:           tenantID,
		StripeCustomerID:   c.ID,
		SubscriptionStatus: "inactive",
	}); err != nil {
		return "", err
	}
	return c.ID, nil
}
