package mybuddy

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/apps/common/batch"
	"buddy/internal/di"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"

	"github.com/spf13/cobra"
)

func NewTxnCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txn [transaction-id-or-e2e-id-or-file]",
		Short: "Query transaction status and generate remediation SQL",
		Long: `Query the status of a transaction by its ID from the payment engine database.
Supports regular transaction IDs, RPP E2E IDs (format: YYYYMMDDGXSPMYXXXXXXXXXXXXXXXX),
and file paths containing multiple transaction IDs.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			processInput(appCtx, clients, input)
		},
	}

	return cmd
}

func processInput(appCtx *common.Context, clients *di.ClientSet, input string) {
	// Check if input is a file
	if _, err := os.Stat(input); err == nil {
		// Process as batch file using the new batch processor
		batch.ProcessTransactionFile(appCtx, clients, input)
	} else {
		// Process as single transaction ID
		processSingleTransaction(appCtx, clients, input)
	}
}

func processSingleTransaction(appCtx *common.Context, clients *di.ClientSet, transactionID string) {
	// Use the injected transaction service
	txnService := clients.TxnSvc

	// 1. Query transaction
	result := txnService.QueryTransactionWithEnv(transactionID, "my")
	if result == nil {
		fmt.Printf("%sError retrieving transaction details for ID: %s\n", appCtx.GetPrefix(), transactionID)
		return
	}

	// 2. Print Status to console
	fmt.Printf("%sQuerying transaction: %s\n", appCtx.GetPrefix(), transactionID)
	adapters.WriteResult(os.Stdout, *result, 1)

	// Check if PaymentEngine or PartnerpayEngine have NotFoundStatus
	paymentEngineNotFound := result.PaymentEngine != nil && result.PaymentEngine.Transfers.Status == domain.NotFoundStatus
	partnerpayEngineNotFound := result.PartnerpayEngine != nil && result.PartnerpayEngine.Charge.Status == domain.NotFoundStatus

	if paymentEngineNotFound || partnerpayEngineNotFound || result.Error != "" {
		fmt.Printf("Failed to retrieve complete transaction details: %+v\n", *result)
		return
	}

	// 3. Generate SQL
	// The txn package handles interactive prompts (for RPP cases) inside GenerateSQLStatements
	results := []domain.TransactionResult{*result}
	statements := adapters.GenerateSQLStatements(results)

	// 4. Output SQL to console
	printSQLToConsole(appCtx, statements)
}

func printSQLToConsole(appCtx *common.Context, statements domain.SQLStatements) {
	hasOutput := false

	if len(statements.PCDeployStatements) > 0 {
		hasOutput = true
		fmt.Printf("\n%s--- PC Deploy SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.PCDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.PCRollbackStatements) > 0 {
		hasOutput = true
		fmt.Printf("\n%s--- PC Rollback SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.PCRollbackStatements {
			fmt.Println(stmt)
		}
	}

	if len(statements.RPPDeployStatements) > 0 {
		hasOutput = true
		fmt.Printf("\n%s--- RPP Deploy SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.RPPDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.RPPRollbackStatements) > 0 {
		hasOutput = true
		fmt.Printf("\n%s--- RPP Rollback SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.RPPRollbackStatements {
			fmt.Println(stmt)
		}
	}

	if len(statements.PEDeployStatements) > 0 {
		hasOutput = true
		fmt.Printf("\n%s--- PE Deploy SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.PEDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.PERollbackStatements) > 0 {
		hasOutput = true
		fmt.Printf("\n%s--- PE Rollback SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.PERollbackStatements {
			fmt.Println(stmt)
		}
	}

	if !hasOutput {
		fmt.Printf("\n%sNo SQL statements generated. Transaction may not require remediation or case conditions were not met.\n", appCtx.GetPrefix())
	}
}
