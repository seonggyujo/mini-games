package service

import (
	"crypto/rand"
	"log"
	"math/big"
	"sync"
	"time"

	"mini-games/model"
)

const (
	TRoundDuration   = 60 * time.Second  // 60 seconds per round
	TRoomTimeout     = 10 * time.Minute  // Room expires after 10 minutes of inactivity
	TPostGameTimeout = 3 * time.Minute   // Time to wait for rematch after game ends
	TDefaultRounds   = 5                 // Default number of rounds
)

// TranslationRoomManager manages all translation battle rooms
type TranslationRoomManager struct {
	rooms      map[string]*model.TranslationRoom
	mu         sync.RWMutex
	groqClient *GroqClient
}

// NewTranslationRoomManager creates a new room manager
func NewTranslationRoomManager() *TranslationRoomManager {
	rm := &TranslationRoomManager{
		rooms:      make(map[string]*model.TranslationRoom),
		groqClient: NewGroqClient(),
	}
	// Start cleanup goroutine
	go rm.cleanupRoutine()
	return rm
}

// generateRoomCode generates a 6-character uppercase room code using crypto/rand
func generateTranslationRoomCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		code[i] = letters[n.Int64()]
	}
	return string(code)
}

// CreateRoom creates a new translation battle room
func (rm *TranslationRoomManager) CreateRoom(player *model.Player) (*model.TranslationRoom, string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Generate unique room code
	var code string
	for {
		code = generateTranslationRoomCode()
		if _, exists := rm.rooms[code]; !exists {
			break
		}
	}

	room := &model.TranslationRoom{
		Code:        code,
		Players:     [2]*model.Player{player, nil},
		State:       model.TStateWaiting,
		Difficulty:  model.DifficultyMedium,
		MaxRounds:   TDefaultRounds,
		IsUnlimited: false,
		Rounds:      make([]model.TranslationRound, 0),
		CreatedAt:   time.Now(),
		StopGame:    make(chan struct{}),
	}

	player.Index = 0
	player.Score = 0
	player.Ready = false
	player.LastActive = time.Now()

	rm.rooms[code] = room
	return room, code
}

// JoinRoom joins an existing translation battle room
func (rm *TranslationRoomManager) JoinRoom(code string, player *model.Player) (*model.TranslationRoom, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[code]
	if !exists {
		return nil, ErrRoomNotFound
	}

	if room.State != model.TStateWaiting {
		return nil, ErrRoomNotAvailable
	}

	if room.Players[1] != nil {
		return nil, ErrRoomFull
	}

	// 닉네임 중복 체크
	if room.Players[0] != nil && room.Players[0].Nickname == player.Nickname {
		return nil, ErrNicknameDuplicate
	}

	player.Index = 1
	player.Score = 0
	player.Ready = false
	player.LastActive = time.Now()

	room.Players[1] = player
	room.State = model.TStateReady
	return room, nil
}

// GetRoom returns a room by code
func (rm *TranslationRoomManager) GetRoom(code string) *model.TranslationRoom {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[code]
}

// RemoveRoom removes a room
func (rm *TranslationRoomManager) RemoveRoom(code string) {
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

// RemovePlayer removes a player from a room
func (rm *TranslationRoomManager) RemovePlayer(code string, playerIndex int) {
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
		room.Players[otherIndex].SendJSON(model.TOpponentLeftMsg{Type: "t_opponent_left"})
	}

	// Check if room is empty
	if room.Players[0] == nil && room.Players[1] == nil {
		rm.RemoveRoom(code)
	} else {
		// Stop any ongoing game
		if room.State == model.TStatePlaying || room.State == model.TStateCountdown || room.State == model.TStateEvaluating {
			select {
			case <-room.StopGame:
			default:
				close(room.StopGame)
			}
			room.StopGame = make(chan struct{})
		}
		room.State = model.TStateWaiting
	}
}

// StartGame starts the translation battle game
func (rm *TranslationRoomManager) StartGame(room *model.TranslationRoom, difficulty model.TranslationDifficulty, maxRounds int) {
	room.Mu.Lock()
	
	if room.State != model.TStateReady || room.Players[0] == nil || room.Players[1] == nil {
		room.Mu.Unlock()
		return
	}

	room.Difficulty = difficulty
	room.MaxRounds = maxRounds
	room.IsUnlimited = (maxRounds == 0)
	room.State = model.TStateCountdown
	room.Rounds = make([]model.TranslationRound, 0)
	room.CurrentRound = 0
	room.Wins = [2]int{0, 0}
	room.TotalScores = [2]int{0, 0}
	room.Mu.Unlock()

	// Notify both players that game is starting
	startMsg := model.TGameStartMsg{Type: "t_game_start", Difficulty: string(difficulty), MaxRounds: maxRounds}
	room.Players[0].SendJSON(startMsg)
	room.Players[1].SendJSON(startMsg)

	// Countdown
	for i := 3; i > 0; i-- {
		msg := model.TCountdownMsg{Type: "t_countdown", Count: i}
		room.Players[0].SendJSON(msg)
		room.Players[1].SendJSON(msg)

		select {
		case <-room.StopGame:
			return
		case <-time.After(1 * time.Second):
		}
	}

	// Start first round
	rm.startRound(room)
}

// startRound starts a new round
func (rm *TranslationRoomManager) startRound(room *model.TranslationRoom) {
	room.Mu.Lock()

	// Generate sentence using AI
	sentenceResult, err := rm.groqClient.GenerateSentenceWithTense(string(room.Difficulty))
	if err != nil {
		log.Printf("[startRound] Groq API error: %v", err)
	}

	round := model.TranslationRound{
		Number:      room.CurrentRound + 1,
		Sentence:    sentenceResult.Sentence,
		Tense:       sentenceResult.Tense,
		Submitted:   [2]bool{false, false},
		Winner:      -1,
		StartTime:   time.Now(),
	}

	room.Rounds = append(room.Rounds, round)
	room.State = model.TStatePlaying
	room.Mu.Unlock()

	// Notify both players
	roundMsg := model.TRoundStartMsg{
		Type:      "t_round_start",
		Round:     round.Number,
		MaxRounds: room.MaxRounds,
		Sentence:  sentenceResult.Sentence,
		Tense:     sentenceResult.Tense,
		TimeLeft:  int(TRoundDuration.Seconds()),
	}
	room.Players[0].SendJSON(roundMsg)
	room.Players[1].SendJSON(roundMsg)

	// Start round timer
	go rm.roundTimer(room, room.CurrentRound)
}

// roundTimer handles the countdown for a round
func (rm *TranslationRoomManager) roundTimer(room *model.TranslationRoom, roundIndex int) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	duration := TRoundDuration

	for {
		select {
		case <-room.StopGame:
			return

		case <-ticker.C:
			room.Mu.RLock()
			// Check if this round is still active
			if room.CurrentRound != roundIndex || room.State != model.TStatePlaying {
				room.Mu.RUnlock()
				return
			}

			// Check if both have submitted
			currentRound := &room.Rounds[roundIndex]
			bothSubmitted := currentRound.Submitted[0] && currentRound.Submitted[1]
			room.Mu.RUnlock()

			if bothSubmitted {
				rm.evaluateRound(room)
				return
			}

			// Calculate remaining time
			elapsed := time.Since(startTime)
			timeLeft := int((duration - elapsed).Seconds())

			if timeLeft <= 0 {
				// Time's up - evaluate round
				rm.evaluateRound(room)
				return
			}

			// Send time update
			timeMsg := model.TTimeUpdateMsg{Type: "t_time_update", TimeLeft: timeLeft}
			room.Mu.RLock()
			if room.Players[0] != nil {
				room.Players[0].SendJSON(timeMsg)
			}
			if room.Players[1] != nil {
				room.Players[1].SendJSON(timeMsg)
			}
			room.Mu.RUnlock()
		}
	}
}

// HandleSubmit handles a player's translation submission
func (rm *TranslationRoomManager) HandleSubmit(room *model.TranslationRoom, playerIndex int, translation string) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	if room.State != model.TStatePlaying {
		return
	}

	if room.CurrentRound >= len(room.Rounds) {
		return
	}

	currentRound := &room.Rounds[room.CurrentRound]

	// Check if already submitted
	if currentRound.Submitted[playerIndex] {
		return
	}

	// Record submission
	currentRound.Translations[playerIndex] = translation
	currentRound.Submitted[playerIndex] = true
	currentRound.SubmitTimes[playerIndex] = time.Now()

	// Notify opponent that this player has submitted
	otherIndex := 1 - playerIndex
	if room.Players[otherIndex] != nil {
		room.Players[otherIndex].SendJSON(model.TOpponentSubmittedMsg{Type: "t_opponent_submitted"})
	}

	// If both have submitted, evaluation will be triggered by the timer goroutine
}

// evaluateRound evaluates the round using AI
func (rm *TranslationRoomManager) evaluateRound(room *model.TranslationRoom) {
	room.Mu.Lock()
	if room.State != model.TStatePlaying {
		room.Mu.Unlock()
		return
	}
	room.State = model.TStateEvaluating
	currentRoundIndex := room.CurrentRound
	currentRound := &room.Rounds[currentRoundIndex]
	sentence := currentRound.Sentence
	trans0 := currentRound.Translations[0]
	trans1 := currentRound.Translations[1]
	room.Mu.Unlock()

	// Notify players that evaluation is in progress
	evalMsg := model.TEvaluatingMsg{Type: "t_evaluating"}
	room.Mu.RLock()
	if room.Players[0] != nil {
		room.Players[0].SendJSON(evalMsg)
	}
	if room.Players[1] != nil {
		room.Players[1].SendJSON(evalMsg)
	}
	room.Mu.RUnlock()

	// Call AI to evaluate
	result, _ := rm.groqClient.EvaluateTranslations(sentence, trans0, trans1)

	room.Mu.Lock()
	// Store scores
	currentRound.Scores[0] = model.TranslationScore{
		Meaning:     result.Translation1.Meaning,
		Grammar:     result.Translation1.Grammar,
		Naturalness: result.Translation1.Naturalness,
		Total:       result.Translation1.Meaning + result.Translation1.Grammar + result.Translation1.Naturalness,
		Feedback:    result.Translation1.Feedback,
	}
	currentRound.Scores[1] = model.TranslationScore{
		Meaning:     result.Translation2.Meaning,
		Grammar:     result.Translation2.Grammar,
		Naturalness: result.Translation2.Naturalness,
		Total:       result.Translation2.Meaning + result.Translation2.Grammar + result.Translation2.Naturalness,
		Feedback:    result.Translation2.Feedback,
	}
	currentRound.ModelAnswer = result.ModelAnswer

	// Determine round winner
	if currentRound.Scores[0].Total > currentRound.Scores[1].Total {
		currentRound.Winner = 0
		room.Wins[0]++
	} else if currentRound.Scores[1].Total > currentRound.Scores[0].Total {
		currentRound.Winner = 1
		room.Wins[1]++
	} else {
		currentRound.Winner = -1 // Draw
	}

	// Update total scores
	room.TotalScores[0] += currentRound.Scores[0].Total
	room.TotalScores[1] += currentRound.Scores[1].Total

	// Check if game is over (지정 라운드 모드에서만)
	var isGameOver bool
	if room.IsUnlimited {
		// 무제한 모드: HandleStopGame 호출 전까지 종료 안함
		isGameOver = false
	} else {
		// 지정 라운드 모드: 모든 라운드 완료 시 종료
		isGameOver = room.CurrentRound >= room.MaxRounds-1
	}
	room.State = model.TStateResult
	room.Mu.Unlock()

	// Send results to both players
	rm.sendRoundResult(room, currentRoundIndex, isGameOver)

	if isGameOver {
		rm.endGame(room)
	}
	// If not game over, wait for both players to click "next round" button
	// HandleNextRound will be called when players send t_next_round message
}

// sendRoundResult sends round results to both players
func (rm *TranslationRoomManager) sendRoundResult(room *model.TranslationRoom, roundIndex int, isGameOver bool) {
	room.Mu.RLock()
	defer room.Mu.RUnlock()

	currentRound := &room.Rounds[roundIndex]

	for i := 0; i < 2; i++ {
		if room.Players[i] == nil {
			continue
		}

		otherIndex := 1 - i
		var roundWinner string
		if currentRound.Winner == i {
			roundWinner = "me"
		} else if currentRound.Winner == otherIndex {
			roundWinner = "opponent"
		} else {
			roundWinner = "draw"
		}

		// For player i, swap wins array so their wins come first
		totalWins := [2]int{room.Wins[i], room.Wins[otherIndex]}

		msg := model.TRoundResultMsg{
			Type:                "t_round_result",
			Round:               currentRound.Number,
			Sentence:            currentRound.Sentence,
			MyTranslation:       currentRound.Translations[i],
			OpponentTranslation: currentRound.Translations[otherIndex],
			MyScore:             currentRound.Scores[i],
			OpponentScore:       currentRound.Scores[otherIndex],
			ModelAnswer:         currentRound.ModelAnswer,
			RoundWinner:         roundWinner,
			TotalWins:           totalWins,
			IsGameOver:          isGameOver,
		}

		room.Players[i].SendJSON(msg)
	}
}

// endGame ends the game and sends final results
func (rm *TranslationRoomManager) endGame(room *model.TranslationRoom) {
	room.Mu.Lock()
	room.State = model.TStateFinished
	room.RematchReady = [2]bool{false, false}

	wins0 := room.Wins[0]
	wins1 := room.Wins[1]
	total0 := room.TotalScores[0]
	total1 := room.TotalScores[1]
	nick0 := ""
	nick1 := ""
	if room.Players[0] != nil {
		nick0 = room.Players[0].Nickname
	}
	if room.Players[1] != nil {
		nick1 = room.Players[1].Nickname
	}

	// 라운드별 점수 요약 생성
	roundSummaries0 := make([]model.RoundSummary, len(room.Rounds))
	roundSummaries1 := make([]model.RoundSummary, len(room.Rounds))
	for i, round := range room.Rounds {
		var winner0, winner1 string
		if round.Winner == 0 {
			winner0 = "me"
			winner1 = "opponent"
		} else if round.Winner == 1 {
			winner0 = "opponent"
			winner1 = "me"
		} else {
			winner0 = "draw"
			winner1 = "draw"
		}

		roundSummaries0[i] = model.RoundSummary{
			Round:         round.Number,
			MyScore:       round.Scores[0].Total,
			OpponentScore: round.Scores[1].Total,
			Winner:        winner0,
		}
		roundSummaries1[i] = model.RoundSummary{
			Round:         round.Number,
			MyScore:       round.Scores[1].Total,
			OpponentScore: round.Scores[0].Total,
			Winner:        winner1,
		}
	}
	room.Mu.Unlock()

	// Determine winner by total score (총점제)
	var winner0, winner1 string
	var winnerNickname string

	if total0 > total1 {
		winner0 = "me"
		winner1 = "opponent"
		winnerNickname = nick0
	} else if total1 > total0 {
		winner0 = "opponent"
		winner1 = "me"
		winnerNickname = nick1
	} else {
		// 총점이 같으면 무승부
		winner0 = "draw"
		winner1 = "draw"
	}

	room.Mu.RLock()
	// Send to player 0
	if room.Players[0] != nil {
		room.Players[0].SendJSON(model.TGameOverMsg{
			Type:           "t_game_over",
			Winner:         winner0,
			MyWins:         wins0,
			OpponentWins:   wins1,
			MyTotalScore:   total0,
			OpponentTotal:  total1,
			WinnerNickname: winnerNickname,
			RoundSummaries: roundSummaries0,
		})
	}

	// Send to player 1
	if room.Players[1] != nil {
		room.Players[1].SendJSON(model.TGameOverMsg{
			Type:           "t_game_over",
			Winner:         winner1,
			MyWins:         wins1,
			OpponentWins:   wins0,
			MyTotalScore:   total1,
			OpponentTotal:  total0,
			WinnerNickname: winnerNickname,
			RoundSummaries: roundSummaries1,
		})
	}
	room.Mu.RUnlock()
}

// HandleRematch handles a rematch request
func (rm *TranslationRoomManager) HandleRematch(room *model.TranslationRoom, playerIndex int) {
	room.Mu.Lock()

	if room.State != model.TStateFinished {
		room.Mu.Unlock()
		return
	}

	room.RematchReady[playerIndex] = true

	// Notify other player
	otherIndex := 1 - playerIndex
	if room.Players[otherIndex] != nil {
		room.Players[otherIndex].SendJSON(model.TOpponentReadyMsg{Type: "t_opponent_ready"})
	}

	// Check if both are ready
	if room.RematchReady[0] && room.RematchReady[1] {
		room.State = model.TStateReady
		room.StopGame = make(chan struct{})
		room.RematchReady = [2]bool{false, false}
		difficulty := room.Difficulty // Copy value before releasing lock
		maxRounds := room.MaxRounds   // 기존 라운드 설정 유지

		// Send rematch start
		room.Players[0].SendJSON(model.TRematchStartMsg{Type: "t_rematch_start"})
		room.Players[1].SendJSON(model.TRematchStartMsg{Type: "t_rematch_start"})

		// Release lock before calling StartGame (avoid deadlock)
		room.Mu.Unlock()
		go rm.StartGame(room, difficulty, maxRounds)
		return
	}

	room.Mu.Unlock()
}

// HandleNextRound handles a player's next round ready signal
func (rm *TranslationRoomManager) HandleNextRound(room *model.TranslationRoom, playerIndex int) {
	room.Mu.Lock()
	
	// Only allow in result state (not game over)
	if room.State != model.TStateResult {
		room.Mu.Unlock()
		return
	}
	
	// Check if game is already over (지정 라운드 모드에서만)
	var isGameOver bool
	if room.IsUnlimited {
		isGameOver = false
	} else {
		isGameOver = room.CurrentRound >= room.MaxRounds-1
	}
	if isGameOver {
		room.Mu.Unlock()
		return
	}
	
	room.NextRoundReady[playerIndex] = true
	
	// Notify opponent
	otherIndex := 1 - playerIndex
	if room.Players[otherIndex] != nil {
		room.Players[otherIndex].SendJSON(model.TOpponentNextReadyMsg{Type: "t_opponent_next_ready"})
	}
	
	// Check if both are ready
	if room.NextRoundReady[0] && room.NextRoundReady[1] {
		room.NextRoundReady = [2]bool{false, false}
		room.CurrentRound++
		room.Mu.Unlock()
		
		rm.startRound(room)
		return
	}
	
	room.Mu.Unlock()
}

// HandleStopGame handles stop game request in unlimited mode
func (rm *TranslationRoomManager) HandleStopGame(room *model.TranslationRoom, playerIndex int) {
	room.Mu.Lock()
	
	// 무제한 모드에서만 동작
	if !room.IsUnlimited {
		room.Mu.Unlock()
		return
	}
	
	// 플레이 중이거나 결과 화면일 때만 종료 가능
	if room.State != model.TStatePlaying && room.State != model.TStateResult {
		room.Mu.Unlock()
		return
	}
	
	// 종료 요청한 플레이어 닉네임
	stoppedByNickname := ""
	if room.Players[playerIndex] != nil {
		stoppedByNickname = room.Players[playerIndex].Nickname
	}
	
	room.Mu.Unlock()
	
	// 양 플레이어에게 게임 종료 알림
	stoppedMsg := model.TGameStoppedMsg{
		Type:      "t_game_stopped",
		StoppedBy: stoppedByNickname,
	}
	room.Mu.RLock()
	if room.Players[0] != nil {
		room.Players[0].SendJSON(stoppedMsg)
	}
	if room.Players[1] != nil {
		room.Players[1].SendJSON(stoppedMsg)
	}
	room.Mu.RUnlock()
	
	// 게임 종료 처리
	rm.endGame(room)
}

// cleanupRoutine periodically cleans up expired rooms
func (rm *TranslationRoomManager) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rm.mu.Lock()
		now := time.Now()
		toDelete := []string{}

		for code, room := range rm.rooms {
			room.Mu.RLock()

			// Delete waiting rooms after timeout
			if room.State == model.TStateWaiting && now.Sub(room.CreatedAt) > TRoomTimeout {
				toDelete = append(toDelete, code)
			}

			// Delete finished rooms after post-game timeout
			if room.State == model.TStateFinished {
				// Check last activity
				lastActivity := room.CreatedAt
				for _, p := range room.Players {
					if p != nil && p.LastActive.After(lastActivity) {
						lastActivity = p.LastActive
					}
				}
				if now.Sub(lastActivity) > TPostGameTimeout {
					toDelete = append(toDelete, code)
				}
			}

			room.Mu.RUnlock()
		}

		for _, code := range toDelete {
			if room, exists := rm.rooms[code]; exists {
				// Notify players before deletion
				room.Mu.RLock()
				for _, p := range room.Players {
					if p != nil {
						p.SendJSON(model.TErrorMsg{Type: "t_error", Message: "방이 시간 초과로 삭제되었습니다."})
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
