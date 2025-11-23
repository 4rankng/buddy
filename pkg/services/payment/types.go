package payment

import "time"

// TransactionStatus represents the status of a transaction
type TransactionStatus int

const (
	StatusPending TransactionStatus = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
	StatusTimeout
	StatusAuthorizing
	StatusSettling
	StatusCancelled
)

// Transaction represents a payment transaction
type Transaction struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CustomerID    string    `json:"customer_id"`
	MerchantID    string    `json:"merchant_id"`
	RetryCount    int       `json:"retry_count"`
}

// StuckTransaction represents a transaction that has been stuck
type StuckTransaction struct {
	TransactionID  string        `json:"transaction_id"`
	Status         string        `json:"status"`
	Amount         float64       `json:"amount"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	RetryCount     int           `json:"retry_count"`
	StuckDuration  time.Duration `json:"stuck_duration"`
	Recommendation string        `json:"recommendation"`
	AvailableFixes []string      `json:"available_fixes"`
}

// FixResult represents the result of applying a fix to a transaction
type FixResult struct {
	TransactionID string    `json:"transaction_id"`
	FixType       string    `json:"fix_type"`
	Success       bool      `json:"success"`
	AffectedRows  int64     `json:"affected_rows"`
	Message       string    `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
}

// TransactionStep represents a step in the transaction flow
type TransactionStep struct {
	Name      string           `json:"name"`
	System    string           `json:"system"`
	Status    TransactionStatus `json:"status"`
	Duration  time.Duration    `json:"duration"`
	Timestamp *time.Time       `json:"timestamp,omitempty"`
}

// TransactionFlow represents the full trace of a transaction
type TransactionFlow struct {
	TransactionID string            `json:"transaction_id"`
	CustomerID    string            `json:"customer_id"`
	TotalAmount   string            `json:"total_amount"`
	Currency      string            `json:"currency"`
	Steps         []TransactionStep `json:"steps"`
}

// SHIPRMResult represents the result of creating a SHIPRM ticket
type SHIPRMResult struct {
	TicketID  string    `json:"ticket_id"`
	TicketURL string    `json:"ticket_url"`
	Accounts  []string  `json:"accounts"`
	CreatedAt time.Time `json:"created_at"`
	Success   bool      `json:"success"`
}

// Helper functions for extracting values from query results

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0.0
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return 0
}

func getTime(data map[string]interface{}, key string) time.Time {
	if val, ok := data[key].(string); ok {
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t
		}
		// Try other common formats
		if t, err := time.Parse("2006-01-02 15:04:05", val); err == nil {
			return t
		}
	}
	return time.Time{}
}