package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
)

type ConnectionService struct {
	connections domain.ConnectionRepository
}

func NewConnectionService(connections domain.ConnectionRepository) *ConnectionService {
	return &ConnectionService{connections: connections}
}

func (s *ConnectionService) Create(ctx context.Context, name, connType, configJSON string, secretRef *string) (*domain.Connection, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	c := &domain.Connection{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Name:       name,
		Type:       connType,
		ConfigJSON: configJSON,
		SecretRef:  secretRef,
	}

	if err := s.connections.Create(ctx, c); err != nil {
		return nil, err
	}

	return s.connections.FindByID(ctx, tenantID, c.ID)
}

func (s *ConnectionService) Get(ctx context.Context, id string) (*domain.Connection, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.connections.FindByID(ctx, tenantID, id)
}

func (s *ConnectionService) List(ctx context.Context) ([]domain.Connection, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.connections.ListByTenant(ctx, tenantID)
}

func (s *ConnectionService) Update(ctx context.Context, id, name, connType, configJSON string, secretRef *string) (*domain.Connection, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	c, err := s.connections.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	c.Name = name
	c.Type = connType
	c.ConfigJSON = configJSON
	c.SecretRef = secretRef

	if err := s.connections.Update(ctx, c); err != nil {
		return nil, err
	}

	return s.connections.FindByID(ctx, tenantID, id)
}

func (s *ConnectionService) Delete(ctx context.Context, id string) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	// Verify exists
	if _, err := s.connections.FindByID(ctx, tenantID, id); err != nil {
		return err
	}

	return s.connections.Delete(ctx, tenantID, id)
}
