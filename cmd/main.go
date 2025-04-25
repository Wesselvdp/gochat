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
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s", cwd)

	envPath := "../.env"
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env file found at %s", envPath)
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
