package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/user/micro-dp/domain"
)

type AuthService struct {
	users     domain.UserRepository
	tenants   domain.TenantRepository
	jwtSecret []byte
}

func NewAuthService(users domain.UserRepository, tenants domain.TenantRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		tenants:   tenants,
		jwtSecret: []byte(jwtSecret),
	}
}

type RegisterResult struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*RegisterResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := uuid.New().String()
	user := &domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	tenantID := uuid.New().String()
	tenant := &domain.Tenant{
		ID:   tenantID,
		Name: displayName + "'s Workspace",
	}
	if err := s.tenants.Create(ctx, tenant); err != nil {
		return nil, err
	}

	ut := &domain.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     "owner",
	}
	if err := s.tenants.AddUserToTenant(ctx, ut); err != nil {
		return nil, err
	}

	return &RegisterResult{UserID: userID, TenantID: tenantID}, nil
}

type LoginResult struct {
	Token string `json:"token"`
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	claims := jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: signed}, nil
}

type MeResult struct {
	UserID      string          `json:"user_id"`
	Email       string          `json:"email"`
	DisplayName string          `json:"display_name"`
	Tenants     []domain.Tenant `json:"tenants"`
}

func (s *AuthService) Me(ctx context.Context, userID string) (*MeResult, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tenants, err := s.tenants.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &MeResult{
		UserID:      user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Tenants:     tenants,
	}, nil
}
