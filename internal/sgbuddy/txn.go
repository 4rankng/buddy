package sgbuddy

import (
	"fmt"
	"os"

	"buddy/internal/app"
	"buddy/sgtxn"

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
			if sgtxn.IsFile(input) {
				// Process batch file
				if err := sgtxn.ProcessSGBatchFile(input); err != nil {
					fmt.Fprintf(os.Stderr, "Error processing batch file: %v\n", err)
					os.Exit(1)
				}
			} else {
				// Process single transaction
				result := sgtxn.QuerySGTransaction(input)

				// Display the result in the required format
				sgtxn.PrintSGTransactionStatus(*result, 1)

				// Exit with error code if transaction not found
				if result.Error != "" {
					os.Exit(1)
				}
			}
		},
	}

	return cmd
}
