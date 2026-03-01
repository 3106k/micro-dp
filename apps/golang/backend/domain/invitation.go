package domain

import (
	"context"
	"errors"
	"time"
)

const (
	TenantRoleOwner  = "owner"
	TenantRoleAdmin  = "admin"
	TenantRoleMember = "member"
)

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusExpired  = "expired"
)

var (
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationExpired     = errors.New("invitation expired")
	ErrInvitationAlreadyUsed = errors.New("invitation already used")
	ErrAlreadyMember         = errors.New("user is already a member")
	ErrDuplicateInvitation   = errors.New("pending invitation already exists")
	ErrCannotRemoveLastOwner = errors.New("cannot remove the last owner")
	ErrInsufficientRole      = errors.New("insufficient role for this action")
	ErrCannotChangeOwnRole   = errors.New("cannot change own role")
)

type TenantInvitation struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Token      string     `json:"token"`
	Status     string     `json:"status"`
	InvitedBy  string     `json:"invited_by"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type TenantMember struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type InvitationRepository interface {
	Create(ctx context.Context, inv *TenantInvitation) error
	FindByToken(ctx context.Context, token string) (*TenantInvitation, error)
	FindPendingByEmail(ctx context.Context, tenantID, email string) (*TenantInvitation, error)
	UpdateStatus(ctx context.Context, id, status string, acceptedAt *time.Time) error
	ListByTenant(ctx context.Context, tenantID string) ([]TenantInvitation, error)
}
