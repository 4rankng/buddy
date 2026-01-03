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

func NewRppRtpCashinCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rtp-cashin [e2e-id-or-file]",
		Short: "Resume stuck RPP RTP cashin workflows (state=200, attempt=0)",
		Long: `Resume RPP RTP cashin workflows that are stuck in state 200 with attempt 0.
Supports both single E2E ID and batch file processing.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			processRppRtpCashin(appCtx, clients, input)
		},
	}
	return cmd
}

func processRppRtpCashin(appCtx *common.Context, clients *di.ClientSet, input string) {
	if _, err := os.Stat(input); err == nil {
		// Process as batch file using the new batch processor
		batch.ProcessRTPCashinFile(appCtx, clients, input)
	} else {
		processSingleE2ERtpCashin(appCtx, clients, input)
	}
}

func processSingleE2ERtpCashin(appCtx *common.Context, clients *di.ClientSet, e2eID string) {
	result := clients.TxnSvc.QueryTransactionWithEnv(e2eID, appCtx.Environment)
	if result.Error != "" {
		fmt.Printf("%sError: %s\n", appCtx.GetPrefix(), result.Error)
		return
	}

	fmt.Printf("\n%s--- RPP Transaction Status ---\n", appCtx.GetPrefix())
	adapters.WriteResult(os.Stdout, *result, 1)

	sopRepo := adapters.SOPRepo
	sopRepo.IdentifyCase(result, appCtx.Environment)

	if result.CaseType != domain.CaseRppRtpCashinStuck200_0 {
		fmt.Printf("%sThis E2E ID does not match RTP cashin stuck criteria (state=200, attempt=0)\n", appCtx.GetPrefix())
		return
	}

	// Generate SQL for the RTP cashin case
	generateRppRtpCashinSQL(appCtx, *result)
}

func generateRppRtpCashinSQL(appCtx *common.Context, result domain.TransactionResult) {
	ticket := adapters.GetDMLTicketForRppRtpCashinStuck200_0(result)
	if ticket == nil {
		fmt.Printf("%sNo SQL generated\n", appCtx.GetPrefix())
		return
	}

	statements, err := adapters.GenerateSQLFromTicket(*ticket)
	if err != nil {
		fmt.Printf("%sError generating SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	// Output SQL to console
	if len(statements.PPEDeployStatements) > 0 {
		fmt.Printf("\n%s--- PPE Deploy SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.PPEDeployStatements {
			fmt.Println(stmt)
		}
	}
	if len(statements.PPERollbackStatements) > 0 {
		fmt.Printf("\n%s--- PPE Rollback SQL ---\n", appCtx.GetPrefix())
		for _, stmt := range statements.PPERollbackStatements {
			fmt.Println(stmt)
		}
	}
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
