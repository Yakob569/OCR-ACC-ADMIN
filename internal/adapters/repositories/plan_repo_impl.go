package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type planRepository struct {
	db *pgxpool.Pool
}

func NewPricingPlanRepository(db *pgxpool.Pool) ports.PricingPlanRepository {
	return &planRepository{
		db: db,
	}
}

func (r *planRepository) Create(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO pricing_plans (pricing_plan_id, name, description, amount, duration_days, status, trial_days, token_per_month, ocr_per_day, is_active, is_default, ocr_lifetime_limit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		plan.PricingPlanID,
		plan.Name,
		plan.Description,
		plan.Amount,
		plan.DurationDays,
		plan.Status,
		plan.TrialDays,
		plan.TokenPerMonth,
		plan.OcrPerDay,
		plan.IsActive,
		plan.IsDefault,
		plan.OcrLifetimeLimit,
	).Scan(&plan.ID, &plan.CreatedAt, &plan.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "pricing_plans_name_key") || strings.Contains(err.Error(), "23505") {
			return nil, errors.New("a pricing plan with this name already exists")
		}
		return nil, err
	}

	// Save and map features relationally
	for _, featDesc := range plan.Features {
		featDescClean := strings.TrimSpace(featDesc)
		if featDescClean == "" {
			continue
		}

		// Generate random feature public ID
		featBusinessID, err := domain.GenerateCustomID("PFE-")
		if err != nil {
			return nil, err
		}

		// Insert feature lookup if missing
		var featIntID int
		featQuery := `
			INSERT INTO features (feature_id, description, status)
			VALUES ($1, $2, 'active')
			ON CONFLICT (description) DO UPDATE SET description = EXCLUDED.description
			RETURNING id
		`
		err = tx.QueryRow(ctx, featQuery, featBusinessID, featDescClean).Scan(&featIntID)
		if err != nil {
			return nil, err
		}

		// Map to plan in junction table
		mapQuery := `
			INSERT INTO plan_features (plan_id, feature_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err = tx.Exec(ctx, mapQuery, plan.ID, featIntID)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return plan, nil
}

func (r *planRepository) Update(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE pricing_plans
		SET name = $1, description = $2, amount = $3, duration_days = $4, status = $5, trial_days = $6, token_per_month = $7, ocr_per_day = $8, is_default = $9, ocr_lifetime_limit = $10
		WHERE id = $11
		RETURNING created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		plan.Name,
		plan.Description,
		plan.Amount,
		plan.DurationDays,
		plan.Status,
		plan.TrialDays,
		plan.TokenPerMonth,
		plan.OcrPerDay,
		plan.IsDefault,
		plan.OcrLifetimeLimit,
		plan.ID,
	).Scan(&plan.CreatedAt, &plan.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "pricing_plans_name_key") || strings.Contains(err.Error(), "23505") {
			return nil, errors.New("a pricing plan with this name already exists")
		}
		return nil, err
	}

	// Delete old junction mappings
	_, err = tx.Exec(ctx, `DELETE FROM plan_features WHERE plan_id = $1`, plan.ID)
	if err != nil {
		return nil, err
	}

	// Save and map updated features relationally
	for _, featDesc := range plan.Features {
		featDescClean := strings.TrimSpace(featDesc)
		if featDescClean == "" {
			continue
		}

		featBusinessID, err := domain.GenerateCustomID("PFE-")
		if err != nil {
			return nil, err
		}

		var featIntID int
		featQuery := `
			INSERT INTO features (feature_id, description, status)
			VALUES ($1, $2, 'active')
			ON CONFLICT (description) DO UPDATE SET description = EXCLUDED.description
			RETURNING id
		`
		err = tx.QueryRow(ctx, featQuery, featBusinessID, featDescClean).Scan(&featIntID)
		if err != nil {
			return nil, err
		}

		mapQuery := `
			INSERT INTO plan_features (plan_id, feature_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err = tx.Exec(ctx, mapQuery, plan.ID, featIntID)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return plan, nil
}

func (r *planRepository) GetByBusinessID(ctx context.Context, planID string) (*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	plan, features, err := r.queryPlanByFilter(ctx, "p.pricing_plan_id = $1", planID)

	if err != nil {
		return nil, err
	}
	plan.Features = features
	return plan, nil
}

func (r *planRepository) List(ctx context.Context) ([]*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT p.id, p.pricing_plan_id, p.name, p.description, p.amount, p.duration_days, p.status, p.trial_days, p.token_per_month, p.ocr_per_day, p.is_active, p.is_default, p.ocr_lifetime_limit, p.created_at, p.updated_at,
		       COALESCE(array_agg(f.description ORDER BY f.description) FILTER (WHERE f.description IS NOT NULL), '{}') as features
		FROM pricing_plans p
		LEFT JOIN plan_features pf ON pf.plan_id = p.id
		LEFT JOIN features f ON pf.feature_id = f.id
		GROUP BY p.id
		ORDER BY p.amount ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*domain.PricingPlan
	for rows.Next() {
		plan, features, scanErr := scanPlanRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		plan.Features = features
		plans = append(plans, plan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if plans == nil {
		plans = []*domain.PricingPlan{}
	}

	return plans, nil
}

func (r *planRepository) ToggleStatus(ctx context.Context, planID string, isActive bool) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	query := `
		UPDATE pricing_plans
		SET is_active = $1
		WHERE pricing_plan_id = $2
	`

	_, err := r.db.Exec(ctx, query, isActive, planID)
	return err
}

func (r *planRepository) HasActiveSubscriptions(ctx context.Context, planID string) (bool, error) {
	if r.db == nil {
		return false, errors.New("database connection is not available")
	}

	query := `
		SELECT COUNT(1) > 0
		FROM user_subscriptions us
		JOIN pricing_plans p ON p.id = us.plan_id
		WHERE p.pricing_plan_id = $1
		  AND us.status = 'active'
	`

	var hasActive bool
	if err := r.db.QueryRow(ctx, query, planID).Scan(&hasActive); err != nil {
		return false, err
	}

	return hasActive, nil
}

func (r *planRepository) GetByInternalID(ctx context.Context, id int) (*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	plan, features, err := r.queryPlanByFilter(ctx, "p.id = $1", id)
	if err != nil {
		return nil, err
	}
	plan.Features = features
	return plan, nil
}

func (r *planRepository) ClearDefaultPlan(ctx context.Context) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}
	_, err := r.db.Exec(ctx, `UPDATE pricing_plans SET is_default = FALSE WHERE is_default = TRUE`)
	return err
}

type planRowScanner interface {
	Scan(dest ...any) error
}

func scanPlanRow(row planRowScanner) (*domain.PricingPlan, []string, error) {
	var plan domain.PricingPlan
	var features []string
	err := row.Scan(
		&plan.ID,
		&plan.PricingPlanID,
		&plan.Name,
		&plan.Description,
		&plan.Amount,
		&plan.DurationDays,
		&plan.Status,
		&plan.TrialDays,
		&plan.TokenPerMonth,
		&plan.OcrPerDay,
		&plan.IsActive,
		&plan.IsDefault,
		&plan.OcrLifetimeLimit,
		&plan.CreatedAt,
		&plan.UpdatedAt,
		&features,
	)
	if err != nil {
		return nil, nil, err
	}
	return &plan, features, nil
}

func (r *planRepository) queryPlanByFilter(ctx context.Context, filter string, arg any) (*domain.PricingPlan, []string, error) {
	query := `
		SELECT p.id, p.pricing_plan_id, p.name, p.description, p.amount, p.duration_days, p.status, p.trial_days, p.token_per_month, p.ocr_per_day, p.is_active, p.is_default, p.ocr_lifetime_limit, p.created_at, p.updated_at,
		       COALESCE(array_agg(f.description ORDER BY f.description) FILTER (WHERE f.description IS NOT NULL), '{}') as features
		FROM pricing_plans p
		LEFT JOIN plan_features pf ON pf.plan_id = p.id
		LEFT JOIN features f ON pf.feature_id = f.id
		WHERE ` + filter + `
		GROUP BY p.id
	`
	plan, features, err := scanPlanRow(r.db.QueryRow(ctx, query, arg))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, sql.ErrNoRows
		}
		return nil, nil, err
	}
	return plan, features, nil
}
