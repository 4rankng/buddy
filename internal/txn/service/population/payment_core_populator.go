package population

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"buddy/internal/txn/utils"
	"fmt"
	"strings"
)

type pcPopulator struct {
	port ports.PaymentCorePort
}

// NewPCPopulator creates a new PaymentCorePopulator
func NewPCPopulator(port ports.PaymentCorePort) PaymentCorePopulator {
	return &pcPopulator{port: port}
}

// QueryInternal fetches internal transactions (AUTH, CAPTURE)
func (p *pcPopulator) QueryInternal(transactionID, createdAt string) ([]domain.PCInternalInfo, error) {
	if createdAt == "" || p.port == nil {
		return nil, nil
	}

	rows, err := p.port.QueryInternalTransactions(transactionID, createdAt)
	if err != nil {
		return nil, err
	}

	results := make([]domain.PCInternalInfo, 0)
	for _, row := range rows {
		txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(row, "tx_type")))
		status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(row, "status")))

		info := domain.PCInternalInfo{
			TxID:      utils.GetStringValue(row, "tx_id"),
			GroupID:   transactionID,
			TxType:    txType,
			TxStatus:  status,
			ErrorCode: utils.GetStringValue(row, "error_code"),
			ErrorMsg:  utils.GetStringValue(row, "error_msg"),
			CreatedAt: utils.GetStringValue(row, "created_at"),
		}

		// Query workflow for this transaction
		info.Workflow = p.QueryWorkflow(info.TxID)

		results = append(results, info)
	}

	return results, nil
}

// QueryExternal fetches external transactions (TRANSFER)
func (p *pcPopulator) QueryExternal(transactionID, createdAt string) ([]domain.PCExternalInfo, error) {
	if createdAt == "" || p.port == nil {
		return nil, nil
	}

	rows, err := p.port.QueryExternalTransactions(transactionID, createdAt)
	if err != nil {
		return nil, err
	}

	results := make([]domain.PCExternalInfo, 0)
	for _, row := range rows {
		txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(row, "tx_type")))
		status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(row, "status")))

		info := domain.PCExternalInfo{
			RefID:     utils.GetStringValue(row, "ref_id"),
			GroupID:   transactionID,
			TxType:    txType,
			TxStatus:  status,
			CreatedAt: utils.GetStringValue(row, "created_at"),
		}

		// Query workflow for this transaction
		info.Workflow = p.QueryWorkflow(info.RefID)

		results = append(results, info)
	}

	return results, nil
}

// QueryWorkflow fetches workflow by run ID
func (p *pcPopulator) QueryWorkflow(runID string) domain.WorkflowInfo {
	workflowInfo := domain.WorkflowInfo{}

	if runID == "" || p.port == nil {
		return workflowInfo
	}

	// Query workflow for this run ID
	workflows, err := p.port.QueryWorkflows([]string{runID})
	if err != nil || len(workflows) == 0 {
		return workflowInfo
	}

	workflow := workflows[0]
	workflowID := utils.GetStringValue(workflow, "workflow_id")
	state := workflow["state"]

	var stateNum int
	if stateInt, ok := state.(float64); ok {
		stateNum = int(stateInt)
	}

	var attempt int
	if attemptFloat, ok := workflow["attempt"].(float64); ok {
		attempt = int(attemptFloat)
	}

	workflowInfo = domain.WorkflowInfo{
		WorkflowID:  workflowID,
		RunID:       runID,
		State:       fmt.Sprintf("%d", stateNum),
		Attempt:     attempt,
		PrevTransID: utils.GetStringValue(workflow, "prev_trans_id"),
	}

	return workflowInfo
}

// PopulatePaymentCoreInfo populates a PaymentCoreInfo from internal and external transactions
func (p *pcPopulator) PopulatePaymentCoreInfo(transactionID, createdAt string) *domain.PaymentCoreInfo {
	if transactionID == "" || createdAt == "" {
		return nil
	}

	pcInfo := &domain.PaymentCoreInfo{}

	// Query internal transactions
	internalTxs, err := p.QueryInternal(transactionID, createdAt)
	if err == nil {
		for _, internalTx := range internalTxs {
			switch internalTx.TxType {
			case "AUTH":
				pcInfo.InternalAuth = internalTx
			case "CAPTURE":
				pcInfo.InternalCapture = internalTx
			}
		}
	}

	// Query external transactions
	externalTxs, err := p.QueryExternal(transactionID, createdAt)
	if err == nil && len(externalTxs) > 0 {
		for _, externalTx := range externalTxs {
			if externalTx.TxType == "TRANSFER" {
				pcInfo.ExternalTransfer = externalTx
				break // Only need first TRANSFER
			}
		}
	}

	return pcInfo
}
