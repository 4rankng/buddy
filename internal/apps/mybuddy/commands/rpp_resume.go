package mybuddy

import (
	"fmt"
	"os"
	"strings"

	"buddy/internal/apps/common"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/service"
	"buddy/internal/txn/utils"

	"github.com/spf13/cobra"
)

func NewRppResumeCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume [e2e-id-or-file]",
		Short: "Resume stuck RPP workflows (state=210, attempt=0)",
		Long: `Resume RPP workflows that are stuck in state 210 with attempt 0.
Supports both wf_ct_cashout and wf_ct_qr_payment workflow types.

The command queries RPP adapter database directly and generates
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

func processRppResume(appCtx *common.Context, input string) {
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

func processSingleE2E(appCtx *common.Context, e2eID string) {
	// Query RPP adapter directly for this E2E ID
	result := queryRPPAdapterForE2E(e2eID)
	if result.Error != "" {
		fmt.Printf("%sError: %s\n", appCtx.GetPrefix(), result.Error)
		return
	}

	// Always display RPP transaction status and workflow info
	fmt.Printf("\n%s--- RPP Transaction Status ---\n", appCtx.GetPrefix())
	adapters.WriteResult(os.Stdout, *result, 1)

	// Check if it matches resume criteria
	sopRepo := adapters.SOPRepo
	sopRepo.IdentifyCase(result, "my") // Malaysia environment
	if result.CaseType != domain.CaseRppNoResponseResume {
		fmt.Printf("%sThis E2E ID does not match resume criteria (state=210, attempt=0, workflow_id in ('wf_ct_cashout', 'wf_ct_qr_payment'))\n", appCtx.GetPrefix())
		return
	}

	// Generate SQL ONLY for the resume case
	generateRppResumeSQL(appCtx, *result)
}

func processBatchFile(appCtx *common.Context, filePath string) {
	// Read E2E IDs from file and process them
	transactionIDs, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("%sError reading file: %v\n", appCtx.GetPrefix(), err)
		return
	}

	fmt.Printf("%sProcessing %d transaction IDs from %s\n", appCtx.GetPrefix(), len(transactionIDs), filePath)

	// Process transactions and collect matching ones
	var matchingResults []domain.TransactionResult
	var allResults []domain.TransactionResult

	for _, id := range transactionIDs {
		result := queryRPPAdapterForE2E(id)
		allResults = append(allResults, *result)

		// Check if it matches resume criteria
		if result.Error == "" {
			sopRepo := adapters.SOPRepo
			sopRepo.IdentifyCase(result, "my") // Malaysia environment
			if result.CaseType == domain.CaseRppNoResponseResume {
				matchingResults = append(matchingResults, *result)
			}
		}
	}

	// Always display RPP transaction status and workflow info for all transactions
	// Write to output file instead of stdout
	outputPath := filePath + "_RPP_Status.txt"
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("%sError creating output file: %v\n", appCtx.GetPrefix(), err)
		return
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Printf("%sError closing output file: %v\n", appCtx.GetPrefix(), err)
		}
	}()

	fmt.Printf("\n%s--- RPP Transaction Status for All Transactions ---\n", appCtx.GetPrefix())
	if _, err := fmt.Fprintf(outputFile, "--- RPP Transaction Status for All Transactions ---\n"); err != nil {
		fmt.Printf("%sError writing to output file: %v\n", appCtx.GetPrefix(), err)
		return
	}

	for i, result := range allResults {
		// Always write the result in proper format, even if there's an error
		adapters.WriteResult(outputFile, result, i+1)

		// Also show error messages on console for visibility
		if result.Error != "" {
			fmt.Printf("%sError for %s: %s\n", appCtx.GetPrefix(), result.TransactionID, result.Error)
		}
	}

	fmt.Printf("%sTransaction status written to %s\n", appCtx.GetPrefix(), outputPath)

	if len(matchingResults) == 0 {
		fmt.Printf("%sNo transactions matched the resume criteria\n", appCtx.GetPrefix())
		return
	}

	fmt.Printf("%sFound %d transactions matching the resume criteria\n", appCtx.GetPrefix(), len(matchingResults))

	// Generate SQL ONLY for the resume case for all matching transactions
	generateRppResumeSQLBatch(appCtx, matchingResults, filePath)
}

func generateRppResumeSQL(appCtx *common.Context, result domain.TransactionResult) {
	// Get the SQL template for the resume case only
	ticket := adapters.GetDMLTicketForRppResume(result)
	if ticket == nil {
		fmt.Printf("%sNo SQL generated\n", appCtx.GetPrefix())
		return
	}

	// Generate SQL statements
	statements, err := adapters.GenerateSQLFromTicket(*ticket)
	if err != nil {
		fmt.Printf("%sError generating SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

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

func generateRppResumeSQLBatch(appCtx *common.Context, results []domain.TransactionResult, filePath string) {
	// Collect all run_ids
	var runIDs []string
	for _, result := range results {
		if result.RPPAdapter.Workflow.RunID != "" {
			runIDs = append(runIDs, result.RPPAdapter.Workflow.RunID)
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
	if err := adapters.WriteSQLFile(deployPath, []string{deploySQL}); err != nil {
		fmt.Printf("%sError writing deploy SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	rollbackPath := filePath + "_RPP_Rollback.sql"
	if err := adapters.WriteSQLFile(rollbackPath, []string{rollbackSQL}); err != nil {
		fmt.Printf("%sError writing rollback SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	fmt.Printf("%sDeploy SQL written to %s\n", appCtx.GetPrefix(), deployPath)
	fmt.Printf("%sRollback SQL written to %s\n", appCtx.GetPrefix(), rollbackPath)
}

func queryRPPAdapterForE2E(e2eID string) *domain.TransactionResult {
	// Use the existing query logic from the txn package
	return service.QueryTransactionStatus(e2eID)
}
