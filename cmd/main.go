package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"gochat/internal/run"
	"log"
	"os"
)

func init() {
	// Try multiple common locations
	envPaths := []string{".env", "../.env", "/app/.env"}

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded env from %s", path)
			return
		}
	}

	log.Print("No .env file found in any of the expected locations")
}

func main() {
	ctx := context.Background()

	//Probably not do this here
	err := os.Setenv("dataSourceName", "analysis.db")

	if err != nil {
		panic(err)
	}

	if err := run.Run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
