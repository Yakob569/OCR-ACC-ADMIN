package services

import (
	"context"
	"errors"
	"time"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
)

type subscriptionService struct {
	subRepo  ports.SubscriptionRepository
	planRepo ports.PricingPlanRepository
}

func NewSubscriptionService(subRepo ports.SubscriptionRepository, planRepo ports.PricingPlanRepository) ports.SubscriptionService {
	return &subscriptionService{
		subRepo:  subRepo,
		planRepo: planRepo,
	}
}

func (s *subscriptionService) ListSubscriptionRequests(ctx context.Context) ([]*domain.SubscriptionRequest, error) {
	return s.subRepo.ListRequests(ctx)
}

func (s *subscriptionService) ApproveSubscriptionRequest(ctx context.Context, id uuid.UUID) error {
	req, err := s.subRepo.GetRequestByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Status != "pending" {
		return errors.New("request is not in pending status")
	}

	plan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return err
	}

	// Update status
	err = s.subRepo.UpdateRequestStatus(ctx, id, "approved", "")
	if err != nil {
		return err
	}

	// Create user subscription record
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, plan.DurationDays)

	sub := &domain.UserSubscription{
		ID:        uuid.New(),
		UserID:    req.UserID,
		PlanID:    req.PlanID,
		Status:    "active",
		StartDate: startDate,
		EndDate:   endDate,
	}

	return s.subRepo.CreateUserSubscription(ctx, sub)
}

func (s *subscriptionService) RejectSubscriptionRequest(ctx context.Context, id uuid.UUID, reason string) error {
	req, err := s.subRepo.GetRequestByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Status != "pending" {
		return errors.New("request is not in pending status")
	}

	if reason == "" {
		return errors.New("rejection reason is required")
	}

	return s.subRepo.UpdateRequestStatus(ctx, id, "rejected", reason)
}
