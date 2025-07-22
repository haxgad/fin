package database

import (
	"testing"
)

// Test constructors without requiring actual database connection
func TestNewAccountRepository_Structure(t *testing.T) {
	// Test with nil database to verify constructor logic
	repo := NewAccountRepository(nil)
	if repo == nil {
		t.Error("NewAccountRepository should not return nil even with nil db")
	}
	if repo.db != nil {
		t.Error("Repository should store the provided db reference")
	}
}

func TestNewTransactionRepository_Structure(t *testing.T) {
	// Test with nil database to verify constructor logic
	repo := NewTransactionRepository(nil)
	if repo == nil {
		t.Error("NewTransactionRepository should not return nil even with nil db")
	}
	if repo.db != nil {
		t.Error("Repository should store the provided db reference")
	}
}

// Test that repository structs have correct field types
func TestRepositoryTypes(t *testing.T) {
	// Create repositories to test their structure
	accountRepo := &AccountRepository{}
	transactionRepo := &TransactionRepository{}

	// Test that they're the expected types
	_ = AccountRepositoryInterface(accountRepo)
	_ = TransactionRepositoryInterface(transactionRepo)

	t.Log("Repository types implement their interfaces correctly")
}

// Test interface method signatures exist
func TestAccountRepositoryInterface_Methods(t *testing.T) {
	// This tests that the interface methods exist with correct signatures
	var _ AccountRepositoryInterface = (*AccountRepository)(nil)
	t.Log("AccountRepositoryInterface methods verified")
}

func TestTransactionRepositoryInterface_Methods(t *testing.T) {
	// This tests that the interface methods exist with correct signatures
	var _ TransactionRepositoryInterface = (*TransactionRepository)(nil)
	t.Log("TransactionRepositoryInterface methods verified")
}
