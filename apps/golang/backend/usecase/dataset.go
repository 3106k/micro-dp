package usecase

import (
	"context"
	"fmt"

	"github.com/user/micro-dp/domain"
)

type DatasetService struct {
	datasets domain.DatasetRepository
}

func NewDatasetService(datasets domain.DatasetRepository) *DatasetService {
	return &DatasetService{datasets: datasets}
}

func (s *DatasetService) Get(ctx context.Context, id string) (*domain.Dataset, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.datasets.FindByID(ctx, tenantID, id)
}

func (s *DatasetService) List(ctx context.Context, filter domain.DatasetListFilter) ([]domain.Dataset, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.datasets.ListByTenant(ctx, tenantID, filter)
}
