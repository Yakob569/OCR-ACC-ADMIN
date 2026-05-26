package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserSubscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PlanID    uuid.UUID `json:"plan_id"`
	Status    string    `json:"status"` // 'active', 'expired', 'canceled'
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriptionRequest struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	PlanID            string    `json:"plan_id"`
	PaymentLink       string    `json:"payment_link"`
	PaymentScreenshot string    `json:"payment_screenshot"`
	Status            string    `json:"status"` // 'pending', 'approved', 'rejected'
	RejectionReason   string    `json:"rejection_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RejectSubscriptionRequest struct {
	RejectionReason string `json:"rejection_reason"`
}
