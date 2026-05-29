package ports

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
)

type MerchantRepository interface {
	ListMerchants(ctx context.Context) ([]*domain.MerchantSummary, error)
}

type MerchantService interface {
	ListMerchants(ctx context.Context) ([]*domain.MerchantSummary, error)
}
