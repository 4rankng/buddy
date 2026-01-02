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
	CaseTypes map[domain.Case]int
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

	fmt.Printf("Processing %d transaction IDs from %s\n", len(ids), filePath)

	// Get the TransactionService singleton for batch processing
	txnService := GetTransactionQueryService()

	// Process each transaction ID
	results := make([]domain.TransactionResult, 0, len(ids))
	for _, id := range ids {
		result := txnService.QueryTransactionWithEnv(id, env)
		results = append(results, *result)
	}

	// Generate output path
	outputPath := generateOutputPath(filePath)

	// Write detailed results to output file
	if err := adapters.WriteBatchResults(results, outputPath); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		return
	}

	// Generate SQL statements
	statements := adapters.GenerateSQLStatements(results)

	// Clear existing SQL files before writing (for batch mode, always start fresh)
	adapters.ClearSQLFiles()

	// Write SQL files
	sqlBasePath := strings.TrimSuffix(outputPath, "-output.txt")
	if err := adapters.WriteSQLFiles(statements, sqlBasePath); err != nil {
		fmt.Printf("Error writing SQL files: %v\n", err)
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
		CaseTypes: make(map[domain.Case]int),
	}

	// Initialize case type counters
	for _, caseType := range domain.GetCaseSummaryOrder() {
		summary.CaseTypes[caseType] = 0
	}

	// Count transactions
	for _, result := range results {

		// Check if transaction is matched or unmatched
		// A transaction is "unmatched" if SOP case is NOT_FOUND (CaseNone), empty, or there's an error
		isUnmatched := result.CaseType == domain.CaseNone || result.CaseType == "" || result.Error != ""

		if isUnmatched {
			summary.Unmatched++

		} else {
			summary.Matched++

		}

		// Count case types (excluding NOT_FOUND and empty cases)
		if result.CaseType != domain.CaseNone && result.CaseType != "" {
			summary.CaseTypes[result.CaseType]++

		}
	}

	return summary
}

// printBatchSummary prints the summary to stdout in the exact format specified
func printBatchSummary(filename string, summary BatchSummary, outputPath string) {

	fmt.Printf("Results written to %s\n", outputPath)
	fmt.Printf("Summary: \n")
	fmt.Printf("  Total: %d\n", summary.Total)
	fmt.Printf("  Unmatched: %d\n", summary.Unmatched)
	fmt.Printf("  Matched: %d\n", summary.Matched)
	fmt.Printf("Case Type Breakdown\n")

	// Print case types in the specific order from the documentation
	for _, caseType := range domain.GetCaseSummaryOrder() {
		fmt.Printf("  %s: %d\n", string(caseType), summary.CaseTypes[caseType])
	}
}
