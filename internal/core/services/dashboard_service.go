package services

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
)

type dashboardService struct {
	repo ports.DashboardRepository
}

func NewDashboardService(repo ports.DashboardRepository) ports.DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetDashboard(ctx context.Context) (*domain.Dashboard, error) {
	return s.repo.GetDashboard(ctx)
}
