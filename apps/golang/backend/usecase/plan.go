package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/edition"
)

type PlanService struct {
	plans       domain.PlanRepository
	tenantPlans domain.TenantPlanRepository
	usage       domain.UsageRepository
}

func NewPlanService(plans domain.PlanRepository, tenantPlans domain.TenantPlanRepository, usage domain.UsageRepository) *PlanService {
	return &PlanService{plans: plans, tenantPlans: tenantPlans, usage: usage}
}

// GetTenantPlan returns the plan for the tenant. Falls back to default plan if not assigned.
func (s *PlanService) GetTenantPlan(ctx context.Context) (*domain.Plan, *domain.TenantPlan, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("tenant id not found in context")
	}

	tp, err := s.tenantPlans.FindByTenantID(ctx, tenantID)
	if err != nil && !errors.Is(err, domain.ErrTenantPlanNotFound) {
		return nil, nil, err
	}

	if tp != nil {
		plan, err := s.plans.FindByID(ctx, tp.PlanID)
		if err != nil {
			return nil, nil, err
		}
		return plan, tp, nil
	}

	// No explicit assignment — use default plan
	plan, err := s.plans.FindDefault(ctx)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	defaultTP := &domain.TenantPlan{
		TenantID:  tenantID,
		PlanID:    plan.ID,
		StartedAt: now,
	}
	return plan, defaultTP, nil
}

// GetUsageSummary returns today's usage + current plan.
func (s *PlanService) GetUsageSummary(ctx context.Context) (*domain.UsageSummary, error) {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant id not found in context")
	}

	date := today()
	daily, err := s.usage.FindDailyByTenantAndDate(ctx, tenantID, date)
	if err != nil {
		return nil, err
	}

	plan, _, err := s.GetTenantPlan(ctx)
	if err != nil && !errors.Is(err, domain.ErrPlanNotFound) {
		return nil, err
	}

	summary := &domain.UsageSummary{
		TenantID: tenantID,
		Date:     date,
		Plan:     plan,
	}
	if daily != nil {
		summary.EventsCount = daily.EventsCount
		summary.StorageBytes = daily.StorageBytes
		summary.RowsCount = daily.RowsCount
		summary.UploadsCount = daily.UploadsCount
	}
	return summary, nil
}

// CheckEventsQuota checks if the tenant is within the events quota.
// OSS edition always returns nil (no quota).
func (s *PlanService) CheckEventsQuota(ctx context.Context) error {
	if edition.IsOSS() {
		return nil
	}
	return s.checkQuota(ctx, func(plan *domain.Plan, daily *domain.UsageDaily) bool {
		if plan.MaxEventsPerDay < 0 {
			return false // unlimited
		}
		current := 0
		if daily != nil {
			current = daily.EventsCount
		}
		return current >= plan.MaxEventsPerDay
	})
}

// CheckUploadQuota checks if the tenant is within the upload quota.
// OSS edition always returns nil (no quota).
func (s *PlanService) CheckUploadQuota(ctx context.Context) error {
	if edition.IsOSS() {
		return nil
	}
	return s.checkQuota(ctx, func(plan *domain.Plan, daily *domain.UsageDaily) bool {
		if plan.MaxUploadsPerDay < 0 {
			return false // unlimited
		}
		current := 0
		if daily != nil {
			current = daily.UploadsCount
		}
		return current >= plan.MaxUploadsPerDay
	})
}

func (s *PlanService) checkQuota(ctx context.Context, exceeded func(*domain.Plan, *domain.UsageDaily) bool) error {
	tenantID, ok := domain.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant id not found in context")
	}

	plan, _, err := s.GetTenantPlan(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return nil // no plan = no quota
		}
		return err
	}

	date := today()
	daily, err := s.usage.FindDailyByTenantAndDate(ctx, tenantID, date)
	if err != nil {
		return err
	}

	if exceeded(plan, daily) {
		return domain.ErrQuotaExceeded
	}
	return nil
}

// --- Admin operations ---

func (s *PlanService) CreatePlan(ctx context.Context, name, displayName string, maxEvents, maxRows, maxUploads int, maxStorage int64) (*domain.Plan, error) {
	p := &domain.Plan{
		ID:               uuid.New().String(),
		Name:             name,
		DisplayName:      displayName,
		MaxEventsPerDay:  maxEvents,
		MaxStorageBytes:  maxStorage,
		MaxRowsPerDay:    maxRows,
		MaxUploadsPerDay: maxUploads,
	}
	if err := s.plans.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlanService) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	return s.plans.ListAll(ctx)
}

func (s *PlanService) UpdatePlan(ctx context.Context, id string, displayName *string, maxEvents, maxRows, maxUploads *int, maxStorage *int64) (*domain.Plan, error) {
	plan, err := s.plans.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if displayName != nil {
		plan.DisplayName = *displayName
	}
	if maxEvents != nil {
		plan.MaxEventsPerDay = *maxEvents
	}
	if maxStorage != nil {
		plan.MaxStorageBytes = *maxStorage
	}
	if maxRows != nil {
		plan.MaxRowsPerDay = *maxRows
	}
	if maxUploads != nil {
		plan.MaxUploadsPerDay = *maxUploads
	}
	if err := s.plans.Update(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanService) AssignPlan(ctx context.Context, tenantID, planID string) (*domain.Plan, *domain.TenantPlan, error) {
	plan, err := s.plans.FindByID(ctx, planID)
	if err != nil {
		return nil, nil, err
	}

	tp := &domain.TenantPlan{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		PlanID:    planID,
		StartedAt: time.Now().UTC(),
	}
	if err := s.tenantPlans.Upsert(ctx, tp); err != nil {
		return nil, nil, err
	}
	return plan, tp, nil
}
