package services

import (
	"context"
	"errors"
	"testing"

	"github.com/cashflow/admin-service/internal/core/domain"
)

type mockPlanRepository struct {
	plans      map[string]*domain.PricingPlan
	activeSubs map[string]bool
}

func newMockPlanRepository() *mockPlanRepository {
	return &mockPlanRepository{
		plans:      make(map[string]*domain.PricingPlan),
		activeSubs: make(map[string]bool),
	}
}

func (m *mockPlanRepository) Create(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error) {
	m.plans[plan.PricingPlanID] = plan
	return plan, nil
}

func (m *mockPlanRepository) Update(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error) {
	m.plans[plan.PricingPlanID] = plan
	return plan, nil
}

func (m *mockPlanRepository) GetByBusinessID(ctx context.Context, planID string) (*domain.PricingPlan, error) {
	plan, exists := m.plans[planID]
	if !exists {
		return nil, errors.New("plan not found")
	}
	return plan, nil
}

func (m *mockPlanRepository) List(ctx context.Context) ([]*domain.PricingPlan, error) {
	var list []*domain.PricingPlan
	for _, p := range m.plans {
		list = append(list, p)
	}
	return list, nil
}

func (m *mockPlanRepository) ToggleStatus(ctx context.Context, planID string, isActive bool) error {
	plan, exists := m.plans[planID]
	if !exists {
		return errors.New("plan not found")
	}
	plan.IsActive = isActive
	return nil
}

func (m *mockPlanRepository) GetByInternalID(ctx context.Context, id int) (*domain.PricingPlan, error) {
	for _, p := range m.plans {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errors.New("plan not found")
}

func (m *mockPlanRepository) HasActiveSubscriptions(ctx context.Context, planID string) (bool, error) {
	return m.activeSubs[planID], nil
}

func (m *mockPlanRepository) ClearDefaultPlan(ctx context.Context) error {
	for _, p := range m.plans {
		p.IsDefault = false
	}
	return nil
}

func TestCreatePlan(t *testing.T) {
	repo := newMockPlanRepository()
	svc := NewPricingPlanService(repo)

	req := &domain.CreatePlanRequest{
		Name:          "Basic",
		Description:   "Basic plan",
		Amount:        9.99,
		DurationDays:  30,
		TrialDays:     7,
		TokenPerMonth: 5000,
		OcrPerDay:     100,
		Features:      []string{"OCR Scanning", "API Access"},
	}

	plan, err := svc.CreatePlan(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Name != "Basic" {
		t.Errorf("expected plan name to be Basic, got %s", plan.Name)
	}
	if plan.Amount != 9.99 {
		t.Errorf("expected plan amount to be 9.99, got %f", plan.Amount)
	}
	if plan.DurationDays != 30 {
		t.Errorf("expected plan duration to be 30, got %d", plan.DurationDays)
	}
	if plan.TrialDays != 7 {
		t.Errorf("expected trial days to be 7, got %d", plan.TrialDays)
	}
	if plan.Status != "active" {
		t.Errorf("expected status to be active, got %s", plan.Status)
	}
	if len(plan.PricingPlanID) == 0 {
		t.Error("expected a generated pricing_plan_id (PPL-XXXXXXX)")
	}
}

func TestCreatePlan_Validation(t *testing.T) {
	repo := newMockPlanRepository()
	svc := NewPricingPlanService(repo)

	// Test missing name
	_, err := svc.CreatePlan(context.Background(), &domain.CreatePlanRequest{
		Amount:       9.99,
		DurationDays: 30,
	})
	if err == nil || err.Error() != "plan name is required" {
		t.Errorf("expected 'plan name is required' error, got: %v", err)
	}

	// Test invalid duration
	_, err = svc.CreatePlan(context.Background(), &domain.CreatePlanRequest{
		Name:         "Test",
		Amount:       9.99,
		DurationDays: 0,
	})
	if err == nil || err.Error() != "duration days must be greater than 0" {
		t.Errorf("expected 'duration days must be greater than 0' error, got: %v", err)
	}

	// Test negative trial days
	_, err = svc.CreatePlan(context.Background(), &domain.CreatePlanRequest{
		Name:         "Test",
		Amount:       9.99,
		DurationDays: 30,
		TrialDays:    -1,
	})
	if err == nil || err.Error() != "trial days must be greater than or equal to 0" {
		t.Errorf("expected 'trial days must be greater than or equal to 0' error, got: %v", err)
	}
}
