package mybuddy

import (
	"fmt"
	"os"

	"buddy/internal/app"
	"buddy/internal/shared/txn"

	"github.com/spf13/cobra"
)

func NewEcoTxnCmd(appCtx *app.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ecotxn [run-id]",
		Short: "Query partnerpay-engine transaction status by run_id",
		Long: `Query the status of a transaction from the partnerpay-engine database using its run_id.
This command specifically queries the charge table and displays the workflow_charge information.

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

func processEcoTransaction(appCtx *app.Context, runID string) {
	// Query the partnerpay-engine database
	info, err := txn.QueryPartnerpayEngine(runID)
	if err != nil {
		fmt.Printf("%sError querying partnerpay-engine: %v\n", appCtx.GetPrefix(), err)
		return
	}

	// Display the result in the required format
	txn.WriteEcoTransactionInfo(os.Stdout, info, runID, 1)
}
