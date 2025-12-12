package main

import (
	"log"
	"os"

	"buddy/internal/cli"
	"buddy/internal/mybuddy"
	"buddy/internal/app"
)

func main() {
	// Create app context for mybuddy
	appCtx, err := app.NewContext("mybuddy")
	if err != nil {
		log.Fatalf("Failed to create app context: %v", err)
	}

	// 1. Get the base command
	rootCmd := cli.NewRootCommand(appCtx)

	// 2. Add mybuddy specific commands to the root
	rootCmd.AddCommand(mybuddy.GetCommands(appCtx)...)

	// 3. Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}