package adapters

import (
	"buddy/internal/txn/ports"
	"fmt"
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
