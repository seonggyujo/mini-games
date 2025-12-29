package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"mini-games/model"
	"mini-games/service"
)

var validNicknameRegexGame = regexp.MustCompile(`^[a-zA-Z0-9가-힣_\-\s]+$`)

// HandleSpeedClickStart handles POST /api/game/speedclick/start
func HandleSpeedClickStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, firstBall := service.CreateSpeedClickSession()

	response := model.StartGameResponse{
		SessionID: session.ID,
		Seed:      session.Seed,
		StartTime: session.StartTime.UnixMilli(),
		NextBall:  firstBall,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSpeedClickClick handles POST /api/game/speedclick/click
func HandleSpeedClickClick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.ClickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.ProcessClick(req.SessionID, req.BallIndex, req.ClickTimeMs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSpeedClickMiss handles POST /api/game/speedclick/miss
func HandleSpeedClickMiss(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.MissRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.ProcessMiss(req.SessionID, req.BallIndex)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSpeedClickEnd handles POST /api/game/speedclick/end
func HandleSpeedClickEnd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.EndGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.EndSpeedClickSession(req.SessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSpeedClickSubmit handles POST /api/game/speedclick/submit
func HandleSpeedClickSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SubmitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Validate nickname
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" || len(req.Nickname) > 20 {
		http.Error(w, "Invalid nickname", http.StatusBadRequest)
		return
	}
	if !validNicknameRegexGame.MatchString(req.Nickname) {
		http.Error(w, "Nickname contains invalid characters", http.StatusBadRequest)
		return
	}

	response := service.SubmitSpeedClickScore(req.SessionID, req.Nickname)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ========== Jump Runner Handlers ==========

// HandleJumpRunnerStart handles POST /api/game/jumprunner/start
func HandleJumpRunnerStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := service.CreateJumpRunnerSession()

	response := model.JumpRunnerStartResponse{
		SessionID: session.ID,
		Seed:      session.Seed,
		StartTime: session.StartTime.UnixMilli(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleJumpRunnerEnd handles POST /api/game/jumprunner/end
func HandleJumpRunnerEnd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.JumpRunnerEndRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.EndJumpRunnerSession(req.SessionID, req.PlayTimeMs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleJumpRunnerSubmit handles POST /api/game/jumprunner/submit
func HandleJumpRunnerSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SubmitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" || len(req.Nickname) > 20 {
		http.Error(w, "Invalid nickname", http.StatusBadRequest)
		return
	}
	if !validNicknameRegexGame.MatchString(req.Nickname) {
		http.Error(w, "Nickname contains invalid characters", http.StatusBadRequest)
		return
	}

	response := service.SubmitJumpRunnerScore(req.SessionID, req.Nickname)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ========== Snake Handlers ==========

// HandleSnakeStart handles POST /api/game/snake/start
func HandleSnakeStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SnakeStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := service.CreateSnakeSession(req.Level)

	response := model.SnakeStartResponse{
		SessionID: session.ID,
		Seed:      session.Seed,
		StartTime: session.StartTime.UnixMilli(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSnakeEat handles POST /api/game/snake/eat
func HandleSnakeEat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SnakeEatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.ProcessSnakeEat(req.SessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSnakeEnd handles POST /api/game/snake/end
func HandleSnakeEnd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SnakeEndRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.EndSnakeSession(req.SessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSnakeSubmit handles POST /api/game/snake/submit
func HandleSnakeSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SubmitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" || len(req.Nickname) > 20 {
		http.Error(w, "Invalid nickname", http.StatusBadRequest)
		return
	}
	if !validNicknameRegexGame.MatchString(req.Nickname) {
		http.Error(w, "Nickname contains invalid characters", http.StatusBadRequest)
		return
	}

	response := service.SubmitSnakeScore(req.SessionID, req.Nickname)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ========== Memory Card Handlers ==========

// HandleMemoryCardStart handles POST /api/game/memorycard/start
func HandleMemoryCardStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.MemoryCardStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, cards := service.CreateMemoryCardSession(req.Level)

	response := model.MemoryCardStartResponse{
		SessionID: session.ID,
		Seed:      session.Seed,
		StartTime: session.StartTime.UnixMilli(),
		Cards:     cards,
		TimeLimit: session.TimeLimit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleMemoryCardMatch handles POST /api/game/memorycard/match
func HandleMemoryCardMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.MemoryCardMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.ProcessMemoryCardMatch(req.SessionID, req.Card1, req.Card2)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleMemoryCardEnd handles POST /api/game/memorycard/end
func HandleMemoryCardEnd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.MemoryCardEndRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	response := service.EndMemoryCardSession(req.SessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleMemoryCardSubmit handles POST /api/game/memorycard/submit
func HandleMemoryCardSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SubmitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Nickname == "" || len(req.Nickname) > 20 {
		http.Error(w, "Invalid nickname", http.StatusBadRequest)
		return
	}
	if !validNicknameRegexGame.MatchString(req.Nickname) {
		http.Error(w, "Nickname contains invalid characters", http.StatusBadRequest)
		return
	}

	response := service.SubmitMemoryCardScore(req.SessionID, req.Nickname)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
