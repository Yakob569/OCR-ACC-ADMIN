package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
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

	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO pricing_plans (id, name, description, amount, duration_days, is_active, features)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		plan.ID,
		plan.Name,
		plan.Description,
		plan.Amount,
		plan.DurationDays,
		plan.IsActive,
		featuresJSON,
	).Scan(&plan.CreatedAt, &plan.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "pricing_plans_name_key") || strings.Contains(err.Error(), "23505") {
			return nil, errors.New("a pricing plan with this name already exists")
		}
		return nil, err
	}

	return plan, nil
}

func (r *planRepository) Update(ctx context.Context, plan *domain.PricingPlan) (*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE pricing_plans
		SET name = $1, description = $2, amount = $3, duration_days = $4, features = $5
		WHERE id = $6
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		plan.Name,
		plan.Description,
		plan.Amount,
		plan.DurationDays,
		featuresJSON,
		plan.ID,
	).Scan(&plan.CreatedAt, &plan.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "pricing_plans_name_key") || strings.Contains(err.Error(), "23505") {
			return nil, errors.New("a pricing plan with this name already exists")
		}
		return nil, err
	}

	return plan, nil
}

func (r *planRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, name, description, amount, duration_days, is_active, features, created_at, updated_at
		FROM pricing_plans
		WHERE id = $1
	`

	var plan domain.PricingPlan
	var featuresJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Description,
		&plan.Amount,
		&plan.DurationDays,
		&plan.IsActive,
		&featuresJSON,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
			return nil, err
		}
	} else {
		plan.Features = []string{}
	}

	return &plan, nil
}

func (r *planRepository) List(ctx context.Context) ([]*domain.PricingPlan, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, name, description, amount, duration_days, is_active, features, created_at, updated_at
		FROM pricing_plans
		ORDER BY amount ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*domain.PricingPlan
	for rows.Next() {
		var plan domain.PricingPlan
		var featuresJSON []byte

		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.Description,
			&plan.Amount,
			&plan.DurationDays,
			&plan.IsActive,
			&featuresJSON,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(featuresJSON) > 0 {
			if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
				return nil, err
			}
		} else {
			plan.Features = []string{}
		}

		plans = append(plans, &plan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *planRepository) ToggleStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	query := `
		UPDATE pricing_plans
		SET is_active = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, isActive, id)
	return err
}
