package main

import (
	"log"
	"net/http"

	"mini-games/database"
	"mini-games/handler"
	"mini-games/middleware"
)

func main() {
	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Create router
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/scores", handler.HandleScores)
	mux.HandleFunc("/api/ranking", handler.HandleRanking)

	// Apply middleware
	handler := middleware.CORS(middleware.RateLimit(mux))

	// Start server
	log.Println("Server starting on :4001")
	if err := http.ListenAndServe(":4001", handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
