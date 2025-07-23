package database

import (
	"github.com/shopspring/decimal"

	"internal-transfers/models"
)

// AccountRepositoryInterface defines the contract for account-related database operations
// This interface abstracts account data access to enable dependency injection and testing
// All methods should handle database errors gracefully and return descriptive error messages
// Implementations must ensure data consistency and proper error handling
// Used by HTTP handlers to interact with account data without direct database coupling
type AccountRepositoryInterface interface {
	// CreateAccount inserts a new account with the specified ID and initial balance
	// Should fail if account ID already exists or if database constraints are violated
	CreateAccount(accountID int64, initialBalance decimal.Decimal) error

	// GetAccount retrieves account information by ID
	// Returns account object with current balance or "account not found" error
	GetAccount(accountID int64) (*models.Account, error)

	// AccountExists checks if an account with the given ID exists
	// Returns boolean result and any database errors that occur during the check
	AccountExists(accountID int64) (bool, error)
}

// TransactionRepositoryInterface defines the contract for transaction-related database operations
// This interface abstracts transaction processing to enable dependency injection and testing
// All implementations must provide ACID transaction guarantees for money transfers
// Critical: Transfer operations must be atomic (all balance updates succeed or all fail)
// Used by HTTP handlers to process money transfers without direct database coupling
type TransactionRepositoryInterface interface {
	// CreateTransaction performs an atomic money transfer between two accounts
	// Must validate account existence, check sufficient balance, and update both accounts
	// Should use database transactions to ensure atomicity and prevent race conditions
	// Returns specific error messages for business rule violations (insufficient funds, etc.)
	CreateTransaction(sourceAccountID, destinationAccountID int64, amount decimal.Decimal) error
}

// Compile-time interface implementation checks
// These lines ensure our concrete repository types implement the required interfaces
// Will cause compilation error if interface contracts are not properly fulfilled
var _ AccountRepositoryInterface = (*AccountRepository)(nil)
var _ TransactionRepositoryInterface = (*TransactionRepository)(nil)
