package txn

import "fmt"

// calculateSummaryStats calculates and prints summary statistics for the processed transactions
func calculateSummaryStats(results []TransactionResult) {
	foundCount := 0
	errorCount := 0
	pcExtPayment200_11MatchCount := 0
	pcExtPayment201_0RPP210MatchCount := 0
	pcExtPayment201_0RPP900MatchCount := 0
	peTransferPayment210_0MatchCount := 0

	// We re-iterate through results. Note: Since identifySOPCase is now potentially
	// interactive or side-effecting (if we hadn't already run it), we should
	// ideally rely on the classification done during GenerateSQLStatements.
	// However, for this summary function, we will re-run identification logic.
	// Since data is already enriched in the 'results' slice (pointers used in GenerateSQL),
	// the interactive prompts won't trigger again if fields are populated.

	for i := range results {
		// Use pointer to access enriched data
		result := &results[i]

		if result.TransferStatus == "NOT_FOUND" {
			errorCount++
		} else if result.Error == "" {
			foundCount++

			// Check classification
			caseType := identifySOPCase(result)
			switch caseType {
			case SOPCasePcExternalPaymentFlow200_11:
				pcExtPayment200_11MatchCount++
			case SOPCasePcExternalPaymentFlow201_0RPP210:
				pcExtPayment201_0RPP210MatchCount++
			case SOPCasePcExternalPaymentFlow201_0RPP900:
				pcExtPayment201_0RPP900MatchCount++
			case SOPCasePeTransferPayment210_0:
				peTransferPayment210_0MatchCount++
			}
		} else {
			errorCount++
		}
	}

	totalMatchCount := pcExtPayment200_11MatchCount + pcExtPayment201_0RPP210MatchCount + pcExtPayment201_0RPP900MatchCount + peTransferPayment210_0MatchCount
	fmt.Printf("Summary: %d found, %d errors/not found\n", foundCount, errorCount)
	fmt.Printf("SQL matches: %d total (pc_external_payment_flow_200_11: %d, pc_external_payment_flow_201_0_RPP_210: %d, pc_external_payment_flow_201_0_RPP_900: %d, pe_transfer_payment_210_0: %d)\n",
		totalMatchCount, pcExtPayment200_11MatchCount, pcExtPayment201_0RPP210MatchCount, pcExtPayment201_0RPP900MatchCount, peTransferPayment210_0MatchCount)
}
