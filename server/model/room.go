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
	Type     string  `json:"type"`
	TimeLeft float64 `json:"timeLeft"`
}

// ========== Translation Battle Room ==========

// TranslationRoom represents a translation battle room
type TranslationRoom struct {
	Code       string                `json:"code"`
	Players    [2]*Player            `json:"players"`
	State      TranslationGameState  `json:"state"`
	Difficulty TranslationDifficulty `json:"difficulty"`

	Rounds       []TranslationRound `json:"rounds"`
	CurrentRound int                `json:"currentRound"` // 0, 1, 2
	Wins         [2]int             `json:"wins"`         // Win count for each player
	TotalScores  [2]int             `json:"totalScores"`  // Total scores across all rounds
	RematchReady [2]bool            `json:"rematchReady"` // Whether each player is ready for rematch

	CreatedAt time.Time
	Mu        sync.RWMutex
	StopGame  chan struct{}
}

// ========== Translation Battle Messages (Server -> Client) ==========

// TRoomCreatedMsg is sent when a room is created
type TRoomCreatedMsg struct {
	Type     string `json:"type"` // "t_room_created"
	RoomCode string `json:"roomCode"`
}

// TOpponentJoinedMsg is sent when opponent joins
type TOpponentJoinedMsg struct {
	Type     string `json:"type"` // "t_opponent_joined"
	Nickname string `json:"nickname"`
}

// TGameStartMsg is sent when game starts
type TGameStartMsg struct {
	Type       string `json:"type"` // "t_game_start"
	Difficulty string `json:"difficulty"`
}

// TCountdownMsg is sent during countdown
type TCountdownMsg struct {
	Type  string `json:"type"` // "t_countdown"
	Count int    `json:"count"`
}

// TRoundStartMsg is sent when a round starts
type TRoundStartMsg struct {
	Type     string `json:"type"` // "t_round_start"
	Round    int    `json:"round"`
	Sentence string `json:"sentence"`
	TimeLeft int    `json:"timeLeft"` // seconds
}

// TTimeUpdateMsg is sent to update remaining time
type TTimeUpdateMsg struct {
	Type     string `json:"type"` // "t_time_update"
	TimeLeft int    `json:"timeLeft"`
}

// TOpponentSubmittedMsg is sent when opponent submits
type TOpponentSubmittedMsg struct {
	Type string `json:"type"` // "t_opponent_submitted"
}

// TEvaluatingMsg is sent when AI is evaluating
type TEvaluatingMsg struct {
	Type string `json:"type"` // "t_evaluating"
}

// TRoundResultMsg is sent with round results
type TRoundResultMsg struct {
	Type                string           `json:"type"` // "t_round_result"
	Round               int              `json:"round"`
	Sentence            string           `json:"sentence"`
	MyTranslation       string           `json:"myTranslation"`
	OpponentTranslation string           `json:"opponentTranslation"`
	MyScore             TranslationScore `json:"myScore"`
	OpponentScore       TranslationScore `json:"opponentScore"`
	ModelAnswer         string           `json:"modelAnswer"`
	RoundWinner         string           `json:"roundWinner"` // "me", "opponent", "draw"
	TotalWins           [2]int           `json:"totalWins"`
	IsGameOver          bool             `json:"isGameOver"`
}

// TGameOverMsg is sent when the game ends
type TGameOverMsg struct {
	Type           string `json:"type"` // "t_game_over"
	Winner         string `json:"winner"` // "me", "opponent", "draw"
	MyWins         int    `json:"myWins"`
	OpponentWins   int    `json:"opponentWins"`
	MyTotalScore   int    `json:"myTotalScore"`
	OpponentTotal  int    `json:"opponentTotal"`
	WinnerNickname string `json:"winnerNickname,omitempty"`
}

// TOpponentReadyMsg is sent when opponent is ready for rematch
type TOpponentReadyMsg struct {
	Type string `json:"type"` // "t_opponent_ready"
}

// TRematchStartMsg is sent when rematch starts
type TRematchStartMsg struct {
	Type string `json:"type"` // "t_rematch_start"
}

// TOpponentLeftMsg is sent when opponent leaves
type TOpponentLeftMsg struct {
	Type string `json:"type"` // "t_opponent_left"
}

// TErrorMsg is sent on error
type TErrorMsg struct {
	Type    string `json:"type"` // "t_error"
	Message string `json:"message"`
}
