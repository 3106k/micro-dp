package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
)

type ChartService struct {
	charts   domain.ChartRepository
	datasets domain.DatasetRepository
}

func NewChartService(charts domain.ChartRepository, datasets domain.DatasetRepository) *ChartService {
	return &ChartService{charts: charts, datasets: datasets}
}

func (s *ChartService) Create(ctx context.Context, name, chartType, datasetID, measure, dimension string, configJSON *string) (*domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify dataset exists and belongs to tenant
	if _, err := s.datasets.FindByID(ctx, tenantID, datasetID); err != nil {
		return nil, err
	}

	c := &domain.Chart{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Name:       name,
		ChartType:  chartType,
		DatasetID:  datasetID,
		Measure:    measure,
		Dimension:  dimension,
		ConfigJSON: configJSON,
	}
	if err := s.charts.Create(ctx, c); err != nil {
		return nil, err
	}
	return s.charts.FindByID(ctx, tenantID, c.ID)
}

func (s *ChartService) Get(ctx context.Context, id string) (*domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.charts.FindByID(ctx, tenantID, id)
}

func (s *ChartService) List(ctx context.Context) ([]domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.charts.ListByTenant(ctx, tenantID)
}

func (s *ChartService) Update(ctx context.Context, id, name, chartType, datasetID, measure, dimension string, configJSON *string) (*domain.Chart, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	c, err := s.charts.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Verify dataset exists if changed
	if datasetID != c.DatasetID {
		if _, err := s.datasets.FindByID(ctx, tenantID, datasetID); err != nil {
			return nil, err
		}
	}

	c.Name = name
	c.ChartType = chartType
	c.DatasetID = datasetID
	c.Measure = measure
	c.Dimension = dimension
	c.ConfigJSON = configJSON

	if err := s.charts.Update(ctx, c); err != nil {
		return nil, err
	}
	return s.charts.FindByID(ctx, tenantID, id)
}

func (s *ChartService) Delete(ctx context.Context, id string) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	if _, err := s.charts.FindByID(ctx, tenantID, id); err != nil {
		return err
	}
	return s.charts.Delete(ctx, tenantID, id)
}
