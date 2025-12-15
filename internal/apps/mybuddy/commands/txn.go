package mybuddy

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"buddy/internal/apps/common"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/service"

	"github.com/spf13/cobra"
)

func NewTxnCmd(appCtx *common.Context) *cobra.Command {
	var (
		interactiveFlag bool
		caseFlag        int
	)

	cmd := &cobra.Command{
		Use:   "txn [transaction-id-or-e2e-id-or-file]",
		Short: "Query transaction status and generate remediation SQL",
		Long: `Query the status of a transaction by its ID from the payment engine database.
Supports regular transaction IDs, RPP E2E IDs (format: YYYYMMDDGXSPMYXXXXXXXXXXXXXXXX),
and file paths containing multiple transaction IDs.

Remediation SQL is automatically generated based on the detected SOP case:
1. pc_external_payment_flow_200_11 - Force workflows stuck at state 200 attempt 11 to fail cleanly
2. pc_external_payment_flow_201_0_RPP_210 - Resume workflows that never received RPP response
3. pc_external_payment_flow_201_0_RPP_900 - Republish from RPP to resume
4. pe_transfer_payment_210_0 - Reject PE workflows stuck at state 210 before Paynet
5. rpp_cashout_reject_101_19 - Force fail RPP cashout/QR payment workflows stuck at state 101 attempt 19`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]

			if interactiveFlag {
				processInteractiveSOP(appCtx, input)
				return
			}

			if caseFlag > 0 {
				processCase(appCtx, caseFlag, input)
				return
			}

			// Default processing
			processInput(appCtx, input)
		},
	}

	// Add flags for SOP processing
	cmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Interactive SOP case selection")
	cmd.Flags().IntVarP(&caseFlag, "case", "c", 0, "Specify SOP case number (1-5)")

	return cmd
}

func processInteractiveSOP(appCtx *common.Context, input string) {
	fmt.Printf("%sSOP Case Selection (Verification only - logic is auto-detected):\n", appCtx.GetPrefix())
	fmt.Println("1. pc_external_payment_flow_200_11 - Force workflows stuck at state 200 attempt 11 to fail cleanly")
	fmt.Println("2. pc_external_payment_flow_201_0_RPP_210 - Resume workflows that never received RPP response")
	fmt.Println("3. pc_external_payment_flow_201_0_RPP_900 - Republish from RPP to resume")
	fmt.Println("4. pe_transfer_payment_210_0 - Reject PE workflows stuck at state 210 before Paynet")
	fmt.Println("5. rpp_cashout_reject_101_19 - Force fail RPP cashout/QR payment workflows stuck at state 101 attempt 19")
	fmt.Print("Select case (1-5): ")

	reader := bufio.NewReader(os.Stdin)
	caseInput, _ := reader.ReadString('\n')
	caseInput = strings.TrimSpace(caseInput)

	caseNum, err := strconv.Atoi(caseInput)
	if err != nil || caseNum < 1 || caseNum > 5 {
		fmt.Printf("Invalid case selection: %s\n", caseInput)
		return
	}

	processCase(appCtx, caseNum, input)
}

func processCase(appCtx *common.Context, caseNum int, input string) {
	// Since detection is automatic in the txn package, the case flag is primarily
	// for user intent verification. We proceed with standard processing.
	fmt.Printf("%sProcessing %s expecting Case %d...\n", appCtx.GetPrefix(), input, caseNum)
	processInput(appCtx, input)
}

func processInput(appCtx *common.Context, input string) {
	// Check if input is a file
	if _, err := os.Stat(input); err == nil {
		// Process as batch file
		// Note: ProcessBatchFile in service package now handles batch processing
		fmt.Printf("%sProcessing batch file: %s\n", appCtx.GetPrefix(), input)
		service.ProcessBatchFile(input)
	} else {
		// Process as single transaction ID
		processSingleTransaction(appCtx, input)
	}
}

func processSingleTransaction(appCtx *common.Context, transactionID string) {
	// 1. Print Status to console
	fmt.Printf("%sQuerying transaction: %s\n", appCtx.GetPrefix(), transactionID)
	service.PrintTransactionStatusWithEnv(transactionID, "my")

	// 2. Fetch Result for processing
	// We query again here to get the struct needed for SQL generation.
	// (Efficiency note: In a larger app, PrintTransactionStatus might return the result to avoid re-query)
	result := service.QueryTransactionStatus(transactionID)
	
	// Check if PaymentEngine or PartnerpayEngine have NotFoundStatus
	paymentEngineNotFound := result.PaymentEngine != nil && result.PaymentEngine.Transfers.Status == domain.NotFoundStatus
	partnerpayEngineNotFound := result.PartnerpayEngine != nil && result.PartnerpayEngine.Transfers.Status == domain.NotFoundStatus
	
	if paymentEngineNotFound || partnerpayEngineNotFound || result.Error != "" {
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
