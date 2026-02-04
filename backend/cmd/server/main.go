package main

import (
	"fmt"
	"log"
	"net/http"

	"camera-dashboard-backend/internal/config"
	"camera-dashboard-backend/internal/database"
	"camera-dashboard-backend/internal/router"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected and migrations completed")

	// Setup router
	r := router.Setup(db, cfg)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
