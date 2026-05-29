package ports

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
)

type DashboardRepository interface {
	GetDashboard(ctx context.Context) (*domain.Dashboard, error)
}

type DashboardService interface {
	GetDashboard(ctx context.Context) (*domain.Dashboard, error)
}
