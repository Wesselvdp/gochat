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

	database, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath))

	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	queries := schema.New(database)
	return queries, database, nil
}
