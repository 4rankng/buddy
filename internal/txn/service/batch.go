package service

import (
	"fmt"

	"buddy/internal/txn/domain"
	"buddy/internal/txn/utils"
)

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

	fmt.Printf("Processing %d transaction IDs from %s\n", len(ids), filePath)

	// Process each transaction ID
	results := make([]domain.TransactionResult, 0, len(ids))
	for _, id := range ids {
		result := QueryTransactionStatusWithEnv(id, env)
		results = append(results, *result)
	}

	// Generate SQL statements and output results
	// For now, just output basic info without SQL generation to avoid circular dependency
	for i, result := range results {
		fmt.Printf("\n### [%d] transaction_id: %s\n", i+1, result.TransactionID)
		// Note: Full output formatting and SQL generation would require the adapters
		// This is a simplified version to avoid circular dependency
	}
}
