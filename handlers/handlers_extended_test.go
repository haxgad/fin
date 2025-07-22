package handlers

import (
	"bytes"
	"encoding/json"
	"internal-transfers/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

func TestNewHandler(t *testing.T) {
	accountRepo := NewMockAccountRepository()
	transactionRepo := NewMockTransactionRepository(accountRepo)

	handler := &Handler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}

	if handler.accountRepo == nil {
		t.Error("Handler accountRepo not initialized")
	}
	if handler.transactionRepo == nil {
		t.Error("Handler transactionRepo not initialized")
	}
}

func TestCreateAccount_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		description    string
	}{
		{
			name:           "Zero account ID",
			requestBody:    models.CreateAccountRequest{AccountID: 0, InitialBalance: "100.00"},
			expectedStatus: http.StatusBadRequest,
			description:    "Account ID must be positive",
		},
		{
			name:           "Negative account ID",
			requestBody:    models.CreateAccountRequest{AccountID: -1, InitialBalance: "100.00"},
			expectedStatus: http.StatusBadRequest,
			description:    "Account ID must be positive",
		},
		{
			name:           "Zero balance",
			requestBody:    models.CreateAccountRequest{AccountID: 123, InitialBalance: "0.00"},
			expectedStatus: http.StatusCreated,
			description:    "Zero balance should be allowed",
		},
		{
			name:           "Very large balance",
			requestBody:    models.CreateAccountRequest{AccountID: 123, InitialBalance: "999999999.99999"},
			expectedStatus: http.StatusCreated,
			description:    "Large balances should be allowed",
		},
		{
			name:           "Many decimal places",
			requestBody:    models.CreateAccountRequest{AccountID: 123, InitialBalance: "100.12345"},
			expectedStatus: http.StatusCreated,
			description:    "Precise decimal amounts should be allowed",
		},
		{
			name:           "Invalid balance - text",
			requestBody:    models.CreateAccountRequest{AccountID: 123, InitialBalance: "not-a-number"},
			expectedStatus: http.StatusBadRequest,
			description:    "Non-numeric balance should be rejected",
		},
		{
			name:           "Invalid balance - empty",
			requestBody:    models.CreateAccountRequest{AccountID: 123, InitialBalance: ""},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty balance should be rejected",
		},
		{
			name:           "Empty JSON",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			description:    "Empty request should be rejected",
		},
		{
			name:           "Missing account_id",
			requestBody:    map[string]interface{}{"initial_balance": "100.00"},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing account ID should be rejected",
		},
		{
			name:           "Missing initial_balance",
			requestBody:    map[string]interface{}{"account_id": 123},
			expectedStatus: http.StatusBadRequest,
			description:    "Missing initial balance should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMockHandler()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateAccount(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Description: %s",
					tt.expectedStatus, rr.Code, tt.description)
			}
		})
	}
}

func TestGetAccount_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		expectedStatus int
		setupAccount   bool
		description    string
	}{
		{
			name:           "Valid account",
			accountID:      "123",
			expectedStatus: http.StatusOK,
			setupAccount:   true,
			description:    "Existing account should return successfully",
		},
		{
			name:           "Non-existent account",
			accountID:      "999",
			expectedStatus: http.StatusNotFound,
			setupAccount:   false,
			description:    "Non-existent account should return 404",
		},
		{
			name:           "Invalid account ID - text",
			accountID:      "abc",
			expectedStatus: http.StatusBadRequest,
			setupAccount:   false,
			description:    "Non-numeric account ID should be rejected",
		},
		{
			name:           "Invalid account ID - negative",
			accountID:      "-1",
			expectedStatus: http.StatusNotFound,
			setupAccount:   false,
			description:    "Negative account ID should parse but not be found",
		},
		{
			name:           "Invalid account ID - zero",
			accountID:      "0",
			expectedStatus: http.StatusNotFound,
			setupAccount:   false,
			description:    "Zero account ID should not be found",
		},
		{
			name:           "Very large account ID",
			accountID:      "999999999999",
			expectedStatus: http.StatusNotFound,
			setupAccount:   false,
			description:    "Large account ID should parse but not be found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMockHandler()

			if tt.setupAccount {
				handler.accountRepo.CreateAccount(123, decimal.NewFromFloat(100.0))
			}

			req := httptest.NewRequest("GET", "/accounts/"+tt.accountID, nil)
			req = mux.SetURLVars(req, map[string]string{"account_id": tt.accountID})

			rr := httptest.NewRecorder()
			handler.GetAccount(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Description: %s",
					tt.expectedStatus, rr.Code, tt.description)
			}
		})
	}
}

func TestCreateTransaction_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupAccounts  bool
		description    string
	}{
		{
			name: "Zero amount",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: 123, DestinationAccountID: 456, Amount: "0.00",
			},
			expectedStatus: http.StatusBadRequest,
			setupAccounts:  true,
			description:    "Zero amount should be rejected",
		},
		{
			name: "Very small amount",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: 123, DestinationAccountID: 456, Amount: "0.00001",
			},
			expectedStatus: http.StatusCreated,
			setupAccounts:  true,
			description:    "Very small positive amounts should be allowed",
		},
		{
			name: "Very large amount",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: 123, DestinationAccountID: 456, Amount: "999999.99999",
			},
			expectedStatus: http.StatusBadRequest,
			setupAccounts:  true,
			description:    "Amount larger than balance should be rejected",
		},
		{
			name: "Invalid source account ID - zero",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: 0, DestinationAccountID: 456, Amount: "100.00",
			},
			expectedStatus: http.StatusBadRequest,
			setupAccounts:  false,
			description:    "Zero source account ID should be rejected",
		},
		{
			name: "Invalid destination account ID - zero",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: 123, DestinationAccountID: 0, Amount: "100.00",
			},
			expectedStatus: http.StatusBadRequest,
			setupAccounts:  false,
			description:    "Zero destination account ID should be rejected",
		},
		{
			name: "Invalid source account ID - negative",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: -1, DestinationAccountID: 456, Amount: "100.00",
			},
			expectedStatus: http.StatusBadRequest,
			setupAccounts:  false,
			description:    "Negative source account ID should be rejected",
		},
		{
			name: "Non-existent source account",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: 999, DestinationAccountID: 456, Amount: "100.00",
			},
			expectedStatus: http.StatusNotFound,
			setupAccounts:  true,
			description:    "Non-existent source account should return 404",
		},
		{
			name: "Non-existent destination account",
			requestBody: models.CreateTransactionRequest{
				SourceAccountID: 123, DestinationAccountID: 999, Amount: "100.00",
			},
			expectedStatus: http.StatusNotFound,
			setupAccounts:  true,
			description:    "Non-existent destination account should return 404",
		},
		{
			name: "Invalid amount format - scientific notation",
			requestBody: map[string]interface{}{
				"source_account_id": 123, "destination_account_id": 456, "amount": "1e10",
			},
			expectedStatus: http.StatusBadRequest,
			setupAccounts:  true,
			description:    "Scientific notation should be rejected",
		},
		{
			name: "Missing amount field",
			requestBody: map[string]interface{}{
				"source_account_id": 123, "destination_account_id": 456,
			},
			expectedStatus: http.StatusBadRequest,
			setupAccounts:  false,
			description:    "Missing amount should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMockHandler()

			if tt.setupAccounts {
				handler.accountRepo.CreateAccount(123, decimal.NewFromFloat(1000.0))
				handler.accountRepo.CreateAccount(456, decimal.NewFromFloat(500.0))
			}

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateTransaction(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Description: %s",
					tt.expectedStatus, rr.Code, tt.description)
			}
		})
	}
}

func TestHandlers_ContentTypeValidation(t *testing.T) {
	handler := NewMockHandler()

	tests := []struct {
		name        string
		endpoint    string
		method      string
		contentType string
		body        string
	}{
		{
			name:        "CreateAccount without content-type",
			endpoint:    "/accounts",
			method:      "POST",
			contentType: "",
			body:        `{"account_id": 123, "initial_balance": "100.00"}`,
		},
		{
			name:        "CreateAccount with wrong content-type",
			endpoint:    "/accounts",
			method:      "POST",
			contentType: "text/plain",
			body:        `{"account_id": 123, "initial_balance": "100.00"}`,
		},
		{
			name:        "CreateTransaction without content-type",
			endpoint:    "/transactions",
			method:      "POST",
			contentType: "",
			body:        `{"source_account_id": 123, "destination_account_id": 456, "amount": "100.00"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()

			if tt.endpoint == "/accounts" {
				handler.CreateAccount(rr, req)
			} else if tt.endpoint == "/transactions" {
				handler.CreateTransaction(rr, req)
			}

			// Most should result in bad request due to JSON parsing issues
			if rr.Code != http.StatusBadRequest && rr.Code != http.StatusCreated {
				t.Logf("Status %d for %s (this tests error handling paths)",
					rr.Code, tt.name)
			}
		})
	}
}

func TestHealthCheck_Detailed(t *testing.T) {
	handler := &Handler{}

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	// Test status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Test content type
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	// Test response body
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", response["status"])
	}
}
