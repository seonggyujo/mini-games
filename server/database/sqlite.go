package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func Init() error {
	dbPath := getEnv("DB_PATH", "./scores.db")
	
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
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

	log.Printf("Database initialized successfully (path: %s)", dbPath)
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
