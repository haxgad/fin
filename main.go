package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"internal-transfers/database"
	"internal-transfers/handlers"
)

// getPort returns the port to listen on, defaulting to 8080
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// setupRoutes configures and returns the HTTP router with all endpoints
func setupRoutes(h *handlers.Handler) *mux.Router {
	r := mux.NewRouter()

	// Account endpoints
	r.HandleFunc("/accounts", h.CreateAccount).Methods("POST")
	r.HandleFunc("/accounts/{account_id}", h.GetAccount).Methods("GET")

	// Transaction endpoint
	r.HandleFunc("/transactions", h.CreateTransaction).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")

	return r
}

// initializeApp initializes the database connection, runs migrations, and returns a handler
func initializeApp() (*handlers.Handler, error) {
	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		return nil, err
	}

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	// Initialize handlers
	h := handlers.NewHandler(db)
	return h, nil
}

func main() {
	// Initialize the application
	h, err := initializeApp()
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Setup routes
	r := setupRoutes(h)

	// Get port
	port := getPort()

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
