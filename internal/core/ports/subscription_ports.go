package ports

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	ListRequests(ctx context.Context) ([]*domain.SubscriptionRequest, error)
	GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionRequest, error)
	UpdateRequestStatus(ctx context.Context, id uuid.UUID, status string, rejectionReason string) error
	CreateUserSubscription(ctx context.Context, sub *domain.UserSubscription) error
}

type SubscriptionService interface {
	ListSubscriptionRequests(ctx context.Context) ([]*domain.SubscriptionRequest, error)
	ApproveSubscriptionRequest(ctx context.Context, id uuid.UUID) error
	RejectSubscriptionRequest(ctx context.Context, id uuid.UUID, reason string) error
}
