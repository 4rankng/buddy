package utils

import (
	"buddy/internal/txn/domain"
	"fmt"
	"regexp"
)

// RppE2EIDPattern matches RPP E2E ID format: first 8 chars are digits (YYYYMMDD) and last 8 chars are digits, total 30 chars
var RppE2EIDPattern = regexp.MustCompile(`^\d{8}.{14}\d{8}$`)

// Helper function to safely extract string values
func GetStringValue(row map[string]interface{}, key string) string {
	if val, ok := row[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// DefaultServiceInstance provides backward compatibility for global functions
type DefaultServiceInstance interface {
	QueryTransaction(transactionID string) *domain.TransactionResult
	QueryTransactionWithEnv(transactionID string, env string) *domain.TransactionResult
	QueryPartnerpayEngine(runID string) (domain.PartnerpayEngineInfo, error)
}

// QueryTransactionStatus returns structured data about a transaction
func QueryTransactionStatus(transactionID string, defaultService DefaultServiceInstance) *domain.TransactionResult {
	return defaultService.QueryTransaction(transactionID)
}

// QueryTransactionStatusWithEnv returns structured data about a transaction with specified environment
func QueryTransactionStatusWithEnv(transactionID string, env string, defaultService DefaultServiceInstance) *domain.TransactionResult {
	return defaultService.QueryTransactionWithEnv(transactionID, env)
}

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func QueryPartnerpayEngine(runID string, defaultService DefaultServiceInstance) (domain.PartnerpayEngineInfo, error) {
	return defaultService.QueryPartnerpayEngine(runID)
}

// PrintTransactionStatusWithEnv prints transaction information in the new format with specified environment
func PrintTransactionStatusWithEnv(transactionID string, env string, defaultService DefaultServiceInstance) {
	result := QueryTransactionStatusWithEnv(transactionID, env, defaultService)
	// Simplified output to avoid circular dependency
	fmt.Printf("\n### [1] transaction_id: %s\n", result.TransactionID)
	fmt.Printf("Environment: %s\n", env)
	// Note: Full formatting would require adapters, using simplified output
}
