package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gochat/internal/schema"
)

func Init() (*schema.Queries, *sql.DB, error) {
	database, err := sql.Open("sqlite3", "database.db")
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
