package main

import (
	"log"
	"os"

	"buddy/internal/apps/common"
	cobraPkg "buddy/internal/apps/common/cobra"
	sgbuddyCmd "buddy/internal/apps/sgbuddy/commands"
	"buddy/internal/clients/doorman"
	clients "buddy/internal/clients/jira"
	"buddy/internal/txn/service"
)

func main() {
	// Create app context for sgbuddy
	appCtx, err := common.NewContext("sgbuddy")
	if err != nil {
		log.Fatalf("Failed to create app context: %v", err)
	}

	// Initialize Doorman client
	if doorman.NewDoormanClient(appCtx.Environment) == nil {
		log.Fatalf("Failed to initialize Doorman client")
	}

	// Initialize Jira client
	if clients.NewJiraClient(appCtx.Environment) == nil {
		log.Fatalf("Failed to initialize Jira client")
	}

	// Initialize TransactionService
	if service.NewTransactionQueryService(appCtx.Environment) == nil {
		log.Fatalf("Failed to initialize TransactionService")
	}

	// 1. Get the base command
	rootCmd := cobraPkg.NewRootCommand(appCtx)

	// 2. Add sgbuddy specific commands to the root
	rootCmd.AddCommand(sgbuddyCmd.GetCommands(appCtx)...)

	// 3. Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
