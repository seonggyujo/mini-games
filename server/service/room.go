package service

import (
	"math/rand"
	"sync"
	"time"

	"mini-games/model"
)

const (
	GameWidth    = 1200
	GameHeight   = 800
	GameDuration = 10.0 // seconds
	RoomTimeout  = 5 * time.Minute
	PostGameTimeout = 2 * time.Minute
)

// RoomManager manages all battle rooms
type RoomManager struct {
	rooms map[string]*model.Room
	mu    sync.RWMutex
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	rm := &RoomManager{
		rooms: make(map[string]*model.Room),
	}
	// Start cleanup goroutine
	go rm.cleanupRoutine()
	return rm
}

// generateRoomCode generates a 6-character uppercase room code
func generateRoomCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	code := make([]byte, 6)
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}

// CreateRoom creates a new room and returns the room code
func (rm *RoomManager) CreateRoom(player *model.Player) (*model.Room, string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Generate unique room code
	var code string
	for {
		code = generateRoomCode()
		if _, exists := rm.rooms[code]; !exists {
			break
		}
	}

	room := &model.Room{
		Code:      code,
		Players:   [2]*model.Player{player, nil},
		State:     model.StateWaiting,
		CreatedAt: time.Now(),
		Duration:  GameDuration,
		StopGame:  make(chan struct{}),
	}

	player.Index = 0
	player.Score = 0
	player.Ready = false
	player.LastActive = time.Now()

	rm.rooms[code] = room
	return room, code
}

// JoinRoom joins an existing room
func (rm *RoomManager) JoinRoom(code string, player *model.Player) (*model.Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[code]
	if !exists {
		return nil, ErrRoomNotFound
	}

	if room.State != model.StateWaiting {
		return nil, ErrRoomNotAvailable
	}

	if room.Players[1] != nil {
		return nil, ErrRoomFull
	}

	player.Index = 1
	player.Score = 0
	player.Ready = false
	player.LastActive = time.Now()

	room.Players[1] = player
	return room, nil
}

// GetRoom returns a room by code
func (rm *RoomManager) GetRoom(code string) *model.Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[code]
}

// RemoveRoom removes a room
func (rm *RoomManager) RemoveRoom(code string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if room, exists := rm.rooms[code]; exists {
		// Signal game loop to stop
		select {
		case <-room.StopGame:
		default:
			close(room.StopGame)
		}
		delete(rm.rooms, code)
	}
}

// RemovePlayerFromRoom removes a player from a room
func (rm *RoomManager) RemovePlayerFromRoom(code string, playerIndex int) {
	rm.mu.Lock()
	room, exists := rm.rooms[code]
	rm.mu.Unlock()

	if !exists {
		return
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	room.Players[playerIndex] = nil

	// Notify other player
	otherIndex := 1 - playerIndex
	if room.Players[otherIndex] != nil {
		room.Players[otherIndex].SendJSON(model.OpponentLeftMsg{Type: "opponent_left"})
	}

	// Check if room is empty
	if room.Players[0] == nil && room.Players[1] == nil {
		rm.RemoveRoom(code)
	} else {
		// Reset room to waiting state if game was in progress
		if room.State == model.StatePlaying || room.State == model.StateCountdown {
			select {
			case <-room.StopGame:
			default:
				close(room.StopGame)
			}
			room.StopGame = make(chan struct{})
		}
		room.State = model.StateWaiting
	}
}

// StartGame starts the game for a room
func (rm *RoomManager) StartGame(room *model.Room) {
	room.Mu.Lock()
	if room.State != model.StateWaiting || room.Players[0] == nil || room.Players[1] == nil {
		room.Mu.Unlock()
		return
	}
	room.State = model.StateCountdown
	room.Mu.Unlock()

	// Countdown
	for i := 3; i > 0; i-- {
		msg := model.CountdownMsg{Type: "countdown", Count: i}
		room.Players[0].SendJSON(msg)
		room.Players[1].SendJSON(msg)
		
		select {
		case <-room.StopGame:
			return
		case <-time.After(1 * time.Second):
		}
	}

	// Start game
	room.Mu.Lock()
	room.State = model.StatePlaying
	room.GameStart = time.Now()
	room.Players[0].Score = 0
	room.Players[1].Score = 0
	room.BallCounter = 0
	room.Mu.Unlock()

	startMsg := model.GameStartMsg{Type: "game_start", Duration: room.Duration}
	room.Players[0].SendJSON(startMsg)
	room.Players[1].SendJSON(startMsg)

	// Game loop
	go rm.gameLoop(room)
}

// gameLoop runs the main game loop
func (rm *RoomManager) gameLoop(room *model.Room) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	timeTicker := time.NewTicker(100 * time.Millisecond)
	defer timeTicker.Stop()

	spawnBall(room)

	for {
		select {
		case <-room.StopGame:
			return

		case <-timeTicker.C:
			// Send time updates
			elapsed := time.Since(room.GameStart).Seconds()
			timeLeft := room.Duration - elapsed

			if timeLeft <= 0 {
				rm.endGame(room)
				return
			}

			timeMsg := model.TimeUpdateMsg{Type: "time_update", TimeLeft: timeLeft}
			room.Mu.RLock()
			if room.Players[0] != nil {
				room.Players[0].SendJSON(timeMsg)
			}
			if room.Players[1] != nil {
				room.Players[1].SendJSON(timeMsg)
			}
			room.Mu.RUnlock()

		case <-ticker.C:
			room.Mu.Lock()
			if room.CurrentBall != nil && !room.CurrentBall.Clicked {
				// Check if ball time expired
				ballElapsed := time.Since(room.CurrentBall.SpawnedAt).Seconds()
				if ballElapsed >= room.CurrentBall.TimeLimit {
					// Ball expired, no one clicked
					sendBallResult(room, -1)
					room.CurrentBall = nil
					room.Mu.Unlock()
					
					// Spawn new ball after delay
					time.Sleep(200 * time.Millisecond)
					spawnBall(room)
					continue
				}
			} else if room.CurrentBall == nil || room.CurrentBall.Clicked {
				room.Mu.Unlock()
				continue
			}
			room.Mu.Unlock()
		}
	}
}

// spawnBall spawns a new ball
func spawnBall(room *model.Room) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	if room.State != model.StatePlaying {
		return
	}

	elapsed := time.Since(room.GameStart).Seconds()
	config := model.GetGameConfig(elapsed)

	padding := float64(config.BallSize)
	x := padding + rand.Float64()*(GameWidth-padding*2)
	y := padding + rand.Float64()*(GameHeight-padding*2)
	isRed := rand.Float64() > config.BlueChance

	room.BallCounter++
	ball := &model.Ball{
		ID:        room.BallCounter,
		X:         x,
		Y:         y,
		IsRed:     isRed,
		Size:      config.BallSize,
		TimeLimit: config.TimeLimit,
		SpawnedAt: time.Now(),
		Clicked:   false,
		ClickedBy: -1,
	}
	room.CurrentBall = ball

	msg := model.BallSpawnMsg{
		Type:      "ball_spawn",
		ID:        ball.ID,
		X:         ball.X,
		Y:         ball.Y,
		IsRed:     ball.IsRed,
		Size:      ball.Size,
		TimeLimit: ball.TimeLimit,
	}

	if room.Players[0] != nil {
		room.Players[0].SendJSON(msg)
	}
	if room.Players[1] != nil {
		room.Players[1].SendJSON(msg)
	}
}

// HandleClick handles a player clicking the ball
func (rm *RoomManager) HandleClick(room *model.Room, playerIndex int) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	if room.State != model.StatePlaying || room.CurrentBall == nil || room.CurrentBall.Clicked {
		return
	}

	// First click wins
	room.CurrentBall.Clicked = true
	room.CurrentBall.ClickedBy = playerIndex

	// Calculate score
	if room.CurrentBall.IsRed {
		room.Players[playerIndex].Score++
	} else {
		room.Players[playerIndex].Score--
	}

	sendBallResult(room, playerIndex)
	room.CurrentBall = nil

	// Spawn new ball after short delay (in separate goroutine)
	go func() {
		time.Sleep(200 * time.Millisecond)
		spawnBall(room)
	}()
}

// sendBallResult sends ball result to both players
func sendBallResult(room *model.Room, clickedBy int) {
	clickedByStr := "none"
	if clickedBy == 0 {
		clickedByStr = "player1"
	} else if clickedBy == 1 {
		clickedByStr = "player2"
	}

	ballID := 0
	if room.CurrentBall != nil {
		ballID = room.CurrentBall.ID
	}

	msg := model.BallResultMsg{
		Type:      "ball_result",
		BallID:    ballID,
		ClickedBy: clickedByStr,
		Scores:    [2]int{room.Players[0].Score, room.Players[1].Score},
	}

	if room.Players[0] != nil {
		room.Players[0].SendJSON(msg)
	}
	if room.Players[1] != nil {
		room.Players[1].SendJSON(msg)
	}
}

// endGame ends the game and sends results
func (rm *RoomManager) endGame(room *model.Room) {
	room.Mu.Lock()
	room.State = model.StateFinished
	room.GameEnd = time.Now()
	room.CurrentBall = nil

	score0 := room.Players[0].Score
	score1 := room.Players[1].Score
	nick0 := room.Players[0].Nickname
	nick1 := room.Players[1].Nickname

	room.Players[0].Ready = false
	room.Players[1].Ready = false
	room.Mu.Unlock()

	// Determine results
	var result0, result1 string
	var winnerNickname string

	if score0 > score1 {
		result0 = "win"
		result1 = "lose"
		winnerNickname = nick0
	} else if score1 > score0 {
		result0 = "lose"
		result1 = "win"
		winnerNickname = nick1
	} else {
		result0 = "draw"
		result1 = "draw"
		winnerNickname = ""
	}

	// Send results to player 0
	room.Players[0].SendJSON(model.GameEndMsg{
		Type:           "game_end",
		MyScore:        score0,
		OpponentScore:  score1,
		Result:         result0,
		WinnerNickname: winnerNickname,
	})

	// Send results to player 1
	room.Players[1].SendJSON(model.GameEndMsg{
		Type:           "game_end",
		MyScore:        score1,
		OpponentScore:  score0,
		Result:         result1,
		WinnerNickname: winnerNickname,
	})
}

// HandleRematchReady handles rematch ready request
func (rm *RoomManager) HandleRematchReady(room *model.Room, playerIndex int) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	if room.State != model.StateFinished {
		return
	}

	room.Players[playerIndex].Ready = true

	// Notify other player
	otherIndex := 1 - playerIndex
	if room.Players[otherIndex] != nil {
		room.Players[otherIndex].SendJSON(model.OpponentReadyMsg{Type: "opponent_ready"})
	}

	// Check if both are ready
	if room.Players[0].Ready && room.Players[1].Ready {
		room.State = model.StateWaiting
		room.StopGame = make(chan struct{})
		
		// Send rematch start
		room.Players[0].SendJSON(model.RematchStartMsg{Type: "rematch_start"})
		room.Players[1].SendJSON(model.RematchStartMsg{Type: "rematch_start"})

		// Start new game
		go rm.StartGame(room)
	}
}

// cleanupRoutine periodically cleans up expired rooms
func (rm *RoomManager) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rm.mu.Lock()
		now := time.Now()
		toDelete := []string{}

		for code, room := range rm.rooms {
			room.Mu.RLock()
			
			// Delete waiting rooms after timeout
			if room.State == model.StateWaiting && now.Sub(room.CreatedAt) > RoomTimeout {
				toDelete = append(toDelete, code)
			}
			
			// Delete finished rooms after post-game timeout
			if room.State == model.StateFinished && now.Sub(room.GameEnd) > PostGameTimeout {
				toDelete = append(toDelete, code)
			}
			
			room.Mu.RUnlock()
		}

		for _, code := range toDelete {
			if room, exists := rm.rooms[code]; exists {
				// Notify players before deletion
				room.Mu.RLock()
				for _, p := range room.Players {
					if p != nil {
						p.SendJSON(model.ErrorMsg{Type: "error", Message: "방이 시간 초과로 삭제되었습니다."})
					}
				}
				room.Mu.RUnlock()
				
				select {
				case <-room.StopGame:
				default:
					close(room.StopGame)
				}
				delete(rm.rooms, code)
			}
		}
		rm.mu.Unlock()
	}
}

// Custom errors
type RoomError struct {
	Message string
}

func (e RoomError) Error() string {
	return e.Message
}

var (
	ErrRoomNotFound    = RoomError{"방을 찾을 수 없습니다"}
	ErrRoomFull        = RoomError{"방이 가득 찼습니다"}
	ErrRoomNotAvailable = RoomError{"참가할 수 없는 방입니다"}
)
