package main

import (
	"internal-transfers/database"
	"internal-transfers/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Setup routes
	r := mux.NewRouter()

	// Account endpoints
	r.HandleFunc("/accounts", h.CreateAccount).Methods("POST")
	r.HandleFunc("/accounts/{account_id}", h.GetAccount).Methods("GET")

	// Transaction endpoint
	r.HandleFunc("/transactions", h.CreateTransaction).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
