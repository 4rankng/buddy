package sgbuddy

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/utils"

	"github.com/spf13/cobra"
)

func NewEcoTxnCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ecotxn [command]",
		Short: "Ecosystem transaction commands for Singapore environment",
		Long:  `Commands for querying and managing ecosystem transactions in Singapore's payment systems`,
	}

	// Add subcommands
	cmd.AddCommand(NewEcoTxnPublishCmd(appCtx))

	return cmd
}

func NewEcoTxnPublishCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish [transaction-id-or-file-path]",
		Short: "Generate SQL deployment and rollback scripts for Grab transactions",
		Long: `Generate Deploy.sql and Rollback.sql scripts to complete Grab transactions.

For a single transaction:
  sgbuddy ecotxn publish fd230a01dcd04282851b7b9dd6260c93

For multiple transactions from a file:
  sgbuddy ecotxn publish TSE-833.txt

Each line in the file should contain a single transaction ID.

The generated scripts will:
1. Update the charge table in partnerpay-engine
2. Update the workflow_execution table in payment-core
3. Set proper status and timestamps`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			processEcoTxnPublish(appCtx, input)
		},
	}

	return cmd
}

func processEcoTxnPublish(appCtx *common.Context, input string) {
	// Check if input is a file or a single transaction ID
	if utils.IsSimpleFilePath(input) {
		// Process batch file
		adapters.ProcessEcoTxnPublishBatch(input, "sg")
	} else {
		// Process single transaction
		if err := adapters.ProcessEcoTxnPublish(input, "sg"); err != nil {
			fmt.Printf("%sError: %v\n", appCtx.GetPrefix(), err)
			os.Exit(1)
		}
	}
}
