package services

import (
	"context"
	"errors"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
)

type planService struct {
	planRepo ports.PricingPlanRepository
}

func NewPricingPlanService(planRepo ports.PricingPlanRepository) ports.PricingPlanService {
	return &planService{
		planRepo: planRepo,
	}
}

func (s *planService) CreatePlan(ctx context.Context, req *domain.CreatePlanRequest) (*domain.PricingPlan, error) {
	if req.Name == "" {
		return nil, errors.New("plan name is required")
	}
	if req.Amount < 0 {
		return nil, errors.New("amount must be greater than or equal to 0")
	}
	if req.DurationDays <= 0 {
		return nil, errors.New("duration days must be greater than 0")
	}

	features := req.Features
	if features == nil {
		features = []string{}
	}

	plan := &domain.PricingPlan{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		Amount:       req.Amount,
		DurationDays: req.DurationDays,
		IsActive:     true,
		Features:     features,
	}

	return s.planRepo.Create(ctx, plan)
}

func (s *planService) UpdatePlan(ctx context.Context, id uuid.UUID, req *domain.UpdatePlanRequest) (*domain.PricingPlan, error) {
	existing, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, errors.New("plan name is required")
	}
	if req.Amount < 0 {
		return nil, errors.New("amount must be greater than or equal to 0")
	}
	if req.DurationDays <= 0 {
		return nil, errors.New("duration days must be greater than 0")
	}

	features := req.Features
	if features == nil {
		features = []string{}
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Amount = req.Amount
	existing.DurationDays = req.DurationDays
	existing.Features = features

	return s.planRepo.Update(ctx, existing)
}

func (s *planService) GetPlan(ctx context.Context, id uuid.UUID) (*domain.PricingPlan, error) {
	return s.planRepo.GetByID(ctx, id)
}

func (s *planService) ListPlans(ctx context.Context) ([]*domain.PricingPlan, error) {
	return s.planRepo.List(ctx)
}

func (s *planService) TogglePlanStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	_, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.planRepo.ToggleStatus(ctx, id, isActive)
}
