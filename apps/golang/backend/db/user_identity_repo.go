package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/user/micro-dp/domain"
)

type UserIdentityRepo struct {
	db DBTX
}

func NewUserIdentityRepo(db DBTX) *UserIdentityRepo {
	return &UserIdentityRepo{db: db}
}

func (r *UserIdentityRepo) Create(ctx context.Context, identity *domain.UserIdentity) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_identities (id, provider, subject, user_id, email, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		identity.ID, identity.Provider, identity.Subject, identity.UserID, identity.Email,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrUserIdentityExists
		}
		return err
	}
	return nil
}

func (r *UserIdentityRepo) FindByProviderSubject(ctx context.Context, provider, subject string) (*domain.UserIdentity, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, provider, subject, user_id, email, created_at, updated_at
		 FROM user_identities
		 WHERE provider = ? AND subject = ?`,
		provider, subject,
	)

	var identity domain.UserIdentity
	if err := row.Scan(
		&identity.ID,
		&identity.Provider,
		&identity.Subject,
		&identity.UserID,
		&identity.Email,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserIdentityNotFound
		}
		return nil, err
	}
	return &identity, nil
}
