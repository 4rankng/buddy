package batch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"buddy/internal/apps/common"
	"buddy/internal/apps/common/doorman"
	"buddy/internal/di"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/utils"
)

// extractJiraIDFromFilename extracts a Jira ID from a filename
// e.g., "TS-4583.txt" -> "TS-4583", "/path/to/TS-4583.txt" -> "TS-4583"
func extractJiraIDFromFilename(filePath string) string {
	basename := filepath.Base(filePath)
	// Remove .txt extension if present
	jiraID := strings.TrimSuffix(basename, ".txt")
	return jiraID
}

// shouldAutoResumeFromTicket checks if the Jira ticket title indicates auto-resume should be applied
// Returns true if the ticket summary contains "Debit Account confirmation" or "Credit Account confirmation"
func shouldAutoResumeFromTicket(clients *di.ClientSet, jiraID string, prefix string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticket, err := clients.Jira.GetIssueDetails(ctx, jiraID)
	if err != nil {
		fmt.Printf("%sWarning: Failed to fetch Jira ticket %s: %v\n", prefix, jiraID, err)
		return false
	}

	summary := strings.ToLower(ticket.Summary)
	containsDebit := strings.Contains(summary, "debit account confirmation")
	containsCredit := strings.Contains(summary, "credit account confirmation")

	return containsDebit || containsCredit
}

// ProcessTransactionFile processes a file containing multiple transaction IDs
func ProcessTransactionFile(appCtx *common.Context, clients *di.ClientSet, filePath string, autoMode bool) {
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

	// Check auto mode: extract Jira ID from filename and check ticket title
	if autoMode {
		jiraID := extractJiraIDFromFilename(filePath)
		fmt.Printf("%sAuto mode: Checking Jira ticket %s for auto-resume keywords...\n", appCtx.GetPrefix(), jiraID)

		if shouldAutoResumeFromTicket(clients, jiraID, appCtx.GetPrefix()) {
			fmt.Printf("%sAuto mode: Ticket contains confirmation keywords - will auto-resume all eligible transactions\n", appCtx.GetPrefix())
			// Pre-populate auto-choice for CaseCashoutRpp210Pe220Pc201
			// Choice 1 = Resume to Success (which is what option 3 sets)
			adapters.PrepopulateAutoChoice(domain.CaseCashoutRpp210Pe220Pc201, 1)
		} else {
			fmt.Printf("%sAuto mode: Ticket does not contain confirmation keywords - will use interactive prompts\n", appCtx.GetPrefix())
		}
	}

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

		// Clear previous SQL files to avoid appending to old runs
		adapters.ClearSQLFiles()

		// Generate SQL statements
		statements := adapters.GenerateSQLStatements(results)

		// Write SQL to database-specific files
		filesCreated, err := adapters.WriteSQLFiles(statements, filePath)
		if err != nil {
			fmt.Printf("%sError writing SQL files: %v\n", appCtx.GetPrefix(), err)
			return
		}

		// Display generated files
		if len(filesCreated) > 0 {
			fmt.Printf("%sSQL DML files generated: %v\n", appCtx.GetPrefix(), filesCreated)
		} else {
			fmt.Printf("%sNo SQL fixes required for these transactions.\n", appCtx.GetPrefix())
		}

		// Prompt to create Doorman DML tickets for all services combined
		doorman.PromptForDoormanTicket(clients.Doorman, statements, false, "")
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
