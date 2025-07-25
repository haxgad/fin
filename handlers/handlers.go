package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"internal-transfers/database"
	"internal-transfers/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// Handler contains the dependencies for HTTP handlers
type Handler struct {
	accountRepo     database.AccountRepositoryInterface
	transactionRepo database.TransactionRepositoryInterface
}

// NewHandler creates a new handler with database repositories
// This is the constructor that injects database dependencies into handlers
// Parameters:
//   - db: SQL database connection used to create repository instances
//
// Returns: Configured Handler with account and transaction repositories
func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		accountRepo:     database.NewAccountRepository(db),
		transactionRepo: database.NewTransactionRepository(db),
	}
}

// CreateAccount handles POST /accounts endpoint for creating new bank accounts
// This endpoint allows creation of new accounts with an initial balance
// Request body: JSON with account_id (int64) and initial_balance (string decimal)
// Validation rules:
//   - Account ID must be positive
//   - Initial balance must be valid decimal format and non-negative
//   - Account ID must not already exist in the system
//
// Response: 201 Created on success, various 4xx/5xx on validation/server errors
// Example request: {"account_id": 123, "initial_balance": "100.50"}
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate account ID
	if req.AccountID <= 0 {
		http.Error(w, "Account ID must be positive", http.StatusBadRequest)
		return
	}

	// Parse initial balance
	initialBalance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		http.Error(w, "Invalid initial balance format", http.StatusBadRequest)
		return
	}

	// Validate initial balance is non-negative
	if initialBalance.IsNegative() {
		http.Error(w, "Initial balance cannot be negative", http.StatusBadRequest)
		return
	}

	// Check if account already exists
	exists, err := h.accountRepo.AccountExists(req.AccountID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Account already exists", http.StatusConflict)
		return
	}

	// Create account
	if err := h.accountRepo.CreateAccount(req.AccountID, initialBalance); err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetAccount handles GET /accounts/{account_id} endpoint for retrieving account information
// This endpoint returns the current balance and details for a specific account
// URL parameter: account_id (int64) - the ID of the account to retrieve
// Validation rules:
//   - Account ID must be a valid integer
//   - Account must exist in the system
//
// Response: JSON with account_id and current balance on success, 404 if not found
// Example response: {"account_id": 123, "balance": "100.50"}
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountIDStr := vars["account_id"]

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	account, err := h.accountRepo.GetAccount(accountID)
	if err != nil {
		if err.Error() == "account not found" {
			http.Error(w, "Account not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := models.AccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateTransaction handles POST /transactions endpoint for transferring money between accounts
// This endpoint performs atomic money transfers with balance validation
// Request body: JSON with source_account_id, destination_account_id, and amount
// Business rules:
//   - Both account IDs must be positive and different from each other
//   - Amount must be positive decimal value
//   - Source account must have sufficient balance
//   - Both accounts must exist in the system
//
// Response: 201 Created on success, various 4xx/5xx on validation/business rule violations
// Example request: {"source_account_id": 123, "destination_account_id": 456, "amount": "50.00"}
// Note: This operation is atomic - either both account balances are updated or neither
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate account IDs
	if req.SourceAccountID <= 0 || req.DestinationAccountID <= 0 {
		http.Error(w, "Account IDs must be positive", http.StatusBadRequest)
		return
	}

	// Validate that source and destination accounts are different
	if req.SourceAccountID == req.DestinationAccountID {
		http.Error(w, "Source and destination accounts must be different", http.StatusBadRequest)
		return
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		http.Error(w, "Invalid amount format", http.StatusBadRequest)
		return
	}

	// Validate amount is positive
	if amount.IsZero() || amount.IsNegative() {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	// Create transaction
	if err := h.transactionRepo.CreateTransaction(req.SourceAccountID, req.DestinationAccountID, amount); err != nil {
		switch err.Error() {
		case "source account not found":
			http.Error(w, "Source account not found", http.StatusNotFound)
		case "destination account not found":
			http.Error(w, "Destination account not found", http.StatusNotFound)
		case "insufficient balance":
			http.Error(w, "Insufficient balance", http.StatusBadRequest)
		default:
			fmt.Printf("Transaction error: %v\n", err)
			http.Error(w, "Failed to process transaction", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// HealthCheck handles GET /health endpoint for service health monitoring
// This endpoint provides a simple health check for load balancers and monitoring systems
// No parameters required
// Response: Always returns 200 OK with JSON status message
// Example response: {"status": "healthy"}
// Note: This is a basic health check - could be enhanced to check database connectivity
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
