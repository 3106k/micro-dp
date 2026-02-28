package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/user/micro-dp/domain"
)

type UserRepo struct {
	db DBTX
}

func NewUserRepo(db DBTX) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, display_name, is_superadmin, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		user.ID, user.Email, user.PasswordHash, user.DisplayName, user.IsSuperadmin,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, display_name, is_superadmin, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	)
	return scanUser(row)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, display_name, is_superadmin, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	)
	return scanUser(row)
}

func (r *UserRepo) SetSuperadmin(ctx context.Context, id string, isSuperadmin bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_superadmin = ?, updated_at = datetime('now') WHERE id = ?`,
		isSuperadmin, id,
	)
	return err
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.IsSuperadmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}
