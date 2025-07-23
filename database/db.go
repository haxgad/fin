package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// InitDB initializes and returns a PostgreSQL database connection
// This function sets up the database connection using environment variables with sensible defaults
// Environment variables used (with defaults):
//   - DB_HOST (localhost): Database server hostname
//   - DB_PORT (5432): Database server port
//   - DB_USER (postgres): Database username
//   - DB_PASSWORD (postgres): Database password
//   - DB_NAME (transfers): Database name
//   - DB_SSLMODE (disable): SSL mode for connection
//
// Returns:
//   - *sql.DB: Active database connection if successful
//   - error: Connection error if database is unreachable or credentials invalid
//
// Note: This function also performs a ping test to verify the connection is working
func InitDB() (*sql.DB, error) {
	host := getEnvWithDefault("DB_HOST", "localhost")
	port := getEnvWithDefault("DB_PORT", "5432")
	user := getEnvWithDefault("DB_USER", "postgres")
	password := getEnvWithDefault("DB_PASSWORD", "postgres")
	dbname := getEnvWithDefault("DB_NAME", "transfers")
	sslmode := getEnvWithDefault("DB_SSLMODE", "disable")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// getEnvWithDefault retrieves an environment variable value or returns a default value if not set
// This utility function provides a clean way to handle optional environment configuration
// Parameters:
//   - key: The environment variable name to look up
//   - defaultValue: The value to return if the environment variable is not set or empty
//
// Returns: The environment variable value if set and non-empty, otherwise the default value
// Note: Empty string environment variables are treated as "not set" and will return the default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
