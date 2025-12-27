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

	session := service.CreateSpeedClickSession()

	response := model.StartGameResponse{
		SessionID: session.ID,
		Seed:      session.Seed,
		StartTime: session.StartTime.UnixMilli(),
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
