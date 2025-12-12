package txn

import (
	"regexp"
)

// RppE2EIDPattern matches RPP E2E ID format: first 8 chars are digits (YYYYMMDD) and last 8 chars are digits, total 30 chars
// Deprecated: Use adapters/my.RppE2EIDPattern instead
var RppE2EIDPattern = regexp.MustCompile(`^\d{8}.{14}\d{8}$`)

// Default service instances for backward compatibility
var defaultService = NewTransactionQueryService("my")

// QueryTransactionStatus returns structured data about a transaction
func QueryTransactionStatus(transactionID string) *TransactionResult {
	return QueryTransactionStatusWithEnv(transactionID, "my")
}

// QueryTransactionStatusWithEnv returns structured data about a transaction with specified environment
func QueryTransactionStatusWithEnv(transactionID string, env string) *TransactionResult {
	// Create service for specified environment
	service := NewTransactionQueryService(env)
	return service.QueryTransaction(transactionID)
}

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func QueryPartnerpayEngine(runID string) (PartnerpayEngineInfo, error) {
	return defaultService.QueryPartnerpayEngine(runID)
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

// getFastAdapterStatusName maps fast adapter status codes to human-readable names
func getFastAdapterStatusName(statusCode int) string {
	// Fast adapter status mapping based on common patterns
	switch statusCode {
	case 0:
		return "INITIATED"
	case 1:
		return "PENDING"
	case 2:
		return "PROCESSING"
	case 3:
		return "SUCCESS"
	case 4:
		return "FAILED"
	case 5:
		return "CANCELLED"
	case 6:
		return "REJECTED"
	case 7:
		return "TIMEOUT"
	case 8:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
