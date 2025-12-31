package mybuddy

import (
	"fmt"
	"os"
	"strings"

	"buddy/internal/apps/common"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/utils"

	"github.com/spf13/cobra"
)

func NewRppRtpCashinCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rtp-cashin [e2e-id-or-file]",
		Short: "Resume stuck RPP RTP cashin workflows (state=200, attempt=0)",
		Long: `Resume RPP RTP cashin workflows that are stuck in state 200 with attempt 0.
Supports both single E2E ID and batch file processing.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input := args[0]
			processRppRtpCashin(appCtx, input)
		},
	}
	return cmd
}

func processRppRtpCashin(appCtx *common.Context, input string) {
	if _, err := os.Stat(input); err == nil {
		processBatchFileRtpCashin(appCtx, input)
	} else {
		processSingleE2ERtpCashin(appCtx, input)
	}
}

func processSingleE2ERtpCashin(appCtx *common.Context, e2eID string) {
	result := queryRPPAdapterForE2E(e2eID)
	if result.Error != "" {
		fmt.Printf("%sError: %s\n", appCtx.GetPrefix(), result.Error)
		return
	}

	fmt.Printf("\n%s--- RPP Transaction Status ---\n", appCtx.GetPrefix())
	adapters.WriteResult(os.Stdout, *result, 1)

	sopRepo := adapters.SOPRepo
	sopRepo.IdentifyCase(result, "my")
	if result.CaseType != domain.CaseRppRtpCashinStuck200_0 {
		fmt.Printf("%sThis E2E ID does not match RTP cashin stuck criteria (wf_ct_rtp_cashin, state=200, attempt=0)\n", appCtx.GetPrefix())
		return
	}

	generateRppRtpCashinSQL(appCtx, *result)
}

func processBatchFileRtpCashin(appCtx *common.Context, filePath string) {
	transactionIDs, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("%sError reading file: %v\n", appCtx.GetPrefix(), err)
		return
	}

	fmt.Printf("%sProcessing %d transaction IDs from %s\n", appCtx.GetPrefix(), len(transactionIDs), filePath)

	var matchingResults []domain.TransactionResult
	var allResults []domain.TransactionResult

	for _, id := range transactionIDs {
		result := queryRPPAdapterForE2E(id)
		allResults = append(allResults, *result)

		if result.Error == "" {
			sopRepo := adapters.SOPRepo
			sopRepo.IdentifyCase(result, "my")
			if result.CaseType == domain.CaseRppRtpCashinStuck200_0 {
				matchingResults = append(matchingResults, *result)
			}
		}
	}

	outputPath := filePath + "_RTP_Cashin_Status.txt"
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
		adapters.WriteResult(outputFile, result, i+1)
		if result.Error != "" {
			fmt.Printf("%sError for %s: %s\n", appCtx.GetPrefix(), result.InputID, result.Error)
		}
	}

	fmt.Printf("%sTransaction status written to %s\n", appCtx.GetPrefix(), outputPath)

	if len(matchingResults) == 0 {
		fmt.Printf("%sNo transactions matched RTP cashin stuck criteria\n", appCtx.GetPrefix())
		return
	}

	fmt.Printf("%sFound %d transactions matching RTP cashin stuck criteria\n", appCtx.GetPrefix(), len(matchingResults))

	generateRppRtpCashinSQLBatch(appCtx, matchingResults, filePath)
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

func generateRppRtpCashinSQLBatch(appCtx *common.Context, results []domain.TransactionResult, filePath string) {
	var partnerTxIDs []string
	var runIDs []string

	for _, result := range results {
		if result.RPPAdapter != nil && result.RPPAdapter.PartnerTxID != "" {
			partnerTxIDs = append(partnerTxIDs, result.RPPAdapter.PartnerTxID)
		}
		for _, wf := range result.RPPAdapter.Workflow {
			if wf.RunID != "" {
				runIDs = append(runIDs, wf.RunID)
			}
		}
	}

	if len(partnerTxIDs) == 0 || len(runIDs) == 0 {
		fmt.Printf("%sNo valid IDs found\n", appCtx.GetPrefix())
		return
	}

	// Generate PPE Deploy SQL
	ppeDeploySQL := fmt.Sprintf(`-- rpp_rtp_cashin_stuck_200_0
UPDATE intent SET status = 'UPDATED'
WHERE intent_id IN ('%s')
AND status = 'CONFIRMED';`, strings.Join(partnerTxIDs, "', '"))

	// Generate PPE Rollback SQL
	ppeRollbackSQL := fmt.Sprintf(`-- rpp_rtp_cashin_stuck_200_0 Rollback
UPDATE intent SET status = 'CONFIRMED'
WHERE intent_id IN ('%s');`, strings.Join(partnerTxIDs, "', '"))

	// Generate RPP Deploy SQL
	rppDeploySQL := fmt.Sprintf(`-- rpp_rtp_cashin_stuck_200_0
UPDATE workflow_execution
SET state = 110,
    attempt = 1,
    data = JSON_SET(data, '$.State', 110)
WHERE run_id IN ('%s')
AND state = 200
AND workflow_id = 'wf_ct_rtp_cashin';`, strings.Join(runIDs, "', '"))

	// Generate RPP Rollback SQL
	rppRollbackSQL := fmt.Sprintf(`-- rpp_rtp_cashin_stuck_200_0 Rollback
UPDATE workflow_execution
SET state = 200,
    attempt = 0,
    data = JSON_SET(data, '$.State', 200)
WHERE run_id IN ('%s')
AND workflow_id = 'wf_ct_rtp_cashin';`, strings.Join(runIDs, "', '"))

	// Write to files
	deployPath := filePath + "_PPE_Deploy.sql"
	if err := adapters.WriteSQLFile(deployPath, []string{ppeDeploySQL}); err != nil {
		fmt.Printf("%sError writing PPE deploy SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	rollbackPath := filePath + "_PPE_Rollback.sql"
	if err := adapters.WriteSQLFile(rollbackPath, []string{ppeRollbackSQL}); err != nil {
		fmt.Printf("%sError writing PPE rollback SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	rppDeployPath := filePath + "_RPP_Deploy.sql"
	if err := adapters.WriteSQLFile(rppDeployPath, []string{rppDeploySQL}); err != nil {
		fmt.Printf("%sError writing RPP deploy SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	rppRollbackPath := filePath + "_RPP_Rollback.sql"
	if err := adapters.WriteSQLFile(rppRollbackPath, []string{rppRollbackSQL}); err != nil {
		fmt.Printf("%sError writing RPP rollback SQL: %v\n", appCtx.GetPrefix(), err)
		return
	}

	fmt.Printf("%sPPE Deploy SQL written to %s\n", appCtx.GetPrefix(), deployPath)
	fmt.Printf("%sPPE Rollback SQL written to %s\n", appCtx.GetPrefix(), rollbackPath)
	fmt.Printf("%sRPP Deploy SQL written to %s\n", appCtx.GetPrefix(), rppDeployPath)
	fmt.Printf("%sRPP Rollback SQL written to %s\n", appCtx.GetPrefix(), rppRollbackPath)
}
