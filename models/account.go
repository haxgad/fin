package models

import (
	"github.com/shopspring/decimal"
)

// Account represents a bank account
type Account struct {
	AccountID int64           `json:"account_id" db:"account_id"`
	Balance   decimal.Decimal `json:"balance" db:"balance"`
}

// CreateAccountRequest represents the request payload for creating an account
type CreateAccountRequest struct {
	AccountID      int64  `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

// AccountResponse represents the response for account queries
type AccountResponse struct {
	AccountID int64  `json:"account_id"`
	Balance   string `json:"balance"`
}
