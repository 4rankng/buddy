package adapters

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"buddy/internal/txn/utils"
	"fmt"
	"strings"
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

func (r *RPPAdapter) QueryByExternalID(externalID string) (*domain.RPPAdapterInfo, error) {
	query := fmt.Sprintf("SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status FROM credit_transfer WHERE end_to_end_id = '%s'", externalID)
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

func (r *RPPAdapter) QueryByE2EID(e2eID string) (*domain.TransactionResult, error) {
	// Validate E2E ID format
	if !r.IsRppE2EID(e2eID) {
		return &domain.TransactionResult{
			TransactionID: e2eID,
			CaseType:      domain.CaseNone,
			RPPAdapter: domain.RPPAdapterInfo{
				EndToEndID: e2eID,
				Status:     domain.NotFoundStatus,
				Info:       "Invalid E2E ID format",
			},
		}, nil
	}

	// Query credit_transfer table
	query := fmt.Sprintf("SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status, created_at FROM credit_transfer WHERE end_to_end_id = '%s'", e2eID)
	rppResults, err := r.client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil || len(rppResults) == 0 {
		return &domain.TransactionResult{
			TransactionID: e2eID,
			CaseType:      domain.CaseNone,
			RPPAdapter: domain.RPPAdapterInfo{
				EndToEndID: e2eID,
				Status:     domain.NotFoundStatus,
				Info:       "No record found in RPP adapter",
			},
		}, nil
	}

	// Extract RPP data
	row := rppResults[0]
	partnerTxID := utils.GetStringValue(row, "partner_tx_id")
	reqBizMsgID := utils.GetStringValue(row, "req_biz_msg_id")
	rppStatus := utils.GetStringValue(row, "status")
	createdAt := utils.GetStringValue(row, "created_at")

	// Initialize result
	result := &domain.TransactionResult{
		TransactionID: e2eID,
		CaseType:      domain.CaseNone,
		RPPAdapter: domain.RPPAdapterInfo{
			ReqBizMsgID: reqBizMsgID,
			PartnerTxID: partnerTxID,
			EndToEndID:  e2eID,
			Status:      rppStatus,
		},
	}

	// Query workflow_execution if partner_tx_id exists
	if partnerTxID != "" {
		workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id = '%s'", partnerTxID)
		if workflows, err := r.client.QueryRppAdapter(workflowQuery); err == nil && len(workflows) > 0 {
			workflow := workflows[0]
			result.RPPAdapter.Workflow.RunID = partnerTxID
			result.RPPAdapter.Workflow.WorkflowID = utils.GetStringValue(workflow, "workflow_id")
			result.RPPAdapter.Workflow.State = utils.GetStringValue(workflow, "state")
			if attemptVal, ok := workflow["attempt"]; ok {
				if attemptFloat, ok := attemptVal.(float64); ok {
					result.RPPAdapter.Workflow.Attempt = int(attemptFloat)
				}
			}
		}
	}

	result.RPPAdapter.Info = fmt.Sprintf("RPP Status: %s", rppStatus)

	// Query Payment Engine using external_id and time window
	if createdAt != "" {
		// Parse created_at and calculate time window
		var startTime time.Time
		var err error

		// Try RFC3339 format first
		if startTime, err = time.Parse(time.RFC3339, createdAt); err != nil {
			// Try alternative date formats
			if startTime, err = time.Parse("2006-01-02 15:04:05", createdAt); err != nil {
				// Try one more format
				if startTime, err = time.Parse("2006-01-02T15:04:05Z", createdAt); err != nil {
					// Continue with empty created_at if parsing fails
					return result, nil
				}
			}
		}

		timeWindow := 30 * time.Minute
		windowStart := startTime.Add(-timeWindow)
		windowEnd := startTime.Add(timeWindow)

		peQuery := fmt.Sprintf(
			"SELECT transaction_id, status, reference_id, external_id, type, txn_subtype, txn_domain, created_at FROM transfer WHERE external_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
			e2eID,
			windowStart.Format(time.RFC3339),
			windowEnd.Format(time.RFC3339),
		)

		if transfers, err := r.client.QueryPaymentEngine(peQuery); err == nil && len(transfers) > 0 {
			transfer := transfers[0]

			// Populate PaymentEngine info
			result.PaymentEngine.Transfers.TransactionID = utils.GetStringValue(transfer, "transaction_id")
			result.PaymentEngine.Transfers.Status = utils.GetStringValue(transfer, "status")
			result.PaymentEngine.Transfers.ReferenceID = utils.GetStringValue(transfer, "reference_id")
			result.PaymentEngine.Transfers.ExternalID = utils.GetStringValue(transfer, "external_id")
			result.PaymentEngine.Transfers.Type = utils.GetStringValue(transfer, "type")
			result.PaymentEngine.Transfers.TxnSubtype = utils.GetStringValue(transfer, "txn_subtype")
			result.PaymentEngine.Transfers.TxnDomain = utils.GetStringValue(transfer, "txn_domain")
			result.PaymentEngine.Transfers.CreatedAt = utils.GetStringValue(transfer, "created_at")

			// Update the main TransactionID with the one from Payment Engine
			if result.PaymentEngine.Transfers.TransactionID != "" {
				result.TransactionID = result.PaymentEngine.Transfers.TransactionID
			}

			// Query Payment Engine workflow if we have reference_id
			if result.PaymentEngine.Transfers.ReferenceID != "" {
				peWorkflowQuery := fmt.Sprintf(
					"SELECT run_id, workflow_id, state, attempt, created_at FROM workflow_execution WHERE run_id = '%s'",
					result.PaymentEngine.Transfers.ReferenceID,
				)
				if peWorkflows, err := r.client.QueryPaymentEngine(peWorkflowQuery); err == nil && len(peWorkflows) > 0 {
					peWorkflow := peWorkflows[0]
					result.PaymentEngine.Workflow.RunID = result.PaymentEngine.Transfers.ReferenceID
					result.PaymentEngine.Workflow.WorkflowID = utils.GetStringValue(peWorkflow, "workflow_id")
					if state, ok := peWorkflow["state"]; ok {
						if stateInt, ok := state.(float64); ok {
							result.PaymentEngine.Workflow.State = fmt.Sprintf("%d", int(stateInt))
						} else {
							result.PaymentEngine.Workflow.State = fmt.Sprintf("%v", state)
						}
					}
					if attemptVal, ok := peWorkflow["attempt"]; ok {
						if attemptFloat, ok := attemptVal.(float64); ok {
							result.PaymentEngine.Workflow.Attempt = int(attemptFloat)
						}
					}
				}
			}

			// Query Payment Core if we have transaction_id and created_at
			if result.PaymentEngine.Transfers.TransactionID != "" && result.PaymentEngine.Transfers.CreatedAt != "" {
				// Query internal transactions
				if internalTxs, err := r.client.QueryPaymentCore(fmt.Sprintf(
					"SELECT tx_id, tx_type, status FROM internal_transaction WHERE tx_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
					result.PaymentEngine.Transfers.TransactionID,
					windowStart.Format(time.RFC3339),
					windowEnd.Format(time.RFC3339),
				)); err == nil {
					for _, internalTx := range internalTxs {
						txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "tx_type")))
						status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "status")))

						result.PaymentCore.InternalTxns = append(result.PaymentCore.InternalTxns, domain.PCInternalTxnInfo{
							TxID:     utils.GetStringValue(internalTx, "tx_id"),
							GroupID:  result.TransactionID,
							TxType:   txType,
							TxStatus: status,
						})
					}
				}

				// Query external transactions
				if externalTxs, err := r.client.QueryPaymentCore(fmt.Sprintf(
					"SELECT ref_id, tx_type, status FROM external_transaction WHERE ref_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
					result.PaymentEngine.Transfers.TransactionID,
					windowStart.Format(time.RFC3339),
					windowEnd.Format(time.RFC3339),
				)); err == nil {
					for _, externalTx := range externalTxs {
						txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "tx_type")))
						status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "status")))

						result.PaymentCore.ExternalTxns = append(result.PaymentCore.ExternalTxns, domain.PCExternalTxnInfo{
							RefID:    utils.GetStringValue(externalTx, "ref_id"),
							GroupID:  result.TransactionID,
							TxType:   txType,
							TxStatus: status,
						})
					}
				}

				// Query PaymentCore workflows for all transactions
				var pcRunIDs []string
				for _, internalTx := range result.PaymentCore.InternalTxns {
					if internalTx.TxID != "" {
						pcRunIDs = append(pcRunIDs, internalTx.TxID)
					}
				}
				for _, externalTx := range result.PaymentCore.ExternalTxns {
					if externalTx.RefID != "" {
						pcRunIDs = append(pcRunIDs, externalTx.RefID)
					}
				}

				if len(pcRunIDs) > 0 {
					// Build WHERE clause for multiple run IDs
					var pcRunIDConditions []string
					for _, runID := range pcRunIDs {
						pcRunIDConditions = append(pcRunIDConditions, fmt.Sprintf("'%s'", runID))
					}
					pcWhereClause := strings.Join(pcRunIDConditions, ",")

					pcWorkflowQuery := fmt.Sprintf(
						"SELECT run_id, workflow_id, state, attempt, created_at FROM workflow_execution WHERE run_id IN (%s)",
						pcWhereClause,
					)
					if pcWorkflows, err := r.client.QueryPaymentCore(pcWorkflowQuery); err == nil {
						for _, pcWorkflow := range pcWorkflows {
							workflowID := utils.GetStringValue(pcWorkflow, "workflow_id")
							runID := utils.GetStringValue(pcWorkflow, "run_id")
							state := pcWorkflow["state"]

							var stateNum int
							if stateInt, ok := state.(float64); ok {
								stateNum = int(stateInt)
							}

							var attempt int
							if attemptFloat, ok := pcWorkflow["attempt"].(float64); ok {
								attempt = int(attemptFloat)
							}

							result.PaymentCore.Workflow = append(result.PaymentCore.Workflow, domain.WorkflowInfo{
								WorkflowID: workflowID,
								RunID:      runID,
								State:      fmt.Sprintf("%d", stateNum),
								Attempt:    attempt,
							})
						}
					}
				}
			}
		}
	}

	return result, nil
}

func (r *RPPAdapter) IsRppE2EID(id string) bool {
	// RppE2EIDPattern from original code
	return len(id) == 30 && id[0:8] >= "20000101" && id[0:8] <= "20991231" && id[22:30] >= "00000000" && id[22:30] <= "99999999"
}
