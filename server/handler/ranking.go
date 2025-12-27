package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"mini-games/model"
	"mini-games/service"
)

const MAX_SCORE = 100000   // 점수 상한
const MAX_BODY_SIZE = 1024 // 1KB max request body

// 허용된 게임 목록 (화이트리스트)
var allowedGames = map[string]bool{
	"jump-runner":  true,
	"speed-click":  true,
	"snake":        true,
	"memory-card":  true,
}

var validNicknameRegex = regexp.MustCompile(`^[a-zA-Z0-9가-힣_\-\s]+$`)

func HandleScores(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createScore(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createScore(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, MAX_BODY_SIZE)

	var input model.ScoreInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and sanitize nickname
	input.Nickname = strings.TrimSpace(input.Nickname)
	if input.Nickname == "" || len(input.Nickname) > 20 {
		http.Error(w, "Invalid nickname", http.StatusBadRequest)
		return
	}
	if !validNicknameRegex.MatchString(input.Nickname) {
		http.Error(w, "Nickname contains invalid characters", http.StatusBadRequest)
		return
	}

	// Validate game (whitelist check)
	input.Game = strings.TrimSpace(input.Game)
	if !allowedGames[input.Game] {
		http.Error(w, "Invalid game", http.StatusBadRequest)
		return
	}

	// Validate score (prevent manipulation)
	if input.Score < 0 || input.Score > MAX_SCORE {
		http.Error(w, "Invalid score", http.StatusBadRequest)
		return
	}

	id, err := service.SaveScore(input)
	if err != nil {
		http.Error(w, "Failed to save score", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func HandleRanking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	game := strings.TrimSpace(r.URL.Query().Get("game"))
	
	// Validate game (whitelist check)
	if !allowedGames[game] {
		http.Error(w, "Invalid game parameter", http.StatusBadRequest)
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	scores, err := service.GetRanking(game, limit)
	if err != nil {
		http.Error(w, "Failed to get ranking", http.StatusInternalServerError)
		return
	}

	if scores == nil {
		scores = []model.Score{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}
