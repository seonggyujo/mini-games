package model

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Player represents a player in a battle room
type Player struct {
	Conn       *websocket.Conn
	Nickname   string
	Score      int
	Ready      bool // Ready for rematch
	Index      int  // 0 or 1
	LastActive time.Time
	mu         sync.Mutex
}

// SendJSON sends a JSON message to the player
func (p *Player) SendJSON(v interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Conn == nil {
		return nil
	}
	return p.Conn.WriteJSON(v)
}

// Ball represents a ball in the game
type Ball struct {
	ID        int     `json:"id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	IsRed     bool    `json:"isRed"`
	Size      int     `json:"size"`
	TimeLimit float64 `json:"timeLimit"`
	SpawnedAt time.Time
	Clicked   bool
	ClickedBy int // -1: none, 0: player1, 1: player2
}

// Room represents a battle room
type Room struct {
	Code      string
	Players   [2]*Player
	State     RoomState
	CreatedAt time.Time
	GameStart time.Time
	GameEnd   time.Time

	CurrentBall *Ball
	BallCounter int
	Duration    float64 // Game duration in seconds

	Mu       sync.RWMutex
	StopGame chan struct{}
}

// RoomState represents the state of a room
type RoomState string

const (
	StateWaiting   RoomState = "waiting"
	StateCountdown RoomState = "countdown"
	StatePlaying   RoomState = "playing"
	StateFinished  RoomState = "finished"
)

// GameConfig represents difficulty settings based on elapsed time
type GameConfig struct {
	BallSize   int
	TimeLimit  float64
	BlueChance float64
}

// GetGameConfig returns game configuration based on elapsed time
func GetGameConfig(elapsedSeconds float64) GameConfig {
	if elapsedSeconds < 3 {
		return GameConfig{BallSize: 80, TimeLimit: 1.0, BlueChance: 0.10}
	} else if elapsedSeconds < 6 {
		return GameConfig{BallSize: 70, TimeLimit: 0.8, BlueChance: 0.15}
	} else if elapsedSeconds < 8 {
		return GameConfig{BallSize: 60, TimeLimit: 0.6, BlueChance: 0.20}
	}
	return GameConfig{BallSize: 50, TimeLimit: 0.5, BlueChance: 0.25}
}

// WebSocket message types
type WSMessage struct {
	Type     string      `json:"type"`
	Data     interface{} `json:"data,omitempty"`
	RoomCode string      `json:"roomCode,omitempty"`
	Nickname string      `json:"nickname,omitempty"`
}

// Server to client messages
type RoomCreatedMsg struct {
	Type     string `json:"type"`
	RoomCode string `json:"roomCode"`
}

type OpponentJoinedMsg struct {
	Type     string `json:"type"`
	Nickname string `json:"nickname"`
}

type CountdownMsg struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type GameStartMsg struct {
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
}

type BallSpawnMsg struct {
	Type      string  `json:"type"`
	ID        int     `json:"id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	IsRed     bool    `json:"isRed"`
	Size      int     `json:"size"`
	TimeLimit float64 `json:"timeLimit"`
}

type BallResultMsg struct {
	Type      string `json:"type"`
	BallID    int    `json:"ballId"`
	ClickedBy string `json:"clickedBy"` // "player1", "player2", "none"
	Scores    [2]int `json:"scores"`
}

type GameEndMsg struct {
	Type           string `json:"type"`
	MyScore        int    `json:"myScore"`
	OpponentScore  int    `json:"opponentScore"`
	Result         string `json:"result"` // "win", "lose", "draw"
	WinnerNickname string `json:"winnerNickname,omitempty"`
}

type OpponentReadyMsg struct {
	Type string `json:"type"`
}

type RematchStartMsg struct {
	Type string `json:"type"`
}

type OpponentLeftMsg struct {
	Type string `json:"type"`
}

type ErrorMsg struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type TimeUpdateMsg struct {
	Type        string  `json:"type"`
	TimeLeft    float64 `json:"timeLeft"`
}
