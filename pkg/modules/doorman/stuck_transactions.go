package doorman

import (
	"fmt"
	"time"

	"oncall/pkg/ports"
)

// StuckTransactionAnalyzer provides methods for analyzing and fixing stuck transactions
type StuckTransactionAnalyzer struct {
	doorman ports.DoormanPort
}

// NewStuckTransactionAnalyzer creates a new analyzer for stuck transactions
func NewStuckTransactionAnalyzer(doorman ports.DoormanPort) *StuckTransactionAnalyzer {
	return &StuckTransactionAnalyzer{
		doorman: doorman,
	}
}

// StuckTransaction represents a stuck transaction
type StuckTransaction struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Amount        float64   `json:"amount"`
	StuckDuration time.Duration `json:"stuck_duration"`
	Recommendation string   `json:"recommendation"`
}

// TransactionFix represents a fix that can be applied to a transaction
type TransactionFix struct {
	Type         string `json:"type"`
	Description  string `json:"description"`
	RiskLevel    string `json:"risk_level"` // low, medium, high
	RequiresApproval bool `json:"requires_approval"`
}

// AnalyzeStuckTransactions analyzes transactions that have been stuck for too long
func (a *StuckTransactionAnalyzer) AnalyzeStuckTransactions(maxHours int) ([]StuckTransaction, error) {
	// Define common stuck states
	stuckStates := []string{"PENDING", "PROCESSING", "AUTHORIZING", "SETTLING"}

	var allStuckTransactions []StuckTransaction

	for _, state := range stuckStates {
		transactions, err := a.doorman.GetStuckTransactions(state, maxHours)
		if err != nil {
			return nil, fmt.Errorf("failed to get stuck transactions for state %s: %w", state, err)
		}

		stuckTxns := a.convertToStuckTransactions(transactions)
		allStuckTransactions = append(allStuckTransactions, stuckTxns...)
	}

	return allStuckTransactions, nil
}

// GetAvailableFixes returns available fixes for a transaction based on its status
func (a *StuckTransactionAnalyzer) GetAvailableFixes(status string) []TransactionFix {
	switch status {
	case "PENDING":
		return []TransactionFix{
			{
				Type:              "retry",
				Description:       "Retry the transaction by setting status to PENDING and incrementing retry count",
				RiskLevel:         "low",
				RequiresApproval:  false,
			},
			{
				Type:              "mark_failed",
				Description:       "Mark transaction as FAILED",
				RiskLevel:         "medium",
				RequiresApproval:  true,
			},
		}
	case "PROCESSING":
		return []TransactionFix{
			{
				Type:              "retry",
				Description:       "Retry the transaction processing",
				RiskLevel:         "medium",
				RequiresApproval:  true,
			},
			{
				Type:              "mark_failed",
				Description:       "Mark transaction as FAILED after investigation",
				RiskLevel:         "medium",
				RequiresApproval:  true,
			},
		}
	case "AUTHORIZING":
		return []TransactionFix{
			{
				Type:              "retry",
				Description:       "Retry the authorization process",
				RiskLevel:         "medium",
				RequiresApproval:  true,
			},
			{
				Type:              "cancel",
				Description:       "Cancel the authorization attempt",
				RiskLevel:         "high",
				RequiresApproval:  true,
			},
		}
	case "SETTLING":
		return []TransactionFix{
			{
				Type:              "retry",
				Description:       "Retry the settlement process",
				RiskLevel:         "medium",
				RequiresApproval:  true,
			},
		}
	default:
		return []TransactionFix{}
	}
}

// ApplyFix applies a fix to a transaction
func (a *StuckTransactionAnalyzer) ApplyFix(transactionID, fixType string) ([]map[string]interface{}, error) {
	// Get transaction details first to validate
	transactions, err := a.doorman.GetTransactionStatus(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction status: %w", err)
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("transaction not found: %s", transactionID)
	}

	// Apply the fix
	result, err := a.doorman.FixStuckTransaction(transactionID, fixType)
	if err != nil {
		return nil, fmt.Errorf("failed to apply fix %s to transaction %s: %w", fixType, transactionID, err)
	}

	return result, nil
}

// GetTransactionDetails retrieves detailed information about a transaction
func (a *StuckTransactionAnalyzer) GetTransactionDetails(transactionID string) (*TransactionDetails, error) {
	transactions, err := a.doorman.GetTransactionStatus(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction details: %w", err)
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("transaction not found: %s", transactionID)
	}

	txn := transactions[0]
	details := &TransactionDetails{
		TransactionID: getStringValue(txn, "transaction_id"),
		Status:        getStringValue(txn, "status"),
		CreatedAt:     getTimeValue(txn, "created_at"),
		UpdatedAt:     getTimeValue(txn, "updated_at"),
		Amount:        getFloat64Value(txn, "amount"),
		RetryCount:    getIntValue(txn, "retry_count"),
		UserID:        getStringValue(txn, "user_id"),
		MerchantID:    getStringValue(txn, "merchant_id"),
		PaymentMethod: getStringValue(txn, "payment_method"),
	}

	// Calculate stuck duration
	if details.UpdatedAt.After(details.CreatedAt) {
		details.StuckDuration = time.Since(details.UpdatedAt)
	}

	// Get available fixes
	details.AvailableFixes = a.GetAvailableFixes(details.Status)

	return details, nil
}

// TransactionDetails contains comprehensive information about a transaction
type TransactionDetails struct {
	TransactionID  string           `json:"transaction_id"`
	Status         string           `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	Amount         float64          `json:"amount"`
	RetryCount     int              `json:"retry_count"`
	UserID         string           `json:"user_id"`
	MerchantID     string           `json:"merchant_id"`
	PaymentMethod  string           `json:"payment_method"`
	StuckDuration  time.Duration    `json:"stuck_duration"`
	AvailableFixes []TransactionFix `json:"available_fixes"`
}

// convertToStuckTransactions converts database results to StuckTransaction structs
func (a *StuckTransactionAnalyzer) convertToStuckTransactions(transactions []map[string]interface{}) []StuckTransaction {
	var stuckTransactions []StuckTransaction
	now := time.Now()

	for _, txn := range transactions {
		updatedAt := getTimeValue(txn, "updated_at")
		stuckDuration := now.Sub(updatedAt)

		stuckTxn := StuckTransaction{
			TransactionID: getStringValue(txn, "transaction_id"),
			Status:        getStringValue(txn, "status"),
			CreatedAt:     getTimeValue(txn, "created_at"),
			UpdatedAt:     updatedAt,
			Amount:        getFloat64Value(txn, "amount"),
			StuckDuration: stuckDuration,
		}

		// Add recommendation based on stuck duration and status
		stuckTxn.Recommendation = a.getRecommendation(stuckTxn.Status, stuckDuration)

		stuckTransactions = append(stuckTransactions, stuckTxn)
	}

	return stuckTransactions
}

// getRecommendation provides a recommendation based on transaction status and stuck duration
func (a *StuckTransactionAnalyzer) getRecommendation(status string, duration time.Duration) string {
	hours := duration.Hours()

	switch status {
	case "PENDING":
		if hours > 24 {
			return "Consider marking as FAILED - transaction has been pending too long"
		}
		return "Safe to retry - increment retry count and reset to PENDING"
	case "PROCESSING":
		if hours > 2 {
			return "Investigate before retrying - processing for unusually long time"
		}
		return "Monitor closely - still within normal processing time"
	case "AUTHORIZING":
		if hours > 1 {
			return "Consider cancellation - authorization taking too long"
		}
		return "Retry authorization - may be temporary issue"
	case "SETTLING":
		if hours > 4 {
			return "Check with payment provider - settlement delay unusual"
		}
		return "Continue monitoring - settlement can take time"
	default:
		return "Manual investigation required"
	}
}

// Helper functions to safely extract values from map[string]interface{}
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getFloat64Value(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func getTimeValue(m map[string]interface{}, key string) time.Time {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			if t, err := time.Parse(time.RFC3339, str); err == nil {
				return t
			}
			// Try other common formats
			if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}