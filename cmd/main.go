package main

import (
	"context"
	"fmt"
	"gochat/internal/run"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	envPath := "/.env" // This is now the absolute path in the container
	if err := godotenv.Load(envPath); err != nil {
		log.Print("No .env file found", envPath)
	}
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
