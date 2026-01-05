package batch

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/utils"
)

// ProcessTransactionFile processes a file containing multiple transaction IDs
func ProcessTransactionFile(appCtx *common.Context, clients *di.ClientSet, filePath string) {
	fmt.Printf("%sProcessing batch file: %s\n", appCtx.GetPrefix(), filePath)

	// Read transaction IDs from file
	transactionIDs, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("%sError reading file %s: %v\n", appCtx.GetPrefix(), filePath, err)
		return
	}

	if len(transactionIDs) == 0 {
		fmt.Printf("%sNo transaction IDs found in file: %s\n", appCtx.GetPrefix(), filePath)
		return
	}

	fmt.Printf("%sFound %d transaction IDs to process\n", appCtx.GetPrefix(), len(transactionIDs))

	// Process each transaction ID
	var results []domain.TransactionResult
	for i, txnID := range transactionIDs {
		fmt.Printf("%sProcessing %d/%d: %s\n", appCtx.GetPrefix(), i+1, len(transactionIDs), txnID)

		result := clients.TxnSvc.QueryTransactionWithEnv(txnID, appCtx.Environment)
		if result != nil {
			results = append(results, *result)
		} else {
			fmt.Printf("%sError processing transaction ID: %s\n", appCtx.GetPrefix(), txnID)
		}
	}

	// Write batch results to file
	if len(results) > 0 {
		outputPath := filePath + "_results.txt"
		fmt.Printf("%s\nWriting batch results to: %s\n", appCtx.GetPrefix(), outputPath)

		if err := adapters.WriteBatchResults(results, outputPath); err != nil {
			fmt.Printf("%sError writing batch results: %v\n", appCtx.GetPrefix(), err)
		} else {
			fmt.Printf("%sBatch processing completed. Results written to %s\n", appCtx.GetPrefix(), outputPath)
		}

		// Generate SQL statements
		statements := adapters.GenerateSQLStatements(results)

		// Write SQL to database-specific files
		if err := adapters.WriteSQLFiles(statements, filePath); err != nil {
			fmt.Printf("%sError writing SQL files: %v\n", appCtx.GetPrefix(), err)
		}
	}
}

// ProcessRPPResumeFile processes a file containing E2E IDs for RPP resume operations
func ProcessRPPResumeFile(appCtx *common.Context, clients *di.ClientSet, filePath string) {
	fmt.Printf("%sProcessing RPP resume batch file: %s\n", appCtx.GetPrefix(), filePath)

	// Read E2E IDs from file
	e2eIDs, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("%sError reading file %s: %v\n", appCtx.GetPrefix(), filePath, err)
		return
	}

	if len(e2eIDs) == 0 {
		fmt.Printf("%sNo E2E IDs found in file: %s\n", appCtx.GetPrefix(), filePath)
		return
	}

	fmt.Printf("%sFound %d E2E IDs to process for RPP resume\n", appCtx.GetPrefix(), len(e2eIDs))

	// Process each E2E ID for RPP resume
	for i, e2eID := range e2eIDs {
		fmt.Printf("%sProcessing RPP resume %d/%d: %s\n", appCtx.GetPrefix(), i+1, len(e2eIDs), e2eID)

		result := clients.TxnSvc.QueryTransactionWithEnv(e2eID, appCtx.Environment)
		if result != nil {
			// Generate and display RPP resume SQL
			fmt.Printf("%sTransaction found, generating RPP resume SQL...\n", appCtx.GetPrefix())
			adapters.WriteResult(os.Stdout, *result, i+1)
		} else {
			fmt.Printf("%sError processing E2E ID: %s\n", appCtx.GetPrefix(), e2eID)
		}
	}
}

// ProcessRTPCashinFile processes a file containing E2E IDs for RTP cashin operations
func ProcessRTPCashinFile(appCtx *common.Context, clients *di.ClientSet, filePath string) {
	fmt.Printf("%sProcessing RTP cashin batch file: %s\n", appCtx.GetPrefix(), filePath)

	// Read E2E IDs from file
	e2eIDs, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("%sError reading file %s: %v\n", appCtx.GetPrefix(), filePath, err)
		return
	}

	if len(e2eIDs) == 0 {
		fmt.Printf("%sNo E2E IDs found in file: %s\n", appCtx.GetPrefix(), filePath)
		return
	}

	fmt.Printf("%sFound %d E2E IDs to process for RTP cashin\n", appCtx.GetPrefix(), len(e2eIDs))

	// Process each E2E ID for RTP cashin
	for i, e2eID := range e2eIDs {
		fmt.Printf("%sProcessing RTP cashin %d/%d: %s\n", appCtx.GetPrefix(), i+1, len(e2eIDs), e2eID)

		result := clients.TxnSvc.QueryTransactionWithEnv(e2eID, appCtx.Environment)
		if result != nil {
			// Generate and display RTP cashin SQL
			fmt.Printf("%sTransaction found, generating RTP cashin SQL...\n", appCtx.GetPrefix())
			adapters.WriteResult(os.Stdout, *result, i+1)
		} else {
			fmt.Printf("%sError processing E2E ID: %s\n", appCtx.GetPrefix(), e2eID)
		}
	}
}
