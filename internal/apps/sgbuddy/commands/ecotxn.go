package sgbuddy

import (
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/service"
	"buddy/internal/txn/utils"

	"github.com/spf13/cobra"
)

func NewEcoTxnCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	var publish bool
	var createDML string

	cmd := &cobra.Command{
		Use:   "ecotxn <transaction-id>",
		Short: "Ecosystem transaction commands for Singapore environment",
		Long: `Commands for querying and managing ecosystem transactions in Singapore's payment systems.

For viewing transaction information:
  sgbuddy ecotxn view <transaction-id>

For publishing transactions (generating SQL scripts):
  sgbuddy ecotxn <transaction-id> --publish

For auto-creating DML tickets:
  sgbuddy ecotxn <transaction-id> --publish --create-dml "TSE-1234"`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]

			if publish {
				// Publish mode - generate SQL scripts
				if utils.IsSimpleFilePath(input) {
					adapters.ProcessEcoTxnPublishBatch(appCtx, clients, input, "sg", createDML)
				} else {
					if err := adapters.ProcessEcoTxnPublish(appCtx, clients, input, "sg", createDML); err != nil {
						os.Exit(1)
					}
				}
			} else {
				// Default to view mode if no subcommand specified
				processEcoTxnView(appCtx, clients, input)
			}
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&publish, "publish", false, "Generate SQL deployment and rollback scripts")
	cmd.Flags().StringVar(&createDML, "create-dml", "", "Auto-create Doorman DML tickets with ticket ID (e.g., \"TSE-1234\")")

	// Add subcommands
	cmd.AddCommand(NewEcoTxnViewCmd(appCtx, clients))

	return cmd
}

func NewEcoTxnViewCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <transaction-id>",
		Short: "View ecosystem transaction information from PartnerPay Engine",
		Long: `Query and display ecosystem transaction information from Singapore's PartnerPay Engine
and Payment Core databases.

For a single transaction:
  sgbuddy ecotxn view de05f9e39aa0485cad6f559f02de9675`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			processEcoTxnView(appCtx, clients, input)
		},
	}

	return cmd
}

func processEcoTxnView(appCtx *common.Context, clients *di.ClientSet, input string) {
	// Check if input is a file or a single transaction ID
	if utils.IsSimpleFilePath(input) {
		// Process batch file with Singapore environment
		service.ProcessEcoBatchFileWithEnv(input, "sg")
	} else {
		// Process single transaction with Singapore environment
		txnService := service.GetTransactionQueryService()
		result := txnService.QueryEcoTransactionWithEnv(input, "sg")

		// Check for errors
		if result.Error != "" {
			os.Exit(1)
		}

		// Print the result
		adapters.WriteResult(os.Stdout, *result, 1)
	}
}
