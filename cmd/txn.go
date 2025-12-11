package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"oncallmy/txn"

	"github.com/spf13/cobra"
)

var (
	interactiveFlag bool
	caseFlag        int
)

var TxnCmd = &cobra.Command{
	Use:   "txn [transaction-id-or-file]",
	Short: "Query transaction status and generate remediation SQL",
	Long: `Query the status of a transaction by its ID from the payment engine database.
If a file path is provided, it will process all transaction IDs in the file and create an output file.

Remediation SQL is automatically generated based on the detected SOP case:
1. pc_external_payment_flow_200_11 - Force workflows stuck at state 200 attempt 11 to fail cleanly
2. pc_external_payment_flow_201_0_RPP_210 - Resume workflows that never received RPP response
3. pc_external_payment_flow_201_0_RPP_900 - Republish from RPP to resume
4. pe_transfer_payment_210_0 - Reject PE workflows stuck at state 210 before Paynet`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		if interactiveFlag {
			processInteractiveSOP(input)
			return
		}

		if caseFlag > 0 {
			processSOPCase(caseFlag, input)
			return
		}

		// Default processing
		processInput(input)
	},
}

func init() {
	// Add flags for SOP processing
	TxnCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Interactive SOP case selection")
	TxnCmd.Flags().IntVarP(&caseFlag, "case", "c", 0, "Specify SOP case number (1-4)")
}

func processInteractiveSOP(input string) {
	fmt.Println("SOP Case Selection (Verification only - logic is auto-detected):")
	fmt.Println("1. pc_external_payment_flow_200_11 - Force workflows stuck at state 200 attempt 11 to fail cleanly")
	fmt.Println("2. pc_external_payment_flow_201_0_RPP_210 - Resume workflows that never received RPP response")
	fmt.Println("3. pc_external_payment_flow_201_0_RPP_900 - Republish from RPP to resume")
	fmt.Println("4. pe_transfer_payment_210_0 - Reject PE workflows stuck at state 210 before Paynet")
	fmt.Print("Select case (1-4): ")

	reader := bufio.NewReader(os.Stdin)
	caseInput, _ := reader.ReadString('\n')
	caseInput = strings.TrimSpace(caseInput)

	caseNum, err := strconv.Atoi(caseInput)
	if err != nil || caseNum < 1 || caseNum > 4 {
		fmt.Printf("Invalid case selection: %s\n", caseInput)
		return
	}

	processSOPCase(caseNum, input)
}

func processSOPCase(caseNum int, input string) {
	// Since detection is automatic in the txn package, the case flag is primarily
	// for user intent verification. We proceed with standard processing.
	fmt.Printf("Processing %s expecting Case %d...\n", input, caseNum)
	processInput(input)
}

func processInput(input string) {
	// Check if input is a file
	if _, err := os.Stat(input); err == nil {
		// Process as batch file
		// Note: ProcessBatchFile in txn package now handles SQL generation automatically
		// and will write the .sql files to disk.
		txn.ProcessBatchFile(input)
	} else {
		// Process as single transaction ID
		processSingleTransaction(input)
	}
}

func processSingleTransaction(transactionID string) {
	// 1. Print Status to console
	txn.PrintTransactionStatus(transactionID)

	// 2. Fetch Result for processing
	// We query again here to get the struct needed for SQL generation.
	// (Efficiency note: In a larger app, PrintTransactionStatus might return the result to avoid re-query)
	result := txn.QueryTransactionStatus(transactionID)
	if result.TransferStatus == "NOT_FOUND" || result.Error != "" {
		return
	}

	// 3. Generate SQL
	// The txn package handles interactive prompts (for RPP cases) inside GenerateSQLStatements
	results := []txn.TransactionResult{*result}
	statements := txn.GenerateSQLStatements(results)

	// 4. Output SQL to console
	printSQLToConsole(statements)
}

func printSQLToConsole(statements txn.SQLStatements) {
	hasOutput := false

	if len(statements.PCDeployStatements) > 0 {
		hasOutput = true
		fmt.Println("\n--- PC Deploy SQL ---")
		for _, stmt := range statements.PCDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.PCRollbackStatements) > 0 {
		hasOutput = true
		fmt.Println("\n--- PC Rollback SQL ---")
		for _, stmt := range statements.PCRollbackStatements {
			fmt.Println(stmt)
		}
	}

	if len(statements.RPPDeployStatements) > 0 {
		hasOutput = true
		fmt.Println("\n--- RPP Deploy SQL ---")
		for _, stmt := range statements.RPPDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.RPPRollbackStatements) > 0 {
		hasOutput = true
		fmt.Println("\n--- RPP Rollback SQL ---")
		for _, stmt := range statements.RPPRollbackStatements {
			fmt.Println(stmt)
		}
	}

	if len(statements.PEDeployStatements) > 0 {
		hasOutput = true
		fmt.Println("\n--- PE Deploy SQL ---")
		for _, stmt := range statements.PEDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.PERollbackStatements) > 0 {
		hasOutput = true
		fmt.Println("\n--- PE Rollback SQL ---")
		for _, stmt := range statements.PERollbackStatements {
			fmt.Println(stmt)
		}
	}

	if !hasOutput {
		fmt.Println("\nNo SQL statements generated. Transaction may not require remediation or case conditions were not met.")
	}
}
