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
	//execPath, err := os.Executable()
	//if err != nil {
	//	log.Printf("Error getting executable path: %v", err)
	//}
	//
	//projectRoot := filepath.Join(filepath.Dir(execPath), "..")
	envPath := "../.env"
	//// loads values from .env into the system
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
