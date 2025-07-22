package database

import (
	"os"
	"testing"
)

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
