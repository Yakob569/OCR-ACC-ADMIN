package domain

import (
	"time"

	"github.com/google/uuid"
)

// MerchantSummary is an admin view of a merchant account with usage and billing aggregates.
type MerchantSummary struct {
	ID              uuid.UUID  `json:"id"`
	PublicID        string     `json:"merchant_id"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	CompanyName     string     `json:"company_name,omitempty"`
	Status          string     `json:"status"`
	CurrentPlanID   string     `json:"current_plan_id,omitempty"`
	CurrentPlanName string     `json:"current_plan_name,omitempty"`
	OcrSuccessCount int        `json:"ocr_success_count"`
	TotalRevenue    float64    `json:"total_revenue"`
	CreatedAt       time.Time  `json:"created_at"`
	LastActiveAt    *time.Time `json:"last_active_at,omitempty"`
}
