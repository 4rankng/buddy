package service

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/utils"
	"os"
	"regexp"
)

// RppE2EIDPattern matches RPP E2E ID format: first 8 chars are digits (YYYYMMDD) and last 8 chars are digits, total 30 chars
// Deprecated: Use adapters/my.RppE2EIDPattern instead
var RppE2EIDPattern = regexp.MustCompile(`^\d{8}.{14}\d{8}$`)

// Default service instances for backward compatibility
var defaultService = NewTransactionQueryService("my")

// QueryTransactionStatus returns structured data about a transaction
func QueryTransactionStatus(transactionID string) *domain.TransactionResult {
	return utils.QueryTransactionStatus(transactionID, defaultService)
}

// QueryTransactionStatusWithEnv returns structured data about a transaction with specified environment
func QueryTransactionStatusWithEnv(transactionID string, env string) *domain.TransactionResult {
	// Create service for specified environment
	service := NewTransactionQueryService(env)
	return utils.QueryTransactionStatus(transactionID, service)
}

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func QueryPartnerpayEngine(runID string) (domain.PartnerpayEngineInfo, error) {
	return utils.QueryPartnerpayEngine(runID, defaultService)
}

// PrintTransactionStatusWithEnv prints transaction information in the new format with specified environment
func PrintTransactionStatusWithEnv(transactionID string, env string) {
	service := NewTransactionQueryService(env)
	result := utils.QueryTransactionStatusWithEnv(transactionID, env, service)
	// Use the full formatting from adapters package
	adapters.WriteResult(os.Stdout, *result, 1)
}
