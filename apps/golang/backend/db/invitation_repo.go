package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/user/micro-dp/domain"
)

type InvitationRepo struct {
	db DBTX
}

func NewInvitationRepo(db DBTX) *InvitationRepo {
	return &InvitationRepo{db: db}
}

func (r *InvitationRepo) Create(ctx context.Context, inv *domain.TenantInvitation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tenant_invitations (id, tenant_id, email, role, token, status, invited_by, expires_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		inv.ID, inv.TenantID, inv.Email, inv.Role, inv.Token, inv.Status, inv.InvitedBy, inv.ExpiresAt,
	)
	return err
}

func (r *InvitationRepo) FindByToken(ctx context.Context, token string) (*domain.TenantInvitation, error) {
	var inv domain.TenantInvitation
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, email, role, token, status, invited_by, expires_at, accepted_at, created_at, updated_at
		 FROM tenant_invitations WHERE token = ?`, token,
	).Scan(&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token, &inv.Status, &inv.InvitedBy, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *InvitationRepo) FindPendingByEmail(ctx context.Context, tenantID, email string) (*domain.TenantInvitation, error) {
	var inv domain.TenantInvitation
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, email, role, token, status, invited_by, expires_at, accepted_at, created_at, updated_at
		 FROM tenant_invitations WHERE tenant_id = ? AND email = ? AND status = 'pending'`, tenantID, email,
	).Scan(&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token, &inv.Status, &inv.InvitedBy, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *InvitationRepo) UpdateStatus(ctx context.Context, id, status string, acceptedAt *time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tenant_invitations SET status = ?, accepted_at = ?, updated_at = datetime('now') WHERE id = ?`,
		status, acceptedAt, id,
	)
	return err
}

func (r *InvitationRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.TenantInvitation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, email, role, token, status, invited_by, expires_at, accepted_at, created_at, updated_at
		 FROM tenant_invitations WHERE tenant_id = ? ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []domain.TenantInvitation
	for rows.Next() {
		var inv domain.TenantInvitation
		if err := rows.Scan(&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token, &inv.Status, &inv.InvitedBy, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}
