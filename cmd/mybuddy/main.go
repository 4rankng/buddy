package main

import (
	"log"
	"os"

	"buddy/clients"
	"buddy/internal/app"
	"buddy/internal/cli"
	"buddy/internal/mybuddy"
)

func main() {
	// Create app context for mybuddy
	appCtx, err := app.NewContext("mybuddy")
	if err != nil {
		log.Fatalf("Failed to create app context: %v", err)
	}

	// Initialize Doorman client
	if err := clients.NewDoormanClient(appCtx.Environment); err != nil {
		log.Fatalf("Failed to initialize Doorman client: %v", err)
	}

	// Initialize Jira client
	if err := clients.NewJiraClientSingleton(appCtx.Environment); err != nil {
		log.Fatalf("Failed to initialize Jira client: %v", err)
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
