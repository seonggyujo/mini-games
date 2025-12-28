package service

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"sync"
	"time"

	"mini-games/model"
)

var (
	memoryCardSessions = make(map[string]*model.MemoryCardSession)
	memoryCardMutex    sync.RWMutex
)

// 레벨별 설정
var memoryCardLevelConfig = map[string]struct {
	Cols      int
	Rows      int
	Pairs     int
	TimeLimit int // seconds
	Bonus     int
}{
	"easy":   {Cols: 4, Rows: 3, Pairs: 6, TimeLimit: 60, Bonus: 0},
	"medium": {Cols: 4, Rows: 4, Pairs: 8, TimeLimit: 90, Bonus: 100},
	"hard":   {Cols: 5, Rows: 4, Pairs: 10, TimeLimit: 120, Bonus: 200},
}

// CreateMemoryCardSession creates a new Memory Card game session
func CreateMemoryCardSession(level string) (*model.MemoryCardSession, []int) {
	// 레벨 검증
	config, ok := memoryCardLevelConfig[level]
	if !ok {
		level = "easy"
		config = memoryCardLevelConfig[level]
	}

	sessionID := generateMemoryCardSessionID()
	seed := time.Now().UnixNano()

	// 카드 배치 생성 (서버에서 결정)
	cards := generateCardLayout(config.Pairs, seed)

	session := &model.MemoryCardSession{
		ID:         sessionID,
		Seed:       seed,
		Level:      level,
		StartTime:  time.Now(),
		TimeLimit:  config.TimeLimit,
		Cards:      cards,
		MatchCount: 0,
		Moves:      0,
		TotalPairs: config.Pairs,
		Score:      0,
		Status:     "playing",
	}

	memoryCardMutex.Lock()
	memoryCardSessions[sessionID] = session
	memoryCardMutex.Unlock()

	// 세션 자동 만료 (30분)
	go func() {
		time.Sleep(30 * time.Minute)
		memoryCardMutex.Lock()
		delete(memoryCardSessions, sessionID)
		memoryCardMutex.Unlock()
	}()

	return session, cards
}

// generateCardLayout generates a shuffled card layout
func generateCardLayout(pairs int, seed int64) []int {
	// pairs 수만큼 카드 쌍 생성 (0 ~ pairs-1 값 2개씩)
	cards := make([]int, pairs*2)
	for i := 0; i < pairs; i++ {
		cards[i*2] = i
		cards[i*2+1] = i
	}

	// Fisher-Yates 셔플 (seed 기반)
	rng := newSeededRandom(seed)
	for i := len(cards) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	return cards
}

// 간단한 시드 기반 난수 생성기
type seededRandom struct {
	state uint64
}

func newSeededRandom(seed int64) *seededRandom {
	return &seededRandom{state: uint64(seed)}
}

func (r *seededRandom) Intn(n int) int {
	// Mulberry32 알고리즘
	r.state += 0x6D2B79F5
	t := r.state
	t = (t ^ (t >> 15)) * (t | 1)
	t ^= t + (t^(t>>7))*(t|61)
	result := t ^ (t >> 14)
	return int(result % uint64(n))
}

// ProcessMemoryCardMatch processes a match attempt
func ProcessMemoryCardMatch(sessionID string, card1, card2 int) model.MemoryCardMatchResponse {
	memoryCardMutex.Lock()
	defer memoryCardMutex.Unlock()

	session, exists := memoryCardSessions[sessionID]
	if !exists {
		return model.MemoryCardMatchResponse{Valid: false}
	}

	if session.Status != "playing" {
		return model.MemoryCardMatchResponse{Valid: false}
	}

	// 시간 초과 체크
	elapsed := time.Since(session.StartTime).Seconds()
	if elapsed > float64(session.TimeLimit) {
		session.Status = "ended"
		return model.MemoryCardMatchResponse{Valid: false}
	}

	// 카드 인덱스 유효성 체크
	if card1 < 0 || card1 >= len(session.Cards) || card2 < 0 || card2 >= len(session.Cards) {
		return model.MemoryCardMatchResponse{Valid: false}
	}

	if card1 == card2 {
		return model.MemoryCardMatchResponse{Valid: false}
	}

	session.Moves++

	// 매칭 체크
	isMatch := session.Cards[card1] == session.Cards[card2]
	if isMatch {
		session.MatchCount++
	}

	return model.MemoryCardMatchResponse{
		Valid:      true,
		IsMatch:    isMatch,
		MatchCount: session.MatchCount,
		Moves:      session.Moves,
	}
}

// EndMemoryCardSession ends a Memory Card game session
func EndMemoryCardSession(sessionID string) model.MemoryCardEndResponse {
	memoryCardMutex.Lock()
	defer memoryCardMutex.Unlock()

	session, exists := memoryCardSessions[sessionID]
	if !exists {
		return model.MemoryCardEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	if session.Status != "playing" {
		return model.MemoryCardEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	// 점수 계산
	elapsed := time.Since(session.StartTime).Seconds()
	remainingTime := session.TimeLimit - int(elapsed)
	if remainingTime < 0 {
		remainingTime = 0
	}

	config := memoryCardLevelConfig[session.Level]

	// 점수 = (남은 시간 × 10) + 레벨 보너스 - (이동 횟수 × 2)
	// 모든 카드를 맞춘 경우에만 점수 부여
	var finalScore int
	if session.MatchCount == session.TotalPairs {
		finalScore = (remainingTime * 10) + config.Bonus - (session.Moves * 2)
		if finalScore < 0 {
			finalScore = 0
		}
	} else {
		finalScore = 0
	}

	session.Score = finalScore
	session.Status = "ended"

	return model.MemoryCardEndResponse{
		Valid:      true,
		FinalScore: finalScore,
		CanSubmit:  finalScore > 0,
	}
}

// SubmitMemoryCardScore submits the score to the ranking
func SubmitMemoryCardScore(sessionID string, nickname string) model.SubmitScoreResponse {
	memoryCardMutex.Lock()
	defer memoryCardMutex.Unlock()

	session, exists := memoryCardSessions[sessionID]
	if !exists {
		return model.SubmitScoreResponse{Success: false}
	}

	if session.Status != "ended" {
		return model.SubmitScoreResponse{Success: false}
	}

	if session.Score <= 0 {
		return model.SubmitScoreResponse{Success: false}
	}

	// 점수 저장
	input := model.ScoreInput{
		Nickname: nickname,
		Game:     "memory-card",
		Score:    session.Score,
	}

	id, err := SaveScore(input)
	if err != nil {
		return model.SubmitScoreResponse{Success: false}
	}

	session.Status = "submitted"

	return model.SubmitScoreResponse{
		Success: true,
		ScoreID: id,
	}
}

func generateMemoryCardSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "mc_" + hex.EncodeToString(bytes)
}

// min helper for int
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max helper for int
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// abs helper for float64
func absFloat64(x float64) float64 {
	return math.Abs(x)
}
