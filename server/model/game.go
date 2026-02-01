package model

import (
	"time"
)

// GameSession represents an active game session
type GameSession struct {
	ID            string        `json:"id"`
	Game          string        `json:"game"`
	Seed          int64         `json:"seed"`
	StartTime     time.Time     `json:"start_time"`
	Score         int           `json:"score"`
	Lives         int           `json:"lives"`
	CurrentBall   int           `json:"current_ball"`    // 현재 공 인덱스
	BallSpawnTime int64         `json:"ball_spawn_time"` // 현재 공 생성 시간 (ms)
	Clicks        []ClickRecord `json:"clicks"`
	Status        string        `json:"status"` // "playing", "ended", "submitted"
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

// NextBallInfo represents ball information sent to client
type NextBallInfo struct {
	Index    int     `json:"index"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	IsRed    bool    `json:"isRed"`
	Duration int64   `json:"duration"` // ms
	Size     int     `json:"size"`
}

// StartGameResponse is the response for game start
type StartGameResponse struct {
	SessionID string       `json:"sessionId"`
	Seed      int64        `json:"seed"`
	StartTime int64        `json:"startTime"`
	NextBall  NextBallInfo `json:"nextBall"`
}

// ClickRequest is the request for reporting a click
type ClickRequest struct {
	SessionID   string `json:"sessionId"`
	BallIndex   int    `json:"ballIndex"`
	ClickTimeMs int64  `json:"clickTimeMs"`
}

// ClickResponse is the response for a click
type ClickResponse struct {
	Valid    bool          `json:"valid"`
	IsRed    bool          `json:"isRed"`
	Points   int           `json:"points"`
	Score    int           `json:"score"`
	Lives    int           `json:"lives"`
	GameOver bool          `json:"gameOver"`
	NextBall *NextBallInfo `json:"nextBall,omitempty"`
	Message  string        `json:"message,omitempty"`
}

// MissRequest is the request for reporting a missed ball
type MissRequest struct {
	SessionID string `json:"sessionId"`
	BallIndex int    `json:"ballIndex"`
}

// MissResponse is the response for a miss
type MissResponse struct {
	Valid    bool          `json:"valid"`
	IsRed    bool          `json:"isRed"`
	Lives    int           `json:"lives"`
	GameOver bool          `json:"gameOver"`
	NextBall *NextBallInfo `json:"nextBall,omitempty"`
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

// ========== Jump Runner Models ==========

// JumpRunnerSession represents an active Jump Runner game session
type JumpRunnerSession struct {
	ID        string    `json:"id"`
	Seed      int64     `json:"seed"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Score     int       `json:"score"`
	Status    string    `json:"status"` // "playing", "ended", "submitted"
}

// JumpRunnerStartResponse is the response for Jump Runner game start
type JumpRunnerStartResponse struct {
	SessionID string `json:"sessionId"`
	Seed      int64  `json:"seed"`
	StartTime int64  `json:"startTime"`
}

// JumpRunnerEndRequest is the request for ending Jump Runner game
type JumpRunnerEndRequest struct {
	SessionID  string `json:"sessionId"`
	PlayTimeMs int64  `json:"playTimeMs"`
	Score      int    `json:"score"`
}

// JumpRunnerEndResponse is the response for Jump Runner game end
type JumpRunnerEndResponse struct {
	Valid      bool `json:"valid"`
	FinalScore int  `json:"finalScore"`
	CanSubmit  bool `json:"canSubmit"`
}

// ========== Snake Models ==========

// SnakeSession represents an active Snake game session
type SnakeSession struct {
	ID          string    `json:"id"`
	Seed        int64     `json:"seed"`
	Level       string    `json:"level"`
	StartTime   time.Time `json:"start_time"`
	LastEatTime time.Time `json:"last_eat_time"` // 마지막 음식 먹은 시간
	EatCount    int       `json:"eat_count"`
	Score       int       `json:"score"`
	Status      string    `json:"status"` // "playing", "ended", "submitted"
}

// SnakeStartRequest is the request for Snake game start
type SnakeStartRequest struct {
	Level string `json:"level"`
}

// SnakeStartResponse is the response for Snake game start
type SnakeStartResponse struct {
	SessionID string `json:"sessionId"`
	Seed      int64  `json:"seed"`
	StartTime int64  `json:"startTime"`
}

// SnakeEatRequest is the request for reporting food eaten
type SnakeEatRequest struct {
	SessionID string `json:"sessionId"`
}

// SnakeEatResponse is the response for food eaten
type SnakeEatResponse struct {
	Valid    bool `json:"valid"`
	EatCount int  `json:"eatCount"`
	Score    int  `json:"score"`
}

// SnakeEndRequest is the request for ending Snake game
type SnakeEndRequest struct {
	SessionID string `json:"sessionId"`
}

// SnakeEndResponse is the response for Snake game end
type SnakeEndResponse struct {
	Valid      bool `json:"valid"`
	FinalScore int  `json:"finalScore"`
	CanSubmit  bool `json:"canSubmit"`
}

// ========== Memory Card Models ==========

// MemoryCardSession represents an active Memory Card game session
type MemoryCardSession struct {
	ID         string    `json:"id"`
	Seed       int64     `json:"seed"`
	Level      string    `json:"level"`
	StartTime  time.Time `json:"start_time"`
	TimeLimit  int       `json:"time_limit"`
	Cards      []int     `json:"cards"` // Card layout (emoji indices)
	MatchCount int       `json:"match_count"`
	Moves      int       `json:"moves"`
	TotalPairs int       `json:"total_pairs"`
	Score      int       `json:"score"`
	Status     string    `json:"status"` // "playing", "ended", "submitted"
}

// MemoryCardStartRequest is the request for Memory Card game start
type MemoryCardStartRequest struct {
	Level string `json:"level"`
}

// MemoryCardStartResponse is the response for Memory Card game start
type MemoryCardStartResponse struct {
	SessionID string `json:"sessionId"`
	Seed      int64  `json:"seed"`
	StartTime int64  `json:"startTime"`
	Cards     []int  `json:"cards"`
	TimeLimit int    `json:"timeLimit"`
}

// MemoryCardMatchRequest is the request for reporting a match attempt
type MemoryCardMatchRequest struct {
	SessionID string `json:"sessionId"`
	Card1     int    `json:"card1"`
	Card2     int    `json:"card2"`
}

// MemoryCardMatchResponse is the response for match attempt
type MemoryCardMatchResponse struct {
	Valid      bool `json:"valid"`
	IsMatch    bool `json:"isMatch"`
	MatchCount int  `json:"matchCount"`
	Moves      int  `json:"moves"`
}

// MemoryCardEndRequest is the request for ending Memory Card game
type MemoryCardEndRequest struct {
	SessionID string `json:"sessionId"`
}

// MemoryCardEndResponse is the response for Memory Card game end
type MemoryCardEndResponse struct {
	Valid      bool `json:"valid"`
	FinalScore int  `json:"finalScore"`
	CanSubmit  bool `json:"canSubmit"`
}

// ========== Translation Battle Models ==========

// TranslationDifficulty represents the difficulty level
type TranslationDifficulty string

const (
	DifficultyEasy   TranslationDifficulty = "easy"
	DifficultyMedium TranslationDifficulty = "medium"
	DifficultyHard   TranslationDifficulty = "hard"
)

// TranslationGameState represents the state of a translation battle game
type TranslationGameState string

const (
	TStateWaiting    TranslationGameState = "waiting"    // Waiting for opponent
	TStateReady      TranslationGameState = "ready"      // Both players joined, ready to start
	TStateCountdown  TranslationGameState = "countdown"  // Countdown before round
	TStatePlaying    TranslationGameState = "playing"    // Round in progress
	TStateEvaluating TranslationGameState = "evaluating" // AI is evaluating
	TStateResult     TranslationGameState = "result"     // Showing round result
	TStateFinished   TranslationGameState = "finished"   // Game over
)

// TranslationScore represents the score breakdown for a translation
type TranslationScore struct {
	Meaning     int    `json:"meaning"`     // 0-40
	Grammar     int    `json:"grammar"`     // 0-30
	Naturalness int    `json:"naturalness"` // 0-30
	Total       int    `json:"total"`       // 0-100
	Feedback    string `json:"feedback"`    // AI feedback
}

// TranslationRound represents a single round in the game
type TranslationRound struct {
	Number       int                `json:"number"`       // 1, 2, 3
	Sentence     string             `json:"sentence"`     // Korean sentence
	Tense        string             `json:"tense"`        // 시제: "현재형", "과거형", "미래형"
	Translations [2]string          `json:"translations"` // Player translations [player0, player1]
	Submitted    [2]bool            `json:"submitted"`    // Whether each player has submitted
	SubmitTimes  [2]time.Time       `json:"submitTimes"`  // When each player submitted
	Scores       [2]TranslationScore `json:"scores"`       // Evaluation scores
	ModelAnswer  string             `json:"modelAnswer"`  // Model answer from AI
	Winner       int                `json:"winner"`       // -1: draw, 0: player0, 1: player1
	StartTime    time.Time          `json:"startTime"`
}
