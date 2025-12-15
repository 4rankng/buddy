package adapters

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"encoding/json"
	"fmt"
)

// PartnerpayEngineAdapter implements the PartnerpayEnginePort interface
type PartnerpayEngineAdapter struct {
	client ports.ClientPort
}

// NewPartnerpayEngineAdapter creates a new PartnerpayEngineAdapter
func NewPartnerpayEngineAdapter(client ports.ClientPort) *PartnerpayEngineAdapter {
	return &PartnerpayEngineAdapter{
		client: client,
	}
}

func (p *PartnerpayEngineAdapter) QueryCharge(transactionID string) (domain.PartnerpayEngineInfo, error) {
	if p.client == nil {
		return domain.PartnerpayEngineInfo{}, fmt.Errorf("QueryCharge: database client is not initialized")
	}
	query := fmt.Sprintf("SELECT status, status_reason, status_reason_description, transaction_id, created_at, updated_at FROM charge WHERE transaction_id='%s'", transactionID)
	charges, err := p.client.QueryPartnerpayEngine(query)
	if err != nil {
		return domain.PartnerpayEngineInfo{}, fmt.Errorf("failed to query charge table: %v", err)
	}
	if len(charges) == 0 {
		return domain.PartnerpayEngineInfo{Transfers: domain.PPEChargeInfo{TransactionID: transactionID, Status: domain.NotFoundStatus}}, nil
	}
	charge := charges[0]
	result := domain.PartnerpayEngineInfo{}
	if txID, ok := charge["transaction_id"].(string); ok && txID != "" {
		result.Transfers.TransactionID = txID
	} else {
		result.Transfers.TransactionID = transactionID
	}
	if status, ok := charge["status"].(string); ok {
		result.Transfers.Status = status
	}
	if statusReason, ok := charge["status_reason"].(string); ok {
		result.Transfers.StatusReason = statusReason
	}
	if statusReasonDesc, ok := charge["status_reason_description"].(string); ok {
		result.Transfers.StatusReasonDescription = statusReasonDesc
	}
	if createdAt, ok := charge["created_at"].(string); ok {
		result.Transfers.CreatedAt = createdAt
	}
	if updatedAt, ok := charge["updated_at"].(string); ok {
		result.Transfers.UpdatedAt = updatedAt
	}
	workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt, data FROM workflow_execution WHERE run_id='%s' AND workflow_id='workflow_charge'", transactionID)
	if workflows, err := p.client.QueryPartnerpayEngine(workflowQuery); err == nil && len(workflows) > 0 {
		workflow := workflows[0]
		if workflowID, ok := workflow["workflow_id"]; ok {
			result.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
		}
		result.Workflow.RunID = transactionID
		if attemptVal, ok := workflow["attempt"]; ok {
			if attemptFloat, ok := attemptVal.(float64); ok {
				result.Workflow.Attempt = int(attemptFloat)
			}
		}
		if state, ok := workflow["state"]; ok {
			if stateInt, ok := state.(float64); ok {
				stateNum := int(stateInt)
				result.Workflow.State = fmt.Sprintf("%d", stateNum)
			} else {
				result.Workflow.State = fmt.Sprintf("%v", state)
			}
		}

		// Extract error information from workflow data
		if dataVal, ok := workflow["data"]; ok {
			if dataStr, ok := dataVal.(string); ok {
				var workflowData map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &workflowData); err == nil {
					// Try to extract error code and message from various possible locations
					if errorCode, ok := workflowData["ErrorCode"].(string); ok {
						result.Transfers.ErrorCode = errorCode
					} else if errorCode, ok := workflowData["error_code"].(string); ok {
						result.Transfers.ErrorCode = errorCode
					}

					if errorMsg, ok := workflowData["ErrorMessage"].(string); ok {
						result.Transfers.ErrorMsg = errorMsg
					} else if errorMsg, ok := workflowData["error_message"].(string); ok {
						result.Transfers.ErrorMsg = errorMsg
					} else if errorMsg, ok := workflowData["error_msg"].(string); ok {
						result.Transfers.ErrorMsg = errorMsg
					}

					// Also check nested ChargeStorage for errors
					if chargeStorage, ok := workflowData["ChargeStorage"].(map[string]interface{}); ok {
						if errorCode, ok := chargeStorage["ErrorCode"].(string); ok {
							result.Transfers.ErrorCode = errorCode
						}
						if errorMsg, ok := chargeStorage["ErrorMessage"].(string); ok {
							result.Transfers.ErrorMsg = errorMsg
						}
					}
				}
			}
		}
	}
	return result, nil
}
