package population

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"buddy/internal/txn/utils"
	"encoding/json"
	"fmt"
)

type pePopulator struct {
	port ports.PaymentEnginePort
}

// NewPEPopulator creates a new PaymentEnginePopulator
func NewPEPopulator(port ports.PaymentEnginePort) PaymentEnginePopulator {
	return &pePopulator{port: port}
}

// QueryByTransactionID fetches transfer and workflow by transaction ID
func (p *pePopulator) QueryByTransactionID(transactionID string) (*domain.PaymentEngineInfo, error) {
	transfer, err := p.port.QueryTransfer(transactionID)
	if err != nil || transfer == nil {
		return nil, fmt.Errorf("failed to query payment engine: %w", err)
	}

	info := &domain.PaymentEngineInfo{
		Transfers: domain.PETransfersInfo{},
		Workflow:  domain.WorkflowInfo{},
	}

	p.populateTransferInfo(info, transfer)

	// Auto-query workflow if reference_id exists
	if info.Transfers.ReferenceID != "" {
		workflow, err := p.port.QueryWorkflow(info.Transfers.ReferenceID)
		if err == nil && workflow != nil {
			p.populateWorkflowInfo(info, workflow)
		}
	}

	return info, nil
}

// QueryByExternalID fetches transfer by external ID within time window
func (p *pePopulator) QueryByExternalID(externalID, createdAt string) (*domain.PaymentEngineInfo, error) {
	transfer, err := p.port.QueryTransferByExternalID(externalID, createdAt)
	if err != nil || transfer == nil {
		return nil, err
	}

	info := &domain.PaymentEngineInfo{
		Transfers: domain.PETransfersInfo{},
		Workflow:  domain.WorkflowInfo{},
	}

	p.populateTransferInfo(info, transfer)

	// Auto-query workflow if reference_id exists
	if info.Transfers.ReferenceID != "" {
		workflow, err := p.port.QueryWorkflow(info.Transfers.ReferenceID)
		if err == nil && workflow != nil {
			p.populateWorkflowInfo(info, workflow)
		}
	}

	return info, nil
}

// QueryWorkflow fetches workflow by reference ID
func (p *pePopulator) QueryWorkflow(referenceID string) (*domain.WorkflowInfo, error) {
	workflow, err := p.port.QueryWorkflow(referenceID)
	if err != nil || workflow == nil {
		return nil, err
	}

	info := &domain.WorkflowInfo{}
	p.populateWorkflowInfoStruct(info, workflow)

	return info, nil
}

// populateTransferInfo extracts transfer data from map to PaymentEngineInfo
func (p *pePopulator) populateTransferInfo(info *domain.PaymentEngineInfo, transfer map[string]interface{}) {
	// Extract status
	if status, ok := transfer["status"].(string); ok {
		info.Transfers.Status = status
	}

	// Extract transaction_id
	if txID, ok := transfer["transaction_id"].(string); ok && txID != "" {
		info.Transfers.TransactionID = txID
	}

	// Extract amount (in cents)
	if amount, ok := transfer["amount"].(float64); ok {
		info.Transfers.Amount = amount
	} else if amountInt, ok := transfer["amount"].(int64); ok {
		info.Transfers.Amount = float64(amountInt)
	}

	// Extract other fields using utils.GetStringValue
	info.Transfers.ReferenceID = utils.GetStringValue(transfer, "reference_id")
	info.Transfers.CreatedAt = utils.GetStringValue(transfer, "created_at")
	info.Transfers.UpdatedAt = utils.GetStringValue(transfer, "updated_at")
	info.Transfers.Type = utils.GetStringValue(transfer, "type")
	info.Transfers.TxnSubtype = utils.GetStringValue(transfer, "txn_subtype")
	info.Transfers.TxnDomain = utils.GetStringValue(transfer, "txn_domain")
	info.Transfers.ExternalID = utils.GetStringValue(transfer, "external_id")
	info.Transfers.SourceAccountID = utils.GetStringValue(transfer, "source_account_id")
	info.Transfers.DestinationAccountID = utils.GetStringValue(transfer, "destination_account_id")
}

// populateWorkflowInfo extracts workflow data from map to PaymentEngineInfo
func (p *pePopulator) populateWorkflowInfo(info *domain.PaymentEngineInfo, workflow map[string]interface{}) {
	p.populateWorkflowInfoStruct(&info.Workflow, workflow)
}

// populateWorkflowInfoStruct extracts workflow data from map to WorkflowInfo struct
func (p *pePopulator) populateWorkflowInfoStruct(workflowInfo *domain.WorkflowInfo, workflow map[string]interface{}) {
	if workflowID, ok := workflow["workflow_id"]; ok {
		workflowInfo.WorkflowID = fmt.Sprintf("%v", workflowID)
	}
	if runID, ok := workflow["run_id"]; ok {
		workflowInfo.RunID = fmt.Sprintf("%v", runID)
	}
	if attemptVal, ok := workflow["attempt"]; ok {
		if attemptFloat, ok := attemptVal.(float64); ok {
			workflowInfo.Attempt = int(attemptFloat)
		}
	}
	if state, ok := workflow["state"]; ok {
		if stateInt, ok := state.(float64); ok {
			workflowInfo.State = fmt.Sprintf("%d", int(stateInt))
		} else {
			workflowInfo.State = fmt.Sprintf("%v", state)
		}
	}

	// Populate prev_trans_id field
	if prevTransID, ok := workflow["prev_trans_id"]; ok {
		workflowInfo.PrevTransID = fmt.Sprintf("%v", prevTransID)
	}

	// Populate data field (full JSON data)
	if data, ok := workflow["data"]; ok {
		if dataStr, ok := data.(string); ok {
			workflowInfo.Data = dataStr
		} else {
			// Convert to JSON string if not already
			if dataBytes, err := json.Marshal(data); err == nil {
				workflowInfo.Data = string(dataBytes)
			}
		}
	}
}
