#!/bin/bash

# Internal Transfers System - API Examples
# Make sure the server is running on http://localhost:8080

set -e

BASE_URL="http://localhost:8080"
echo "🚀 Testing Internal Transfers System API"
echo "========================================"

# Function to make HTTP requests with error handling
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo ""
    echo "📡 $description"
    echo "Method: $method"
    echo "Endpoint: $endpoint"
    if [ ! -z "$data" ]; then
        echo "Data: $data"
    fi
    echo "Response:"
    
    if [ -z "$data" ]; then
        curl -s -w "\nHTTP Status: %{http_code}\n" -X "$method" "$BASE_URL$endpoint"
    else
        curl -s -w "\nHTTP Status: %{http_code}\n" -X "$method" \
             -H "Content-Type: application/json" \
             -d "$data" \
             "$BASE_URL$endpoint"
    fi
    echo ""
    echo "----------------------------------------"
}

# Health Check
make_request "GET" "/health" "" "Health Check"

# Create first account
make_request "POST" "/accounts" \
    '{"account_id": 123, "initial_balance": "1000.00"}' \
    "Create Account 123 with $1000"

# Create second account  
make_request "POST" "/accounts" \
    '{"account_id": 456, "initial_balance": "500.50"}' \
    "Create Account 456 with $500.50"

# Create third account for more tests
make_request "POST" "/accounts" \
    '{"account_id": 789, "initial_balance": "0.00"}' \
    "Create Account 789 with $0"

# Check initial balances
make_request "GET" "/accounts/123" "" "Get Account 123 Balance"
make_request "GET" "/accounts/456" "" "Get Account 456 Balance" 
make_request "GET" "/accounts/789" "" "Get Account 789 Balance"

# Transfer money from 123 to 456
make_request "POST" "/transactions" \
    '{"source_account_id": 123, "destination_account_id": 456, "amount": "100.25"}' \
    "Transfer $100.25 from Account 123 to 456"

# Check balances after first transfer
make_request "GET" "/accounts/123" "" "Account 123 Balance After Transfer"
make_request "GET" "/accounts/456" "" "Account 456 Balance After Transfer"

# Transfer money from 456 to 789
make_request "POST" "/transactions" \
    '{"source_account_id": 456, "destination_account_id": 789, "amount": "50.75"}' \
    "Transfer $50.75 from Account 456 to 789"

# Check final balances
make_request "GET" "/accounts/123" "" "Account 123 Final Balance"
make_request "GET" "/accounts/456" "" "Account 456 Final Balance"
make_request "GET" "/accounts/789" "" "Account 789 Final Balance"

echo ""
echo "🎯 Testing Error Cases"
echo "====================="

# Test insufficient balance
make_request "POST" "/transactions" \
    '{"source_account_id": 789, "destination_account_id": 123, "amount": "1000.00"}' \
    "Try to transfer $1000 from Account 789 (insufficient balance)"

# Test non-existent account
make_request "GET" "/accounts/999" "" "Try to get non-existent Account 999"

# Test invalid account creation (duplicate)
make_request "POST" "/accounts" \
    '{"account_id": 123, "initial_balance": "100.00"}' \
    "Try to create duplicate Account 123"

# Test invalid amount format
make_request "POST" "/transactions" \
    '{"source_account_id": 123, "destination_account_id": 456, "amount": "invalid"}' \
    "Try transaction with invalid amount"

# Test negative amount
make_request "POST" "/transactions" \
    '{"source_account_id": 123, "destination_account_id": 456, "amount": "-50.00"}' \
    "Try transaction with negative amount"

# Test same source and destination
make_request "POST" "/transactions" \
    '{"source_account_id": 123, "destination_account_id": 123, "amount": "50.00"}' \
    "Try transaction to same account"

echo ""
echo "✅ API Testing Complete!"
echo ""
echo "Summary of Expected Final Balances:"
echo "- Account 123: $899.75 (1000 - 100.25)"
echo "- Account 456: $549.75 (500.50 + 100.25 - 50.75)"  
echo "- Account 789: $50.75 (0 + 50.75)"
echo ""
echo "Use 'curl http://localhost:8080/accounts/{id}' to verify balances" 