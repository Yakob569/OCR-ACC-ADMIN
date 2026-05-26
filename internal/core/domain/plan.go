package domain

import (
	"crypto/rand"
	"math/big"
	"time"
)

type PricingPlan struct {
	ID            int       `json:"id"`
	PricingPlanID string    `json:"pricing_plan_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Amount        float64   `json:"amount"`
	DurationDays  int       `json:"duration_days"`
	Status        string    `json:"status"`
	TrialDays     int       `json:"trial_days"`
	TokenPerMonth float64   `json:"token_per_month"`
	OcrPerDay     int       `json:"ocr_per_day"`
	IsActive      bool      `json:"is_active"`
	Features      []string  `json:"features"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Feature struct {
	ID          int       `json:"id"`
	FeatureID   string    `json:"feature_id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePlanRequest struct {
	Name          string   `json:"name" validate:"required"`
	Description   string   `json:"description" validate:"required"`
	Status        string   `json:"status" validate:"required"`
	TrialDays     int      `json:"trial_days" validate:"required"`
	TokenPerMonth float64  `json:"token_per_month" validate:"required"`
	OcrPerDay     int      `json:"ocr_per_day" validate:"required"`
	Amount        float64  `json:"amount" validate:"required"`
	DurationDays  int      `json:"duration_days" validate:"required"`
	Features      []string `json:"features" validate:"required"`
}

type UpdatePlanRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Status        string   `json:"status"`
	TrialDays     int      `json:"trial_days"`
	TokenPerMonth float64  `json:"token_per_month"`
	OcrPerDay     int      `json:"ocr_per_day"`
	Amount        float64  `json:"amount"`
	DurationDays  int      `json:"duration_days"`
	Features      []string `json:"features"`
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateCustomID(prefix string) (string, error) {
	b := make([]byte, 7)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return prefix + string(b), nil
}
