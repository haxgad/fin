package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Transaction represents a money transfer between accounts
type Transaction struct {
	ID                   int64           `json:"id" db:"id"`
	SourceAccountID      int64           `json:"source_account_id" db:"source_account_id"`
	DestinationAccountID int64           `json:"destination_account_id" db:"destination_account_id"`
	Amount               decimal.Decimal `json:"amount" db:"amount"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
}

// CreateTransactionRequest represents the request payload for creating a transaction
type CreateTransactionRequest struct {
	SourceAccountID      int64  `json:"source_account_id"`
	DestinationAccountID int64  `json:"destination_account_id"`
	Amount               string `json:"amount"`
}
