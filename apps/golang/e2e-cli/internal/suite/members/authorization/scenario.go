package authorization

import (
	"context"
	"fmt"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/openapi"
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
	client.SetTenantID(regA.TenantId)

	invB, err := createInvitation(ctx, client, emailB, openapi.Member)
	if err != nil {
		return fmt.Errorf("invite B: %w", err)
	}

	invC, err := createInvitation(ctx, client, emailC, openapi.Admin)
	if err != nil {
		return fmt.Errorf("invite C: %w", err)
	}

	// B accepts
	client.SetToken(tokenB)
	client.SetTenantID("")
	code, body, err := client.PostJSON(ctx, "/api/v1/tenants/current/invitations/"+*invB.Token+"/accept", nil, nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("B accept: status=%d body=%s", code, string(body))
	}

	// C accepts
	client.SetToken(tokenC)
	client.SetTenantID("")
	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations/"+*invC.Token+"/accept", nil, nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("C accept: status=%d body=%s", code, string(body))
	}

	// Find user IDs by listing members as A
	client.SetToken(tokenA)
	client.SetTenantID(regA.TenantId)

	members, err := listMembers(ctx, client)
	if err != nil {
		return err
	}

	var userAID, userBID string
	for _, m := range members {
		switch m.Role {
		case openapi.Owner:
			userAID = m.UserId
		case openapi.Member:
			userBID = m.UserId
		}
	}

	// 1. User B (member): create invitation -> 403
	client.SetToken(tokenB)
	client.SetTenantID(regA.TenantId)

	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations", openapi.CreateInvitationRequest{
		Email: openapi.Email("someone@example.com"), Role: openapi.Member,
	}, nil)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("member invite: expected 403, got %d body=%s", code, string(body))
	}

	// 2. User B (member): update role -> 403
	code, body, err = client.PatchJSON(ctx, "/api/v1/tenants/current/members/"+userAID, openapi.UpdateMemberRoleRequest{
		Role: openapi.Member,
	}, nil)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("member update role: expected 403, got %d body=%s", code, string(body))
	}

	// 3. User B (member): delete other -> 403
	code, body, err = client.Delete(ctx, "/api/v1/tenants/current/members/"+userAID)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("member delete other: expected 403, got %d body=%s", code, string(body))
	}

	// 4. User B (member): self-leave -> 204
	code, body, err = client.Delete(ctx, "/api/v1/tenants/current/members/"+userBID)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("member self-leave: expected 204, got %d body=%s", code, string(body))
	}

	// 5. User C (admin): invite as owner -> 403
	client.SetToken(tokenC)
	client.SetTenantID(regA.TenantId)

	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations", openapi.CreateInvitationRequest{
		Email: openapi.Email("owner-test@example.com"), Role: openapi.Owner,
	}, nil)
	if err != nil {
		return err
	}
	if code != 403 {
		return fmt.Errorf("admin invite owner: expected 403, got %d body=%s", code, string(body))
	}

	// 6. User C (admin): invite as member -> 201
	code, body, err = client.PostJSON(ctx, "/api/v1/tenants/current/invitations", openapi.CreateInvitationRequest{
		Email: openapi.Email(emailD), Role: openapi.Member,
	}, nil)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("admin invite member: expected 201, got %d body=%s", code, string(body))
	}

	// 7. User A (owner): self-delete as last owner -> 409
	client.SetToken(tokenA)
	client.SetTenantID(regA.TenantId)

	code, body, err = client.Delete(ctx, "/api/v1/tenants/current/members/"+userAID)
	if err != nil {
		return err
	}
	if code != 409 {
		return fmt.Errorf("last owner self-delete: expected 409, got %d body=%s", code, string(body))
	}

	return nil
}

func registerAndLogin(ctx context.Context, client *httpclient.Client, email, password, displayName string) (*openapi.RegisterResponse, string, error) {
	var regResp openapi.RegisterResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", openapi.RegisterRequest{
		Email: openapi.Email(email), Password: password, DisplayName: openapi.Ptr(displayName),
	}, &regResp)
	if err != nil {
		return nil, "", err
	}
	if code != 201 {
		return nil, "", fmt.Errorf("register: status=%d body=%s", code, string(body))
	}

	var loginResp openapi.LoginResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", openapi.LoginRequest{
		Email: openapi.Email(email), Password: password,
	}, &loginResp)
	if err != nil {
		return nil, "", err
	}
	if code != 200 {
		return nil, "", fmt.Errorf("login: status=%d body=%s", code, string(body))
	}

	return &regResp, loginResp.Token, nil
}

func listMembers(ctx context.Context, client *httpclient.Client) ([]openapi.TenantMember, error) {
	var resp openapi.ListResponse[openapi.TenantMember]
	code, body, err := client.GetJSON(ctx, "/api/v1/tenants/current/members", &resp)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("list members: status=%d body=%s", code, string(body))
	}
	return resp.Items, nil
}

func createInvitation(ctx context.Context, client *httpclient.Client, email string, role openapi.TenantRole) (*openapi.TenantInvitation, error) {
	var resp openapi.TenantInvitation
	code, body, err := client.PostJSON(ctx, "/api/v1/tenants/current/invitations", openapi.CreateInvitationRequest{
		Email: openapi.Email(email), Role: role,
	}, &resp)
	if err != nil {
		return nil, err
	}
	if code != 201 {
		return nil, fmt.Errorf("create invitation: status=%d body=%s", code, string(body))
	}
	return &resp, nil
}
