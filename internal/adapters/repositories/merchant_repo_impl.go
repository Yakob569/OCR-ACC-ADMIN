package repositories

import (
	"context"
	"fmt"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type merchantRepository struct {
	db *pgxpool.Pool
}

func NewMerchantRepository(db *pgxpool.Pool) ports.MerchantRepository {
	return &merchantRepository{db: db}
}

func (r *merchantRepository) ListMerchants(ctx context.Context) ([]*domain.MerchantSummary, error) {
	query := `
		SELECT
			u.id,
			COALESCE(u.user_id, ''),
			u.email,
			u.full_name,
			COALESCE(u.company_name, ''),
			u.created_at,
			COALESCE(p.pricing_plan_id, ''),
			COALESCE(p.name, ''),
			COALESCE(ocr.cnt, 0),
			COALESCE(rev.total, 0),
			act.last_active
		FROM users u
		LEFT JOIN LATERAL (
			SELECT us.plan_id
			FROM user_subscriptions us
			WHERE us.user_id = u.id
			  AND us.status = 'active'
			  AND us.end_date > NOW()
			ORDER BY us.start_date DESC
			LIMIT 1
		) active_sub ON true
		LEFT JOIN pricing_plans p ON p.id = active_sub.plan_id
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int AS cnt
			FROM receipt_images ri
			WHERE ri.user_id = u.id AND ri.ocr_status = 'completed'
		) ocr ON true
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(pp.amount), 0)::float8 AS total
			FROM subscription_requests sr
			JOIN pricing_plans pp ON pp.id = sr.plan_id
			WHERE sr.user_id = u.id AND sr.status = 'approved'
		) rev ON true
		LEFT JOIN LATERAL (
			SELECT GREATEST(
				u.updated_at,
				COALESCE((SELECT MAX(ri.updated_at) FROM receipt_images ri WHERE ri.user_id = u.id), u.updated_at)
			) AS last_active
		) act ON true
		WHERE u.role <> 'admin'
		ORDER BY act.last_active DESC NULLS LAST, u.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list merchants: %w", err)
	}
	defer rows.Close()

	var merchants []*domain.MerchantSummary
	for rows.Next() {
		var m domain.MerchantSummary
		if err := rows.Scan(
			&m.ID,
			&m.PublicID,
			&m.Email,
			&m.FullName,
			&m.CompanyName,
			&m.CreatedAt,
			&m.CurrentPlanID,
			&m.CurrentPlanName,
			&m.OcrSuccessCount,
			&m.TotalRevenue,
			&m.LastActiveAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan merchant: %w", err)
		}
		merchants = append(merchants, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if merchants == nil {
		merchants = []*domain.MerchantSummary{}
	}
	return merchants, nil
}
