package service

import (
	"time"

	"mini-games/database"
	"mini-games/model"
)

func SaveScore(input model.ScoreInput) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO scores (nickname, game, score) VALUES (?, ?, ?)",
		input.Nickname, input.Game, input.Score,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetRanking(game string, limit int) ([]model.Score, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// 닉네임별 최고 점수만 조회 (중복 제거)
	rows, err := database.DB.Query(
		`SELECT MIN(id) as id, nickname, game, MAX(score) as score, MAX(created_at) as created_at 
		 FROM scores 
		 WHERE game = ? 
		 GROUP BY nickname
		 ORDER BY score DESC 
		 LIMIT ?`,
		game, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []model.Score
	for rows.Next() {
		var s model.Score
		var createdAtStr string
		err := rows.Scan(&s.ID, &s.Nickname, &s.Game, &s.Score, &createdAtStr)
		if err != nil {
			return nil, err
		}
		// SQLite에서 반환된 문자열을 time.Time으로 파싱
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		scores = append(scores, s)
	}

	return scores, nil
}
