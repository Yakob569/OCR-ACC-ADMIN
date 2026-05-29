package ports

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/google/uuid"
)

type PricingPlanRepository interface {
	Create(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error)
	Update(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error)
	GetByBusinessID(ctx context.Context, planID string) (*domain.PricingPlan, error)
	List(ctx context.Context) ([]*domain.PricingPlan, error)
	ToggleStatus(ctx context.Context, planID string, isActive bool) error
	HasActiveSubscriptions(ctx context.Context, planID string) (bool, error)
	GetByInternalID(ctx context.Context, id int) (*domain.PricingPlan, error)
}

type PricingPlanService interface {
	CreatePlan(ctx context.Context, req *domain.CreatePlanRequest) (*domain.PricingPlan, error)
	UpdatePlan(ctx context.Context, planID string, req *domain.UpdatePlanRequest) (*domain.PricingPlan, error)
	GetPlan(ctx context.Context, planID string) (*domain.PricingPlan, error)
	ListPlans(ctx context.Context) ([]*domain.PricingPlan, error)
	TogglePlanStatus(ctx context.Context, planID string, isActive bool) error
}

type AuthService interface {
	ValidateToken(token string) (uuid.UUID, string, error) // Returns userID, role, error
	ValidateRefreshToken(token string) (uuid.UUID, string, error)
	GenerateTokenPair(userID uuid.UUID, role string) (string, string, error)
}
