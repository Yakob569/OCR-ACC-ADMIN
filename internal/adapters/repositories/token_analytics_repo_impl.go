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

type tokenAnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewTokenAnalyticsRepository(db *pgxpool.Pool) ports.TokenAnalyticsRepository {
	return &tokenAnalyticsRepository{db: db}
}

func pctDelta(current, previous float64) float64 {
	if previous <= 0 {
		if current <= 0 {
			return 0
		}
		return 100
	}
	return math.Round(((current-previous)/previous)*1000) / 10
}

func usagePct(used int64, limit float64) float64 {
	if limit <= 0 {
		return 0
	}
	return math.Round((float64(used)/limit)*1000) / 10
}

func (r *tokenAnalyticsRepository) GetAnalytics(ctx context.Context) (*domain.TokenAnalytics, error) {
	summary, err := r.fetchSummary(ctx)
	if err != nil {
		return nil, err
	}

	series, err := r.fetchSeries14d(ctx)
	if err != nil {
		return nil, err
	}

	merchants, err := r.fetchMerchantUsage(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.TokenAnalytics{
		Summary:   summary,
		Series14d: series,
		Merchants: merchants,
	}, nil
}

func (r *tokenAnalyticsRepository) fetchSummary(ctx context.Context) (domain.TokenAnalyticsSummary, error) {
	const q = `
		WITH merchant_count AS (
			SELECT COUNT(*)::int AS cnt FROM users WHERE role <> 'admin'
		),
		tokens_24h AS (
			SELECT COALESCE(SUM(otl.total_tokens), 0)::bigint AS cur
			FROM ocr_token_logs otl
			WHERE otl.created_at >= NOW() - INTERVAL '24 hours'
		),
		tokens_prev_24h AS (
			SELECT COALESCE(SUM(otl.total_tokens), 0)::bigint AS prev
			FROM ocr_token_logs otl
			WHERE otl.created_at >= NOW() - INTERVAL '48 hours'
			  AND otl.created_at < NOW() - INTERVAL '24 hours'
		),
		tokens_month AS (
			SELECT COALESCE(SUM(otl.total_tokens), 0)::bigint AS cur
			FROM ocr_token_logs otl
			WHERE otl.created_at >= DATE_TRUNC('month', NOW())
		),
		tokens_prev_month AS (
			SELECT COALESCE(SUM(otl.total_tokens), 0)::bigint AS prev
			FROM ocr_token_logs otl
			WHERE otl.created_at >= DATE_TRUNC('month', NOW()) - INTERVAL '1 month'
			  AND otl.created_at < DATE_TRUNC('month', NOW())
		)
		SELECT
			t24.cur,
			tp24.prev,
			tm.cur,
			tpm.prev,
			mc.cnt
		FROM tokens_24h t24, tokens_prev_24h tp24, tokens_month tm, tokens_prev_month tpm, merchant_count mc
	`

	var tokens24h, prev24h, tokensMonth, prevMonth int64
	var merchantCnt int
	err := r.db.QueryRow(ctx, q).Scan(&tokens24h, &prev24h, &tokensMonth, &prevMonth, &merchantCnt)
	if err != nil {
		return domain.TokenAnalyticsSummary{}, fmt.Errorf("failed to fetch token summary: %w", err)
	}

	avg24h := float64(0)
	prevAvg24h := float64(0)
	if merchantCnt > 0 {
		avg24h = float64(tokens24h) / float64(merchantCnt)
		prevAvg24h = float64(prev24h) / float64(merchantCnt)
	}

	return domain.TokenAnalyticsSummary{
		Tokens24h:              tokens24h,
		Tokens24hDeltaPct:      pctDelta(float64(tokens24h), float64(prev24h)),
		AvgPerMerchant24h:      math.Round(avg24h*10) / 10,
		AvgPerMerchantDeltaPct: pctDelta(avg24h, prevAvg24h),
		TokensMonth:            tokensMonth,
		TokensMonthDeltaPct:    pctDelta(float64(tokensMonth), float64(prevMonth)),
		ActiveMerchants:        merchantCnt,
	}, nil
}

func (r *tokenAnalyticsRepository) fetchSeries14d(ctx context.Context) ([]domain.TokenDailyPoint, error) {
	const q = `
		SELECT
			TO_CHAR(DATE(otl.created_at), 'Mon DD') AS day_label,
			COALESCE(SUM(otl.total_tokens), 0)::bigint AS tokens
		FROM ocr_token_logs otl
		WHERE otl.created_at >= NOW() - INTERVAL '14 days'
		GROUP BY DATE(otl.created_at)
		ORDER BY DATE(otl.created_at) ASC
	`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token series: %w", err)
	}
	defer rows.Close()

	var points []domain.TokenDailyPoint
	for rows.Next() {
		var p domain.TokenDailyPoint
		if err := rows.Scan(&p.Day, &p.Tokens); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if points == nil {
		// Fill last 14 days with zeros for stable chart
		for i := 13; i >= 0; i-- {
			d := time.Now().AddDate(0, 0, -i)
			points = append(points, domain.TokenDailyPoint{
				Day:    d.Format("Jan 02"),
				Tokens: 0,
			})
		}
	}
	return points, rows.Err()
}

func (r *tokenAnalyticsRepository) fetchMerchantUsage(ctx context.Context) ([]domain.MerchantTokenUsage, error) {
	const q = `
		SELECT
			u.id,
			COALESCE(u.user_id, ''),
			u.full_name,
			COALESCE(u.company_name, ''),
			COALESCE(p.name, ''),
			COALESCE(p.token_per_month, 0),
			COALESCE(cur.tokens_used, 0),
			COALESCE(prev.tokens_used, 0)
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
			SELECT COALESCE(SUM(otl.total_tokens), 0)::bigint AS tokens_used
			FROM ocr_token_logs otl
			JOIN receipt_images ri ON ri.id = otl.receipt_image_id
			WHERE ri.user_id = u.id
			  AND otl.created_at >= DATE_TRUNC('month', NOW())
		) cur ON true
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(otl.total_tokens), 0)::bigint AS tokens_used
			FROM ocr_token_logs otl
			JOIN receipt_images ri ON ri.id = otl.receipt_image_id
			WHERE ri.user_id = u.id
			  AND otl.created_at >= DATE_TRUNC('month', NOW()) - INTERVAL '1 month'
			  AND otl.created_at < DATE_TRUNC('month', NOW())
		) prev ON true
		WHERE u.role <> 'admin'
		ORDER BY cur.tokens_used DESC, u.created_at DESC
	`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch merchant token usage: %w", err)
	}
	defer rows.Close()

	var merchants []domain.MerchantTokenUsage
	for rows.Next() {
		var m domain.MerchantTokenUsage
		var curUsed, prevUsed int64
		var tokenLimit float64
		if err := rows.Scan(
			&m.ID,
			&m.PublicID,
			&m.FullName,
			&m.CompanyName,
			&m.CurrentPlanName,
			&tokenLimit,
			&curUsed,
			&prevUsed,
		); err != nil {
			return nil, err
		}
		m.TokensUsed = curUsed
		m.TokenLimit = tokenLimit
		m.UsagePct = usagePct(curUsed, tokenLimit)
		m.PeriodDeltaPct = pctDelta(float64(curUsed), float64(prevUsed))
		merchants = append(merchants, m)
	}
	if merchants == nil {
		merchants = []domain.MerchantTokenUsage{}
	}
	return merchants, rows.Err()
}
