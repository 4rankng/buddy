package service

import (
	"fmt"
	"strings"

	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/utils"
)

// BatchSummary holds summary statistics for batch processing
type BatchSummary struct {
	Total     int
	Matched   int
	Unmatched int
	CaseTypes map[domain.SOPCase]int
}

// ProcessBatchFile processes a file containing multiple transaction IDs
func ProcessBatchFile(filePath string) {
	processBatchFileWithEnv(filePath, "my")
}

// ProcessBatchFileWithEnv processes a file with specified environment
func ProcessBatchFileWithEnv(filePath, env string) {
	processBatchFileWithEnv(filePath, env)
}

// processBatchFileWithEnv is the internal implementation
func processBatchFileWithEnv(filePath, env string) {
	// Read transaction IDs from file
	ids, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	if len(ids) == 0 {
		fmt.Printf("No transaction IDs found in %s\n", filePath)
		return
	}

	fmt.Printf("[MY] Processing batch file: %s\n", filePath)
	if env == "sg" {
		fmt.Printf("[SG] Processing batch file: %s\n", filePath)
	}
	fmt.Printf("Processing %d transaction IDs from %s\n", len(ids), filePath)

	// Process each transaction ID
	results := make([]domain.TransactionResult, 0, len(ids))
	for _, id := range ids {
		result := QueryTransactionStatusWithEnv(id, env)
		results = append(results, *result)
	}

	// Generate output path
	outputPath := generateOutputPath(filePath)

	// Write detailed results to output file
	if err := adapters.WriteBatchResults(results, outputPath); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	// Generate and display summary
	summary := generateBatchSummary(results)
	printBatchSummary(filePath, summary, outputPath)
}

// generateOutputPath creates the output file path by appending "-output.txt"
func generateOutputPath(inputPath string) string {
	if strings.HasSuffix(strings.ToLower(inputPath), ".txt") {
		return inputPath + "-output.txt"
	}
	return inputPath + ".txt-output.txt"
}

// generateBatchSummary creates summary statistics from transaction results
func generateBatchSummary(results []domain.TransactionResult) BatchSummary {
	summary := BatchSummary{
		Total:     len(results),
		Matched:   0,
		Unmatched: 0,
		CaseTypes: make(map[domain.SOPCase]int),
	}

	// Initialize case type counters
	caseTypes := []domain.SOPCase{
		domain.SOPCasePcExternalPaymentFlow200_11,
		domain.SOPCasePcExternalPaymentFlow201_0RPP210,
		domain.SOPCasePcExternalPaymentFlow201_0RPP900,
		domain.SOPCasePeTransferPayment210_0,
		domain.SOPCasePe2200FastCashinFailed,
		domain.SOPCaseRppCashoutReject101_19,
		domain.SOPCaseRppQrPaymentReject210_0,
		domain.SOPCaseRppNoResponseResume,
	}

	for _, caseType := range caseTypes {
		summary.CaseTypes[caseType] = 0
	}

	// Count transactions
	for _, result := range results {
		// Check if transaction is matched or unmatched
		// A transaction is "unmatched" if SOP case is NOT_FOUND (SOPCaseNone) or there's an error
		isUnmatched := result.CaseType == domain.SOPCaseNone || result.Error != ""

		if isUnmatched {
			summary.Unmatched++
		} else {
			summary.Matched++
		}

		// Count case types (excluding NOT_FOUND cases)
		if result.CaseType != domain.SOPCaseNone && result.CaseType != "" {
			summary.CaseTypes[result.CaseType]++
		}
	}

	return summary
}

// printBatchSummary prints the summary to stdout in the exact format specified
func printBatchSummary(filename string, summary BatchSummary, outputPath string) {
	fmt.Printf("\n--- Generating SQL Statements ---\n")
	fmt.Printf("Results written to %s\n", outputPath)
	fmt.Printf("Summary: \n")
	fmt.Printf("  Total: %d\n", summary.Total)
	fmt.Printf("  Unmatched: %d\n", summary.Unmatched)
	fmt.Printf("  Matched: %d\n", summary.Matched)
	fmt.Printf("Case Type Breakdown\n")

	// Print case types in the specific order from the documentation
	caseOrder := []struct {
		caseType domain.SOPCase
		name     string
	}{
		{domain.SOPCasePcExternalPaymentFlow200_11, "pc_external_payment_flow_200_11"},
		{domain.SOPCasePcExternalPaymentFlow201_0RPP210, "pc_external_payment_flow_201_0_RPP_210"},
		{domain.SOPCasePcExternalPaymentFlow201_0RPP900, "pc_external_payment_flow_201_0_RPP_900"},
		{domain.SOPCasePeTransferPayment210_0, "pe_transfer_payment_210_0"},
		{domain.SOPCasePe2200FastCashinFailed, "pe_220_0_fast_cashin_failed"},
		{domain.SOPCaseRppCashoutReject101_19, "rpp_cashout_reject_101_19"},
		{domain.SOPCaseRppQrPaymentReject210_0, "rpp_qr_payment_reject_210_0"},
		{domain.SOPCaseRppNoResponseResume, "rpp_no_response_resume"},
	}

	for _, ct := range caseOrder {
		fmt.Printf("  %s: %d\n", ct.name, summary.CaseTypes[ct.caseType])
	}
}
