package adapters

import (
	"buddy/internal/txn/ports"
	"fmt"
	"strings"
	"time"
)

// PaymentCoreAdapter implements the PaymentCorePort interface
type PaymentCoreAdapter struct {
	client ports.ClientPort
}

// NewPaymentCoreAdapter creates a new PaymentCoreAdapter
func NewPaymentCoreAdapter(client ports.ClientPort) *PaymentCoreAdapter {
	return &PaymentCoreAdapter{
		client: client,
	}
}

func (p *PaymentCoreAdapter) QueryInternalTransactions(transactionID string, createdAt string) ([]map[string]interface{}, error) {
	if p.client == nil {
		return nil, fmt.Errorf("QueryInternalTransactions: database client is not initialized")
	}
	if createdAt == "" {
		return nil, nil
	}
	startTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	// Use 1-hour window on both sides for ecological transactions
	queryStartTime := startTime.Add(-1 * time.Hour)
	queryEndTime := startTime.Add(1 * time.Hour)
	query := fmt.Sprintf("SELECT tx_id, tx_type, status, error_code, error_msg, created_at FROM internal_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'", transactionID, queryStartTime.Format(time.RFC3339), queryEndTime.Format(time.RFC3339))
	return p.client.QueryPaymentCore(query)
}

func (p *PaymentCoreAdapter) QueryExternalTransactions(transactionID string, createdAt string) ([]map[string]interface{}, error) {
	if p.client == nil {
		return nil, fmt.Errorf("database client is not initialized")
	}
	if createdAt == "" {
		return nil, nil
	}
	startTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	// Use 1-hour window on both sides for ecological transactions
	queryStartTime := startTime.Add(-1 * time.Hour)
	queryEndTime := startTime.Add(1 * time.Hour)
	query := fmt.Sprintf("SELECT ref_id, tx_type, status, created_at FROM external_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'", transactionID, queryStartTime.Format(time.RFC3339), queryEndTime.Format(time.RFC3339))
	return p.client.QueryPaymentCore(query)
}

func (p *PaymentCoreAdapter) QueryWorkflows(runIDs []string) ([]map[string]interface{}, error) {
	if p.client == nil {
		return nil, fmt.Errorf("QueryWorkflows: database client is not initialized")
	}
	if len(runIDs) == 0 {
		return nil, nil
	}
	quotedRunIDs := make([]string, len(runIDs))
	for i, id := range runIDs {
		quotedRunIDs[i] = "'" + id + "'"
	}
	runIDsStr := strings.Join(quotedRunIDs, ", ")
	query := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id IN (%s)", runIDsStr)
	return p.client.QueryPaymentCore(query)
}

// QueryPaymentCore executes a custom query against Payment Core
func (p *PaymentCoreAdapter) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	if p.client == nil {
		return nil, fmt.Errorf("QueryPaymentCore: database client is not initialized")
	}
	return p.client.QueryPaymentCore(query)
}
