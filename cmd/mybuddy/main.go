package main

import (
	"os"
	"strings"

	"buddy/internal/apps/common"
	cobraPkg "buddy/internal/apps/common/cobra"
	mybuddyCmd "buddy/internal/apps/mybuddy/commands"
	"buddy/internal/di"
	"buddy/internal/logging"
)

func main() {
	logger := logging.NewDefaultLogger("mybuddy")

	// Create app context for mybuddy
	appCtx, err := common.NewContext("mybuddy")
	if err != nil {
		logger.Error("Failed to create app context: %v", err)
		os.Exit(1)
	}

	// Initialize dependency injection container
	container := di.NewContainer()
	if err := container.InitializeForEnvironment(appCtx.Environment); err != nil {
		// Check if this is a 401 Doorman authentication failure - if so, stop immediately
		if strings.Contains(err.Error(), "DOORMAN_AUTH_FAILURE_401") {
			logger.Error("CRITICAL: Doorman authentication failed (401 Unauthorized)")
			logger.Error("%v", err)
			logger.Error("Please verify your Doorman credentials and try again.")
			os.Exit(1)
		}
		logger.Error("Failed to initialize services: %v", err)
		os.Exit(1)
	}

	// Get the base command
	rootCmd := cobraPkg.NewRootCommand(appCtx)

	// Add mybuddy specific commands to the root
	clientSet := container.GetClientSet()
	rootCmd.AddCommand(mybuddyCmd.GetCommands(appCtx, clientSet)...)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		// Check if this is a 401 Doorman authentication failure - if so, stop immediately
		if strings.Contains(err.Error(), "DOORMAN_AUTH_FAILURE_401") {
			logger.Error("CRITICAL: Doorman authentication failed (401 Unauthorized)")
			logger.Error("%v", err)
			logger.Error("Please verify your Doorman credentials and try again.")
			os.Exit(1)
		}
		logger.Error("Command execution failed: %v", err)
		os.Exit(1)
	}
}
