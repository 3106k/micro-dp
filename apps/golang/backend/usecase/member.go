package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/notification"
)

type MemberService struct {
	tenants     domain.TenantRepository
	users       domain.UserRepository
	invitations domain.InvitationRepository
	emailSender notification.EmailSender
	appBaseURL  string
}

func NewMemberService(
	tenants domain.TenantRepository,
	users domain.UserRepository,
	invitations domain.InvitationRepository,
	emailSender notification.EmailSender,
	appBaseURL string,
) *MemberService {
	return &MemberService{
		tenants:     tenants,
		users:       users,
		invitations: invitations,
		emailSender: emailSender,
		appBaseURL:  appBaseURL,
	}
}

func (s *MemberService) ListMembers(ctx context.Context) ([]domain.TenantMember, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return s.tenants.ListMembersByTenantID(ctx, tenantID)
}

func (s *MemberService) CreateInvitation(ctx context.Context, email, role string) (*domain.TenantInvitation, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	actorID, ok := domain.UserIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrInsufficientRole
	}

	// Authorization: actor must be owner or admin
	actorRole, err := s.tenants.GetUserRole(ctx, actorID, tenantID)
	if err != nil {
		return nil, err
	}
	if actorRole != domain.TenantRoleOwner && actorRole != domain.TenantRoleAdmin {
		return nil, domain.ErrInsufficientRole
	}
	// Admin cannot invite as owner
	if actorRole == domain.TenantRoleAdmin && role == domain.TenantRoleOwner {
		return nil, domain.ErrInsufficientRole
	}

	// Check if user is already a member
	user, _ := s.users.FindByEmail(ctx, email)
	if user != nil {
		isMember, err := s.tenants.IsUserInTenant(ctx, user.ID, tenantID)
		if err != nil {
			return nil, err
		}
		if isMember {
			return nil, domain.ErrAlreadyMember
		}
	}

	// Check for pending invitation
	existing, err := s.invitations.FindPendingByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrDuplicateInvitation
	}

	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	inv := &domain.TenantInvitation{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		Token:     token,
		Status:    domain.InvitationStatusPending,
		InvitedBy: actorID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.invitations.Create(ctx, inv); err != nil {
		return nil, err
	}

	// Send invitation email (fire-and-forget)
	go func() {
		actor, err := s.users.FindByID(context.Background(), actorID)
		if err != nil {
			log.Printf("invitation email: find inviter: %v", err)
			return
		}
		tenant, err := s.tenants.FindByID(context.Background(), tenantID)
		if err != nil {
			log.Printf("invitation email: find tenant: %v", err)
			return
		}
		acceptURL := s.appBaseURL + "/api/v1/tenants/current/invitations/" + token + "/accept"
		subj, html, text, err := notification.RenderInvitation(actor.DisplayName, tenant.Name, role, token, acceptURL)
		if err != nil {
			log.Printf("invitation email: render: %v", err)
			return
		}
		msg := &notification.EmailMessage{
			To:      email,
			Subject: subj,
			HTML:    html,
			Text:    text,
		}
		if err := s.emailSender.Send(context.Background(), msg); err != nil {
			log.Printf("invitation email: send: %v", err)
		}
	}()

	return inv, nil
}

func (s *MemberService) AcceptInvitation(ctx context.Context, token string) (*domain.TenantInvitation, error) {
	actorID, ok := domain.UserIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrInsufficientRole
	}

	inv, err := s.invitations.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Verify the invitation email matches the authenticated user
	actor, err := s.users.FindByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if actor.Email != inv.Email {
		return nil, domain.ErrInvitationNotFound
	}

	if inv.Status != domain.InvitationStatusPending {
		return nil, domain.ErrInvitationAlreadyUsed
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, domain.ErrInvitationExpired
	}

	// Check if already a member
	isMember, err := s.tenants.IsUserInTenant(ctx, actorID, inv.TenantID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, domain.ErrAlreadyMember
	}

	// Add user to tenant
	ut := &domain.UserTenant{
		UserID:   actorID,
		TenantID: inv.TenantID,
		Role:     inv.Role,
	}
	if err := s.tenants.AddUserToTenant(ctx, ut); err != nil {
		return nil, err
	}

	// Update invitation status
	now := time.Now()
	if err := s.invitations.UpdateStatus(ctx, inv.ID, domain.InvitationStatusAccepted, &now); err != nil {
		return nil, err
	}

	inv.Status = domain.InvitationStatusAccepted
	inv.AcceptedAt = &now
	return inv, nil
}

func (s *MemberService) UpdateMemberRole(ctx context.Context, targetUserID, newRole string) (*domain.TenantMember, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	actorID, ok := domain.UserIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrInsufficientRole
	}

	if actorID == targetUserID {
		return nil, domain.ErrCannotChangeOwnRole
	}

	actorRole, err := s.tenants.GetUserRole(ctx, actorID, tenantID)
	if err != nil {
		return nil, err
	}

	targetRole, err := s.tenants.GetUserRole(ctx, targetUserID, tenantID)
	if err != nil {
		return nil, err
	}

	// Authorization checks
	if actorRole != domain.TenantRoleOwner && actorRole != domain.TenantRoleAdmin {
		return nil, domain.ErrInsufficientRole
	}
	// Admin cannot change owner's role or promote to owner
	if actorRole == domain.TenantRoleAdmin {
		if targetRole == domain.TenantRoleOwner || newRole == domain.TenantRoleOwner {
			return nil, domain.ErrInsufficientRole
		}
	}

	if err := s.tenants.UpdateUserRole(ctx, targetUserID, tenantID, newRole); err != nil {
		return nil, err
	}

	// Return updated member info
	user, err := s.users.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	return &domain.TenantMember{
		UserID:      user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        newRole,
	}, nil
}

func (s *MemberService) RemoveMember(ctx context.Context, targetUserID string) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return domain.ErrTenantNotFound
	}
	actorID, ok := domain.UserIDFromContext(ctx)
	if !ok {
		return domain.ErrInsufficientRole
	}

	targetRole, err := s.tenants.GetUserRole(ctx, targetUserID, tenantID)
	if err != nil {
		return err
	}

	if actorID == targetUserID {
		// Self-removal: allowed unless last owner
		if targetRole == domain.TenantRoleOwner {
			count, err := s.tenants.CountOwners(ctx, tenantID)
			if err != nil {
				return err
			}
			if count <= 1 {
				return domain.ErrCannotRemoveLastOwner
			}
		}
	} else {
		// Removing someone else: need owner or admin role
		actorRole, err := s.tenants.GetUserRole(ctx, actorID, tenantID)
		if err != nil {
			return err
		}
		if actorRole != domain.TenantRoleOwner && actorRole != domain.TenantRoleAdmin {
			return domain.ErrInsufficientRole
		}
		// Admin cannot remove owner
		if actorRole == domain.TenantRoleAdmin && targetRole == domain.TenantRoleOwner {
			return domain.ErrInsufficientRole
		}
	}

	return s.tenants.RemoveUserFromTenant(ctx, targetUserID, tenantID)
}
