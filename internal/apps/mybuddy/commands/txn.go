package mybuddy

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"

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

	// 5. Prompt to create Doorman DML
	promptForDoormanTicket(appCtx, clients, statements)
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

func promptForDoormanTicket(appCtx *common.Context, clients *di.ClientSet, statements domain.SQLStatements) {
	// Skip if Doorman client is not initialized
	if clients.Doorman == nil {
		return
	}

	processServiceDML(appCtx, clients, "payment_core", statements.PCDeployStatements, statements.PCRollbackStatements)
	processServiceDML(appCtx, clients, "rpp_adapter", statements.RPPDeployStatements, statements.RPPRollbackStatements)
	processServiceDML(appCtx, clients, "payment_engine", statements.PEDeployStatements, statements.PERollbackStatements)
}

func processServiceDML(appCtx *common.Context, clients *di.ClientSet, serviceName string, deployStmts, rollbackStmts []string) {
	if len(deployStmts) == 0 {
		return
	}

	fmt.Println()
	prompt := promptui.Select{
		Label: fmt.Sprintf("Create Doorman DML ticket for %s?", serviceName),
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if result == "Yes" {
		promptNote := promptui.Prompt{
			Label: "Ticket Note",
		}
		note, err := promptNote.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		originalQuery := strings.Join(deployStmts, "\n")
		rollbackQuery := strings.Join(rollbackStmts, "\n")

		fmt.Printf("Creating ticket for %s...\n", serviceName)
		ticketID, err := clients.Doorman.CreateTicket(serviceName, originalQuery, rollbackQuery, note)
		if err != nil {
			fmt.Printf("Failed to create ticket: %v\n", err)
			return
		}

		ticketURL := fmt.Sprintf("https://doorman.infra.prd.g-bank.app/rds/dml/%s", ticketID)
		fmt.Printf("Ticket created successfully!\nTicket ID: %s\nTicket URL: %s\n", ticketID, ticketURL)
	}
}
