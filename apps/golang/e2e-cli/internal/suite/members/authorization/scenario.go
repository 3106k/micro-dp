package authorization

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
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
	return "members/authorization/role_enforcement"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	ts := time.Now().UnixNano()

	// Setup: Register User A (owner), B, C
	emailA := fmt.Sprintf("e2e_auth_a_%d@example.com", ts)
	emailB := fmt.Sprintf("e2e_auth_b_%d@example.com", ts)
	emailC := fmt.Sprintf("e2e_auth_c_%d@example.com", ts)
	emailD := fmt.Sprintf("e2e_auth_d_%d@example.com", ts)

	regA, tokenA, err := registerAndLogin(ctx, client, emailA, s.password, "Auth A")
	if err != nil {
		return fmt.Errorf("setup A: %w", err)
	}
	_, tokenB, err := registerAndLogin(ctx, client, emailB, s.password, "Auth B")
	if err != nil {
		return fmt.Errorf("setup B: %w", err)
	}
	_, tokenC, err := registerAndLogin(ctx, client, emailC, s.password, "Auth C")
	if err != nil {
		return fmt.Errorf("setup C: %w", err)
	}
	// Register D for later use (admin invite test)
	registerAndLogin(ctx, client, emailD, s.password, "Auth D")

	// User A invites B as member and C as admin
	client.SetToken(tokenA)
	client.SetTenantID(regA.TenantID)

	invB, err := createInvitation(ctx, client, emailB, "member")
	if err != nil {
		return fmt.Errorf("invite B: %w", err)
	}

	invC, err := createInvitation(ctx, client, emailC, "admin")
	if err != nil {
		return fmt.Errorf("invite C: %w", err)
	}

	// B accepts
	client.SetToken(tokenB)
	client.SetTenantID("")
	code, body, err := client.PostJSON(ctx, "/api/v1/tenants/current/invitations/"+invB.Token+"/accept", nil, nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("B accept: status=%d body=%s", code, string(body))
	}

	// C accepts
	client.SetToken(tokenC)
	client.SetTenantID("")
	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations/"+invC.Token+"/accept", nil, nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("C accept: status=%d body=%s", code, string(body))
	}

	// Find user IDs by listing members as A
	client.SetToken(tokenA)
	client.SetTenantID(regA.TenantID)

	members, err := listMembers(ctx, client)
	if err != nil {
		return err
	}

	var userAID, userBID string
	for _, m := range members {
		switch m.Role {
		case "owner":
			userAID = m.UserID
		case "member":
			userBID = m.UserID
		}
	}

	// 1. User B (member): create invitation → 403
	client.SetToken(tokenB)
	client.SetTenantID(regA.TenantID)

	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations", map[string]string{
		"email": "someone@example.com", "role": "member",
	}, nil)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("member invite: expected 403, got %d body=%s", code, string(body))
	}

	// 2. User B (member): update role → 403
	code, body, err = client.PatchJSON(ctx, "/api/v1/tenants/current/members/"+userAID, map[string]string{
		"role": "member",
	}, nil)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("member update role: expected 403, got %d body=%s", code, string(body))
	}

	// 3. User B (member): delete other → 403
	code, body, err = client.Delete(ctx, "/api/v1/tenants/current/members/"+userAID)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("member delete other: expected 403, got %d body=%s", code, string(body))
	}

	// 4. User B (member): self-leave → 204
	code, body, err = client.Delete(ctx, "/api/v1/tenants/current/members/"+userBID)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("member self-leave: expected 204, got %d body=%s", code, string(body))
	}

	// 5. User C (admin): invite as owner → 403
	client.SetToken(tokenC)
	client.SetTenantID(regA.TenantID)

	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations", map[string]string{
		"email": "owner-test@example.com", "role": "owner",
	}, nil)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("admin invite owner: expected 403, got %d body=%s", code, string(body))
	}

	// 6. User C (admin): invite as member → 201
	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations", map[string]string{
		"email": emailD, "role": "member",
	}, nil)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("admin invite member: expected 201, got %d body=%s", code, string(body))
	}

	// 7. User A (owner): self-delete as last owner → 409
	client.SetToken(tokenA)
	client.SetTenantID(regA.TenantID)

	code, body, err = client.Delete(ctx, "/api/v1/tenants/current/members/"+userAID)
	if err != nil {
		return err
	}
	if code != 409 {
		return fmt.Errorf("last owner self-delete: expected 409, got %d body=%s", code, string(body))
	}

	return nil
}

type regResult struct {
	UserID   string
	TenantID string
}

func registerAndLogin(ctx context.Context, client *httpclient.Client, email, password, displayName string) (*regResult, string, error) {
	var regResp struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", map[string]string{
		"email": email, "password": password, "display_name": displayName,
	}, &regResp)
	if err != nil {
		return nil, "", err
	}
	if code != 201 {
		return nil, "", fmt.Errorf("register: status=%d body=%s", code, string(body))
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", map[string]string{
		"email": email, "password": password,
	}, &loginResp)
	if err != nil {
		return nil, "", err
	}
	if code != 200 {
		return nil, "", fmt.Errorf("login: status=%d body=%s", code, string(body))
	}

	return &regResult{UserID: regResp.UserID, TenantID: regResp.TenantID}, loginResp.Token, nil
}

type memberInfo struct {
	UserID string
	Role   string
}

func listMembers(ctx context.Context, client *httpclient.Client) ([]memberInfo, error) {
	var resp struct {
		Items []struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		} `json:"items"`
	}
	code, body, err := client.GetJSON(ctx, "/api/v1/tenants/current/members", &resp)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("list members: status=%d body=%s", code, string(body))
	}
	out := make([]memberInfo, len(resp.Items))
	for i, item := range resp.Items {
		out[i] = memberInfo{UserID: item.UserID, Role: item.Role}
	}
	return out, nil
}

type invInfo struct {
	ID     string
	Token  string
	Status string
}

func createInvitation(ctx context.Context, client *httpclient.Client, email, role string) (*invInfo, error) {
	var resp struct {
		ID     string `json:"id"`
		Token  string `json:"token"`
		Status string `json:"status"`
	}
	code, body, err := client.PostJSON(ctx, "/api/v1/tenants/current/invitations", map[string]string{
		"email": email, "role": role,
	}, &resp)
	if err != nil {
		return nil, err
	}
	if code != 201 {
		return nil, fmt.Errorf("create invitation: status=%d body=%s", code, string(body))
	}
	return &invInfo{ID: resp.ID, Token: resp.Token, Status: resp.Status}, nil
}
