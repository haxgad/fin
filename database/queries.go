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

// NewAccountRepository creates a new account repository instance
// This constructor initializes the repository with a database connection
// Parameters:
//   - db: Active SQL database connection for executing account operations
//
// Returns: Configured AccountRepository ready for use
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// CreateAccount inserts a new account record into the database
// This method creates a new account with the specified ID and initial balance
// Parameters:
//   - accountID: Unique identifier for the new account (must be positive)
//   - initialBalance: Starting balance for the account (should be non-negative)
//
// Returns:
//   - error: Database error if insertion fails, nil on success
//
// Database behavior:
//   - Inserts into accounts table with provided ID and balance
//   - Will fail if account ID already exists (database constraint violation)
//   - Uses precise decimal arithmetic for monetary values
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

// GetAccount retrieves account information by account ID
// This method fetches the current account details including balance
// Parameters:
//   - accountID: The unique identifier of the account to retrieve
//
// Returns:
//   - *models.Account: Account object with ID and current balance if found
//   - error: "account not found" if ID doesn't exist, other database errors possible
//
// Database behavior:
//   - Performs single SELECT query on accounts table
//   - Returns sql.ErrNoRows if account doesn't exist (converted to friendly error)
//   - Balance is returned as precise decimal value
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

// AccountExists checks whether an account with the given ID exists in the database
// This method is used for validation before creating accounts or processing transactions
// Parameters:
//   - accountID: The account ID to check for existence
//
// Returns:
//   - bool: true if account exists, false if it doesn't exist
//   - error: Database error if query fails, nil on successful check
//
// Database behavior:
//   - Uses efficient EXISTS query to check presence without retrieving data
//   - Returns boolean result without loading full account details
//   - Fast operation suitable for validation checks
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

// NewTransactionRepository creates a new transaction repository instance
// This constructor initializes the repository with a database connection
// Parameters:
//   - db: Active SQL database connection for executing transaction operations
//
// Returns: Configured TransactionRepository ready for use
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction performs an atomic money transfer between two accounts
// This method implements a complete transfer operation with balance validation and record keeping
// Parameters:
//   - sourceAccountID: Account ID to debit the amount from
//   - destinationAccountID: Account ID to credit the amount to
//   - amount: Amount to transfer (must be positive)
//
// Returns:
//   - error: Specific error messages for business rule violations or database issues
//
// Business rules enforced:
//   - Source account must exist and have sufficient balance
//   - Destination account must exist
//   - Amount must be positive (validated by caller)
//
// Database behavior:
//   - Uses database transaction for atomicity (all operations succeed or all fail)
//   - Locks both account rows with FOR UPDATE to prevent race conditions
//   - Updates both account balances and creates transaction record
//   - Automatically rolls back on any error, commits only on complete success
//
// Possible error returns:
//   - "source account not found": Source account doesn't exist
//   - "destination account not found": Destination account doesn't exist
//   - "insufficient balance": Source account has less than transfer amount
//   - Various database errors for connection/constraint issues
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
