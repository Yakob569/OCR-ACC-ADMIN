package domain

import (
	"time"

	"github.com/google/uuid"
)

type PricingPlan struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Amount       float64   `json:"amount"`
	DurationDays int       `json:"duration_days"`
	IsActive     bool      `json:"is_active"`
	Features     []string  `json:"features"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreatePlanRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Amount       float64  `json:"amount"`
	DurationDays int      `json:"duration_days"`
	Features     []string `json:"features"`
}

type UpdatePlanRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Amount       float64  `json:"amount"`
	DurationDays int      `json:"duration_days"`
	Features     []string `json:"features"`
}
