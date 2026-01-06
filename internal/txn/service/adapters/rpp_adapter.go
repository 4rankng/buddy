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

// Query uses fallback logic to find credit_transfer record
// Priority 1: If EndToEndID is provided, query by end_to_end_id
// Priority 2: If PartnerTxID is provided, query by partner_tx_id
// Priority 3: If SourceAccountID, DestinationAccountID, Amount and Timestamp are provided, query by these fields
func (r *RPPAdapter) Query(params domain.RPPQueryParams) (*domain.RPPAdapterInfo, error) {
	if r.client == nil {
		return nil, fmt.Errorf("Query: database client is not initialized")
	}

	// Priority 1: Query by end_to_end_id if provided
	if params.EndToEndID != "" {
		return r.queryByE2EID(params.EndToEndID)
	}

	// Priority 2: Query by partner_tx_id if provided
	if params.PartnerTxID != "" {
		info, err := r.queryByPartnerTxID(params.PartnerTxID)
		if err == nil && info != nil {
			return info, nil
		}
	}

	// Priority 3: Query by account details, amount and timestamp
	if params.SourceAccountID != "" && params.DestinationAccountID != "" && params.Timestamp != "" {
		return r.queryByAccountsAmountAndTimestamp(params)
	}

	return nil, fmt.Errorf("insufficient parameters provided for RPP query")
}

func (r *RPPAdapter) queryByE2EID(externalID string) (*domain.RPPAdapterInfo, error) {
	query := fmt.Sprintf("SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status, created_at FROM credit_transfer WHERE end_to_end_id = '%s'", externalID)
	rppResults, err := r.client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil || len(rppResults) == 0 {
		// Fallback: Query wf_process_registry workflow using date extracted from EndToEndID
		return r.queryProcessRegistryByE2EID(externalID)
	}
	row := rppResults[0]
	info := &domain.RPPAdapterInfo{
		ReqBizMsgID: utils.GetStringValue(row, "req_biz_msg_id"),
		PartnerTxID: utils.GetStringValue(row, "partner_tx_id"),
		EndToEndID:  externalID,
		Status:      utils.GetStringValue(row, "status"),
		CreatedAt:   utils.GetStringValue(row, "created_at"),
	}
	r.populateWorkflowInfo(info)
	info.Info = fmt.Sprintf("RPP Status: %s", info.Status)
	return info, nil
}

func (r *RPPAdapter) queryByPartnerTxID(partnerTxID string) (*domain.RPPAdapterInfo, error) {
	query := fmt.Sprintf(
		"SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status, end_to_end_id, created_at FROM credit_transfer WHERE partner_tx_id = '%s'",
		partnerTxID,
	)

	rppResults, err := r.client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil {
		return nil, err
	}

	if len(rppResults) == 0 {
		return nil, nil
	}

	row := rppResults[0]
	info := &domain.RPPAdapterInfo{
		ReqBizMsgID: utils.GetStringValue(row, "req_biz_msg_id"),
		PartnerTxID: utils.GetStringValue(row, "partner_tx_id"),
		EndToEndID:  utils.GetStringValue(row, "end_to_end_id"),
		Status:      utils.GetStringValue(row, "status"),
		CreatedAt:   utils.GetStringValue(row, "created_at"),
	}
	r.populateWorkflowInfo(info)
	info.Info = fmt.Sprintf("RPP Status: %s", info.Status)
	return info, nil
}

func (r *RPPAdapter) queryByAccountsAmountAndTimestamp(params domain.RPPQueryParams) (*domain.RPPAdapterInfo, error) {
	createdAt, err := time.Parse(time.RFC3339, params.Timestamp)
	if err != nil {
		createdAt, err = time.Parse(time.RFC3339Nano, params.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %w", err)
		}
	}

	timeWindowStart := createdAt.Add(-2 * time.Minute)
	timeWindowEnd := createdAt.Add(2 * time.Minute)

	var query string
	if params.Amount > 0 {
		amountDollars := params.Amount / 100.0
		query = fmt.Sprintf(
			"SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status, end_to_end_id, created_at FROM credit_transfer WHERE "+
				"dbtr_acct_id='%s' AND cdtr_acct_id='%s' AND amount=%.2f AND created_at >= '%s' AND created_at <= '%s'",
			params.SourceAccountID,
			params.DestinationAccountID,
			amountDollars,
			timeWindowStart.Format(time.RFC3339Nano),
			timeWindowEnd.Format(time.RFC3339Nano),
		)
	} else {
		query = fmt.Sprintf(
			"SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status, end_to_end_id, created_at FROM credit_transfer WHERE "+
				"dbtr_acct_id='%s' AND cdtr_acct_id='%s' AND created_at >= '%s' AND created_at <= '%s'",
			params.SourceAccountID,
			params.DestinationAccountID,
			timeWindowStart.Format(time.RFC3339Nano),
			timeWindowEnd.Format(time.RFC3339Nano),
		)
	}

	rppResults, err := r.client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil {
		return nil, err
	}

	if len(rppResults) == 0 {
		return nil, nil
	}

	row := rppResults[0]
	info := &domain.RPPAdapterInfo{
		ReqBizMsgID: utils.GetStringValue(row, "req_biz_msg_id"),
		PartnerTxID: utils.GetStringValue(row, "partner_tx_id"),
		EndToEndID:  utils.GetStringValue(row, "end_to_end_id"),
		Status:      utils.GetStringValue(row, "status"),
		CreatedAt:   utils.GetStringValue(row, "created_at"),
	}
	r.populateWorkflowInfo(info)
	info.Info = fmt.Sprintf("RPP Status: %s", info.Status)
	return info, nil
}

// populateWorkflowInfo populates workflow data for RPPAdapterInfo
func (r *RPPAdapter) populateWorkflowInfo(info *domain.RPPAdapterInfo) {
	info.Workflow = make([]domain.WorkflowInfo, 0)

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

	if info.PartnerTxID != "" {
		workflowQuery := fmt.Sprintf(
			"SELECT run_id, workflow_id, state, attempt, prev_trans_id, data FROM workflow_execution "+
				"WHERE run_id = '%s'",
			info.PartnerTxID,
		)

		if workflowRows, err := r.client.QueryRppAdapter(workflowQuery); err == nil && len(workflowRows) > 0 {
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

// queryProcessRegistryByE2EID queries wf_process_registry workflow using date extracted from EndToEndID
// This is a fallback method when the main credit_transfer query fails
func (r *RPPAdapter) queryProcessRegistryByE2EID(externalID string) (*domain.RPPAdapterInfo, error) {
	// Extract date from EndToEndID (first 8 characters should be YYYYMMDD)
	if len(externalID) < 8 {
		return nil, fmt.Errorf("EndToEndID too short to extract date: %s", externalID)
	}

	dateStr := externalID[:8]

	// Parse the date string (YYYYMMDD format)
	startDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date from EndToEndID %s: %w", externalID, err)
	}

	// Create time window: start of day to start of day + 1 hour
	timeWindowStart := startDate
	timeWindowEnd := startDate.Add(1 * time.Hour)

	// Query workflow_execution table for wf_process_registry workflows
	workflowQuery := fmt.Sprintf(
		"SELECT run_id, workflow_id, state, attempt, prev_trans_id, data FROM workflow_execution "+
			"WHERE created_at >= '%s' "+
			"AND created_at <= '%s' "+
			"AND workflow_id = 'wf_process_registry' "+
			"AND data LIKE '%%%s%%'",
		timeWindowStart.Format(time.RFC3339),
		timeWindowEnd.Format(time.RFC3339),
		externalID,
	)

	workflowRows, err := r.client.QueryRppAdapter(workflowQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query wf_process_registry workflow: %w", err)
	}

	if len(workflowRows) == 0 {
		return nil, fmt.Errorf("no wf_process_registry workflow found for EndToEndID: %s", externalID)
	}

	// Create RPPAdapterInfo with workflow data
	info := &domain.RPPAdapterInfo{
		EndToEndID: externalID,
		Status:     "STUCK_IN_PROCESS_REGISTRY",
		Info:       fmt.Sprintf("Found in wf_process_registry workflow (stuck transaction)"),
		Workflow:   make([]domain.WorkflowInfo, 0),
	}

	// Populate workflow information
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

	return info, nil
}
