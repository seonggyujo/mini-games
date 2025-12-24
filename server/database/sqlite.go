package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("sqlite3", "./scores.db")
	if err != nil {
		return err
	}

	// Create scores table
	createTable := `
	CREATE TABLE IF NOT EXISTS scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nickname TEXT NOT NULL,
		game TEXT NOT NULL,
		score INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_game_score ON scores(game, score DESC);
	`

	_, err = DB.Exec(createTable)
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
