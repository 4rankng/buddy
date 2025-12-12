package sgbuddy

import (
	"os"

	"buddy/internal/app"
	"buddy/txn"

	"github.com/spf13/cobra"
)

func NewTxnCmd(appCtx *app.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txn [transaction-id-or-file-path]",
		Short: "Query Singapore transaction status from payment systems",
		Long: `Query transaction status from Singapore payment-engine, payment-core, and fast-adapter databases.

For a single transaction:
  sgbuddy txn 9392fb12b6c64db18e779ae60bdf4307

For multiple transactions from a file:
  sgbuddy txn file-path.txt

Each line in the file should contain a single transaction ID.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]

			// Check if input is a file or a single transaction ID
			if txn.IsFilePath(input) {
				// Process batch file
				txn.ProcessBatchFile(input)
			} else {
				// Process single transaction
				txn.PrintTransactionStatus(input)

				// Query again to check for errors
				result := txn.QueryTransactionStatus(input)
				if result.Error != "" {
					os.Exit(1)
				}
			}
		},
	}

	return cmd
}
