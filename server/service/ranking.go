package service

import (
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
		err := rows.Scan(&s.ID, &s.Nickname, &s.Game, &s.Score, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}

	return scores, nil
}
