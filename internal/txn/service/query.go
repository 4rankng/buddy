package service

import (
	"buddy/internal/txn/domain"
	"fmt"
	"regexp"
)

// RppE2EIDPattern matches RPP E2E ID format: first 8 chars are digits (YYYYMMDD) and last 8 chars are digits, total 30 chars
// Deprecated: Use adapters/my.RppE2EIDPattern instead
var RppE2EIDPattern = regexp.MustCompile(`^\d{8}.{14}\d{8}$`)

// Default service instances for backward compatibility
var defaultService = NewTransactionQueryService("my")

// QueryTransactionStatus returns structured data about a transaction
func QueryTransactionStatus(transactionID string) *domain.TransactionResult {
	return QueryTransactionStatusWithEnv(transactionID, "my")
}

// QueryTransactionStatusWithEnv returns structured data about a transaction with specified environment
func QueryTransactionStatusWithEnv(transactionID string, env string) *domain.TransactionResult {
	// Create service for specified environment
	service := NewTransactionQueryService(env)
	return service.QueryTransaction(transactionID)
}

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func QueryPartnerpayEngine(runID string) (domain.PartnerpayEngineInfo, error) {
	return defaultService.QueryPartnerpayEngine(runID)
}

// PrintTransactionStatusWithEnv prints transaction information in the new format with specified environment
func PrintTransactionStatusWithEnv(transactionID string, env string) {
	result := QueryTransactionStatusWithEnv(transactionID, env)
	// Simplified output to avoid circular dependency
	fmt.Printf("\n### [1] transaction_id: %s\n", result.TransactionID)
	fmt.Printf("Environment: %s\n", env)
	// Note: Full formatting would require adapters, using simplified output
}

// Helper function to safely extract string values
func getStringValue(row map[string]interface{}, key string) string {
	if val, ok := row[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
