package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"mini-games/model"
)

var (
	snakeSessions = make(map[string]*model.SnakeSession)
	snakeMutex    sync.RWMutex
)

// 레벨별 설정
var snakeLevelConfig = map[string]struct {
	Speed    int // ms per move
	MinSpeed int
}{
	"easy":   {Speed: 200, MinSpeed: 200},
	"medium": {Speed: 150, MinSpeed: 80},
	"hard":   {Speed: 100, MinSpeed: 50},
}

const (
	snakePointsPerFood = 10
	snakeMaxEatCount   = 1000 // 최대 음식 먹기 횟수
)

// CreateSnakeSession creates a new Snake game session
func CreateSnakeSession(level string) *model.SnakeSession {
	// 레벨 검증
	if _, ok := snakeLevelConfig[level]; !ok {
		level = "easy"
	}

	sessionID := generateSnakeSessionID()
	seed := time.Now().UnixNano()

	session := &model.SnakeSession{
		ID:        sessionID,
		Seed:      seed,
		Level:     level,
		StartTime: time.Now(),
		EatCount:  0,
		Score:     0,
		Status:    "playing",
	}

	snakeMutex.Lock()
	snakeSessions[sessionID] = session
	snakeMutex.Unlock()

	// 세션 자동 만료 (30분)
	go func() {
		time.Sleep(30 * time.Minute)
		snakeMutex.Lock()
		delete(snakeSessions, sessionID)
		snakeMutex.Unlock()
	}()

	return session
}

// ProcessSnakeEat processes a food eaten event
func ProcessSnakeEat(sessionID string) model.SnakeEatResponse {
	snakeMutex.Lock()
	defer snakeMutex.Unlock()

	session, exists := snakeSessions[sessionID]
	if !exists {
		return model.SnakeEatResponse{Valid: false}
	}

	if session.Status != "playing" {
		return model.SnakeEatResponse{Valid: false}
	}

	// 레벨별 설정 가져오기
	config := snakeLevelConfig[session.Level]

	// 최소 간격 검증: 레벨 속도 * 3칸 이동 시간 (최소 3칸은 이동해야 음식 도달)
	minInterval := time.Duration(config.MinSpeed*3) * time.Millisecond
	if !session.LastEatTime.IsZero() {
		timeSinceLastEat := time.Since(session.LastEatTime)
		if timeSinceLastEat < minInterval {
			// 너무 빠른 eat 요청 - 스팸 의심
			return model.SnakeEatResponse{Valid: false, EatCount: session.EatCount, Score: session.Score}
		}
	}

	// 합리성 검증: 플레이 시간 대비 최대 먹기 횟수
	playTimeMs := time.Since(session.StartTime).Milliseconds()

	// 최소 이동 시간 기준 최대 먹기 횟수 계산
	// 음식까지 평균 5칸 이동 가정
	avgMovesToFood := 5
	maxEatCount := int(playTimeMs / int64(config.MinSpeed*avgMovesToFood))
	if maxEatCount < 1 {
		maxEatCount = 1
	}

	// 현재 먹기 횟수가 최대치를 초과하면 거부
	if session.EatCount >= maxEatCount+5 { // 5회 여유
		return model.SnakeEatResponse{Valid: false, EatCount: session.EatCount, Score: session.Score}
	}

	// 최대 먹기 횟수 제한
	if session.EatCount >= snakeMaxEatCount {
		return model.SnakeEatResponse{Valid: false, EatCount: session.EatCount, Score: session.Score}
	}

	session.EatCount++
	session.Score = session.EatCount * snakePointsPerFood
	session.LastEatTime = time.Now() // 마지막 먹은 시간 업데이트

	return model.SnakeEatResponse{
		Valid:    true,
		EatCount: session.EatCount,
		Score:    session.Score,
	}
}

// EndSnakeSession ends a Snake game session
func EndSnakeSession(sessionID string) model.SnakeEndResponse {
	snakeMutex.Lock()
	defer snakeMutex.Unlock()

	session, exists := snakeSessions[sessionID]
	if !exists {
		return model.SnakeEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	if session.Status != "playing" {
		return model.SnakeEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	session.Status = "ended"

	return model.SnakeEndResponse{
		Valid:      true,
		FinalScore: session.Score,
		CanSubmit:  session.Score > 0,
	}
}

// SubmitSnakeScore submits the score to the ranking
func SubmitSnakeScore(sessionID string, nickname string) model.SubmitScoreResponse {
	snakeMutex.Lock()
	defer snakeMutex.Unlock()

	session, exists := snakeSessions[sessionID]
	if !exists {
		return model.SubmitScoreResponse{Success: false}
	}

	if session.Status != "ended" {
		return model.SubmitScoreResponse{Success: false}
	}

	// 점수 저장
	input := model.ScoreInput{
		Nickname: nickname,
		Game:     "snake",
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

func generateSnakeSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "sn_" + hex.EncodeToString(bytes)
}
