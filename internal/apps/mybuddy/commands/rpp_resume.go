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

func NewRppResumeCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
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
			processRppResume(appCtx, clients, input)
		},
	}

	return cmd
}

func processRppResume(appCtx *common.Context, clients *di.ClientSet, input string) {
	// Check if input is a file
	if _, err := os.Stat(input); err == nil {
		// Process as batch file using the new batch processor
		batch.ProcessRPPResumeFile(appCtx, clients, input)
	} else {
		// Process as single E2E ID
		fmt.Printf("%sResuming RPP workflow for E2E ID: %s\n", appCtx.GetPrefix(), input)
		processSingleE2E(appCtx, clients, input)
	}
}

func processSingleE2E(appCtx *common.Context, clients *di.ClientSet, e2eID string) {
	// Query RPP adapter directly for this E2E ID using the injected service
	result := clients.TxnSvc.QueryTransactionWithEnv(e2eID, appCtx.Environment)
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
