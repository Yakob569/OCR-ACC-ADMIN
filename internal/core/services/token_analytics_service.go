package services

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
)

type tokenAnalyticsService struct {
	repo ports.TokenAnalyticsRepository
}

func NewTokenAnalyticsService(repo ports.TokenAnalyticsRepository) ports.TokenAnalyticsService {
	return &tokenAnalyticsService{repo: repo}
}

func (s *tokenAnalyticsService) GetAnalytics(ctx context.Context) (*domain.TokenAnalytics, error) {
	return s.repo.GetAnalytics(ctx)
}
