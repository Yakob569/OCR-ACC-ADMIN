package repositories

import (
	"context"
	"fmt"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type subscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) ports.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

const subscriptionRequestSelect = `
	SELECT sr.id, sr.request_id, sr.user_id, sr.plan_id,
	       p.pricing_plan_id, p.name, p.amount, p.duration_days,
	       u.email, u.full_name, COALESCE(u.company_name, ''),
	       sr.payment_method_id, COALESCE(pm.name, ''), COALESCE(pm.account_number, ''),
	       sr.payment_link, sr.payment_screenshot, sr.status, sr.rejection_reason,
	       sr.created_at, sr.updated_at
	FROM subscription_requests sr
	JOIN pricing_plans p ON p.id = sr.plan_id
	JOIN users u ON u.id = sr.user_id
	LEFT JOIN payment_methods pm ON pm.id = sr.payment_method_id
`

func scanSubscriptionRequest(
	rows interface {
		Scan(dest ...any) error
	},
) (*domain.SubscriptionRequest, error) {
	var req domain.SubscriptionRequest
	var rejReason *string
	var payLink *string
	var payScreen *string

	err := rows.Scan(
		&req.ID,
		&req.RequestID,
		&req.UserID,
		&req.PlanID,
		&req.PlanPublicID,
		&req.PlanName,
		&req.PlanAmount,
		&req.PlanDurationDays,
		&req.UserEmail,
		&req.UserFullName,
		&req.CompanyName,
		&req.PaymentMethodID,
		&req.PaymentMethodName,
		&req.PaymentMethodAccount,
		&payLink,
		&payScreen,
		&req.Status,
		&rejReason,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, err
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

func (r *subscriptionRepository) CountPendingRequests(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM subscription_requests WHERE status = 'pending'`,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending subscription requests: %w", err)
	}
	return count, nil
}

func (r *subscriptionRepository) ListRequests(ctx context.Context) ([]*domain.SubscriptionRequest, error) {
	query := subscriptionRequestSelect + ` ORDER BY sr.created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscription requests: %w", err)
	}
	defer rows.Close()

	var requests []*domain.SubscriptionRequest
	for rows.Next() {
		req, err := scanSubscriptionRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription request: %w", err)
		}
		requests = append(requests, req)
	}

	if requests == nil {
		requests = []*domain.SubscriptionRequest{}
	}

	return requests, nil
}

func (r *subscriptionRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionRequest, error) {
	query := subscriptionRequestSelect + ` WHERE sr.id = $1`

	row := r.db.QueryRow(ctx, query, id)
	req, err := scanSubscriptionRequest(row)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription request: %w", err)
	}
	return req, nil
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
	deactivateQuery := `UPDATE user_subscriptions SET status = 'expired', updated_at = NOW() WHERE user_id = $1 AND status = 'active'`
	if _, err := r.db.Exec(ctx, deactivateQuery, sub.UserID); err != nil {
		return fmt.Errorf("failed to deactivate current subscriptions: %w", err)
	}

	query := `INSERT INTO user_subscriptions (id, user_id, plan_id, status, start_date, end_date, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

	_, err := r.db.Exec(ctx, query, sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.StartDate, sub.EndDate)
	if err != nil {
		return fmt.Errorf("failed to create user subscription: %w", err)
	}
	return nil
}
