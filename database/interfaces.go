package database

import (
	"github.com/shopspring/decimal"

	"internal-transfers/models"
)

// AccountRepositoryInterface defines the contract for account operations
type AccountRepositoryInterface interface {
	CreateAccount(accountID int64, initialBalance decimal.Decimal) error
	GetAccount(accountID int64) (*models.Account, error)
	AccountExists(accountID int64) (bool, error)
}

// TransactionRepositoryInterface defines the contract for transaction operations
type TransactionRepositoryInterface interface {
	CreateTransaction(sourceAccountID, destinationAccountID int64, amount decimal.Decimal) error
}

// Ensure our concrete types implement the interfaces
var _ AccountRepositoryInterface = (*AccountRepository)(nil)
var _ TransactionRepositoryInterface = (*TransactionRepository)(nil)
