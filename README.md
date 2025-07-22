# Internal Transfers System

A high-performance financial transaction system built with Go that facilitates secure money transfers between accounts through HTTP endpoints.

## Features

- **Account Management**: Create accounts with initial balances and query account information
- **Money Transfers**: Secure atomic transactions between accounts with balance validation
- **Data Integrity**: ACID-compliant transactions using PostgreSQL
- **High Precision**: Decimal arithmetic for accurate financial calculations
- **Comprehensive Error Handling**: Detailed validation and error responses
- **Health Monitoring**: Built-in health check endpoint

## API Endpoints

### Account Management

#### Create Account
```http
POST /accounts
Content-Type: application/json

{
  "account_id": 123,
  "initial_balance": "100.23344"
}
```

#### Get Account Balance
```http
GET /accounts/{account_id}
```

Response:
```json
{
  "account_id": 123,
  "balance": "100.23344"
}
```

### Transactions

#### Transfer Money
```http
POST /transactions
Content-Type: application/json

{
  "source_account_id": 123,
  "destination_account_id": 456,
  "amount": "100.12345"
}
```

### Health Check
```http
GET /health
```

## Installation & Setup

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Git

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd internal-transfers
   ```

2. **Start PostgreSQL database**
   ```bash
   docker-compose up -d
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the application**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

### Environment Variables

The application supports the following environment variables for database configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `transfers` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |

### Custom Database Setup

If you prefer to use your own PostgreSQL instance:

```bash
# Set environment variables
export DB_HOST=your-db-host
export DB_PORT=5432
export DB_USER=your-username
export DB_PASSWORD=your-password
export DB_NAME=transfers

# Run the application
go run main.go
```

## Usage Examples

### Create Two Accounts
```bash
# Create first account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 123, "initial_balance": "1000.00"}'

# Create second account
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 456, "initial_balance": "500.00"}'
```

### Check Account Balances
```bash
# Check first account
curl http://localhost:8080/accounts/123

# Check second account
curl http://localhost:8080/accounts/456
```

### Transfer Money
```bash
# Transfer $100 from account 123 to account 456
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 123, "destination_account_id": 456, "amount": "100.00"}'
```

### Verify Transfer
```bash
# Check balances after transfer
curl http://localhost:8080/accounts/123  # Should show 900.00
curl http://localhost:8080/accounts/456  # Should show 600.00
```

## Architecture

### Database Schema

**Accounts Table**
```sql
CREATE TABLE accounts (
    account_id BIGINT PRIMARY KEY,
    balance DECIMAL(15,5) NOT NULL CHECK (balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Transactions Table**
```sql
CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    source_account_id BIGINT NOT NULL,
    destination_account_id BIGINT NOT NULL,
    amount DECIMAL(15,5) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (source_account_id) REFERENCES accounts(account_id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(account_id),
    CHECK (source_account_id != destination_account_id)
);
```

### Project Structure
```
internal-transfers/
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── docker-compose.yml      # PostgreSQL setup
├── handlers/               # HTTP request handlers
│   └── handlers.go
├── models/                 # Data models
│   ├── account.go
│   └── transaction.go
└── database/               # Database layer
    ├── db.go              # Database connection
    ├── migrations.go      # Schema migrations
    └── queries.go         # Repository pattern
```

## Technical Implementation

### Concurrency & Data Safety
- **Row-level locking** with `SELECT ... FOR UPDATE` prevents race conditions
- **Database transactions** ensure atomic operations
- **Proper error handling** for all edge cases
- **Decimal precision** using `shopspring/decimal` for financial accuracy

### Error Handling
The system provides comprehensive error handling for:
- Invalid request formats
- Account not found scenarios
- Insufficient balance conditions
- Duplicate account creation
- Database connection issues
- Transaction failures

### Performance Features
- **Database indexes** on frequently queried fields
- **Connection pooling** for efficient database usage
- **Minimal dependencies** for fast startup
- **Clean architecture** for maintainability

## Assumptions

1. **Single Currency**: All accounts operate in the same currency
2. **Account IDs**: Positive integers used as account identifiers
3. **Precision**: Financial amounts support up to 5 decimal places
4. **Authentication**: No authentication/authorization implemented (internal system)
5. **Idempotency**: Transactions are not idempotent (each request creates a new transaction)

## Development

### IDE Setup (VS Code)

The project is fully configured for VS Code development:

1. **Install recommended extensions** when prompted, or install manually:
   - Go extension
   - REST Client (for API testing)
   - Docker extension

2. **Run from IDE**:
   - Press `F5` or go to Run and Debug view
   - Select "Launch Internal Transfers Server"
   - The application will start with database automatically

3. **Available debug configurations**:
   - **Launch Internal Transfers Server**: Default setup with localhost database
   - **Launch with Custom DB**: Prompts for database connection details
   - **Debug Tests**: Run tests in debug mode

4. **Tasks available** (Ctrl+Shift+P → "Tasks: Run Task"):
   - `Go: Build Application` - Build the binary
   - `Go: Run Application` - Run with automatic database startup
   - `Go: Test All` - Run all tests
   - `Docker: Start Database` - Start PostgreSQL container
   - `Docker: Stop Database` - Stop PostgreSQL container
   - `Test API` - Run the API test script

5. **API Testing**:
   - Open `api_test.http` file
   - Click "Send Request" on any HTTP request to test endpoints
   - Make sure the server is running first

### Environment Configuration

Copy `env.example` to `.env` and modify as needed:
```bash
cp env.example .env
```

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o transfers main.go
./transfers
```

### Docker Build
```bash
docker build -t internal-transfers .
docker run -p 8080:8080 --env-file .env internal-transfers
```

## Troubleshooting

### Debug Issues in VS Code

If you encounter **"Failed to launch dlv: Error: timed out while waiting for DAP"**:

1. **Try different debug configurations**:
   - Use **"Launch Server (Simple)"** instead of the main one
   - Try **"Run Without Debug"** for basic execution

2. **Update Go tools**:
   ```bash
   go install github.com/go-delve/delve/cmd/dlv@latest
   go install golang.org/x/tools/gopls@latest
   ```

3. **Restart VS Code** after installing updates

4. **Alternative: Use Terminal**:
   ```bash
   # Start database
   docker-compose up -d

   # Run application
   go run main.go
   ```

### Database Connection Issues
1. Ensure PostgreSQL is running: `docker-compose ps`
2. Check database logs: `docker-compose logs postgres`
3. Verify connection parameters

### Application Errors
1. Check application logs for detailed error messages
2. Verify database schema is properly migrated
3. Ensure all required environment variables are set

### Health Check
Use the health endpoint to verify system status:
```bash
curl http://localhost:8080/health
```
