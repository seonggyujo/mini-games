package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"mini-games/model"
)

var (
	jumpRunnerSessions = make(map[string]*model.JumpRunnerSession)
	jumpRunnerMutex    sync.RWMutex
)

const (
	// 최대 점수 상한 (60fps 기준, 1시간 플레이)
	jumpRunnerMaxScore = 216000
	// 최소 플레이 시간 (ms)
	jumpRunnerMinPlayTime = 1000
	// 플레이 시간 허용 오차 (ms)
	jumpRunnerTimeAllowance = 5000
)

// CreateJumpRunnerSession creates a new Jump Runner game session
func CreateJumpRunnerSession() *model.JumpRunnerSession {
	sessionID := generateJumpRunnerSessionID()
	seed := time.Now().UnixNano()

	session := &model.JumpRunnerSession{
		ID:        sessionID,
		Seed:      seed,
		StartTime: time.Now(),
		Score:     0,
		Status:    "playing",
	}

	jumpRunnerMutex.Lock()
	jumpRunnerSessions[sessionID] = session
	jumpRunnerMutex.Unlock()

	// 세션 자동 만료 (30분)
	go func() {
		time.Sleep(30 * time.Minute)
		jumpRunnerMutex.Lock()
		delete(jumpRunnerSessions, sessionID)
		jumpRunnerMutex.Unlock()
	}()

	return session
}

// EndJumpRunnerSession ends a Jump Runner game session and validates the score
func EndJumpRunnerSession(sessionID string, playTimeMs int64, clientScore int) model.JumpRunnerEndResponse {
	jumpRunnerMutex.Lock()
	defer jumpRunnerMutex.Unlock()

	session, exists := jumpRunnerSessions[sessionID]
	if !exists {
		return model.JumpRunnerEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	if session.Status != "playing" {
		return model.JumpRunnerEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	// 실제 경과 시간 계산
	actualPlayTime := time.Since(session.StartTime).Milliseconds()

	// 검증 1: 최소 플레이 시간 체크
	if playTimeMs < jumpRunnerMinPlayTime {
		return model.JumpRunnerEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	// 검증 2: 플레이 시간과 실제 경과 시간 비교 (허용 오차 내)
	timeDiff := actualPlayTime - playTimeMs
	if timeDiff < -jumpRunnerTimeAllowance {
		// 클라이언트가 보고한 시간이 실제보다 너무 길면 의심
		return model.JumpRunnerEndResponse{Valid: false, FinalScore: 0, CanSubmit: false}
	}

	// 검증 3: 점수 상한 체크 (60fps 기준, 프레임당 1점)
	// score = frameCount, maxFrames = playTimeMs / 16.67
	maxScore := int(float64(playTimeMs) / 16.67)
	if maxScore > jumpRunnerMaxScore {
		maxScore = jumpRunnerMaxScore
	}

	// 최종 점수 결정 (클라이언트 점수와 최대 가능 점수 중 작은 값)
	finalScore := clientScore
	if finalScore > maxScore {
		finalScore = maxScore
	}
	if finalScore < 0 {
		finalScore = 0
	}

	session.Score = finalScore
	session.EndTime = time.Now()
	session.Status = "ended"

	return model.JumpRunnerEndResponse{
		Valid:      true,
		FinalScore: finalScore,
		CanSubmit:  finalScore > 0,
	}
}

// SubmitJumpRunnerScore submits the score to the ranking
func SubmitJumpRunnerScore(sessionID string, nickname string) model.SubmitScoreResponse {
	jumpRunnerMutex.Lock()
	defer jumpRunnerMutex.Unlock()

	session, exists := jumpRunnerSessions[sessionID]
	if !exists {
		return model.SubmitScoreResponse{Success: false}
	}

	if session.Status != "ended" {
		return model.SubmitScoreResponse{Success: false}
	}

	// 점수 저장
	input := model.ScoreInput{
		Nickname: nickname,
		Game:     "jump-runner",
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

func generateJumpRunnerSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "jr_" + hex.EncodeToString(bytes)
}
