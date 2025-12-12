package txn

import (
	"fmt"
)

// getAllSOPCases returns all defined SOP cases except SOPCaseNone
func getAllSOPCases() []SOPCase {
	return []SOPCase{
		SOPCasePcExternalPaymentFlow200_11,
		SOPCasePcExternalPaymentFlow201_0RPP210,
		SOPCasePcExternalPaymentFlow201_0RPP900,
		SOPCasePeTransferPayment210_0,
		SOPCasePe2200FastCashinFailed,
		SOPCaseRppCashoutReject101_19,
		SOPCaseRppQrPaymentReject210_0,
		SOPCaseRppNoResponseResume,
	}
}

// calculateSummaryStats calculates and prints summary statistics for the processed transactions
func calculateSummaryStats(results []TransactionResult) {
	foundCount := 0
	errorCount := 0

	// Use a map to track case counts - this makes it easier to add new cases
	caseCounts := make(map[SOPCase]int)

	// Initialize all known cases with 0 count to ensure they appear in output
	allCases := getAllSOPCases()
	for _, caseType := range allCases {
		caseCounts[caseType] = 0
	}

	// We re-iterate through results. Note: Since identifySOPCase is now potentially
	// interactive or side-effecting (if we hadn't already run it), we should
	// ideally rely on the classification done during GenerateSQLStatements.
	// However, for this summary function, we will re-run identification logic.
	// Since data is already enriched in the 'results' slice (pointers used in GenerateSQL),
	// the interactive prompts won't trigger again if fields are populated.

	for i := range results {
		// Use pointer to access enriched data
		result := &results[i]

		notFound := result.PaymentEngine.Transfers.Status == NotFoundStatus || result.PartnerpayEngine.Transfers.Status == NotFoundStatus

		if notFound {
			errorCount++
		} else if result.Error == "" {
			foundCount++

			// Check classification
			caseType := result.CaseType
			if caseType == SOPCaseNone {
				sopRepo := NewSOPRepository()
				sopRepo.IdentifySOPCase(result, "my") // Default to MY for backward compatibility
				caseType = result.CaseType
			}
			if caseType != SOPCaseNone {
				caseCounts[caseType]++
			}
		} else {
			errorCount++
		}
	}

	// Calculate total matches
	totalMatchCount := 0
	for _, count := range caseCounts {
		totalMatchCount += count
	}

	// Build the case details string dynamically using the helper function
	var caseDetails []string
	for _, caseType := range allCases {
		caseDetails = append(caseDetails, fmt.Sprintf("%s: %d", getCaseDisplayName(caseType), caseCounts[caseType]))
	}

	fmt.Printf("Summary: %d found, %d errors/not found\n", foundCount, errorCount)
	fmt.Printf("SQL matches: %d total\n", totalMatchCount)
	for _, detail := range caseDetails {
		fmt.Printf("  %s\n", detail)
	}
}

// getCaseDisplayName returns a user-friendly display name for a SOP case
func getCaseDisplayName(caseType SOPCase) string {
	switch caseType {
	case SOPCasePcExternalPaymentFlow200_11:
		return "pc_external_payment_flow_200_11"
	case SOPCasePcExternalPaymentFlow201_0RPP210:
		return "pc_external_payment_flow_201_0_RPP_210"
	case SOPCasePcExternalPaymentFlow201_0RPP900:
		return "pc_external_payment_flow_201_0_RPP_900"
	case SOPCasePeTransferPayment210_0:
		return "pe_transfer_payment_210_0"
	case SOPCaseRppCashoutReject101_19:
		return "rpp_cashout_reject_101_19"
	case SOPCaseRppQrPaymentReject210_0:
		return "rpp_qr_payment_reject_210_0"
	case SOPCaseRppNoResponseResume:
		return "rpp_no_response_resume"
	default:
		return string(caseType)
	}
}
