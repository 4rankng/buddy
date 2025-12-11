package cmd

import (
	"fmt"
	"os"
	"strings"

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

	// Generate SQL ONLY for the resume case
	generateRppResumeSQL(appCtx, *result)
}

func processBatchFile(appCtx *app.Context, filePath string) {
	// Read E2E IDs from file and process them
	transactionIDs, err := txn.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("%sError reading file: %v\n", appCtx.GetPrefix(), err)
		return
	}

	fmt.Printf("%sProcessing %d transaction IDs from %s\n", appCtx.GetPrefix(), len(transactionIDs), filePath)

	// Process transactions and collect matching ones
	var matchingResults []txn.TransactionResult
	for _, id := range transactionIDs {
		result := queryRPPAdapterForE2E(id)
		if result.Error == "" && txn.MatchSOPCaseRppNoResponseResume(*result) {
			matchingResults = append(matchingResults, *result)
		}
	}

	if len(matchingResults) == 0 {
		fmt.Printf("%sNo transactions matched the resume criteria\n", appCtx.GetPrefix())
		return
	}

	fmt.Printf("%sFound %d transactions matching the resume criteria\n", appCtx.GetPrefix(), len(matchingResults))

	// Generate SQL ONLY for the resume case for all matching transactions
	generateRppResumeSQLBatch(appCtx, matchingResults, filePath)
}

func generateRppResumeSQL(appCtx *app.Context, result txn.TransactionResult) {
	// Get the SQL template for the resume case only
	ticket := txn.GetDMLTicketForRppResume(result)
	if ticket == nil {
		fmt.Printf("%sNo SQL generated\n", appCtx.GetPrefix())
		return
	}

	// Generate SQL statements
	statements := txn.GenerateSQLFromTicket(*ticket)

	// Output to console
	if len(statements.RPPDeployStatements) > 0 {
		fmt.Printf("\n%s--- RPP Deploy SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.RPPDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.RPPRollbackStatements) > 0 {
		fmt.Printf("\n%s--- RPP Rollback SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.RPPRollbackStatements {
			fmt.Println(stmt)
		}
	}
}

func generateRppResumeSQLBatch(appCtx *app.Context, results []txn.TransactionResult, filePath string) {
	// Collect all run_ids
	var runIDs []string
	for _, result := range results {
		if result.RPPWorkflow.RunID != "" {
			runIDs = append(runIDs, result.RPPWorkflow.RunID)
		}
	}

	if len(runIDs) == 0 {
		fmt.Printf("%sNo valid run IDs found\n", appCtx.GetPrefix())
		return
	}

	// Generate SQL using the resume template
	deploySQL := fmt.Sprintf(`-- rpp_no_response_resume_acsp
-- RPP did not respond in time, but status at Paynet is ACSP (Accepted Settlement in Process) or ACTC (Accepted Technical Validation)
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.State', 222)
WHERE run_id IN ('%s')
AND state = 210
AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');`, strings.Join(runIDs, "', '"))

	rollbackSQL := fmt.Sprintf(`-- RPP Rollback: Move workflows back to state 210
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.State', 210)
WHERE run_id IN ('%s')
AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');`, strings.Join(runIDs, "', '"))

	// Write to files
	deployPath := filePath + "_RPP_Deploy.sql"
	if err := txn.WriteSQLFile(deployPath, []string{deploySQL}); err != nil {
		fmt.Printf("%sError writing deploy SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	rollbackPath := filePath + "_RPP_Rollback.sql"
	if err := txn.WriteSQLFile(rollbackPath, []string{rollbackSQL}); err != nil {
		fmt.Printf("%sError writing rollback SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	fmt.Printf("%sDeploy SQL written to %s\n", appCtx.GetPrefix(), deployPath)
	fmt.Printf("%sRollback SQL written to %s\n", appCtx.GetPrefix(), rollbackPath)
}

func queryRPPAdapterForE2E(e2eID string) *txn.TransactionResult {
	// Use the existing query logic from the txn package
	return txn.QueryTransactionStatus(e2eID)
}
