package main

import (
	"os"

	"buddy/internal/apps/common"
	cobraPkg "buddy/internal/apps/common/cobra"
	sgbuddyCmd "buddy/internal/apps/sgbuddy/commands"
	"buddy/internal/di"
	"buddy/internal/logging"
)

func main() {
	logger := logging.NewDefaultLogger("sgbuddy")

	// Create app context for sgbuddy
	appCtx, err := common.NewContext("sgbuddy")
	if err != nil {
		logger.Error("Failed to create app context: %v", err)
		os.Exit(1)
	}

	// Initialize dependency injection container
	container := di.NewContainer()
	if err := container.InitializeForEnvironment(appCtx.Environment); err != nil {
		logger.Error("Failed to initialize services: %v", err)
		os.Exit(1)
	}

	// Get the base command
	rootCmd := cobraPkg.NewRootCommand(appCtx)

	// Add sgbuddy specific commands to the root
	clientSet := container.GetClientSet()
	rootCmd.AddCommand(sgbuddyCmd.GetCommands(appCtx, clientSet)...)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed: %v", err)
		os.Exit(1)
	}
}
