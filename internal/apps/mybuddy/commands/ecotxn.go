package mybuddy

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/service"

	"github.com/spf13/cobra"
)

func NewEcoTxnCmd(appCtx *common.Context) *cobra.Command {
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
			processEcoTransaction(appCtx, runID)
		},
	}

	return cmd
}

func processEcoTransaction(appCtx *common.Context, runID string) {
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

	// Generate and write DML files if case is identified
	if result.CaseType != "" && result.CaseType != "CaseNone" {
		// Generate SQL statements based on the identified case
		statements := adapters.GenerateSQLStatements([]domain.TransactionResult{*result})

		// Write SQL files to current directory
		basePath := fmt.Sprintf("ecotxn_%s", runID)
		if err := adapters.WriteSQLFiles(statements, basePath); err != nil {
			fmt.Printf("\nWarning: Failed to write DML files: %v\n", err)
		} else {
			fmt.Printf("\nDML files written successfully for case: %s\n", result.CaseType)
		}
	}
}
