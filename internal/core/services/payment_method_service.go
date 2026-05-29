package services

import (
	"context"
	"errors"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
)

type paymentMethodService struct {
	methodRepo ports.PaymentMethodRepository
}

func NewPaymentMethodService(methodRepo ports.PaymentMethodRepository) ports.PaymentMethodService {
	return &paymentMethodService{
		methodRepo: methodRepo,
	}
}

func (s *paymentMethodService) CreatePaymentMethod(ctx context.Context, name, imageURL, accountNumber, status string) (*domain.PaymentMethod, error) {
	if name == "" {
		return nil, errors.New("payment method name is required")
	}
	if accountNumber == "" {
		return nil, errors.New("account number is required")
	}
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		return nil, errors.New("invalid status: must be 'active' or 'inactive'")
	}

	method := &domain.PaymentMethod{
		ID:            uuid.New(),
		Name:          name,
		ImageURL:      imageURL,
		AccountNumber: accountNumber,
		Status:        status,
	}

	return s.methodRepo.Create(ctx, method)
}

func (s *paymentMethodService) UpdatePaymentMethod(ctx context.Context, id uuid.UUID, name, imageURL, accountNumber, status string) (*domain.PaymentMethod, error) {
	existing, err := s.methodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, errors.New("payment method name is required")
	}
	if accountNumber == "" {
		return nil, errors.New("account number is required")
	}
	if status != "active" && status != "inactive" {
		return nil, errors.New("invalid status: must be 'active' or 'inactive'")
	}

	existing.Name = name
	if imageURL != "" {
		existing.ImageURL = imageURL
	}
	existing.AccountNumber = accountNumber
	existing.Status = status

	return s.methodRepo.Update(ctx, existing)
}

func (s *paymentMethodService) ListPaymentMethods(ctx context.Context) ([]*domain.PaymentMethod, error) {
	return s.methodRepo.List(ctx)
}

func (s *paymentMethodService) DeletePaymentMethod(ctx context.Context, id uuid.UUID) error {
	_, err := s.methodRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.methodRepo.Delete(ctx, id)
}
