package main

import (
	"context"
	"fmt"
	"gochat/internal/run"
	"os"
)

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
