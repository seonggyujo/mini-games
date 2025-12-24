package model

import "time"

type Score struct {
	ID        int64     `json:"id"`
	Nickname  string    `json:"nickname"`
	Game      string    `json:"game"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

type ScoreInput struct {
	Nickname string `json:"nickname"`
	Game     string `json:"game"`
	Score    int    `json:"score"`
}
