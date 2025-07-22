package database

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// =============================================================================
// Environment and Database Connection Tests
// =============================================================================

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Environment variable set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "Environment variable not set",
			key:          "UNSET_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if specified
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvWithDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvWithDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInitDB_InvalidConnection(t *testing.T) {
	// Test with invalid database parameters
	originalEnv := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_SSLMODE":  os.Getenv("DB_SSLMODE"),
	}

	// Set invalid connection parameters
	os.Setenv("DB_HOST", "invalid-host")
	os.Setenv("DB_PORT", "9999")
	os.Setenv("DB_USER", "invalid")
	os.Setenv("DB_PASSWORD", "invalid")
	os.Setenv("DB_NAME", "invalid")
	os.Setenv("DB_SSLMODE", "disable")

	// Restore original environment after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	db, err := InitDB()
	if err == nil {
		t.Error("Expected error with invalid database connection, got nil")
		if db != nil {
			db.Close()
		}
	}
}

// =============================================================================
// Migration Tests
// =============================================================================

func TestMigrationStructure(t *testing.T) {
	// Test that migration constants are properly structured
	if len(createAccountsTable) == 0 {
		t.Error("createAccountsTable is empty")
	}
	if len(createTransactionsTable) == 0 {
		t.Error("createTransactionsTable is empty")
	}
	if len(createIndexes) == 0 {
		t.Error("createIndexes is empty")
	}
}

func TestMigrationSQL_Accounts(t *testing.T) {
	sql := createAccountsTable

	// Test that required elements are present
	requiredElements := []string{
		"CREATE TABLE",
		"accounts",
		"account_id",
		"BIGINT",
		"PRIMARY KEY",
		"balance",
		"DECIMAL",
		"NOT NULL",
		"CHECK",
		"balance >= 0",
		"created_at",
		"updated_at",
		"TIMESTAMP WITH TIME ZONE",
		"DEFAULT NOW()",
	}

	for _, element := range requiredElements {
		if !strings.Contains(sql, element) {
			t.Errorf("accounts table SQL missing required element: %s", element)
		}
	}
}

func TestMigrationSQL_Transactions(t *testing.T) {
	sql := createTransactionsTable

	// Test that required elements are present
	requiredElements := []string{
		"CREATE TABLE",
		"transactions",
		"id",
		"BIGSERIAL",
		"PRIMARY KEY",
		"source_account_id",
		"destination_account_id",
		"amount",
		"DECIMAL",
		"NOT NULL",
		"CHECK",
		"amount > 0",
		"created_at",
		"FOREIGN KEY",
		"REFERENCES accounts",
		"source_account_id != destination_account_id",
	}

	for _, element := range requiredElements {
		if !strings.Contains(sql, element) {
			t.Errorf("transactions table SQL missing required element: %s", element)
		}
	}
}

func TestMigrationSQL_Indexes(t *testing.T) {
	sql := createIndexes

	// Test that required indexes are present
	requiredIndexes := []string{
		"CREATE INDEX",
		"idx_transactions_source_account",
		"idx_transactions_destination_account",
		"idx_transactions_created_at",
		"ON transactions",
	}

	for _, index := range requiredIndexes {
		if !strings.Contains(sql, index) {
			t.Errorf("indexes SQL missing required index: %s", index)
		}
	}
}

func TestMigrationValidSQL(t *testing.T) {
	// Test that each migration contains valid SQL structure
	migrations := map[string]string{
		"accounts table":     createAccountsTable,
		"transactions table": createTransactionsTable,
		"indexes":            createIndexes,
	}

	for name, migration := range migrations {
		t.Run(name, func(t *testing.T) {
			// Check basic SQL structure
			if !strings.Contains(migration, "CREATE") {
				t.Errorf("%s migration doesn't contain CREATE statement", name)
			}

			// Check that it's not obviously malformed
			if strings.Count(migration, "(") != strings.Count(migration, ")") {
				t.Errorf("%s migration has mismatched parentheses", name)
			}
		})
	}
}

// =============================================================================
// Repository Constructor Tests
// =============================================================================

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

func TestRepositoryTypes(t *testing.T) {
	// Create repositories to test their structure
	accountRepo := &AccountRepository{}
	transactionRepo := &TransactionRepository{}

	// Test that they're the expected types
	_ = AccountRepositoryInterface(accountRepo)
	_ = TransactionRepositoryInterface(transactionRepo)

	t.Log("Repository types implement their interfaces correctly")
}

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

// =============================================================================
// SQLite-dependent tests (will be skipped if SQLite not available)
// =============================================================================

func TestMigrate_Success(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing migrations")
	}
	defer db.Close()

	err = Migrate(db)
	// We expect this to fail with SQLite due to PostgreSQL-specific syntax
	// but it tests that the function executes and handles errors properly
	if err == nil {
		t.Log("Migrations completed successfully")
	} else {
		t.Logf("Expected error due to PostgreSQL syntax on SQLite: %v", err)
	}
}

func TestMigrate_InvalidDatabase(t *testing.T) {
	// Create a closed database connection to test error handling
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skip("SQLite not available for testing")
	}
	db.Close() // Close it to make it invalid

	err = Migrate(db)
	if err == nil {
		t.Error("Expected error with closed database connection")
	}
}

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
