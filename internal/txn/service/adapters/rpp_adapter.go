package adapters

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"buddy/internal/txn/utils"
	"fmt"
)

// RPPAdapter implements the RPPAdapterPort interface (Malaysia only)
type RPPAdapter struct {
	client ports.ClientPort
}

// NewRPPAdapter creates a new RPPAdapter
func NewRPPAdapter(client ports.ClientPort) *RPPAdapter {
	return &RPPAdapter{
		client: client,
	}
}

func (r *RPPAdapter) QueryByE2EID(externalID string) (*domain.RPPAdapterInfo, error) {
	query := fmt.Sprintf("SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status, created_at FROM credit_transfer WHERE end_to_end_id = '%s'", externalID)
	rppResults, err := r.client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil || len(rppResults) == 0 {
		return nil, err
	}
	row := rppResults[0]
	info := &domain.RPPAdapterInfo{
		ReqBizMsgID: utils.GetStringValue(row, "req_biz_msg_id"),
		PartnerTxID: utils.GetStringValue(row, "partner_tx_id"),
		EndToEndID:  externalID,
		Status:      utils.GetStringValue(row, "status"),
		CreatedAt:   utils.GetStringValue(row, "created_at"),
	}
	if info.PartnerTxID != "" {
		workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id='%s'", info.PartnerTxID)
		if workflows, err := r.client.QueryRppAdapter(workflowQuery); err == nil && len(workflows) > 0 {
			workflow := workflows[0]
			info.Workflow.RunID = info.PartnerTxID
			if workflowID, ok := workflow["workflow_id"]; ok {
				info.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
			}
			if state, ok := workflow["state"]; ok {
				if stateInt, ok := state.(float64); ok {
					info.Workflow.State = fmt.Sprintf("%d", int(stateInt))
				} else {
					info.Workflow.State = fmt.Sprintf("%v", state)
				}
			}
			if attempt, ok := workflow["attempt"]; ok {
				if attemptFloat, ok := attempt.(float64); ok {
					info.Workflow.Attempt = int(attemptFloat)
				}
			}
		}
	}
	info.Info = fmt.Sprintf("RPP Status: %s", info.Status)
	return info, nil
}
