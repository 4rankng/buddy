package mybuddy

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/service"

	"github.com/spf13/cobra"
)

// isEmptyStatements checks if all SQL statement slices are empty
func isEmptyStatements(statements domain.SQLStatements) bool {
	return len(statements.PCDeployStatements) == 0 &&
		len(statements.PCRollbackStatements) == 0 &&
		len(statements.PEDeployStatements) == 0 &&
		len(statements.PERollbackStatements) == 0 &&
		len(statements.PPEDeployStatements) == 0 &&
		len(statements.PPERollbackStatements) == 0 &&
		len(statements.RPPDeployStatements) == 0 &&
		len(statements.RPPRollbackStatements) == 0
}

func NewEcoTxnCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ecotxn [run-id]",
		Short: "Query partnerpay-engine transaction status by run_id",
		Long: `Querys status of a transaction froms partnerpay-engine database using its run_id.
This command specifically queries to charge table and displays workflow_charge information.

Example:
  mybuddy ecotxn fd230a01dcd04282851b7b9dd6260c93`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runID := args[0]
			processEcoTransaction(appCtx, clients, runID)
		},
	}

	return cmd
}

func processEcoTransaction(appCtx *common.Context, clients *di.ClientSet, runID string) {
	// Get the TransactionService singleton
	txnService := service.GetTransactionQueryService()

	// Use the dedicated QueryEcoTransactionWithEnv method for ecological transactions
	result := txnService.QueryEcoTransactionWithEnv(runID, "my")

	// Check for errors
	if result.Error != "" {
		fmt.Printf("%sError querying transaction: %s\n", appCtx.GetPrefix(), result.Error)
		return
	}

	// Display the result in the required format
	adapters.WriteEcoTransactionInfo(os.Stdout, *result, runID, 1)

	// Generate and write DML files if case is identified and not NOT_FOUND
	if result.CaseType != "" && result.CaseType != domain.CaseNone {
		// Generate SQL statements based on the identified case
		statements := adapters.GenerateSQLStatements([]domain.TransactionResult{*result})

		// Debug: Check if any statements were generated
		if isEmptyStatements(statements) {
			fmt.Printf("\nWarning: No SQL statements were generated for case: %s\n", result.CaseType)
			if result.Error != "" {
				fmt.Printf("Error details: %s\n", result.Error)
			}
		}

		// Write SQL files to current directory
		if err := adapters.WriteSQLFiles(statements, ""); err != nil {
			fmt.Printf("\nWarning: Failed to write DML files: %v\n", err)
		} else {
			fmt.Printf("\nDML files written successfully for case: %s\n", result.CaseType)
		}

		// Prompt to create Doorman DML tickets
		PromptForDoormanTicket(appCtx, clients, statements)
	} else if result.CaseType == domain.CaseNone {
		fmt.Printf("\nSkipping DML generation for NOT_FOUND case\n")
	}
}
