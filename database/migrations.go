package database

import (
	"database/sql"
	"fmt"
)

// Migrate executes all database schema migrations in the correct order
// This function sets up the complete database schema for the transfers service
// Parameters:
//   - db: Active database connection to execute migrations against
//
// Returns:
//   - error: Migration error if any step fails, nil on complete success
//
// Migration sequence:
//  1. Creates accounts table with balance constraints
//  2. Creates transactions table with foreign key relationships
//  3. Creates performance indexes on transaction lookups
//
// Note: Uses IF NOT EXISTS to make migrations idempotent (safe to run multiple times)
// Important: Migrations are run in order and will stop on first failure
func Migrate(db *sql.DB) error {
	migrations := []string{
		createAccountsTable,
		createTransactionsTable,
		createIndexes,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	return nil
}

// createAccountsTable defines the schema for storing bank account information
// Key design decisions:
//   - BIGINT account_id for large scale account numbering
//   - DECIMAL(15,5) for precise monetary calculations (up to 999,999,999.99999)
//   - CHECK constraint prevents negative balances at database level
//   - Timestamps for audit trail with timezone awareness
//   - Primary key on account_id for unique identification and fast lookups
const createAccountsTable = `
CREATE TABLE IF NOT EXISTS accounts (
    account_id BIGINT PRIMARY KEY,
    balance DECIMAL(15,5) NOT NULL CHECK (balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

// createTransactionsTable defines the schema for storing transfer transaction records
// Key design decisions:
//   - BIGSERIAL id for unique transaction identification and ordering
//   - Foreign keys ensure referential integrity with accounts table
//   - CHECK constraints enforce business rules (positive amounts, different accounts)
//   - DECIMAL(15,5) matches account balance precision for consistency
//   - Timestamps for transaction audit trail and ordering
//   - Source/destination pattern supports directional money transfers
const createTransactionsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    source_account_id BIGINT NOT NULL,
    destination_account_id BIGINT NOT NULL,
    amount DECIMAL(15,5) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (source_account_id) REFERENCES accounts(account_id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(account_id),
    CHECK (source_account_id != destination_account_id)
);
`

// createIndexes defines performance indexes for efficient transaction queries
// Index strategy:
//   - idx_transactions_source_account: Fast lookup of outgoing transfers for an account
//   - idx_transactions_destination_account: Fast lookup of incoming transfers for an account
//   - idx_transactions_created_at: Fast time-based queries and transaction history ordering
//
// These indexes support common query patterns like account transaction history,
// balance calculations, and time-based reporting without full table scans
const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_transactions_source_account ON transactions(source_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_destination_account ON transactions(destination_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
`
