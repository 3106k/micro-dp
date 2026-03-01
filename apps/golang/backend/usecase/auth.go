package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/notification"
)

type AuthService struct {
	users       domain.UserRepository
	tenants     domain.TenantRepository
	jwtSecret   []byte
	emailSender notification.EmailSender
}

func NewAuthService(users domain.UserRepository, tenants domain.TenantRepository, jwtSecret string, emailSender notification.EmailSender) *AuthService {
	return &AuthService{
		users:       users,
		tenants:     tenants,
		jwtSecret:   []byte(jwtSecret),
		emailSender: emailSender,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (userID, tenantID string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	userID = uuid.New().String()
	user := &domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
		PlatformRole: domain.PlatformRoleUser,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return "", "", err
	}

	tenantID = uuid.New().String()
	tenant := &domain.Tenant{
		ID:       tenantID,
		Name:     displayName + "'s Workspace",
		IsActive: true,
	}
	if err := s.tenants.Create(ctx, tenant); err != nil {
		return "", "", err
	}

	ut := &domain.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     "owner",
	}
	if err := s.tenants.AddUserToTenant(ctx, ut); err != nil {
		return "", "", err
	}

	if s.emailSender != nil {
		subject, html, text, err := notification.RenderWelcome(displayName)
		if err == nil {
			if sendErr := s.emailSender.Send(ctx, &notification.EmailMessage{
				To: email, Subject: subject, HTML: html, Text: text,
			}); sendErr != nil {
				log.Printf("welcome email failed to=%s: %v", email, sendErr)
			}
		}
	}

	return userID, tenantID, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (token string, err error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	claims := jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, []domain.Tenant, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	tenants, err := s.tenants.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return user, tenants, nil
}
