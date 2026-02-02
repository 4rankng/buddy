package sgbuddy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/logging"
	"buddy/internal/clients/doorman"

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
	doormanCmd.AddCommand(NewDoormanQueryCmd(appCtx, clients))

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
				var authErr doorman.AuthenticationUnauthorizedError
				if errors.As(err, &authErr) {
					logger.Warn("Authentication failed (401). Login attempt aborted.")
					if authErr.RequestID != "" {
						logger.Warn("request_id=%s", authErr.RequestID)
					}
					os.Exit(1)
				}
				logger.Error("Failed to create ticket: %v", err)
				os.Exit(1)
			}

			// Construct ticket URL (hardcoded for Singapore environment)
			ticketURL := fmt.Sprintf("https://doorman.sgbank.pr/rds/dml/%s", ticketID)

			logger.Info("Ticket created successfully!")
			logger.Info("Ticket ID: %s", ticketID)
			logger.Info("Ticket URL: %s", ticketURL)
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&serviceName, "service", "s", "", "Service name (payment_engine, payment_core, fast_adapter, partnerpay_engine)")
	cmd.Flags().StringVarP(&originalQuery, "original", "o", "", "Original DML query")
	cmd.Flags().StringVarP(&rollbackQuery, "rollback", "r", "", "Rollback query")
	cmd.Flags().StringVarP(&note, "note", "n", "", "Note/description for the ticket")

	return cmd
}

// NewDoormanQueryCmd creates a command to execute SQL queries via Doorman
func NewDoormanQueryCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	var (
		serviceName string
		query       string
		format      string
	)

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Execute SQL query via Doorman",
		Long:  `Execute a SQL query against the specified service database through Doorman.`,
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
			if query == "" {
				logger.Error("query is required (use --query flag)")
				os.Exit(1)
			}

			logger.Info("Executing query on service: %s", serviceName)

			// Execute query
			var rows []map[string]interface{}
			var err error

			switch serviceName {
			case "payment_engine":
				rows, err = clients.Doorman.QueryPaymentEngine(query)
			case "payment_core":
				rows, err = clients.Doorman.QueryPaymentCore(query)
			case "fast_adapter":
				rows, err = clients.Doorman.QueryFastAdapter(query)
			case "partnerpay_engine":
				rows, err = clients.Doorman.QueryPartnerpayEngine(query)
			default:
				logger.Error("unknown service: %s", serviceName)
				os.Exit(1)
			}

			if err != nil {
				logger.Error("Failed to execute query: %v", err)
				os.Exit(1)
			}

			// Display results
			if len(rows) == 0 {
				fmt.Println("No results found.")
				return
			}

			printQueryResults(rows, format)
		},
	}

	// Define flags
	cmd.Flags().StringVarP(&serviceName, "service", "s", "", "Service name (payment_engine, payment_core, fast_adapter, partnerpay_engine)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "SQL query to execute")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json)")

	return cmd
}

// printQueryResults displays query results in the specified format
func printQueryResults(rows []map[string]interface{}, format string) {
	switch format {
	case "json":
		printJSONResults(rows)
	default:
		printTableResults(rows)
	}
}

// printTableResults displays query results in table format
func printTableResults(rows []map[string]interface{}) {
	if len(rows) == 0 {
		return
	}

	// Get column names from first row
	var columns []string
	for col := range rows[0] {
		columns = append(columns, col)
	}

	// Calculate column widths
	colWidths := make(map[string]int)
	for _, col := range columns {
		colWidths[col] = len(col)
		for _, row := range rows {
			val := fmt.Sprintf("%v", row[col])
			if len(val) > colWidths[col] {
				colWidths[col] = len(val)
			}
		}
	}

	// Print header
	for i, col := range columns {
		fmt.Printf("%-*s", colWidths[col], col)
		if i < len(columns)-1 {
			fmt.Printf(" | ")
		}
	}
	fmt.Println()

	// Print separator
	for i, col := range columns {
		fmt.Printf("%s", strings.Repeat("-", colWidths[col]))
		if i < len(columns)-1 {
			fmt.Printf("-+-")
		}
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, col := range columns {
			fmt.Printf("%-*s", colWidths[col], fmt.Sprintf("%v", row[col]))
			if i < len(columns)-1 {
				fmt.Printf(" | ")
			}
		}
		fmt.Println()
	}
}

// printJSONResults displays query results in JSON format
func printJSONResults(rows []map[string]interface{}) {
	for _, row := range rows {
		jsonBytes, _ := json.Marshal(row)
		fmt.Println(string(jsonBytes))
	}
}
