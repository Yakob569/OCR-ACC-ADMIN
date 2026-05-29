package repositories

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dashboardRepository struct {
	db *pgxpool.Pool
}

func NewDashboardRepository(db *pgxpool.Pool) ports.DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetDashboard(ctx context.Context) (*domain.Dashboard, error) {
	stats, err := r.fetchStats(ctx)
	if err != nil {
		return nil, err
	}

	volume, err := r.fetchExtractionVolume(ctx)
	if err != nil {
		return nil, err
	}

	revenueMix, err := r.fetchRevenueMix(ctx)
	if err != nil {
		return nil, err
	}

	queue, err := r.fetchApprovalQueue(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.Dashboard{
		Stats:         stats,
		ExtractionVol: volume,
		RevenueMix:    revenueMix,
		ApprovalQueue: queue,
	}, nil
}

func (r *dashboardRepository) fetchStats(ctx context.Context) (domain.DashboardStats, error) {
	const q = `
		SELECT
			(SELECT COUNT(*)::bigint FROM receipt_images
			 WHERE ocr_status = 'completed'
			   AND COALESCE(processed_at, updated_at) >= NOW() - INTERVAL '24 hours'),
			(SELECT COUNT(*)::int FROM users WHERE role <> 'admin'),
			(SELECT COALESCE(SUM(p.amount), 0) FROM subscription_requests sr
			 JOIN pricing_plans p ON p.id = sr.plan_id
			 WHERE sr.status = 'approved'
			   AND sr.updated_at >= DATE_TRUNC('month', NOW())),
			(SELECT COALESCE(SUM(p.amount), 0) FROM subscription_requests sr
			 JOIN pricing_plans p ON p.id = sr.plan_id
			 WHERE sr.status = 'approved'),
			(SELECT COUNT(*)::int FROM subscription_requests WHERE status = 'pending'),
			(SELECT COUNT(*)::bigint FROM receipt_images
			 WHERE COALESCE(processed_at, updated_at) >= NOW() - INTERVAL '24 hours')
	`

	var stats domain.DashboardStats
	var ocr24h, total24h int64
	err := r.db.QueryRow(ctx, q).Scan(
		&ocr24h,
		&stats.ActiveMerchants,
		&stats.MonthlyRevenue,
		&stats.TotalRevenue,
		&stats.PendingApprovals,
		&total24h,
	)
	if err != nil {
		return domain.DashboardStats{}, fmt.Errorf("failed to fetch dashboard stats: %w", err)
	}

	stats.OcrExtractions24h = ocr24h
	if total24h > 0 {
		stats.OcrSuccessRate24h = math.Round((float64(ocr24h)/float64(total24h))*1000) / 10
	}

	return stats, nil
}

func (r *dashboardRepository) fetchExtractionVolume(ctx context.Context) ([]domain.ExtractionVolumePoint, error) {
	const q = `
		SELECT
			TO_CHAR(DATE(COALESCE(ri.processed_at, ri.updated_at)), 'Mon DD') AS day_label,
			COUNT(*)::bigint AS cnt
		FROM receipt_images ri
		WHERE ri.ocr_status = 'completed'
		  AND COALESCE(ri.processed_at, ri.updated_at) >= NOW() - INTERVAL '14 days'
		GROUP BY DATE(COALESCE(ri.processed_at, ri.updated_at))
		ORDER BY DATE(COALESCE(ri.processed_at, ri.updated_at)) ASC
	`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch extraction volume: %w", err)
	}
	defer rows.Close()

	var points []domain.ExtractionVolumePoint
	for rows.Next() {
		var p domain.ExtractionVolumePoint
		if err := rows.Scan(&p.Day, &p.Extractions); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if points == nil {
		for i := 13; i >= 0; i-- {
			d := time.Now().AddDate(0, 0, -i)
			points = append(points, domain.ExtractionVolumePoint{
				Day:         d.Format("Jan 02"),
				Extractions: 0,
			})
		}
	}
	return points, nil
}

func (r *dashboardRepository) fetchRevenueMix(ctx context.Context) ([]domain.RevenueMixItem, error) {
	const q = `
		SELECT
			p.pricing_plan_id,
			p.name,
			COUNT(sr.id)::int,
			COALESCE(SUM(p.amount), 0)
		FROM pricing_plans p
		LEFT JOIN subscription_requests sr ON sr.plan_id = p.id AND sr.status = 'approved'
		WHERE p.is_active = true
		GROUP BY p.id, p.pricing_plan_id, p.name
		ORDER BY COALESCE(SUM(p.amount), 0) DESC, p.name ASC
	`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch revenue mix: %w", err)
	}
	defer rows.Close()

	var items []domain.RevenueMixItem
	for rows.Next() {
		var item domain.RevenueMixItem
		if err := rows.Scan(&item.PlanID, &item.PlanName, &item.RequestCount, &item.Revenue); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []domain.RevenueMixItem{}
	}
	return items, rows.Err()
}

func (r *dashboardRepository) fetchApprovalQueue(ctx context.Context) ([]*domain.SubscriptionRequest, error) {
	query := subscriptionRequestSelect + `
		WHERE sr.status = 'pending'
		ORDER BY sr.created_at ASC
		LIMIT 8
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch approval queue: %w", err)
	}
	defer rows.Close()

	var requests []*domain.SubscriptionRequest
	for rows.Next() {
		req, err := scanSubscriptionRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if requests == nil {
		requests = []*domain.SubscriptionRequest{}
	}
	return requests, rows.Err()
}
