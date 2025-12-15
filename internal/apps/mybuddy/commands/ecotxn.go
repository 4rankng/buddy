package mybuddy

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/txn/adapters"
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
	// Create a TransactionService instance
	txnService := service.NewTransactionQueryService("my")

	// Querys partnerpay-engine database
	info, err := txnService.QueryPartnerpayEngine(runID)
	if err != nil {
		fmt.Printf("%sError querying partnerpay-engine: %v\n", appCtx.GetPrefix(), err)
		return
	}

	// Display the result in the required format
	adapters.WriteEcoTransactionInfo(os.Stdout, info, runID, 1)
}
