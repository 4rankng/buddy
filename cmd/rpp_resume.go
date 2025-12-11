package cmd

import (
	"fmt"
	"os"

	"buddy/internal/app"
	"buddy/txn"

	"github.com/spf13/cobra"
)

func NewRppResumeCmd(appCtx *app.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume [e2e-id-or-file]",
		Short: "Resume stuck RPP workflows (state=210, attempt=0)",
		Long: `Resume RPP workflows that are stuck in state 210 with attempt 0.
Supports both wf_ct_cashout and wf_ct_qr_payment workflow types.

The command queries the RPP adapter database directly and generates
SQL to move workflows from state 210 to state 222 for resuming.

Supported inputs:
- Single RPP E2E ID (format: YYYYMMDDGXSPMYXXXXXXXXXXXXXXXX)
- File path containing multiple E2E IDs (one per line)`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			processRppResume(appCtx, input)
		},
	}

	return cmd
}

func processRppResume(appCtx *app.Context, input string) {
	// Check if input is a file
	if _, err := os.Stat(input); err == nil {
		// Process as batch file
		fmt.Printf("%sProcessing batch file: %s\n", appCtx.GetPrefix(), input)
		processBatchFile(appCtx, input)
	} else {
		// Process as single E2E ID
		fmt.Printf("%sResuming RPP workflow for E2E ID: %s\n", appCtx.GetPrefix(), input)
		processSingleE2E(appCtx, input)
	}
}

func processSingleE2E(appCtx *app.Context, e2eID string) {
	// Query the RPP adapter directly for this E2E ID
	result := queryRPPAdapterForE2E(e2eID)
	if result.Error != "" {
		fmt.Printf("%sError: %s\n", appCtx.GetPrefix(), result.Error)
		return
	}

	// Check if it matches the resume criteria
	if !txn.MatchSOPCaseRppNoResponseResume(*result) {
		fmt.Printf("%sThis E2E ID does not match the resume criteria (state=210, attempt=0, workflow_id in ('wf_ct_cashout', 'wf_ct_qr_payment'))\n", appCtx.GetPrefix())
		return
	}

	// Generate SQL for the resume case
	results := []txn.TransactionResult{*result}
	statements := txn.GenerateSQLStatements(results)

	// Output SQL to console
	printSQLToConsole(appCtx, statements)
}

func processBatchFile(appCtx *app.Context, filePath string) {
	// Read E2E IDs from file and process them
	// For now, delegate to the existing batch processing logic
	// but filtered for the resume case only
	txn.ProcessBatchFileWithFilter(filePath, func(result txn.TransactionResult) bool {
		return txn.MatchSOPCaseRppNoResponseResume(result)
	})
}

func queryRPPAdapterForE2E(e2eID string) *txn.TransactionResult {
	// Use the existing query logic from the txn package
	return txn.QueryTransactionStatus(e2eID)
}
