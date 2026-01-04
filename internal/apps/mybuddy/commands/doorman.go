package mybuddy

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/logging"

	"github.com/spf13/cobra"
)

// NewDoormanCmd creates the doorman command group
func NewDoormanCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	doormanCmd := &cobra.Command{
		Use:   "doorman",
		Short: "Doorman DML ticket operations",
		Long:  `Create and manage Doorman DML tickets for database changes`,
	}

	doormanCmd.AddCommand(NewDoormanCreateDMLCmd(appCtx, clients))

	return doormanCmd
}

// NewDoormanCreateDMLCmd creates a command to create DML tickets in Doorman
func NewDoormanCreateDMLCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	var (
		serviceName   string
		originalQuery string
		rollbackQuery string
		note          string
	)

	cmd := &cobra.Command{
		Use:   "create-dml",
		Short: "Create a DML ticket in Doorman",
		Long:  `Create a DML ticket in Doorman for the specified service with original and rollback queries.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logging.NewDefaultLogger("doorman")

			// Check if Doorman client is initialized
			if clients.Doorman == nil {
				logger.Error("Doorman client not initialized")
				os.Exit(1)
			}

			// Validate required parameters
			if serviceName == "" {
				logger.Error("service name is required (use --service flag)")
				os.Exit(1)
			}
			if originalQuery == "" {
				logger.Error("original query is required (use --original flag)")
				os.Exit(1)
			}
			if rollbackQuery == "" {
				logger.Error("rollback query is required (use --rollback flag)")
				os.Exit(1)
			}
			if note == "" {
				logger.Error("note is required (use --note flag)")
				os.Exit(1)
			}

			logger.Info("Creating DML ticket for service: %s", serviceName)
			logger.Info("Original query: %s", originalQuery)
			logger.Info("Rollback query: %s", rollbackQuery)

			// Call CreateTicket
			ticketID, err := clients.Doorman.CreateTicket(serviceName, originalQuery, rollbackQuery, note)
			if err != nil {
				logger.Error("Failed to create ticket: %v", err)
				os.Exit(1)
			}

			// Construct ticket URL (hardcoded for Malaysia environment)
			ticketURL := fmt.Sprintf("https://doorman.infra.prd.g-bank.app/rds/dml/%s", ticketID)

			logger.Info("Ticket created successfully!")
			logger.Info("Ticket ID: %s", ticketID)
			logger.Info("Ticket URL: %s", ticketURL)
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&serviceName, "service", "s", "", "Service name (payment_engine, payment_core, fast_adapter, rpp_adapter, partnerpay_engine)")
	cmd.Flags().StringVarP(&originalQuery, "original", "o", "", "Original DML query")
	cmd.Flags().StringVarP(&rollbackQuery, "rollback", "r", "", "Rollback query")
	cmd.Flags().StringVarP(&note, "note", "n", "", "Note/description for the ticket")

	return cmd
}
