package services

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
)

type merchantService struct {
	repo ports.MerchantRepository
}

func NewMerchantService(repo ports.MerchantRepository) ports.MerchantService {
	return &merchantService{repo: repo}
}

func (s *merchantService) ListMerchants(ctx context.Context) ([]*domain.MerchantSummary, error) {
	merchants, err := s.repo.ListMerchants(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range merchants {
		m.Status = "active"
	}
	if merchants == nil {
		merchants = []*domain.MerchantSummary{}
	}
	return merchants, nil
}
