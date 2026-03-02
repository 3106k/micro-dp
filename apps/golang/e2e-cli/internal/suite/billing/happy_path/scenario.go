package happy_path

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/runner"
)

type Scenario struct {
	password    string
	displayName string
}

func NewScenario(password, displayName string) *Scenario {
	return &Scenario{
		password:    password,
		displayName: displayName,
	}
}

func (s *Scenario) ID() string {
	return "billing/happy_path/endpoints"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	// 1. Register new user
	email := fmt.Sprintf("e2e_billing_%d@example.com", time.Now().UnixNano())
	registerReq := map[string]string{
		"email":        email,
		"password":     s.password,
		"display_name": s.displayName,
	}
	var registerResp struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: status=%d body=%s", code, string(body))
	}

	// 2. Login
	loginReq := map[string]string{
		"email":    email,
		"password": s.password,
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login: status=%d body=%s", code, string(body))
	}
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantID)

	// 3. GET /api/v1/billing/subscription (#84)
	//    200 with subscription data is valid for any tenant.
	//    500 with "billing is not configured" means Stripe is not set up -> skip.
	code, body, err = client.GetJSON(ctx, "/api/v1/billing/subscription", nil)
	if err != nil {
		return err
	}
	if code == 500 && isBillingNotConfigured(body) {
		return runner.Skip("stripe is not configured")
	}
	if code != 200 {
		return fmt.Errorf("get subscription: expected 200, got %d body=%s", code, string(body))
	}

	// 4. POST /api/v1/billing/checkout-session (#82)
	//    Use a dummy price_id; 400 is expected (invalid price).
	//    500 with billing error means Stripe is not configured -> skip.
	checkoutReq := map[string]string{
		"price_id": "price_test",
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/billing/checkout-session", checkoutReq, nil)
	if err != nil {
		return err
	}
	if code == 500 && isBillingNotConfigured(body) {
		return runner.Skip("stripe is not configured")
	}
	// 400 is acceptable (bad price_id), 200 means it actually created a session
	if code != 200 && code != 400 {
		return fmt.Errorf("checkout-session: expected 200 or 400, got %d body=%s", code, string(body))
	}

	// 5. POST /api/v1/billing/portal-session (#83)
	//    For a new tenant without an active Stripe subscription, 400 is acceptable.
	//    500 with billing error means Stripe is not configured -> skip.
	portalReq := map[string]string{}
	code, body, err = client.PostJSON(ctx, "/api/v1/billing/portal-session", portalReq, nil)
	if err != nil {
		return err
	}
	if code == 500 && isBillingNotConfigured(body) {
		return runner.Skip("stripe is not configured")
	}
	// 200 = session created, 400 = no active subscription (both valid)
	if code != 200 && code != 400 {
		return fmt.Errorf("portal-session: expected 200 or 400, got %d body=%s", code, string(body))
	}

	// 6. POST /api/v1/billing/webhook (#85)
	//    Webhook endpoint is public (no auth required).
	//    Without a valid Stripe-Signature header, should get 400.
	//    Save and restore token/tenant to test as unauthenticated.
	savedToken := loginResp.Token
	savedTenant := registerResp.TenantID
	client.SetToken("")
	client.SetTenantID("")

	// Send empty body with no Stripe-Signature -> expect 400 (missing signature)
	code, body, err = client.PostJSON(ctx, "/api/v1/billing/webhook", map[string]any{}, nil)
	if err != nil {
		// Restore credentials before returning
		client.SetToken(savedToken)
		client.SetTenantID(savedTenant)
		return err
	}

	// Restore credentials
	client.SetToken(savedToken)
	client.SetTenantID(savedTenant)

	if code == 500 && isBillingNotConfigured(body) {
		return runner.Skip("stripe is not configured")
	}
	if code != 400 {
		return fmt.Errorf("webhook (no signature): expected 400, got %d body=%s", code, string(body))
	}

	return nil
}

// isBillingNotConfigured checks if the response body indicates that
// Stripe/billing is not configured on the server.
func isBillingNotConfigured(body []byte) bool {
	s := strings.ToLower(string(body))
	return strings.Contains(s, "billing is not configured") ||
		strings.Contains(s, "stripe")
}
