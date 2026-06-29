package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID            uuid.UUID  `json:"id"`
	WalletID      uuid.UUID  `json:"wallet_id"`
	Amount        int64      `json:"amount"`
	BalanceBefore int64      `json:"balance_before"`
	BalanceAfter  int64      `json:"balance_after"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Description   string     `json:"description"`
	ReferenceType string     `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CreateTransactionRequest struct {
	Amount        int64      `json:"amount"`
	Type          string     `json:"type"`
	Description   string     `json:"description,omitempty"`
	ReferenceType string     `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
}
