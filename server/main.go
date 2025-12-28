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

	// Create router for API routes
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

	// JumpRunner game API routes
	mux.HandleFunc("/api/game/jumprunner/start", handler.HandleJumpRunnerStart)
	mux.HandleFunc("/api/game/jumprunner/end", handler.HandleJumpRunnerEnd)
	mux.HandleFunc("/api/game/jumprunner/submit", handler.HandleJumpRunnerSubmit)

	// Snake game API routes
	mux.HandleFunc("/api/game/snake/start", handler.HandleSnakeStart)
	mux.HandleFunc("/api/game/snake/eat", handler.HandleSnakeEat)
	mux.HandleFunc("/api/game/snake/end", handler.HandleSnakeEnd)
	mux.HandleFunc("/api/game/snake/submit", handler.HandleSnakeSubmit)

	// MemoryCard game API routes
	mux.HandleFunc("/api/game/memorycard/start", handler.HandleMemoryCardStart)
	mux.HandleFunc("/api/game/memorycard/match", handler.HandleMemoryCardMatch)
	mux.HandleFunc("/api/game/memorycard/end", handler.HandleMemoryCardEnd)
	mux.HandleFunc("/api/game/memorycard/submit", handler.HandleMemoryCardSubmit)

	// Apply middleware to API routes (order: Logging -> CORS -> RateLimit)
	apiHandler := middleware.Logging(middleware.CORS(middleware.RateLimit(mux)))

	// Main router - WebSocket without middleware, API with middleware
	mainMux := http.NewServeMux()
	mainMux.HandleFunc("/ws/battle", handler.HandleBattleWS)
	mainMux.Handle("/api/", apiHandler)

	// Start server
	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mainMux); err != nil {
		log.Fatal("Server failed:", err)
	}
}
