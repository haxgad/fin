package database

import (
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewAccountRepository(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	defer db.Close()

	repo := NewAccountRepository(db)
	if repo == nil {
		t.Error("NewAccountRepository returned nil")
	}
	if repo.db != db {
		t.Error("Repository not initialized with correct database")
	}
}

func TestNewTransactionRepository(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	defer db.Close()

	repo := NewTransactionRepository(db)
	if repo == nil {
		t.Error("NewTransactionRepository returned nil")
	}
	if repo.db != db {
		t.Error("Repository not initialized with correct database")
	}
}

// Test repository methods with mock/invalid database to test error paths
func TestAccountRepository_ErrorPaths(t *testing.T) {
	// Create a closed database to test error handling
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	db.Close() // Close to make operations fail

	repo := NewAccountRepository(db)

	t.Run("CreateAccount with closed DB", func(t *testing.T) {
		err := repo.CreateAccount(123, decimal.NewFromFloat(100.0))
		if err == nil {
			t.Error("Expected error with closed database")
		}
	})

	t.Run("GetAccount with closed DB", func(t *testing.T) {
		_, err := repo.GetAccount(123)
		if err == nil {
			t.Error("Expected error with closed database")
		}
	})

	t.Run("AccountExists with closed DB", func(t *testing.T) {
		_, err := repo.AccountExists(123)
		if err == nil {
			t.Error("Expected error with closed database")
		}
	})
}

func TestTransactionRepository_ErrorPaths(t *testing.T) {
	// Create a closed database to test error handling
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	db.Close() // Close to make operations fail

	repo := NewTransactionRepository(db)

	t.Run("CreateTransaction with closed DB", func(t *testing.T) {
		err := repo.CreateTransaction(123, 456, decimal.NewFromFloat(100.0))
		if err == nil {
			t.Error("Expected error with closed database")
		}
	})
}

// Test with in-memory database for actual SQL operations
func TestAccountRepository_WithMemoryDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	defer db.Close()

	// Create a simplified accounts table for SQLite
	_, err = db.Exec(`
		CREATE TABLE accounts (
			account_id INTEGER PRIMARY KEY,
			balance DECIMAL(15,5) NOT NULL CHECK (balance >= 0)
		)
	`)
	if err != nil {
		t.Skip("Could not create test table")
	}

	repo := NewAccountRepository(db)

	t.Run("CreateAccount success", func(t *testing.T) {
		err := repo.CreateAccount(123, decimal.NewFromFloat(100.50))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("GetAccount success", func(t *testing.T) {
		account, err := repo.GetAccount(123)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if account == nil {
			t.Error("Expected account, got nil")
		}
		if account != nil && account.AccountID != 123 {
			t.Errorf("Expected account ID 123, got %d", account.AccountID)
		}
	})

	t.Run("GetAccount not found", func(t *testing.T) {
		_, err := repo.GetAccount(999)
		if err == nil {
			t.Error("Expected error for non-existent account")
		}
	})

	t.Run("AccountExists true", func(t *testing.T) {
		exists, err := repo.AccountExists(123)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !exists {
			t.Error("Expected account to exist")
		}
	})

	t.Run("AccountExists false", func(t *testing.T) {
		exists, err := repo.AccountExists(999)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if exists {
			t.Error("Expected account to not exist")
		}
	})
}

// Test interface compliance
func TestRepositoryInterfaces(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	defer db.Close()

	// Test that our concrete types implement the interfaces
	var _ AccountRepositoryInterface = NewAccountRepository(db)
	var _ TransactionRepositoryInterface = NewTransactionRepository(db)

	t.Log("Repository interfaces implemented correctly")
}

// Test repository method signatures
func TestRepositoryMethodSignatures(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	defer db.Close()

	accountRepo := NewAccountRepository(db)
	transactionRepo := NewTransactionRepository(db)

	// Test that methods exist and have correct signatures
	t.Run("AccountRepository methods", func(t *testing.T) {
		// These calls will fail but test that methods exist with correct signatures
		accountRepo.CreateAccount(0, decimal.Zero)
		accountRepo.GetAccount(0)
		accountRepo.AccountExists(0)
	})

	t.Run("TransactionRepository methods", func(t *testing.T) {
		// This call will fail but tests that method exists with correct signature
		transactionRepo.CreateTransaction(0, 0, decimal.Zero)
	})
}
