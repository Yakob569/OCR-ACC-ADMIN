package ports

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
)

type TokenAnalyticsRepository interface {
	GetAnalytics(ctx context.Context) (*domain.TokenAnalytics, error)
}

type TokenAnalyticsService interface {
	GetAnalytics(ctx context.Context) (*domain.TokenAnalytics, error)
}
