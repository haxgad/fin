package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"internal-transfers/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// MockAccountRepository implements AccountRepository interface for testing
type MockAccountRepository struct {
	accounts map[int64]*models.Account
}

func NewMockAccountRepository() *MockAccountRepository {
	return &MockAccountRepository{
		accounts: make(map[int64]*models.Account),
	}
}

func (m *MockAccountRepository) CreateAccount(accountID int64, initialBalance decimal.Decimal) error {
	if _, exists := m.accounts[accountID]; exists {
		return sql.ErrNoRows // Simulate duplicate error
	}
	m.accounts[accountID] = &models.Account{
		AccountID: accountID,
		Balance:   initialBalance,
	}
	return nil
}

func (m *MockAccountRepository) GetAccount(accountID int64) (*models.Account, error) {
	if account, exists := m.accounts[accountID]; exists {
		return account, nil
	}
	return nil, fmt.Errorf("account not found")
}

func (m *MockAccountRepository) AccountExists(accountID int64) (bool, error) {
	_, exists := m.accounts[accountID]
	return exists, nil
}

// MockTransactionRepository implements TransactionRepository interface for testing
type MockTransactionRepository struct {
	accountRepo *MockAccountRepository
}

func NewMockTransactionRepository(accountRepo *MockAccountRepository) *MockTransactionRepository {
	return &MockTransactionRepository{
		accountRepo: accountRepo,
	}
}

func (m *MockTransactionRepository) CreateTransaction(sourceAccountID, destinationAccountID int64, amount decimal.Decimal) error {
	sourceAccount, exists := m.accountRepo.accounts[sourceAccountID]
	if !exists {
		return fmt.Errorf("source account not found")
	}

	_, exists = m.accountRepo.accounts[destinationAccountID]
	if !exists {
		return fmt.Errorf("destination account not found")
	}

	if sourceAccount.Balance.LessThan(amount) {
		return fmt.Errorf("insufficient balance")
	}

	// Update balances
	sourceAccount.Balance = sourceAccount.Balance.Sub(amount)
	m.accountRepo.accounts[destinationAccountID].Balance = m.accountRepo.accounts[destinationAccountID].Balance.Add(amount)

	return nil
}

// MockHandler creates a handler with mock repositories for testing
func NewMockHandler() *Handler {
	accountRepo := NewMockAccountRepository()
	transactionRepo := NewMockTransactionRepository(accountRepo)

	return &Handler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

func TestCreateAccountHandler_Success(t *testing.T) {
	handler := NewMockHandler()

	reqBody := models.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "100.50",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateAccount(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}
}

func TestCreateAccountHandler_InvalidJSON(t *testing.T) {
	handler := NewMockHandler()

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateAccount(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestCreateAccountHandler_NegativeBalance(t *testing.T) {
	handler := NewMockHandler()

	reqBody := models.CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "-100.00",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateAccount(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestGetAccountHandler_Success(t *testing.T) {
	handler := NewMockHandler()

	// First create an account
	handler.accountRepo.CreateAccount(123, decimal.NewFromFloat(100.50))

	req := httptest.NewRequest("GET", "/accounts/123", nil)
	req = mux.SetURLVars(req, map[string]string{"account_id": "123"})

	rr := httptest.NewRecorder()
	handler.GetAccount(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response models.AccountResponse
	json.NewDecoder(rr.Body).Decode(&response)

	if response.AccountID != 123 {
		t.Errorf("Expected AccountID 123, got %d", response.AccountID)
	}

	if response.Balance != "100.5" {
		t.Errorf("Expected balance '100.5', got '%s'", response.Balance)
	}
}

func TestGetAccountHandler_NotFound(t *testing.T) {
	handler := NewMockHandler()

	req := httptest.NewRequest("GET", "/accounts/999", nil)
	req = mux.SetURLVars(req, map[string]string{"account_id": "999"})

	rr := httptest.NewRecorder()
	handler.GetAccount(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestCreateTransactionHandler_Success(t *testing.T) {
	handler := NewMockHandler()

	// Create source and destination accounts
	handler.accountRepo.CreateAccount(123, decimal.NewFromFloat(1000.00))
	handler.accountRepo.CreateAccount(456, decimal.NewFromFloat(500.00))

	reqBody := models.CreateTransactionRequest{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               "100.25",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTransaction(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	// Verify balances were updated
	sourceAccount, _ := handler.accountRepo.GetAccount(123)
	destAccount, _ := handler.accountRepo.GetAccount(456)

	expectedSourceBalance := decimal.NewFromFloat(899.75)
	expectedDestBalance := decimal.NewFromFloat(600.25)

	if !sourceAccount.Balance.Equal(expectedSourceBalance) {
		t.Errorf("Expected source balance %s, got %s", expectedSourceBalance, sourceAccount.Balance)
	}

	if !destAccount.Balance.Equal(expectedDestBalance) {
		t.Errorf("Expected destination balance %s, got %s", expectedDestBalance, destAccount.Balance)
	}
}

func TestCreateTransactionHandler_InsufficientBalance(t *testing.T) {
	handler := NewMockHandler()

	// Create accounts with insufficient balance
	handler.accountRepo.CreateAccount(123, decimal.NewFromFloat(50.00))
	handler.accountRepo.CreateAccount(456, decimal.NewFromFloat(500.00))

	reqBody := models.CreateTransactionRequest{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               "100.00", // More than available
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestCreateTransactionHandler_SameAccount(t *testing.T) {
	handler := NewMockHandler()

	reqBody := models.CreateTransactionRequest{
		SourceAccountID:      123,
		DestinationAccountID: 123, // Same account
		Amount:               "100.00",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestCreateTransactionHandler_InvalidAmount(t *testing.T) {
	handler := NewMockHandler()

	reqBody := map[string]interface{}{
		"source_account_id":      123,
		"destination_account_id": 456,
		"amount":                 "invalid",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateTransaction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}
