package services

import (
	"context"
	"errors"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
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

	features := req.Features
	if features == nil {
		features = []string{}
	}

	businessID, err := domain.GenerateCustomID("PPL-")
	if err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	plan := &domain.PricingPlan{
		PricingPlanID: businessID,
		Name:          req.Name,
		Description:   req.Description,
		Amount:        req.Amount,
		DurationDays:  req.DurationDays,
		Status:        status,
		TrialDays:     req.TrialDays,
		TokenPerMonth: req.TokenPerMonth,
		OcrPerDay:     req.OcrPerDay,
		IsActive:      true,
		Features:      features,
	}

	return s.planRepo.Create(ctx, plan)
}

func (s *planService) UpdatePlan(ctx context.Context, planID string, req *domain.UpdatePlanRequest) (*domain.PricingPlan, error) {
	existing, err := s.planRepo.GetByBusinessID(ctx, planID)
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
	if req.TrialDays < 0 {
		return nil, errors.New("trial days must be greater than or equal to 0")
	}
	if req.TokenPerMonth < 0 {
		return nil, errors.New("token per month must be greater than or equal to 0")
	}
	if req.OcrPerDay < 0 {
		return nil, errors.New("ocr per day must be greater than or equal to 0")
	}

	features := req.Features
	if features == nil {
		features = []string{}
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Amount = req.Amount
	existing.DurationDays = req.DurationDays
	existing.Status = status
	existing.TrialDays = req.TrialDays
	existing.TokenPerMonth = req.TokenPerMonth
	existing.OcrPerDay = req.OcrPerDay
	existing.Features = features

	return s.planRepo.Update(ctx, existing)
}

func (s *planService) GetPlan(ctx context.Context, planID string) (*domain.PricingPlan, error) {
	if planID == "" {
		return nil, errors.New("plan ID is required")
	}
	return s.planRepo.GetByBusinessID(ctx, planID)
}

func (s *planService) ListPlans(ctx context.Context) ([]*domain.PricingPlan, error) {
	return s.planRepo.List(ctx)
}

func (s *planService) TogglePlanStatus(ctx context.Context, planID string, isActive bool) error {
	_, err := s.planRepo.GetByBusinessID(ctx, planID)
	if err != nil {
		return err
	}
	return s.planRepo.ToggleStatus(ctx, planID, isActive)
}
