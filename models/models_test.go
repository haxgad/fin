package models

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestAccountModel(t *testing.T) {
	account := Account{
		AccountID: 123,
		Balance:   decimal.NewFromFloat(100.50),
	}

	if account.AccountID != 123 {
		t.Errorf("Expected AccountID 123, got %d", account.AccountID)
	}

	expectedBalance := decimal.NewFromFloat(100.50)
	if !account.Balance.Equal(expectedBalance) {
		t.Errorf("Expected balance %s, got %s", expectedBalance, account.Balance)
	}
}

func TestCreateAccountRequest(t *testing.T) {
	req := CreateAccountRequest{
		AccountID:      456,
		InitialBalance: "250.75",
	}

	if req.AccountID != 456 {
		t.Errorf("Expected AccountID 456, got %d", req.AccountID)
	}

	if req.InitialBalance != "250.75" {
		t.Errorf("Expected InitialBalance '250.75', got '%s'", req.InitialBalance)
	}
}

func TestAccountResponse(t *testing.T) {
	resp := AccountResponse{
		AccountID: 789,
		Balance:   "1000.00",
	}

	if resp.AccountID != 789 {
		t.Errorf("Expected AccountID 789, got %d", resp.AccountID)
	}

	if resp.Balance != "1000.00" {
		t.Errorf("Expected Balance '1000.00', got '%s'", resp.Balance)
	}
}

func TestCreateTransactionRequest(t *testing.T) {
	req := CreateTransactionRequest{
		SourceAccountID:      123,
		DestinationAccountID: 456,
		Amount:               "100.50",
	}

	if req.SourceAccountID != 123 {
		t.Errorf("Expected SourceAccountID 123, got %d", req.SourceAccountID)
	}

	if req.DestinationAccountID != 456 {
		t.Errorf("Expected DestinationAccountID 456, got %d", req.DestinationAccountID)
	}

	if req.Amount != "100.50" {
		t.Errorf("Expected Amount '100.50', got '%s'", req.Amount)
	}
}
