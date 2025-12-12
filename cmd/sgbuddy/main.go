package main

import (
	"log"
	"os"

	"buddy/internal/app"
	"buddy/internal/cli"
	"buddy/internal/sgbuddy"
)

func main() {
	// Create app context for sgbuddy
	appCtx, err := app.NewContext("sgbuddy")
	if err != nil {
		log.Fatalf("Failed to create app context: %v", err)
	}

	// 1. Get the base command
	rootCmd := cli.NewRootCommand(appCtx)

	// 2. Add sgbuddy specific commands to the root
	rootCmd.AddCommand(sgbuddy.GetCommands(appCtx)...)

	// 3. Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
