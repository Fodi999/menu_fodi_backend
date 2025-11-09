package dto

import (
	"time"

	"github.com/google/uuid"
)

// PurchaseRequest represents a token purchase request
type PurchaseRequest struct {
	Amount        int    `json:"amount" validate:"required,gt=0"`
	PaymentMethod string `json:"paymentMethod"` // card, paypal, crypto
	Description   string `json:"description"`
}

// SpendRequest represents a token spend request
type SpendRequest struct {
	Amount      int    `json:"amount" validate:"required,gt=0"`
	Description string `json:"description"`
	RelatedID   string `json:"relatedId,omitempty"` // Optional: course ID, recipe ID, etc.
}

// BalanceResponse represents wallet balance response
type BalanceResponse struct {
	UserID  uuid.UUID `json:"userId"`
	Balance int       `json:"balance"`
}

// TransactionResponse represents a wallet transaction response
type TransactionResponse struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"userId"`
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
	Amount          int       `json:"amount"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	NewBalance      int       `json:"newBalance"`
	PreviousBalance int       `json:"previousBalance,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}
