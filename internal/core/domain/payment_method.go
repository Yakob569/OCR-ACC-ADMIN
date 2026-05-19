package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	ImageURL      string    `json:"image_url"`
	AccountNumber string    `json:"account_number"`
	Status        string    `json:"status"` // 'active', 'inactive'
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
