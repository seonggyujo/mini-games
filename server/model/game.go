package model

import (
	"time"
)

// GameSession represents an active game session
type GameSession struct {
	ID           string        `json:"id"`
	Game         string        `json:"game"`
	Seed         int64         `json:"seed"`
	StartTime    time.Time     `json:"start_time"`
	Score        int           `json:"score"`
	Lives        int           `json:"lives"`
	CurrentBall  int           `json:"current_ball"`  // 현재 공 인덱스
	BallSpawnTime int64        `json:"ball_spawn_time"` // 현재 공 생성 시간 (ms)
	Clicks       []ClickRecord `json:"clicks"`
	Status       string        `json:"status"` // "playing", "ended", "submitted"
}

// ClickRecord represents a single click during a game
type ClickRecord struct {
	BallIndex int   `json:"ball_index"`
	ClickTime int64 `json:"click_time"` // ms since game start
	Valid     bool  `json:"valid"`
	Points    int   `json:"points"`
}

// SpeedClickBall represents a ball in SpeedClick game
type SpeedClickBall struct {
	Index     int     `json:"index"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	IsRed     bool    `json:"is_red"`
	SpawnTime int64   `json:"spawn_time"` // ms since game start
	Duration  int64   `json:"duration"`   // ms
	Size      int     `json:"size"`
	Level     int     `json:"level"`
}

// SpeedClickLevel configuration
type SpeedClickLevel struct {
	Level         int
	TimeLimit     float64 // seconds
	BallSize      int
	BlueChance    float64
	RequiredScore int
}

// StartGameResponse is the response for game start
type StartGameResponse struct {
	SessionID string `json:"sessionId"`
	Seed      int64  `json:"seed"`
	StartTime int64  `json:"startTime"`
}

// ClickRequest is the request for reporting a click
type ClickRequest struct {
	SessionID   string `json:"sessionId"`
	BallIndex   int    `json:"ballIndex"`
	ClickTimeMs int64  `json:"clickTimeMs"`
}

// ClickResponse is the response for a click
type ClickResponse struct {
	Valid    bool   `json:"valid"`
	Points   int    `json:"points"`
	Score    int    `json:"score"`
	Lives    int    `json:"lives"`
	GameOver bool   `json:"gameOver"`
	Message  string `json:"message,omitempty"`
}

// MissRequest is the request for reporting a missed ball
type MissRequest struct {
	SessionID string `json:"sessionId"`
	BallIndex int    `json:"ballIndex"`
}

// MissResponse is the response for a miss
type MissResponse struct {
	Valid    bool `json:"valid"`
	Lives    int  `json:"lives"`
	GameOver bool `json:"gameOver"`
}

// EndGameRequest is the request for ending a game
type EndGameRequest struct {
	SessionID string `json:"sessionId"`
}

// EndGameResponse is the response for game end
type EndGameResponse struct {
	FinalScore int  `json:"finalScore"`
	CanSubmit  bool `json:"canSubmit"`
}

// SubmitScoreRequest is the request for submitting score
type SubmitScoreRequest struct {
	SessionID string `json:"sessionId"`
	Nickname  string `json:"nickname"`
}

// SubmitScoreResponse is the response for score submission
type SubmitScoreResponse struct {
	Success bool  `json:"success"`
	ScoreID int64 `json:"scoreId,omitempty"`
}
