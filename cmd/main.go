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

	envPath := "/.env"  // This is now the absolute path in the container

	if err := godotenv.Load(envPath); err != nil {
		log.Print("No .env file found", envPath)
		if err := godotenv.Load(".env"); err != nil {
			log.Print("No .env file found in second layer" +
				"", "")
		}
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
