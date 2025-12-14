package adapters

import (
	"buddy/internal/txn/ports"
	"fmt"
	"time"
)

// PaymentEngineAdapter implements the PaymentEnginePort interface
type PaymentEngineAdapter struct {
	client ports.ClientPort
}

// NewPaymentEngineAdapter creates a new PaymentEngineAdapter
func NewPaymentEngineAdapter(client ports.ClientPort) *PaymentEngineAdapter {
	return &PaymentEngineAdapter{
		client: client,
	}
}

func (p *PaymentEngineAdapter) QueryTransfer(transactionID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT transaction_id, status, reference_id, created_at, updated_at, type, txn_subtype, txn_domain, external_id FROM transfer WHERE transaction_id='%s'", transactionID)
	transfers, err := p.client.QueryPaymentEngine(query)
	if err != nil || len(transfers) == 0 {
		return nil, err
	}
	return transfers[0], nil
}

func (p *PaymentEngineAdapter) QueryWorkflow(referenceID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT run_id, workflow_id, prev_trans_id, state, attempt, created_at, updated_at FROM workflow_execution WHERE run_id='%s'", referenceID)
	workflows, err := p.client.QueryPaymentEngine(query)
	if err != nil {
		return nil, err
	}
	if len(workflows) == 0 {
		return nil, fmt.Errorf("no workflow found")
	}
	return workflows[0], nil
}

func (p *PaymentEngineAdapter) QueryTransferByExternalID(externalID, createdAt string) (map[string]interface{}, error) {
	// Parse the created_at timestamp
	parsedTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		// Try alternative formats if RFC3339 fails
		parsedTime, err = time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at timestamp: %w", err)
		}
	}
	
	// Calculate time window (±30 minutes)
	startTime := parsedTime.Add(-30 * time.Minute)
	endTime := parsedTime.Add(30 * time.Minute)
	
	query := fmt.Sprintf(
		"SELECT transaction_id, status, reference_id, created_at, updated_at, type, txn_subtype, txn_domain, external_id "+
		"FROM transfer "+
		"WHERE external_id='%s' "+
		"AND created_at >= '%s' "+
		"AND created_at <= '%s'",
		externalID,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)
	
	transfers, err := p.client.QueryPaymentEngine(query)
	if err != nil || len(transfers) == 0 {
		return nil, err
	}
	
	return transfers[0], nil
}
