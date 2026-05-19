package ports

import (
	"context"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/google/uuid"
)

type PaymentMethodRepository interface {
	Create(ctx context.Context, method *domain.PaymentMethod) (*domain.PaymentMethod, error)
	Update(ctx context.Context, method *domain.PaymentMethod) (*domain.PaymentMethod, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PaymentMethod, error)
	List(ctx context.Context) ([]*domain.PaymentMethod, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PaymentMethodService interface {
	CreatePaymentMethod(ctx context.Context, name, imageURL, accountNumber string) (*domain.PaymentMethod, error)
	UpdatePaymentMethod(ctx context.Context, id uuid.UUID, name, imageURL, accountNumber, status string) (*domain.PaymentMethod, error)
	ListPaymentMethods(ctx context.Context) ([]*domain.PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, id uuid.UUID) error
}
