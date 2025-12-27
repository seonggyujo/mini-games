package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"mini-games/model"
)

const (
	GAME_WIDTH  = 1200
	GAME_HEIGHT = 800
	SPAWN_DELAY = 300 // ms delay between balls

	// 네트워크 지연 허용 범위 (ms)
	CLICK_TIME_TOLERANCE = 200
)

// SpeedClick 레벨 설정 (클라이언트와 동일해야 함)
var speedClickLevels = []model.SpeedClickLevel{
	{Level: 1, TimeLimit: 1.00, BallSize: 80, BlueChance: 0.10, RequiredScore: 0},
	{Level: 2, TimeLimit: 0.90, BallSize: 75, BlueChance: 0.13, RequiredScore: 5},
	{Level: 3, TimeLimit: 0.80, BallSize: 70, BlueChance: 0.16, RequiredScore: 10},
	{Level: 4, TimeLimit: 0.70, BallSize: 65, BlueChance: 0.20, RequiredScore: 15},
	{Level: 5, TimeLimit: 0.60, BallSize: 60, BlueChance: 0.23, RequiredScore: 20},
	{Level: 6, TimeLimit: 0.50, BallSize: 55, BlueChance: 0.26, RequiredScore: 25},
	{Level: 7, TimeLimit: 0.40, BallSize: 50, BlueChance: 0.30, RequiredScore: 30},
}

// 세션 저장소
var (
	sessions   = make(map[string]*model.GameSession)
	sessionsMu sync.RWMutex
)

// Mulberry32 PRNG state
type Mulberry32 struct {
	state uint32
}

// NewMulberry32 creates a new PRNG with the given seed
func NewMulberry32(seed int64) *Mulberry32 {
	return &Mulberry32{state: uint32(seed)}
}

// Next returns the next random number between 0 and 1
func (m *Mulberry32) Next() float64 {
	m.state += 0x6D2B79F5
	t := m.state
	t = (t ^ (t >> 15)) * (t | 1)
	t ^= t + (t^(t>>7))*(t|61)
	return float64((t^(t>>14))>>0) / 4294967296.0
}

// generateSessionID creates a random session ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getLevelConfig returns the level config for a given score
func getLevelConfig(score int) model.SpeedClickLevel {
	for i := len(speedClickLevels) - 1; i >= 0; i-- {
		if score >= speedClickLevels[i].RequiredScore {
			return speedClickLevels[i]
		}
	}
	return speedClickLevels[0]
}

// GenerateBall generates a deterministic ball based on seed and index
func GenerateBall(seed int64, index int, score int, prevBallEndTime int64) model.SpeedClickBall {
	// 각 공마다 고유한 시드 생성
	ballSeed := seed + int64(index)*12345
	rng := NewMulberry32(ballSeed)

	config := getLevelConfig(score)
	padding := float64(config.BallSize)

	x := padding + rng.Next()*(GAME_WIDTH-padding*2)
	y := padding + rng.Next()*(GAME_HEIGHT-padding*2)
	isRed := rng.Next() > config.BlueChance

	// 공 생성 시간 = 이전 공 종료 시간 + 딜레이
	spawnTime := prevBallEndTime + SPAWN_DELAY

	return model.SpeedClickBall{
		Index:     index,
		X:         x,
		Y:         y,
		IsRed:     isRed,
		SpawnTime: spawnTime,
		Duration:  int64(config.TimeLimit * 1000),
		Size:      config.BallSize,
		Level:     config.Level,
	}
}

// CreateSpeedClickSession creates a new game session
func CreateSpeedClickSession() *model.GameSession {
	sessionID := generateSessionID()
	seed := time.Now().UnixNano()

	session := &model.GameSession{
		ID:            sessionID,
		Game:          "speed-click",
		Seed:          seed,
		StartTime:     time.Now(),
		Score:         0,
		Lives:         3,
		CurrentBall:   0,
		BallSpawnTime: 0, // 첫 공은 즉시 생성
		Clicks:        []model.ClickRecord{},
		Status:        "playing",
	}

	sessionsMu.Lock()
	sessions[sessionID] = session
	sessionsMu.Unlock()

	// 10분 후 자동 삭제
	go func() {
		time.Sleep(10 * time.Minute)
		sessionsMu.Lock()
		delete(sessions, sessionID)
		sessionsMu.Unlock()
	}()

	return session
}

// GetSession retrieves a session by ID
func GetSession(sessionID string) *model.GameSession {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	return sessions[sessionID]
}

// ProcessClick handles a click event
func ProcessClick(sessionID string, ballIndex int, clickTimeMs int64) model.ClickResponse {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	session, exists := sessions[sessionID]
	if !exists {
		return model.ClickResponse{Valid: false, Message: "Session not found"}
	}

	if session.Status != "playing" {
		return model.ClickResponse{Valid: false, Message: "Game not in progress"}
	}

	// 현재 공 인덱스 확인
	if ballIndex != session.CurrentBall {
		return model.ClickResponse{Valid: false, Message: "Invalid ball index"}
	}

	// 이미 클릭한 공인지 확인
	for _, click := range session.Clicks {
		if click.BallIndex == ballIndex {
			return model.ClickResponse{Valid: false, Message: "Ball already clicked"}
		}
	}

	// 공 생성 (서버 측에서 검증용)
	ball := GenerateBall(session.Seed, ballIndex, session.Score, session.BallSpawnTime)

	// 클릭 시간 검증 (네트워크 지연 허용)
	ballEndTime := ball.SpawnTime + ball.Duration
	
	// 클릭이 너무 빠름 (공 생성 전)
	if clickTimeMs < ball.SpawnTime-CLICK_TIME_TOLERANCE {
		return model.ClickResponse{Valid: false, Message: "Click too early"}
	}

	// 클릭이 너무 늦음 (공 만료 후)
	if clickTimeMs > ballEndTime+CLICK_TIME_TOLERANCE {
		return model.ClickResponse{Valid: false, Message: "Click too late"}
	}

	// 점수 계산
	points := 0
	if ball.IsRed {
		// 빨간 공: 점수 획득
		timeRatio := float64(ballEndTime-clickTimeMs) / float64(ball.Duration)
		if timeRatio < 0 {
			timeRatio = 0
		}
		if timeRatio > 1 {
			timeRatio = 1
		}

		points = ball.Level
		if timeRatio >= 0.75 {
			points += 2
		} else if timeRatio >= 0.50 {
			points += 1
		}
		session.Score += points
	} else {
		// 파란 공: 목숨 감소
		session.Lives--
	}

	// 클릭 기록
	session.Clicks = append(session.Clicks, model.ClickRecord{
		BallIndex: ballIndex,
		ClickTime: clickTimeMs,
		Valid:     true,
		Points:    points,
	})

	// 다음 공으로 이동
	session.CurrentBall++
	session.BallSpawnTime = ballEndTime // 이 공이 끝나는 시간 기준

	gameOver := session.Lives <= 0
	if gameOver {
		session.Status = "ended"
	}

	return model.ClickResponse{
		Valid:    true,
		Points:   points,
		Score:    session.Score,
		Lives:    session.Lives,
		GameOver: gameOver,
	}
}

// ProcessMiss handles a missed ball (time expired without click)
func ProcessMiss(sessionID string, ballIndex int) model.MissResponse {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	session, exists := sessions[sessionID]
	if !exists {
		return model.MissResponse{Valid: false}
	}

	if session.Status != "playing" {
		return model.MissResponse{Valid: false}
	}

	if ballIndex != session.CurrentBall {
		return model.MissResponse{Valid: false}
	}

	// 공 생성 (검증용)
	ball := GenerateBall(session.Seed, ballIndex, session.Score, session.BallSpawnTime)

	// 빨간 공을 놓친 경우만 목숨 감소
	if ball.IsRed {
		session.Lives--
	}

	// 다음 공으로 이동
	session.CurrentBall++
	session.BallSpawnTime = ball.SpawnTime + ball.Duration

	gameOver := session.Lives <= 0
	if gameOver {
		session.Status = "ended"
	}

	return model.MissResponse{
		Valid:    true,
		Lives:    session.Lives,
		GameOver: gameOver,
	}
}

// EndSpeedClickSession ends a game session
func EndSpeedClickSession(sessionID string) model.EndGameResponse {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	session, exists := sessions[sessionID]
	if !exists {
		return model.EndGameResponse{FinalScore: 0, CanSubmit: false}
	}

	if session.Status == "playing" {
		session.Status = "ended"
	}

	return model.EndGameResponse{
		FinalScore: session.Score,
		CanSubmit:  session.Status == "ended" && session.Score > 0,
	}
}

// SubmitSpeedClickScore saves the score to database
func SubmitSpeedClickScore(sessionID string, nickname string) model.SubmitScoreResponse {
	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	if !exists {
		sessionsMu.Unlock()
		return model.SubmitScoreResponse{Success: false}
	}

	if session.Status != "ended" {
		sessionsMu.Unlock()
		return model.SubmitScoreResponse{Success: false}
	}

	// 상태를 submitted로 변경 (중복 제출 방지)
	session.Status = "submitted"
	score := session.Score
	sessionsMu.Unlock()

	// DB에 점수 저장
	input := model.ScoreInput{
		Nickname: nickname,
		Game:     "speed-click",
		Score:    score,
	}

	scoreID, err := SaveScore(input)
	if err != nil {
		return model.SubmitScoreResponse{Success: false}
	}

	return model.SubmitScoreResponse{
		Success: true,
		ScoreID: scoreID,
	}
}
