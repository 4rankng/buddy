package sgbuddy

import (
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/txn/service"
	"buddy/internal/txn/utils"

	"github.com/spf13/cobra"
)

func NewTxnCmd(appCtx *common.Context) *cobra.Command {
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
			if utils.IsSimpleFilePath(input) {
				// Process batch file with Singapore environment
				service.ProcessBatchFileWithEnv(input, "sg")
			} else {
				// Process single transaction with Singapore environment
				service.PrintTransactionStatusWithEnv(input, "sg")

				// Query again to check for errors
				result := service.QueryTransactionStatusWithEnv(input, "sg")
				if result.Error != "" {
					os.Exit(1)
				}
			}
		},
	}

	return cmd
}
