package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
)

type WriteKeyService struct {
	repo    domain.WriteKeyRepository
	tenants domain.TenantRepository
}

func NewWriteKeyService(repo domain.WriteKeyRepository, tenants domain.TenantRepository) *WriteKeyService {
	return &WriteKeyService{repo: repo, tenants: tenants}
}

func (s *WriteKeyService) checkOwnerOrAdmin(ctx context.Context) error {
	userID, ok := domain.UserIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("user id not found in context")
	}
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}
	role, err := s.tenants.GetUserRole(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	if role != domain.TenantRoleOwner && role != domain.TenantRoleAdmin {
		return domain.ErrInsufficientRole
	}
	return nil
}

// Generate creates a new write key and returns (WriteKey, rawKey, error).
func (s *WriteKeyService) Generate(ctx context.Context, name string) (*domain.WriteKey, string, error) {
	if err := s.checkOwnerOrAdmin(ctx); err != nil {
		return nil, "", err
	}

	tenantID, _ := domain.TenantIDFromContext(ctx)

	rawKey, err := generateRawKey()
	if err != nil {
		return nil, "", fmt.Errorf("generate key: %w", err)
	}

	keyHash := hashKey(rawKey)
	keyPrefix := rawKey[:12]

	wk := &domain.WriteKey{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		IsActive:  true,
	}

	if err := s.repo.Create(ctx, wk); err != nil {
		return nil, "", fmt.Errorf("create write key: %w", err)
	}

	return wk, rawKey, nil
}

// Regenerate replaces the key hash for an existing write key.
func (s *WriteKeyService) Regenerate(ctx context.Context, id string) (*domain.WriteKey, string, error) {
	if err := s.checkOwnerOrAdmin(ctx); err != nil {
		return nil, "", err
	}

	tenantID, _ := domain.TenantIDFromContext(ctx)

	if _, err := s.repo.FindByID(ctx, tenantID, id); err != nil {
		return nil, "", err
	}

	rawKey, err := generateRawKey()
	if err != nil {
		return nil, "", fmt.Errorf("generate key: %w", err)
	}

	keyHash := hashKey(rawKey)
	keyPrefix := rawKey[:12]

	if err := s.repo.UpdateKeyHash(ctx, tenantID, id, keyHash, keyPrefix); err != nil {
		return nil, "", fmt.Errorf("update key hash: %w", err)
	}

	wk, err := s.repo.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, "", err
	}

	return wk, rawKey, nil
}

func (s *WriteKeyService) List(ctx context.Context) ([]domain.WriteKey, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.repo.FindByTenantID(ctx, tenantID)
}

func (s *WriteKeyService) Delete(ctx context.Context, id string) error {
	if err := s.checkOwnerOrAdmin(ctx); err != nil {
		return err
	}
	tenantID, _ := domain.TenantIDFromContext(ctx)
	return s.repo.Delete(ctx, tenantID, id)
}

// Authenticate validates a raw write key and returns the associated WriteKey.
func (s *WriteKeyService) Authenticate(ctx context.Context, rawKey string) (*domain.WriteKey, error) {
	keyHash := hashKey(rawKey)
	wk, err := s.repo.FindByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, domain.ErrWriteKeyInvalid
	}
	if !wk.IsActive {
		return nil, domain.ErrWriteKeyInvalid
	}
	return wk, nil
}

// generateRawKey produces "dk_live_" + 32 random hex characters.
func generateRawKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "dk_live_" + hex.EncodeToString(b), nil
}

func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
