package database

import (
	"database/sql"
	"fmt"
	"internal-transfers/models"

	"github.com/shopspring/decimal"
)

// AccountRepository handles account-related database operations
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// CreateAccount creates a new account with the given ID and initial balance
func (r *AccountRepository) CreateAccount(accountID int64, initialBalance decimal.Decimal) error {
	query := `
		INSERT INTO accounts (account_id, balance)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(query, accountID, initialBalance)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

// GetAccount retrieves an account by ID
func (r *AccountRepository) GetAccount(accountID int64) (*models.Account, error) {
	query := `
		SELECT account_id, balance
		FROM accounts
		WHERE account_id = $1
	`

	var account models.Account
	err := r.db.QueryRow(query, accountID).Scan(&account.AccountID, &account.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// AccountExists checks if an account exists
func (r *AccountRepository) AccountExists(accountID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id = $1)`

	var exists bool
	err := r.db.QueryRow(query, accountID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}

	return exists, nil
}

// TransactionRepository handles transaction-related database operations
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction creates a new transaction and updates account balances atomically
func (r *TransactionRepository) CreateTransaction(sourceAccountID, destinationAccountID int64, amount decimal.Decimal) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check source account balance and lock the row
	var sourceBalance decimal.Decimal
	err = tx.QueryRow("SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE", sourceAccountID).Scan(&sourceBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("source account not found")
		}
		return fmt.Errorf("failed to get source account: %w", err)
	}

	// Check if source account has sufficient balance
	if sourceBalance.LessThan(amount) {
		return fmt.Errorf("insufficient balance")
	}

	// Lock destination account
	var destinationBalance decimal.Decimal
	err = tx.QueryRow("SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE", destinationAccountID).Scan(&destinationBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("destination account not found")
		}
		return fmt.Errorf("failed to get destination account: %w", err)
	}

	// Update source account balance
	_, err = tx.Exec("UPDATE accounts SET balance = balance - $1, updated_at = NOW() WHERE account_id = $2", amount, sourceAccountID)
	if err != nil {
		return fmt.Errorf("failed to update source account: %w", err)
	}

	// Update destination account balance
	_, err = tx.Exec("UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE account_id = $2", amount, destinationAccountID)
	if err != nil {
		return fmt.Errorf("failed to update destination account: %w", err)
	}

	// Insert transaction record
	_, err = tx.Exec(
		"INSERT INTO transactions (source_account_id, destination_account_id, amount) VALUES ($1, $2, $3)",
		sourceAccountID, destinationAccountID, amount,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
