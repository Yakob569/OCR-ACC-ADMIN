package services

import (
	"context"
	"errors"
	"testing"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/google/uuid"
)

type mockPlanRepository struct {
	plans map[uuid.UUID]*domain.PricingPlan
}

func newMockPlanRepository() *mockPlanRepository {
	return &mockPlanRepository{
		plans: make(map[uuid.UUID]*domain.PricingPlan),
	}
}

func (m *mockPlanRepository) Create(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error) {
	m.plans[plan.ID] = plan
	return plan, nil
}

func (m *mockPlanRepository) Update(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error) {
	m.plans[plan.ID] = plan
	return plan, nil
}

func (m *mockPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PricingPlan, error) {
	plan, exists := m.plans[id]
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

func (m *mockPlanRepository) ToggleStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	plan, exists := m.plans[id]
	if !exists {
		return errors.New("plan not found")
	}
	plan.IsActive = isActive
	return nil
}

func TestCreatePlan(t *testing.T) {
	repo := newMockPlanRepository()
	svc := NewPricingPlanService(repo)

	req := &domain.CreatePlanRequest{
		Name:         "Basic",
		Description:  "Basic plan",
		Amount:       9.99,
		DurationDays: 30,
		Features:     []string{"scan"},
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
}
