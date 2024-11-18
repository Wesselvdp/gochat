package database

import (
	"database/sql"
	_ "embed"
	_ "github.com/mattn/go-sqlite3"
	"gochat/internal/schema"
	"log"
	"os"
	"path/filepath"
)

func Init() (*schema.Queries, *sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	// When opening the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			log.Fatal(err)
		}
		// Retry opening the database
		db, err = sql.Open("sqlite3", os.Getenv("DB_PATH"))
		if err != nil {
			log.Fatal(err)
		}
	}
	queries := schema.New(database)
	return queries, db, nil
}
