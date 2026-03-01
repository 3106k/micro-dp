package happy_path

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
	return "members/happy_path/invite_accept_manage"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	ts := time.Now().UnixNano()

	// 1. Register User A (owner)
	emailA := fmt.Sprintf("e2e_members_a_%d@example.com", ts)
	regA, tokenA, err := registerAndLogin(ctx, client, emailA, s.password, "User A")
	if err != nil {
		return fmt.Errorf("setup A: %w", err)
	}

	// 2. Register User B (gets own tenant)
	emailB := fmt.Sprintf("e2e_members_b_%d@example.com", ts)
	_, tokenB, err := registerAndLogin(ctx, client, emailB, s.password, "User B")
	if err != nil {
		return fmt.Errorf("setup B: %w", err)
	}

	// 3. User A: list members → 1 (owner)
	client.SetToken(tokenA)
	client.SetTenantID(regA.TenantID)

	members, err := listMembers(ctx, client)
	if err != nil {
		return err
	}
	if len(members) != 1 {
		return fmt.Errorf("expected 1 member, got %d", len(members))
	}
	if members[0].Role != "owner" {
		return fmt.Errorf("expected owner role, got %s", members[0].Role)
	}

	// 4. User A: invite User B as member → 201
	inv, err := createInvitation(ctx, client, emailB, "member")
	if err != nil {
		return err
	}
	if inv.Status != "pending" {
		return fmt.Errorf("expected pending status, got %s", inv.Status)
	}
	if inv.Token == "" {
		return fmt.Errorf("invitation token is empty")
	}
	invTokenB := inv.Token

	// 5. User A: duplicate invitation → 409
	code, body, err := client.PostJSON(ctx, "/api/v1/tenants/current/invitations", map[string]string{
		"email": emailB, "role": "member",
	}, nil)
	if err != nil {
		return err
	}
	if code != 409 {
		return fmt.Errorf("duplicate invitation: expected 409, got %d body=%s", code, string(body))
	}

	// 6. User B: accept invitation
	client.SetToken(tokenB)
	client.SetTenantID("")

	var acceptResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations/"+invTokenB+"/accept", nil, &acceptResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("accept invitation: status=%d body=%s", code, string(body))
	}
	if acceptResp.Status != "accepted" {
		return fmt.Errorf("expected accepted status, got %s", acceptResp.Status)
	}

	// 7. User A: list members → 2
	client.SetToken(tokenA)
	client.SetTenantID(regA.TenantID)

	members, err = listMembers(ctx, client)
	if err != nil {
		return err
	}
	if len(members) != 2 {
		return fmt.Errorf("expected 2 members, got %d", len(members))
	}

	// 8. User A: change B to admin → 200
	var userBID string
	for _, m := range members {
		if m.Role == "member" {
			userBID = m.UserID
			break
		}
	}
	if userBID == "" {
		return fmt.Errorf("cannot find member user B")
	}

	var updateResp struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	code, body, err = client.PatchJSON(ctx, "/api/v1/tenants/current/members/"+userBID, map[string]string{
		"role": "admin",
	}, &updateResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("update role: status=%d body=%s", code, string(body))
	}
	if updateResp.Role != "admin" {
		return fmt.Errorf("expected admin role, got %s", updateResp.Role)
	}

	// 9. User A: remove B → 204
	code, body, err = client.Delete(ctx, "/api/v1/tenants/current/members/"+userBID)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("remove member: expected 204, got %d body=%s", code, string(body))
	}

	// 10. User A: list members → 1 again
	members, err = listMembers(ctx, client)
	if err != nil {
		return err
	}
	if len(members) != 1 {
		return fmt.Errorf("expected 1 member after remove, got %d", len(members))
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
