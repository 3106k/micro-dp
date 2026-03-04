package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/user/micro-dp/domain"
)

type CredentialRepo struct {
	db DBTX
}

func NewCredentialRepo(db DBTX) *CredentialRepo {
	return &CredentialRepo{db: db}
}

func (r *CredentialRepo) Create(ctx context.Context, c *domain.Credential) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO credentials (id, user_id, tenant_id, provider, provider_label, access_token, refresh_token, token_expiry, scopes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		c.ID, c.UserID, c.TenantID, c.Provider, c.ProviderLabel, c.AccessToken, c.RefreshToken, c.TokenExpiry, c.Scopes,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrCredentialAlreadyExists
		}
		return err
	}
	return nil
}

func (r *CredentialRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Credential, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, tenant_id, provider, provider_label, access_token, refresh_token, token_expiry, scopes, created_at, updated_at
		 FROM credentials WHERE tenant_id = ? AND id = ?`, tenantID, id,
	)
	var c domain.Credential
	if err := row.Scan(&c.ID, &c.UserID, &c.TenantID, &c.Provider, &c.ProviderLabel, &c.AccessToken, &c.RefreshToken, &c.TokenExpiry, &c.Scopes, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCredentialNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CredentialRepo) FindByUserAndProvider(ctx context.Context, userID, tenantID, provider string) (*domain.Credential, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, tenant_id, provider, provider_label, access_token, refresh_token, token_expiry, scopes, created_at, updated_at
		 FROM credentials WHERE user_id = ? AND tenant_id = ? AND provider = ?`, userID, tenantID, provider,
	)
	var c domain.Credential
	if err := row.Scan(&c.ID, &c.UserID, &c.TenantID, &c.Provider, &c.ProviderLabel, &c.AccessToken, &c.RefreshToken, &c.TokenExpiry, &c.Scopes, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCredentialNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CredentialRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Credential, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, tenant_id, provider, provider_label, access_token, refresh_token, token_expiry, scopes, created_at, updated_at
		 FROM credentials WHERE tenant_id = ?
		 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []domain.Credential
	for rows.Next() {
		var c domain.Credential
		if err := rows.Scan(&c.ID, &c.UserID, &c.TenantID, &c.Provider, &c.ProviderLabel, &c.AccessToken, &c.RefreshToken, &c.TokenExpiry, &c.Scopes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		credentials = append(credentials, c)
	}
	return credentials, rows.Err()
}

func (r *CredentialRepo) Update(ctx context.Context, c *domain.Credential) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE credentials SET access_token = ?, refresh_token = ?, token_expiry = ?, provider_label = ?, scopes = ?, updated_at = datetime('now')
		 WHERE tenant_id = ? AND id = ?`,
		c.AccessToken, c.RefreshToken, c.TokenExpiry, c.ProviderLabel, c.Scopes, c.TenantID, c.ID,
	)
	return err
}

func (r *CredentialRepo) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM credentials WHERE tenant_id = ? AND id = ?`,
		tenantID, id,
	)
	return err
}
