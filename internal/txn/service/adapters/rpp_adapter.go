package adapters

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"buddy/internal/txn/utils"
	"fmt"
	"time"
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
	if r.client == nil {
		return nil, fmt.Errorf("QueryByE2EID: database client is not initialized")
	}
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

	// Query workflows using new pattern: time window + req_biz_msg_id filter
	if info.ReqBizMsgID != "" && info.CreatedAt != "" {
		createdAt, err := time.Parse(time.RFC3339Nano, info.CreatedAt)
		if err == nil {
			timeWindowStart := createdAt.Add(-5 * time.Minute)
			timeWindowEnd := createdAt.Add(5 * time.Minute)

			workflowQuery := fmt.Sprintf(
				"SELECT run_id, workflow_id, state, attempt, prev_trans_id, data FROM workflow_execution "+
					"WHERE created_at >= '%s' "+
					"AND created_at <= '%s' "+
					"AND workflow_id = 'wf_process_registry' "+
					"AND data LIKE '%%%s%%'",
				timeWindowStart.Format(time.RFC3339Nano),
				timeWindowEnd.Format(time.RFC3339Nano),
				info.ReqBizMsgID,
			)

			if workflowRows, err := r.client.QueryRppAdapter(workflowQuery); err == nil && len(workflowRows) > 0 {
				info.Workflow = make([]domain.WorkflowInfo, 0, len(workflowRows))
				for _, workflow := range workflowRows {
					wf := domain.WorkflowInfo{}
					if runID, ok := workflow["run_id"]; ok {
						wf.RunID = fmt.Sprintf("%v", runID)
					}
					if workflowID, ok := workflow["workflow_id"]; ok {
						wf.WorkflowID = fmt.Sprintf("%v", workflowID)
					}
					if state, ok := workflow["state"]; ok {
						if stateInt, ok := state.(float64); ok {
							wf.State = fmt.Sprintf("%d", int(stateInt))
						} else {
							wf.State = fmt.Sprintf("%v", state)
						}
					}
					if attempt, ok := workflow["attempt"]; ok {
						if attemptFloat, ok := attempt.(float64); ok {
							wf.Attempt = int(attemptFloat)
						}
					}
					if prevTransID, ok := workflow["prev_trans_id"]; ok {
						wf.PrevTransID = fmt.Sprintf("%v", prevTransID)
					}
					if data, ok := workflow["data"]; ok {
						if dataStr, ok := data.(string); ok {
							wf.Data = dataStr
						}
					}
					info.Workflow = append(info.Workflow, wf)
				}
			}
		}
	}
	info.Info = fmt.Sprintf("RPP Status: %s", info.Status)
	return info, nil
}
