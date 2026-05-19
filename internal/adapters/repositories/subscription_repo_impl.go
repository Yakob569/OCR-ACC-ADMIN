package repositories

import (
	"context"
	"fmt"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type subscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) ports.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) ListRequests(ctx context.Context) ([]*domain.SubscriptionRequest, error) {
	query := `SELECT id, user_id, plan_id, payment_link, payment_screenshot, status, rejection_reason, created_at, updated_at 
	          FROM subscription_requests ORDER BY created_at DESC`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscription requests: %w", err)
	}
	defer rows.Close()

	var requests []*domain.SubscriptionRequest
	for rows.Next() {
		var req domain.SubscriptionRequest
		var rejReason *string
		var payLink *string
		var payScreen *string

		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.PlanID,
			&payLink,
			&payScreen,
			&req.Status,
			&rejReason,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription request: %w", err)
		}

		if payLink != nil {
			req.PaymentLink = *payLink
		}
		if payScreen != nil {
			req.PaymentScreenshot = *payScreen
		}
		if rejReason != nil {
			req.RejectionReason = *rejReason
		}

		requests = append(requests, &req)
	}

	return requests, nil
}

func (r *subscriptionRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionRequest, error) {
	query := `SELECT id, user_id, plan_id, payment_link, payment_screenshot, status, rejection_reason, created_at, updated_at 
	          FROM subscription_requests WHERE id = $1`
	
	var req domain.SubscriptionRequest
	var rejReason *string
	var payLink *string
	var payScreen *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.UserID,
		&req.PlanID,
		&payLink,
		&payScreen,
		&req.Status,
		&rejReason,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription request: %w", err)
	}

	if payLink != nil {
		req.PaymentLink = *payLink
	}
	if payScreen != nil {
		req.PaymentScreenshot = *payScreen
	}
	if rejReason != nil {
		req.RejectionReason = *rejReason
	}

	return &req, nil
}

func (r *subscriptionRepository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, status string, rejectionReason string) error {
	var err error
	if rejectionReason != "" {
		query := `UPDATE subscription_requests SET status = $1, rejection_reason = $2, updated_at = NOW() WHERE id = $3`
		_, err = r.db.Exec(ctx, query, status, rejectionReason, id)
	} else {
		query := `UPDATE subscription_requests SET status = $1, updated_at = NOW() WHERE id = $2`
		_, err = r.db.Exec(ctx, query, status, id)
	}
	if err != nil {
		return fmt.Errorf("failed to update subscription request status: %w", err)
	}
	return nil
}

func (r *subscriptionRepository) CreateUserSubscription(ctx context.Context, sub *domain.UserSubscription) error {
	// First deactivate any currently active subscriptions for this user
	deactivateQuery := `UPDATE user_subscriptions SET status = 'expired', updated_at = NOW() WHERE user_id = $1 AND status = 'active'`
	_, err := r.db.Exec(ctx, deactivateQuery, sub.UserID)
	if err != nil {
		return fmt.Errorf("failed to deactivate current subscriptions: %w", err)
	}

	// Insert the new subscription
	query := `INSERT INTO user_subscriptions (id, user_id, plan_id, status, start_date, end_date, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	
	_, err = r.db.Exec(ctx, query, sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.StartDate, sub.EndDate)
	if err != nil {
		return fmt.Errorf("failed to create user subscription: %w", err)
	}
	return nil
}
