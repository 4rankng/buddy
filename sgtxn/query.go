package sgtxn

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"buddy/clients"
)

// QuerySGTransaction queries Singapore databases for transaction status
func QuerySGTransaction(transactionID string) *SGTransactionResult {
	result := &SGTransactionResult{
		TransactionID:        transactionID,
		PaymentCoreWorkflows: []SGWorkflowInfo{},
	}

	// Initialize Singapore doorman client factory
	factory := clients.NewDoormanClientFactory("sg")
	_, err := factory.CreateClient(30 * time.Second)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create doorman client: %v", err)
		return result
	}

	// Query payment-engine transfer table
	transferQuery := fmt.Sprintf(
		"SELECT transaction_id, reference_id, status, external_id, created_at FROM transfer WHERE transaction_id='%s'",
		transactionID,
	)
	transferResults, err := factory.QueryDatabase("payment-engine", "prod_payment_engine_db01", transferQuery)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to query payment-engine: %v", err)
		return result
	}

	if len(transferResults) == 0 {
		result.Error = "Transaction not found in payment-engine"
		return result
	}

	// Extract transfer information
	transfer := transferResults[0]
	result.TransactionID = getString(transfer, "transaction_id")
	result.TransferStatus = getString(transfer, "status")
	result.CreatedAt = getString(transfer, "created_at")
	result.ExternalID = getString(transfer, "external_id")
	result.ReferenceID = getString(transfer, "reference_id")

	// Query payment-engine workflow if reference_id exists
	if result.ReferenceID != "" {
		workflowQuery := fmt.Sprintf(
			"SELECT workflow_id, attempt, state FROM workflow_execution WHERE run_id='%s'",
			result.ReferenceID,
		)
		workflowResults, err := factory.QueryDatabase("payment-engine", "prod_payment_engine_db01", workflowQuery)
		if err == nil && len(workflowResults) > 0 {
			workflow := workflowResults[0]
			result.PaymentEngineWorkflow = SGWorkflowInfo{
				Type:    getString(workflow, "workflow_id"),
				RunID:   result.ReferenceID,
				State:   formatWorkflowState(getString(workflow, "state"), "workflow_transfer_payment"),
				Attempt: getInt(workflow, "attempt"),
			}
		}
	}

	// Parse created_at for time-based queries
	if result.CreatedAt != "" {
		createdAt, err := time.Parse(time.RFC3339Nano, result.CreatedAt)
		if err == nil {
			endTime := createdAt.Add(time.Hour)
			timeStart := createdAt.Format(time.RFC3339Nano)
			timeEnd := endTime.Format(time.RFC3339Nano)

			// Query payment-core internal_transaction
			internalQuery := fmt.Sprintf(
				"SELECT tx_id, tx_type, status FROM internal_transaction WHERE created_at >= '%s' AND created_at <= '%s' AND group_id = '%s'",
				timeStart, timeEnd, transactionID,
			)
			internalResults, err := factory.QueryDatabase("payment-core", "", internalQuery)
			if err == nil && len(internalResults) > 0 {
				// Collect all internal transaction statuses
				for _, internal := range internalResults {
					result.InternalTxStatuses = append(result.InternalTxStatuses, getString(internal, "status"))
				}
			}

			// Query payment-core external_transaction
			externalQuery := fmt.Sprintf(
				"SELECT ref_id, tx_type, status FROM external_transaction WHERE created_at >= '%s' AND created_at <= '%s' AND group_id = '%s'",
				timeStart, timeEnd, transactionID,
			)
			externalResults, err := factory.QueryDatabase("payment-core", "", externalQuery)
			if err == nil && len(externalResults) > 0 {
				// Collect all external transaction statuses
				for _, external := range externalResults {
					result.ExternalTxStatuses = append(result.ExternalTxStatuses, getString(external, "status"))
				}
			}

			// Query payment-core workflows
			var runIDs []string
			if len(internalResults) > 0 {
				for _, internal := range internalResults {
					runIDs = append(runIDs, "'"+getString(internal, "tx_id")+"'")
				}
			}
			if len(externalResults) > 0 {
				for _, external := range externalResults {
					runIDs = append(runIDs, "'"+getString(external, "ref_id")+"'")
				}
			}

			if len(runIDs) > 0 {
				workflowQuery := fmt.Sprintf(
					"SELECT workflow_id, run_id, state, attempt FROM workflow_execution WHERE run_id IN (%s)",
					strings.Join(runIDs, ","),
				)
				workflowResults, err := factory.QueryDatabase("payment-core", "payment_core", workflowQuery)
				if err == nil {
					for _, w := range workflowResults {
						workflowID := getString(w, "workflow_id")
						mappedWorkflowID := mapWorkflowName(workflowID)
						result.PaymentCoreWorkflows = append(result.PaymentCoreWorkflows, SGWorkflowInfo{
							Type:    mappedWorkflowID,
							RunID:   getString(w, "run_id"),
							State:   formatWorkflowState(getString(w, "state"), workflowID),
							Attempt: getInt(w, "attempt"),
						})
					}
				}
			}

			// Query fast-adapter if external_id exists
			if result.ExternalID != "" {
				fastQuery := fmt.Sprintf(
					"SELECT type, instruction_id, status, cancel_reason_code, reject_reason_code FROM transactions WHERE instruction_id='%s' AND created_at >= '%s' AND created_at <= '%s'",
					result.ExternalID, timeStart, timeEnd,
				)
				fastResults, err := factory.QueryDatabase("fast-adapter", "", fastQuery)
				if err == nil && len(fastResults) > 0 {
					fast := fastResults[0]
					result.FastAdapterType = getString(fast, "type")
					result.FastAdapterStatus = formatFastAdapterStatus(
						getString(fast, "type"),
						getString(fast, "status"),
					)
					result.FastAdapterCancelCode = getString(fast, "cancel_reason_code")
					result.FastAdapterRejectCode = getString(fast, "reject_reason_code")
				}
			}
		}
	}

	return result
}

// getString safely extracts string value from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// getInt safely extracts int value from map
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// formatWorkflowState formats state with name and number
func formatWorkflowState(stateStr, workflowType string) string {
	if stateNum, err := strconv.Atoi(stateStr); err == nil {
		var stateName string
		if workflowType == "workflow_transfer_payment" {
			stateName = GetPaymentEngineWorkflowStateName(stateNum)
		} else {
			stateName = GetPaymentCoreWorkflowStateName(stateNum)
		}
		return fmt.Sprintf("%s(%d)", stateName, stateNum)
	}
	return stateStr
}

// formatFastAdapterStatus formats fast-adapter status based on type
func formatFastAdapterStatus(txType, statusStr string) string {
	if statusNum, err := strconv.Atoi(statusStr); err == nil {
		var stateName string
		switch txType {
		case "cashin":
			stateName = GetFastAdapterCashinStateName(statusNum)
		case "cashout":
			stateName = GetFastAdapterCashoutStateName(statusNum)
		default:
			stateName = "unknown"
		}
		return stateName
	}
	return statusStr
}

// mapWorkflowName maps workflow IDs to their display names
func mapWorkflowName(workflowID string) string {
	switch workflowID {
	case "payment_core_workflow_internal_payment_flow":
		return "workflow_internal_payment_flow"
	case "payment_core_workflow_external_payment_flow":
		return "workflow_external_payment_flow"
	default:
		return workflowID
	}
}
