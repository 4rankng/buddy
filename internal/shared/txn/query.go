package txn

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"buddy/clients"
)

// RppE2EIDPattern matches RPP E2E ID format: first 8 chars are digits (YYYYMMDD) and last 8 chars are digits, total 30 chars
var RppE2EIDPattern = regexp.MustCompile(`^\d{8}.{14}\d{8}$`)

// QueryTransactionStatus returns structured data about a transaction
func QueryTransactionStatus(transactionID string) *TransactionResult {
	return QueryTransactionStatusWithEnv(transactionID, "my")
}

// QueryTransactionStatusWithEnv returns structured data about a transaction with specified environment
func QueryTransactionStatusWithEnv(transactionID string, env string) *TransactionResult {
	// Get singleton doorman client
	client, err := clients.GetDoormanClient(env)
	if err != nil {
		return &TransactionResult{
			TransactionID: transactionID,
			Error:         fmt.Sprintf("failed to get doorman client: %v", err),
		}
	}

	// Check if input is an RPP E2E ID
	if RppE2EIDPattern.MatchString(transactionID) {
		return queryRPPE2EID(client, transactionID)
	}

	// Query transfer table with specific fields including timestamps and type details
	transferQuery := fmt.Sprintf("SELECT transaction_id, status, reference_id, created_at, updated_at, type, txn_subtype, txn_domain, external_id FROM transfer WHERE transaction_id='%s'", transactionID)
	transfers, err := client.QueryPaymentEngine(transferQuery)
	if err != nil {
		return &TransactionResult{
			TransactionID: transactionID,
			Error:         fmt.Sprintf("failed to query payment engine: %v", err),
		}
	}

	if len(transfers) == 0 {
		return &TransactionResult{
			TransactionID: transactionID,
			PaymentEngine: PaymentEngineInfo{
				Transfers: PETransfersInfo{
					TransactionID: transactionID,
					Status:        NotFoundStatus,
				},
			},
			Error: "No transaction found with the given ID",
		}
	}

	// Get transfer information
	transfer := transfers[0]

	result := &TransactionResult{
		TransactionID: transactionID,
	}

	// Check status
	status, ok := transfer["status"].(string)
	if !ok {
		return &TransactionResult{
			TransactionID: transactionID,
			Error:         "could not determine transaction status",
		}
	}
	result.PaymentEngine.Transfers.Status = status

	// Extract transaction_id if available
	if txID, ok := transfer["transaction_id"].(string); ok && txID != "" {
		result.TransactionID = txID
		result.PaymentEngine.Transfers.TransactionID = txID
	} else {
		result.PaymentEngine.Transfers.TransactionID = transactionID
	}

	// Extract created_at and updated_at from transfer
	var createdAtStr string
	if createdAt, createdAtOk := transfer["created_at"]; createdAtOk {
		createdAtStr = fmt.Sprintf("%v", createdAt)
		result.PaymentEngine.Transfers.CreatedAt = createdAtStr
	}
	// UpdatedAt field removed - no longer needed

	// Extract type details for new format
	if txType, ok := transfer["type"].(string); ok {
		result.PaymentEngine.Transfers.Type = txType
	}
	if txSubtype, ok := transfer["txn_subtype"].(string); ok {
		result.PaymentEngine.Transfers.TxnSubtype = txSubtype
	}
	if txDomain, ok := transfer["txn_domain"].(string); ok {
		result.PaymentEngine.Transfers.TxnDomain = txDomain
	}

	// Extract external_id for fast adapter lookup
	if externalID, ok := transfer["external_id"].(string); ok {
		result.PaymentEngine.Transfers.ExternalID = externalID
	}

	// Get reference_id if available
	referenceID, ok := transfer["reference_id"].(string)
	if ok && referenceID != "" {
		result.PaymentEngine.Transfers.ReferenceID = referenceID
		// Query workflow_execution table with specific fields including timestamps
		workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt, created_at, updated_at FROM workflow_execution WHERE run_id='%s'", referenceID)
		workflows, err := client.QueryPaymentEngine(workflowQuery)
		if err != nil {
			result.Error = fmt.Sprintf("failed to query workflow_execution table: %v", err)
			return result
		}

		if len(workflows) == 0 {
			result.Error = "No workflow execution found for this reference ID"
			return result
		}

		workflow := workflows[0]

		if workflowID, workflowIDOk := workflow["workflow_id"]; workflowIDOk {
			result.PaymentEngine.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
		}
		if runID, runIDOk := workflow["run_id"]; runIDOk {
			result.PaymentEngine.Workflow.RunID = fmt.Sprintf("%v", runID)
		}

		if attemptVal, attemptOk := workflow["attempt"]; attemptOk {
			if attemptFloat, ok := attemptVal.(float64); ok {
				result.PaymentEngine.Workflow.Attempt = int(attemptFloat)
			}
		}

		if state, stateOk := workflow["state"]; stateOk {
			if stateInt, ok := state.(float64); ok {
				stateNum := int(stateInt)
				result.PaymentEngine.Workflow.State = fmt.Sprintf("%d", stateNum)
			} else {
				result.PaymentEngine.Workflow.State = fmt.Sprintf("%v", state)
			}
		}

		// If we don't have created_at from transfer, use workflow timestamps
		if result.PaymentEngine.Transfers.CreatedAt == "" {
			if createdAt, createdAtOk := workflow["created_at"]; createdAtOk {
				result.PaymentEngine.Transfers.CreatedAt = fmt.Sprintf("%v", createdAt)
			}
		}
	}

	// Query payment-core database if we have created_at timestamp
	if result.PaymentEngine.Transfers.CreatedAt != "" {
		// Parse created_at timestamp to add 1 hour
		createdAt, err := time.Parse(time.RFC3339, result.PaymentEngine.Transfers.CreatedAt)
		if err == nil {
			endTime := createdAt.Add(1 * time.Hour)
			endTimeStr := endTime.Format(time.RFC3339)
			internalQuery := fmt.Sprintf("SELECT tx_id, tx_type, status FROM internal_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'",
				transactionID, result.PaymentEngine.Transfers.CreatedAt, endTimeStr)
			externalQuery := fmt.Sprintf("SELECT ref_id, tx_type, status FROM external_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'",
				transactionID, result.PaymentEngine.Transfers.CreatedAt, endTimeStr)
			// Query internal_transaction table
			internalTxs, err := client.QueryPaymentCore(internalQuery)
			if err == nil && len(internalTxs) > 0 {
				for _, internalTx := range internalTxs {
					txType, _ := internalTx["tx_type"].(string)
					status, _ := internalTx["status"].(string)
					txType = strings.TrimSpace(strings.ToUpper(txType))
					status = strings.TrimSpace(strings.ToUpper(status))

					info := PCInternalTxnInfo{
						TxID:     fmt.Sprintf("%v", internalTx["tx_id"]),
						GroupID:  transactionID,
						TxType:   txType,
						TxStatus: status,
					}

					result.PaymentCore.InternalTxns = append(result.PaymentCore.InternalTxns, info)
				}
			}

			// Query external_transaction table
			externalTxs, err := client.QueryPaymentCore(externalQuery)
			if err == nil && len(externalTxs) > 0 {
				for _, externalTx := range externalTxs {
					txType, _ := externalTx["tx_type"].(string)
					status, _ := externalTx["status"].(string)
					txType = strings.TrimSpace(strings.ToUpper(txType))
					status = strings.TrimSpace(strings.ToUpper(status))
					info := PCExternalTxnInfo{
						RefID:    fmt.Sprintf("%v", externalTx["ref_id"]),
						GroupID:  transactionID,
						TxType:   txType,
						TxStatus: status,
					}

					result.PaymentCore.ExternalTxns = append(result.PaymentCore.ExternalTxns, info)
				}
			}

			// Query workflow_execution table for payment-core using tx_id and ref_id
			var runIDs []string
			for _, internalTx := range result.PaymentCore.InternalTxns {
				if internalTx.TxID != "" {
					runIDs = append(runIDs, internalTx.TxID)
				}
			}
			for _, externalTx := range result.PaymentCore.ExternalTxns {
				if externalTx.RefID != "" {
					runIDs = append(runIDs, externalTx.RefID)
				}
			}

			if len(runIDs) > 0 {
				// Create IN clause for multiple run_ids
				runIDsStr := "'" + strings.Join(runIDs, "', '") + "'"
				workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id IN (%s)", runIDsStr)
				paymentCoreWorkflows, err := client.QueryPaymentCore(workflowQuery)

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
							WorkflowID: workflowID,
							RunID:      runID,
							State:      fmt.Sprintf("%d", stateNum),
							Attempt:    attempt,
						}

						// Add to payment-core workflow slice
						result.PaymentCore.Workflow = append(result.PaymentCore.Workflow, workflowInfo)
					}
				}
			}
		}
	}

	// Query fast adapter if we have external_id
	if result.PaymentEngine.Transfers.ExternalID != "" {
		if fastAdapterInfo, err := queryFastAdapter(client, result.PaymentEngine.Transfers.ExternalID, result.PaymentEngine.Transfers.CreatedAt); err == nil && fastAdapterInfo != nil {
			result.FastAdapter = *fastAdapterInfo
		}
	}

	return result
}

// queryRPPE2EID handles RPP E2E ID lookups
func queryRPPE2EID(client clients.DoormanInterface, e2eID string) *TransactionResult {
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
			RPPAdapter:    RPPAdapterInfo{Status: NotFoundStatus},
			Error:         "No transaction found with the given E2E ID",
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
	workflows, err := client.QueryRppAdapter(workflowQuery)
	if err != nil {
		return &TransactionResult{
			TransactionID: e2eID,
			Error:         fmt.Sprintf("failed to query workflow_execution table: %v", err),
		}
	}
	if len(workflows) == 0 {
		// If workflow execution is missing but credit_transfer exists,
		// still return the credit_transfer information
		result := &TransactionResult{
			TransactionID: e2eID,
			RPPAdapter: RPPAdapterInfo{
				PartnerTxID: partnerTxID,
				Status:      creditTransferStatus,
				Workflow: WorkflowInfo{
					RunID: partnerTxID,
				},
			},
		}

		// Set RPP info for display
		result.RPPAdapter.Info = fmt.Sprintf("RPP Status: %s (workflow execution not found)", creditTransferStatus)

		return result
	}

	// Build result
	workflow := workflows[0]
	result := &TransactionResult{
		TransactionID: e2eID,
		RPPAdapter: RPPAdapterInfo{
			PartnerTxID: partnerTxID,
			Status:      creditTransferStatus, // Set credit_transfer status
		},
	}

	// Set RPP workflow info
	if workflowID, workflowIDOk := workflow["workflow_id"]; workflowIDOk {
		result.RPPAdapter.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
	}
	result.RPPAdapter.Workflow.RunID = partnerTxID

	// Extract attempt if available
	if attemptVal, attemptOk := workflow["attempt"]; attemptOk {
		if attemptFloat, ok := attemptVal.(float64); ok {
			result.RPPAdapter.Workflow.Attempt = int(attemptFloat)
		}
	}

	// Extract state
	if state, stateOk := workflow["state"]; stateOk {
		if stateInt, ok := state.(float64); ok {
			stateNum := int(stateInt)
			result.RPPAdapter.Workflow.State = fmt.Sprintf("%d", stateNum)
		} else {
			result.RPPAdapter.Workflow.State = fmt.Sprintf("%v", state)
		}
	}
	// Keep the original creditTransferStatus from the database query
	// Don't overwrite it with the workflow state

	// Also set RPP info for display
	result.RPPAdapter.Info = fmt.Sprintf("RPP Workflow: %s, State: %s, Attempt: %d",
		result.RPPAdapter.Workflow.WorkflowID, result.RPPAdapter.Workflow.State, result.RPPAdapter.Workflow.Attempt)

	return result
}

// QueryPartnerpayEngineTransaction queries the partnerpay-engine database for a transaction by run_id
func QueryPartnerpayEngineTransaction(runID string) *TransactionResult {
	return QueryPartnerpayEngineTransactionWithEnv(runID, "my")
}

// QueryPartnerpayEngineTransactionWithEnv queries the partnerpay-engine database for a transaction by run_id with specified environment
func QueryPartnerpayEngineTransactionWithEnv(runID string, env string) *TransactionResult {
	// Get singleton doorman client
	client, err := clients.GetDoormanClient(env)
	if err != nil {
		return &TransactionResult{
			TransactionID: runID,
			Error:         fmt.Sprintf("failed to get doorman client: %v", err),
		}
	}

	// Query the charge table with specific fields
	chargeQuery := fmt.Sprintf("SELECT status, status_reason, status_reason_description, transaction_id FROM charge WHERE transaction_id='%s'", runID)
	charges, err := client.QueryPartnerpayEngine(chargeQuery)
	if err != nil {
		return &TransactionResult{
			TransactionID: runID,
			Error:         fmt.Sprintf("failed to query charge table: %v", err),
		}
	}

	if len(charges) == 0 {
		return &TransactionResult{
			TransactionID: runID,
			PartnerpayEngine: PartnerpayEngineInfo{
				Transfers: PPEChargeInfo{
					TransactionID: runID,
					Status:        NotFoundStatus,
				},
			},
			Error: "No transaction found with the given run_id",
		}
	}

	// Get charge information
	charge := charges[0]
	result := &TransactionResult{
		TransactionID: runID,
	}

	// Extract transaction_id if available
	if transactionID, ok := charge["transaction_id"].(string); ok && transactionID != "" {
		result.TransactionID = transactionID
		result.PartnerpayEngine.Transfers.TransactionID = transactionID
	} else {
		result.PartnerpayEngine.Transfers.TransactionID = runID
	}

	// Extract status
	if status, ok := charge["status"].(string); ok {
		result.PartnerpayEngine.Transfers.Status = status
	}

	// Extract status_reason
	if statusReason, ok := charge["status_reason"].(string); ok {
		result.PartnerpayEngine.Transfers.StatusReason = statusReason
		result.Error = statusReason
	}

	// Extract status_reason_description
	if statusReasonDesc, ok := charge["status_reason_description"].(string); ok {
		result.PartnerpayEngine.Transfers.StatusReasonDescription = statusReasonDesc
	}

	// Query workflow_execution table for workflow_charge information
	workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id='%s' AND workflow_id='workflow_charge'", runID)
	workflows, err := client.QueryPartnerpayEngine(workflowQuery)
	if err != nil {
		// Don't fail the whole operation if workflow query fails
		if result.PartnerpayEngine.Transfers.StatusReasonDescription == "" {
			result.PartnerpayEngine.Transfers.StatusReasonDescription = fmt.Sprintf("Failed to query workflow: %v", err)
		}
		return result
	}

	if len(workflows) > 0 {
		// Get workflow information
		workflow := workflows[0]

		// Set PartnerpayEngineWorkflow info
		if workflowID, workflowIDOk := workflow["workflow_id"]; workflowIDOk {
			result.PartnerpayEngine.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
		}
		result.PartnerpayEngine.Workflow.RunID = runID

		// Extract attempt if available
		if attemptVal, attemptOk := workflow["attempt"]; attemptOk {
			if attemptFloat, ok := attemptVal.(float64); ok {
				result.PartnerpayEngine.Workflow.Attempt = int(attemptFloat)
			}
		}

		// Extract state
		if state, stateOk := workflow["state"]; stateOk {
			if stateInt, ok := state.(float64); ok {
				stateNum := int(stateInt)
				result.PartnerpayEngine.Workflow.State = fmt.Sprintf("%d", stateNum)
			} else {
				result.PartnerpayEngine.Workflow.State = fmt.Sprintf("%v", state)
			}
		}
	}

	return result
}

// queryFastAdapter retrieves fast adapter information using external_id
func queryFastAdapter(client clients.DoormanInterface, externalID, createdAt string) (*FastAdapterInfo, error) {
	if externalID == "" {
		return nil, nil // Skip if no external_id
	}

	// Base query
	query := fmt.Sprintf(`
	SELECT type, instruction_id, status, cancel_reason_code, reject_reason_code, created_at
	       FROM transactions
	       WHERE instruction_id = '%s'`, externalID)

	// Add time filtering if created_at is provided
	if createdAt != "" {
		// Parse created_at timestamp to add 1 hour
		startTime, err := time.Parse(time.RFC3339, createdAt)
		if err == nil {
			endTime := startTime.Add(1 * time.Hour)
			endTimeStr := endTime.Format(time.RFC3339)
			query += fmt.Sprintf(" AND created_at >= '%s' AND created_at <= '%s'", createdAt, endTimeStr)
		}
	}

	query += " LIMIT 1"

	results, err := client.QueryFastAdapter(query)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil // No fast adapter data found
	}

	row := results[0]
	info := &FastAdapterInfo{
		InstructionID: getStringValue(row, "instruction_id"),
		Type:          getStringValue(row, "type"),
		Status:        getStringValue(row, "status"),
		CreatedAt:     getStringValue(row, "created_at"),
	}

	// Parse numeric status code
	if statusCode, ok := row["status_code"].(float64); ok {
		info.StatusCode = int(statusCode)
	}

	info.CancelReasonCode = getStringValue(row, "cancel_reason_code")
	info.RejectReasonCode = getStringValue(row, "reject_reason_code")

	return info, nil
}

// Helper function to safely extract string values
func getStringValue(row map[string]interface{}, key string) string {
	if val, ok := row[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
