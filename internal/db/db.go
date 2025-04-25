package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gochat/internal/schema"
	"os"
)

func Init() (*schema.Queries, *sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")

	if dbPath == "" {
		return nil, nil, fmt.Errorf("DB_PATH environment variable not set")
	}

	// Check if file exists before connecting
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("database file does not exist at path: %s", dbPath)
	}

	database, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed tos open database: %w", err)
	}

	// Verify connection with a ping
	if err := database.Ping(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Verify essential tables exist
	var tableName string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='user' LIMIT 1").Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("essential table 'user' not found in database")
		}
		return nil, nil, fmt.Errorf("failed to verify database schema: %w", err)
	}

	queries := schema.New(database)
	return queries, database, nil
}
