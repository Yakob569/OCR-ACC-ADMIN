package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserSubscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PlanID    int       `json:"plan_id"`
	Status    string    `json:"status"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriptionRequest struct {
	ID                uuid.UUID `json:"id"`
	RequestID         string    `json:"request_id"`
	UserID            uuid.UUID `json:"user_id"`
	PlanID            int       `json:"plan_id"`
	PlanPublicID      string    `json:"pricing_plan_id"`
	PlanName          string    `json:"plan_name"`
	PlanAmount        float64   `json:"plan_amount"`
	PlanDurationDays  int       `json:"plan_duration_days"`
	UserEmail         string    `json:"user_email"`
	UserFullName      string    `json:"user_full_name"`
	CompanyName            string     `json:"company_name,omitempty"`
	PaymentMethodID        *uuid.UUID `json:"payment_method_id,omitempty"`
	PaymentMethodName      string     `json:"payment_method_name,omitempty"`
	PaymentMethodAccount   string     `json:"payment_method_account,omitempty"`
	PaymentLink            string     `json:"payment_link"`
	PaymentScreenshot string    `json:"payment_screenshot"`
	Status            string    `json:"status"`
	RejectionReason   string    `json:"rejection_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RejectSubscriptionRequest struct {
	RejectionReason string `json:"rejection_reason"`
}
