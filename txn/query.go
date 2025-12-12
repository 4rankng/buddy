package txn

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"buddy/clients"
)

// RppE2EIDPattern matches RPP E2E ID format: YYYYMMDDGXSPMY + 16 characters (total 30)
var RppE2EIDPattern = regexp.MustCompile(`^\d{8}GXSPMY[A-Za-z0-9]{16}$`)

// QueryTransactionStatus returns structured data about a transaction
func QueryTransactionStatus(transactionID string) *TransactionResult {
	// Initialize doorman client
	client, err := clients.NewDoormanClient(30 * time.Second)
	if err != nil {
		return &TransactionResult{
			TransactionID: transactionID,
			Error:         fmt.Sprintf("failed to create doorman client: %v", err),
		}
	}

	// Check if input is an RPP E2E ID
	if RppE2EIDPattern.MatchString(transactionID) {
		return queryRPPE2EID(client, transactionID)
	}

	// Query transfer table with specific fields including timestamps
	transferQuery := fmt.Sprintf("SELECT transaction_id, status, reference_id, created_at, updated_at FROM transfer WHERE transaction_id='%s'", transactionID)
	transfers, err := client.QueryPrdPaymentsPaymentEngine(transferQuery)
	if err != nil {
		return &TransactionResult{
			TransactionID: transactionID,
			Error:         fmt.Sprintf("failed to query transfer table: %v", err),
		}
	}

	if len(transfers) == 0 {
		return &TransactionResult{
			TransactionID:  transactionID,
			TransferStatus: "NOT_FOUND",
			Error:          "No transaction found with the given ID",
		}
	}

	// Get transfer information
	transfer := transfers[0]

	// Check status
	status, ok := transfer["status"].(string)
	if !ok {
		return &TransactionResult{
			TransactionID: transactionID,
			Error:         "could not determine transaction status",
		}
	}

	result := &TransactionResult{
		TransactionID:  transactionID,
		TransferStatus: status,
	}

	// Extract created_at and updated_at from transfer
	var createdAtStr string
	if createdAt, createdAtOk := transfer["created_at"]; createdAtOk {
		createdAtStr = fmt.Sprintf("%v", createdAt)
		result.CreatedAt = createdAtStr
	}
	// UpdatedAt field removed - no longer needed

	// Get reference_id if available
	referenceID, ok := transfer["reference_id"].(string)
	if ok && referenceID != "" {
		// Query workflow_execution table with specific fields including timestamps
		workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt, created_at, updated_at FROM workflow_execution WHERE run_id='%s'", referenceID)
		workflows, err := client.QueryPrdPaymentsPaymentEngine(workflowQuery)
		if err != nil {
			result.Error = fmt.Sprintf("failed to query workflow_execution table: %v", err)
			return result
		}

		if len(workflows) == 0 {
			result.Error = "No workflow execution found for this reference ID"
			return result
		}

		// Get workflow information
		workflow := workflows[0]

		// Extract workflow_id
		if workflowID, workflowIDOk := workflow["workflow_id"]; workflowIDOk {
			result.PaymentEngineWorkflow.Type = fmt.Sprintf("%v", workflowID)
		}
		result.PaymentEngineWorkflow.RunID = ""
		if runID, runIDOk := workflow["run_id"]; runIDOk {
			result.PaymentEngineWorkflow.RunID = fmt.Sprintf("%v", runID)
		}

		// Extract attempt if available
		var workflowAttempt int
		if attemptVal, attemptOk := workflow["attempt"]; attemptOk {
			if attemptFloat, ok := attemptVal.(float64); ok {
				workflowAttempt = int(attemptFloat)
			}
		}
		result.PaymentEngineWorkflow.Attempt = workflowAttempt

		// Extract workflow_id
		if workflowID, workflowIDOk := workflow["workflow_id"]; workflowIDOk {
			result.PaymentEngineWorkflow.Type = fmt.Sprintf("%v", workflowID)
		}

		// Extract state
		if state, stateOk := workflow["state"]; stateOk {
			// Try to convert state to int for mapping
			if stateInt, ok := state.(float64); ok {
				stateNum := int(stateInt)
				result.PaymentEngineWorkflow.State = fmt.Sprintf("%d", stateNum)
			} else {
				result.PaymentEngineWorkflow.State = fmt.Sprintf("%v", state)
			}
		}

		// If we don't have created_at from transfer, use workflow timestamps
		if result.CreatedAt == "" {
			if createdAt, createdAtOk := workflow["created_at"]; createdAtOk {
				result.CreatedAt = fmt.Sprintf("%v", createdAt)
			}
		}
	}

	// Query payment-core database if we have created_at timestamp
	if result.CreatedAt != "" {
		// Parse created_at timestamp to add 1 hour
		createdAt, err := time.Parse(time.RFC3339, result.CreatedAt)
		if err == nil {
			endTime := createdAt.Add(1 * time.Hour)
			endTimeStr := endTime.Format(time.RFC3339)
			internalQuery := fmt.Sprintf("SELECT tx_id, tx_type, status FROM internal_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'",
				transactionID, result.CreatedAt, endTimeStr)
			externalQuery := fmt.Sprintf("SELECT ref_id, tx_type, status FROM external_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'",
				transactionID, result.CreatedAt, endTimeStr)
			// Query internal_transaction table
			internalTxs, err := client.QueryPrdPaymentsPaymentCore(internalQuery)
			if err == nil && len(internalTxs) > 0 {
				var internalStatuses []string
				for _, internalTx := range internalTxs {
					txType, _ := internalTx["tx_type"].(string)
					status, _ := internalTx["status"].(string)
					txType = strings.TrimSpace(strings.ToUpper(txType))
					status = strings.TrimSpace(strings.ToUpper(status))
					var parts []string
					if txType != "" {
						parts = append(parts, txType)
					}
					if status != "" {
						parts = append(parts, status)
					}
					if len(parts) > 0 {
						internalStatuses = append(internalStatuses, strings.Join(parts, " "))
					}
				}
				if len(internalStatuses) > 0 {
					result.InternalTxStatus = strings.Join(internalStatuses, " , ")
				}
			}

			// Query external_transaction table
			externalTxs, err := client.QueryPrdPaymentsPaymentCore(externalQuery)
			if err == nil && len(externalTxs) > 0 {
				var externalStatuses []string
				for _, externalTx := range externalTxs {
					txType, _ := externalTx["tx_type"].(string)
					status, _ := externalTx["status"].(string)
					txType = strings.TrimSpace(strings.ToUpper(txType))
					status = strings.TrimSpace(strings.ToUpper(status))
					var parts []string
					if txType != "" {
						parts = append(parts, txType)
					}
					if status != "" {
						parts = append(parts, status)
					}
					if len(parts) > 0 {
						externalStatuses = append(externalStatuses, strings.Join(parts, " "))
					}
				}
				if len(externalStatuses) > 0 {
					result.ExternalTxStatus = strings.Join(externalStatuses, " , ")
				}
			}

			// Query workflow_execution table for payment-core using tx_id and ref_id
			var runIDs []string
			if len(internalTxs) > 0 {
				for _, internalTx := range internalTxs {
					if txID, ok := internalTx["tx_id"].(string); ok && txID != "" {
						runIDs = append(runIDs, txID)
					}
				}
			}
			if len(externalTxs) > 0 {
				for _, externalTx := range externalTxs {
					if refID, ok := externalTx["ref_id"].(string); ok && refID != "" {
						runIDs = append(runIDs, refID)
					}
				}
			}

			if len(runIDs) > 0 {
				// Create IN clause for multiple run_ids
				runIDsStr := "'" + strings.Join(runIDs, "', '") + "'"
				workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id IN (%s)", runIDsStr)
				paymentCoreWorkflows, err := client.QueryPrdPaymentsPaymentCore(workflowQuery)

				if err == nil {
					for _, workflow := range paymentCoreWorkflows {
						workflowID, _ := workflow["workflow_id"].(string)
						runID, _ := workflow["run_id"].(string)
						state := workflow["state"]

						// Convert state to int for mapping
						var stateNum int
						if stateInt, ok := state.(float64); ok {
							stateNum = int(stateInt)
						}

						// Get attempt count
						var attempt int
						if attemptFloat, ok := workflow["attempt"].(float64); ok {
							attempt = int(attemptFloat)
						}

						// Create WorkflowInfo for this payment core workflow
						workflowInfo := WorkflowInfo{
							Type:    workflowID,
							RunID:   runID,
							State:   fmt.Sprintf("%d", stateNum),
							Attempt: attempt,
						}

						// Add to PaymentCoreWorkflows slice
						result.PaymentCoreWorkflows = append(result.PaymentCoreWorkflows, workflowInfo)
					}
				}
			}
		}
	}

	return result
}

// queryRPPE2EID handles RPP E2E ID lookups
func queryRPPE2EID(client *clients.DoormanClient, e2eID string) *TransactionResult {
	// First query RPP adapter database to get partner_tx_id and credit_transfer status
	rppQuery := fmt.Sprintf("SELECT partner_tx_id, partner_tx_sts AS status FROM credit_transfer WHERE end_to_end_id = '%s'", e2eID)
	rppResults, err := client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", rppQuery)
	if err != nil {
		return &TransactionResult{
			TransactionID: e2eID,
			Error:         fmt.Sprintf("failed to query RPP adapter credit_transfer table: %v", err),
		}
	}

	if len(rppResults) == 0 {
		return &TransactionResult{
			TransactionID: e2eID,
			// TransferStatus remains empty since we're not querying payment-engine
			Error: "No transaction found with the given E2E ID",
		}
	}

	// Extract partner_tx_id which is the workflow run_id
	partnerTxID, ok := rppResults[0]["partner_tx_id"].(string)
	if !ok || partnerTxID == "" {
		return &TransactionResult{
			TransactionID: e2eID,
			Error:         "could not extract partner_tx_id from RPP adapter",
		}
	}

	// Extract credit_transfer status if available
	var creditTransferStatus string
	if status, statusOk := rppResults[0]["status"]; statusOk {
		if statusStr, ok := status.(string); ok {
			creditTransferStatus = statusStr
		}
	}

	// Now query the workflow_execution table using partner_tx_id as run_id
	workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt, created_at, updated_at FROM workflow_execution WHERE run_id='%s'", partnerTxID)
	workflows, err := client.QueryPrdPaymentsRppAdapter(workflowQuery)
	if err != nil {
		return &TransactionResult{
			TransactionID: e2eID,
			Error:         fmt.Sprintf("failed to query workflow_execution table: %v", err),
		}
	}

	if len(workflows) == 0 {
		return &TransactionResult{
			TransactionID: e2eID,
			Error:         "No workflow execution found for this E2E ID",
		}
	}

	// Build result
	workflow := workflows[0]
	result := &TransactionResult{
		TransactionID: e2eID,
		// Note: TransferStatus is not set here because we're not querying payment-engine
		// It should remain empty unless we query payment-engine
		PartnerTxID: partnerTxID,
		RPPStatus:   creditTransferStatus, // Set credit_transfer status
	}

	// Set RPP workflow info
	if workflowID, workflowIDOk := workflow["workflow_id"]; workflowIDOk {
		result.RPPWorkflow.Type = fmt.Sprintf("%v", workflowID)
	}
	result.RPPWorkflow.RunID = partnerTxID

	// Extract attempt if available
	if attemptVal, attemptOk := workflow["attempt"]; attemptOk {
		if attemptFloat, ok := attemptVal.(float64); ok {
			result.RPPWorkflow.Attempt = int(attemptFloat)
		}
	}

	// Extract state
	if state, stateOk := workflow["state"]; stateOk {
		if stateInt, ok := state.(float64); ok {
			stateNum := int(stateInt)
			result.RPPWorkflow.State = fmt.Sprintf("%d", stateNum)
		} else {
			result.RPPWorkflow.State = fmt.Sprintf("%v", state)
		}
	}
	// Keep the original creditTransferStatus from the database query
	// Don't overwrite it with the workflow state

	// Set created_at timestamp
	if createdAt, createdAtOk := workflow["created_at"]; createdAtOk {
		result.CreatedAt = fmt.Sprintf("%v", createdAt)
	}

	// Also set RPP info for display
	result.RPPInfo = fmt.Sprintf("RPP Workflow: %s, State: %s, Attempt: %d",
		result.RPPWorkflow.Type, result.RPPWorkflow.State, result.RPPWorkflow.Attempt)

	return result
}
