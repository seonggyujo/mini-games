package main

import (
	"log"
	"net/http"
	"os"

	"mini-games/database"
	"mini-games/handler"
	"mini-games/middleware"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Get port from environment variable
	port := getEnv("PORT", "4001")

	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Initialize WebSocket handler
	handler.InitWebSocket()

	// Create router
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/scores", handler.HandleScores)
	mux.HandleFunc("/api/ranking", handler.HandleRanking)

	// SpeedClick game API routes
	mux.HandleFunc("/api/game/speedclick/start", handler.HandleSpeedClickStart)
	mux.HandleFunc("/api/game/speedclick/click", handler.HandleSpeedClickClick)
	mux.HandleFunc("/api/game/speedclick/miss", handler.HandleSpeedClickMiss)
	mux.HandleFunc("/api/game/speedclick/end", handler.HandleSpeedClickEnd)
	mux.HandleFunc("/api/game/speedclick/submit", handler.HandleSpeedClickSubmit)

	// WebSocket route (no rate limiting for WebSocket)
	mux.HandleFunc("/ws/battle", handler.HandleBattleWS)

	// Apply middleware (order: Logging -> CORS -> RateLimit)
	h := middleware.Logging(middleware.CORS(middleware.RateLimit(mux)))

	// Start server
	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Fatal("Server failed:", err)
	}
}
