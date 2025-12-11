package txn

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mybuddy/output"
)

// ReadTransactionIDsFromFile reads transaction IDs from a file, one per line
func ReadTransactionIDsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

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
	results := make([]TransactionResult, 0, len(transactionIDs))

	for _, transactionID := range transactionIDs {
		result := QueryTransactionStatus(transactionID)
		results = append(results, *result)
	}

	return results
}

// ProcessBatchFile processes a file containing transaction IDs and writes results to an output file
func ProcessBatchFile(filePath string) {
	// Read transaction IDs from file
	transactionIDs, err := ReadTransactionIDsFromFile(filePath)
	if err != nil {
		output.PrintError(fmt.Errorf("failed to read transaction IDs from file: %v", err))
		return
	}

	fmt.Printf("Processing %d transaction IDs from %s\n", len(transactionIDs), filePath)

	// Process transactions in batch
	results := ProcessBatchTransactions(transactionIDs)

	// Generate output filename
	outputPath := filePath + "-output.txt"

	// Write results to output file
	if err := WriteBatchResults(results, outputPath); err != nil {
		output.PrintError(fmt.Errorf("failed to write results to output file: %v", err))
		return
	}

	fmt.Printf("Results written to %s\n", outputPath)

	// Process all transactions using the unified handler system
	// This will trigger interactive prompts for relevant cases
	sqlStatements := GenerateSQLStatements(results)

	// Write SQL files
	if err := WriteSQLFiles(sqlStatements, filePath); err != nil {
		output.PrintError(fmt.Errorf("failed to write SQL files: %v", err))
		return
	}

	// Print summary
	calculateSummaryStats(results)
}
