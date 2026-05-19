package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type paymentMethodRepository struct {
	db *pgxpool.Pool
}

func NewPaymentMethodRepository(db *pgxpool.Pool) ports.PaymentMethodRepository {
	return &paymentMethodRepository{
		db: db,
	}
}

func (r *paymentMethodRepository) Create(ctx context.Context, method *domain.PaymentMethod) (*domain.PaymentMethod, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		INSERT INTO payment_methods (id, name, image_url, account_number, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		method.ID,
		method.Name,
		method.ImageURL,
		method.AccountNumber,
		method.Status,
	).Scan(&method.CreatedAt, &method.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return method, nil
}

func (r *paymentMethodRepository) Update(ctx context.Context, method *domain.PaymentMethod) (*domain.PaymentMethod, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		UPDATE payment_methods
		SET name = $1, image_url = $2, account_number = $3, status = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		method.Name,
		method.ImageURL,
		method.AccountNumber,
		method.Status,
		method.ID,
	).Scan(&method.CreatedAt, &method.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return method, nil
}

func (r *paymentMethodRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PaymentMethod, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, name, image_url, account_number, status, created_at, updated_at
		FROM payment_methods
		WHERE id = $1
	`

	var method domain.PaymentMethod
	var imageURL sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&method.ID,
		&method.Name,
		&imageURL,
		&method.AccountNumber,
		&method.Status,
		&method.CreatedAt,
		&method.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	if imageURL.Valid {
		method.ImageURL = imageURL.String
	}

	return &method, nil
}

func (r *paymentMethodRepository) List(ctx context.Context) ([]*domain.PaymentMethod, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, name, image_url, account_number, status, created_at, updated_at
		FROM payment_methods
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []*domain.PaymentMethod
	for rows.Next() {
		var method domain.PaymentMethod
		var imageURL sql.NullString

		err := rows.Scan(
			&method.ID,
			&method.Name,
			&imageURL,
			&method.AccountNumber,
			&method.Status,
			&method.CreatedAt,
			&method.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if imageURL.Valid {
			method.ImageURL = imageURL.String
		}

		methods = append(methods, &method)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return methods, nil
}

func (r *paymentMethodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	query := `
		DELETE FROM payment_methods
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}
