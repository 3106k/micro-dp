package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/user/micro-dp/domain"
)

type DashboardService struct {
	dashboards domain.DashboardRepository
	widgets    domain.DashboardWidgetRepository
	charts     domain.ChartRepository
}

func NewDashboardService(
	dashboards domain.DashboardRepository,
	widgets domain.DashboardWidgetRepository,
	charts domain.ChartRepository,
) *DashboardService {
	return &DashboardService{
		dashboards: dashboards,
		widgets:    widgets,
		charts:     charts,
	}
}

func (s *DashboardService) Create(ctx context.Context, name string, description *string) (*domain.Dashboard, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	d := &domain.Dashboard{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
	}
	if err := s.dashboards.Create(ctx, d); err != nil {
		return nil, err
	}
	return s.dashboards.FindByID(ctx, tenantID, d.ID)
}

func (s *DashboardService) Get(ctx context.Context, id string) (*domain.Dashboard, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.dashboards.FindByID(ctx, tenantID, id)
}

func (s *DashboardService) List(ctx context.Context) ([]domain.Dashboard, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}
	return s.dashboards.ListByTenant(ctx, tenantID)
}

func (s *DashboardService) Update(ctx context.Context, id string, name string, description *string) (*domain.Dashboard, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	d, err := s.dashboards.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	d.Name = name
	d.Description = description

	if err := s.dashboards.Update(ctx, d); err != nil {
		return nil, err
	}
	return s.dashboards.FindByID(ctx, tenantID, id)
}

func (s *DashboardService) Delete(ctx context.Context, id string) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	if _, err := s.dashboards.FindByID(ctx, tenantID, id); err != nil {
		return err
	}
	return s.dashboards.Delete(ctx, tenantID, id)
}

func (s *DashboardService) CreateWidget(ctx context.Context, dashboardID, chartID string, position int) (*domain.DashboardWidget, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify dashboard exists and belongs to tenant
	if _, err := s.dashboards.FindByID(ctx, tenantID, dashboardID); err != nil {
		return nil, err
	}

	// Verify chart exists and belongs to same tenant
	if _, err := s.charts.FindByID(ctx, tenantID, chartID); err != nil {
		return nil, err
	}

	w := &domain.DashboardWidget{
		ID:          uuid.New().String(),
		DashboardID: dashboardID,
		ChartID:     chartID,
		Position:    position,
	}
	if err := s.widgets.Create(ctx, w); err != nil {
		return nil, err
	}
	return s.widgets.FindByID(ctx, dashboardID, w.ID)
}

func (s *DashboardService) ListWidgets(ctx context.Context, dashboardID string) ([]domain.DashboardWidget, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	// Verify dashboard exists and belongs to tenant
	if _, err := s.dashboards.FindByID(ctx, tenantID, dashboardID); err != nil {
		return nil, err
	}

	return s.widgets.ListByDashboard(ctx, dashboardID)
}

func (s *DashboardService) DeleteWidget(ctx context.Context, dashboardID, widgetID string) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	// Verify dashboard exists and belongs to tenant
	if _, err := s.dashboards.FindByID(ctx, tenantID, dashboardID); err != nil {
		return err
	}

	if _, err := s.widgets.FindByID(ctx, dashboardID, widgetID); err != nil {
		return err
	}
	return s.widgets.Delete(ctx, dashboardID, widgetID)
}
