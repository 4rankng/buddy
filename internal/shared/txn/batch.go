package txn

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"buddy/output"
)

// ReadTransactionIDsFromFile reads transaction IDs from a file, one per line
func ReadTransactionIDsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			output.LogEvent("batch_file_close_error", map[string]any{"error": err.Error(), "filePath": filePath})
		}
	}()

	var transactionIDs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			transactionIDs = append(transactionIDs, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return transactionIDs, nil
}

// ProcessBatchTransactions processes multiple transaction IDs and returns results
func ProcessBatchTransactions(transactionIDs []string) []TransactionResult {
	return ProcessBatchTransactionsWithEnv(transactionIDs, "my")
}

// ProcessBatchTransactionsWithEnv processes multiple transaction IDs and returns results with specified environment
func ProcessBatchTransactionsWithEnv(transactionIDs []string, env string) []TransactionResult {
	results := make([]TransactionResult, 0, len(transactionIDs))

	for _, transactionID := range transactionIDs {
		result := QueryTransactionStatusWithEnv(transactionID, env)
		results = append(results, *result)
	}

	return results
}

// ProcessBatchFile processes a file containing transaction IDs and writes results to an output file
func ProcessBatchFile(filePath string) {
	ProcessBatchFileWithEnv(filePath, "my")
}

// ProcessBatchFileWithEnv processes a file containing transaction IDs and writes results to an output file with specified environment
func ProcessBatchFileWithEnv(filePath string, env string) {
	// Read transaction IDs from file
	transactionIDs, err := ReadTransactionIDsFromFile(filePath)
	if err != nil {
		output.PrintError(fmt.Errorf("failed to read transaction IDs from file: %v", err))
		return
	}

	fmt.Printf("Processing %d transaction IDs from %s\n", len(transactionIDs), filePath)

	// Process transactions in batch
	results := ProcessBatchTransactionsWithEnv(transactionIDs, env)

	// Identify SOP cases for all transactions FIRST (before any output generation)
	IdentifySOPCasesWithEnv(results, env)

	// Process all transactions using the unified handler system
	// Generate SQL statements from pre-identified SOP cases
	sqlStatements := GenerateSQLStatements(results)

	// Write SQL files
	if err := WriteSQLFiles(sqlStatements, filePath); err != nil {
		output.PrintError(fmt.Errorf("failed to write SQL files: %v", err))
		return
	}

	// Generate output filename
	outputPath := filePath + "-output.txt"

	// Write results to output file AFTER SQL generation (now with SOP cases identified)
	if err := WriteBatchResults(results, outputPath); err != nil {
		output.PrintError(fmt.Errorf("failed to write results to output file: %v", err))
		return
	}

	fmt.Printf("Results written to %s\n", outputPath)

	// Print summary
	calculateSummaryStats(results)
}

// ProcessBatchFileWithFilter processes a file containing transaction IDs with a filter function
func ProcessBatchFileWithFilter(filePath string, filter func(TransactionResult) bool) {
	ProcessBatchFileWithFilterAndEnv(filePath, filter, "my")
}

// ProcessBatchFileWithFilterAndEnv processes a file containing transaction IDs with a filter function and environment
func ProcessBatchFileWithFilterAndEnv(filePath string, filter func(TransactionResult) bool, env string) {
	// Read transaction IDs from file
	transactionIDs, err := ReadTransactionIDsFromFile(filePath)
	if err != nil {
		output.PrintError(fmt.Errorf("failed to read transaction IDs from file: %v", err))
		return
	}

	fmt.Printf("Processing %d transaction IDs from %s\n", len(transactionIDs), filePath)

	// Process transactions in batch
	results := ProcessBatchTransactionsWithEnv(transactionIDs, env)

	// Filter results based on the provided filter function
	filteredResults := make([]TransactionResult, 0)
	for _, result := range results {
		if filter(result) {
			filteredResults = append(filteredResults, result)
		}
	}

	if len(filteredResults) == 0 {
		fmt.Printf("No transactions matched the filter criteria\n")
		return
	}

	fmt.Printf("Found %d transactions matching the filter criteria\n", len(filteredResults))

	// Identify SOP cases for filtered transactions FIRST (before any output generation)
	IdentifySOPCasesWithEnv(filteredResults, env)

	// Process filtered transactions using the unified handler system
	// Generate SQL statements from pre-identified SOP cases
	sqlStatements := GenerateSQLStatements(filteredResults)

	// Write SQL files
	if err := WriteSQLFiles(sqlStatements, filePath); err != nil {
		output.PrintError(fmt.Errorf("failed to write SQL files: %v", err))
		return
	}

	// Generate output filename
	outputPath := filePath + "-output.txt"

	// Write results to output file AFTER SQL generation (now with SOP cases identified)
	if err := WriteBatchResults(filteredResults, outputPath); err != nil {
		output.PrintError(fmt.Errorf("failed to write results to output file: %v", err))
		return
	}

	fmt.Printf("Results written to %s\n", outputPath)

	// Print summary
	calculateSummaryStats(filteredResults)
}
