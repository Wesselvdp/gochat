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
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	// create tables
	//ctx := context.Background()
	//_, err = database.ExecContext(ctx, ddl)
	//if err != nil {
	//	return nil, nil, fmt.Errorf("failed creating tables: %w", err)
	//}

	queries := schema.New(database)
	return queries, database, nil
}
