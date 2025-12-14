package service

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	svcAdapters "buddy/internal/txn/service/adapters"
	"buddy/internal/txn/utils"
	"fmt"
	"strings"
	"time"

	"buddy/internal/txn/adapters"
)

// AdapterSet contains all adapters needed for transaction queries
type AdapterSet struct {
	PaymentEngine    ports.PaymentEnginePort
	PaymentCore      ports.PaymentCorePort
	RPPAdapter       ports.RPPAdapterPort
	FastAdapter      ports.FastAdapterPort
	PartnerpayEngine ports.PartnerpayEnginePort
}

// TransactionQueryService orchestrates transaction queries across multiple data sources
type TransactionQueryService struct {
	adapters AdapterSet
	sopRepo  *adapters.SOPRepository
	env      string
}

// NewTransactionQueryService creates a new transaction query service
func NewTransactionQueryService(env string) *TransactionQueryService {
	var adapterSet AdapterSet

	switch env {
	case "my":
		adapterSet = createMalaysiaAdapters()
	case "sg":
		adapterSet = createSingaporeAdapters()
	default:
		// Default to Malaysia for backward compatibility
		adapterSet = createMalaysiaAdapters()
	}

	return &TransactionQueryService{
		adapters: adapterSet,
		sopRepo:  adapters.SOPRepo,
		env:      env,
	}
}

// NewTransactionQueryServiceWithAdapters creates a new service with custom adapters
func NewTransactionQueryServiceWithAdapters(adapterSet AdapterSet, env string) *TransactionQueryService {
	return &TransactionQueryService{
		adapters: adapterSet,
		sopRepo:  adapters.SOPRepo,
		env:      env,
	}
}

// createMalaysiaAdapters creates adapters for Malaysia environment
func createMalaysiaAdapters() AdapterSet {
	client := NewDoormanClient()
	return AdapterSet{
		PaymentEngine:    svcAdapters.NewPaymentEngineAdapter(client),
		PaymentCore:      svcAdapters.NewPaymentCoreAdapter(client),
		RPPAdapter:       svcAdapters.NewRPPAdapter(client),
		FastAdapter:      svcAdapters.NewFastAdapter(client),
		PartnerpayEngine: svcAdapters.NewPartnerpayEngineAdapter(client),
	}
}

// createSingaporeAdapters creates adapters for Singapore environment
func createSingaporeAdapters() AdapterSet {
	client := NewDoormanClient()
	return AdapterSet{
		PaymentEngine:    svcAdapters.NewPaymentEngineAdapter(client),
		PaymentCore:      svcAdapters.NewPaymentCoreAdapter(client),
		RPPAdapter:       nil, // Singapore doesn't use RPP
		FastAdapter:      svcAdapters.NewFastAdapter(client),
		PartnerpayEngine: svcAdapters.NewPartnerpayEngineAdapter(client),
	}
}

// QueryTransaction retrieves complete transaction information by ID
func (s *TransactionQueryService) QueryTransaction(transactionID string) *domain.TransactionResult {
	return s.QueryTransactionWithEnv(transactionID, s.env)
}

// QueryTransactionWithEnv retrieves complete transaction information by ID with specified environment
func (s *TransactionQueryService) QueryTransactionWithEnv(transactionID string, env string) *domain.TransactionResult {
	result := &domain.TransactionResult{
		TransactionID: transactionID,
		CaseType:      domain.CaseNone,
	}

	// Check if it's an RPP E2E ID and handle separately (Malaysia only)
	if env == "my" && s.adapters.RPPAdapter != nil && domain.IsRppE2EID(transactionID) {
		if rppResult, err := s.adapters.RPPAdapter.QueryByE2EID(transactionID); err == nil && rppResult != nil {
			if rppResult.CaseType == "" {
				rppResult.CaseType = domain.CaseNone
			}
			s.sopRepo.IdentifyCase(rppResult, env)
			return rppResult
		}
	}

	// Query Payment Engine transfer information
	if transfer, err := s.adapters.PaymentEngine.QueryTransfer(transactionID); err == nil && transfer != nil {
		s.populatePaymentEngineInfo(result, transfer)
	} else {
		if err != nil {
			result.Error = fmt.Sprintf("failed to query payment engine: %v", err)
		} else {
			result.Error = "No transaction found with the given ID"
		}
		return result
	}

	// Query Payment Engine workflow information if we have reference ID
	if result.PaymentEngine.Transfers.ReferenceID != "" {
		if workflow, err := s.adapters.PaymentEngine.QueryWorkflow(result.PaymentEngine.Transfers.ReferenceID); err == nil && workflow != nil {
			s.populatePaymentEngineWorkflow(result, workflow)
		}
	}

	// Query RPP adapter if available and we have external ID (Malaysia only)
	if s.adapters.RPPAdapter != nil && result.PaymentEngine.Transfers.ExternalID != "" {
		if rppInfo, err := s.adapters.RPPAdapter.QueryByExternalID(result.PaymentEngine.Transfers.ExternalID); err == nil && rppInfo != nil {
			result.RPPAdapter = *rppInfo
		}
	}

	// Query Payment Core transactions if we have created_at timestamp
	if result.PaymentEngine.Transfers.CreatedAt != "" {
		s.populatePaymentCoreInfo(result)
	}

	// Query Fast Adapter if we have external ID
	if result.PaymentEngine.Transfers.ExternalID != "" {
		if fastInfo, err := s.adapters.FastAdapter.QueryByInstructionID(
			result.PaymentEngine.Transfers.ExternalID,
			result.PaymentEngine.Transfers.CreatedAt); err == nil && fastInfo != nil {
			result.FastAdapter = *fastInfo
		}
	}

	// Identify SOP case using rule engine
	s.sopRepo.IdentifyCase(result, env)

	return result
}

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func (s *TransactionQueryService) QueryPartnerpayEngine(runID string) (domain.PartnerpayEngineInfo, error) {
	return s.adapters.PartnerpayEngine.QueryCharge(runID)
}

// GetAdapters returns the adapter set for this service
func (s *TransactionQueryService) GetAdapters() AdapterSet {
	return s.adapters
}

// populatePaymentEngineInfo populates payment engine transfer information
func (s *TransactionQueryService) populatePaymentEngineInfo(result *domain.TransactionResult, transfer map[string]interface{}) {
	// Extract status
	if status, ok := transfer["status"].(string); ok {
		result.PaymentEngine.Transfers.Status = status
	}

	// Extract transaction_id
	if txID, ok := transfer["transaction_id"].(string); ok && txID != "" {
		result.TransactionID = txID
		result.PaymentEngine.Transfers.TransactionID = txID
	} else {
		result.PaymentEngine.Transfers.TransactionID = result.TransactionID
	}

	// Extract other fields
	result.PaymentEngine.Transfers.ReferenceID = utils.GetStringValue(transfer, "reference_id")
	result.PaymentEngine.Transfers.CreatedAt = utils.GetStringValue(transfer, "created_at")
	result.PaymentEngine.Transfers.Type = utils.GetStringValue(transfer, "type")
	result.PaymentEngine.Transfers.TxnSubtype = utils.GetStringValue(transfer, "txn_subtype")
	result.PaymentEngine.Transfers.TxnDomain = utils.GetStringValue(transfer, "txn_domain")
	result.PaymentEngine.Transfers.ExternalID = utils.GetStringValue(transfer, "external_id")
}

// populatePaymentEngineWorkflow populates payment engine workflow information
func (s *TransactionQueryService) populatePaymentEngineWorkflow(result *domain.TransactionResult, workflow map[string]interface{}) {
	if workflowID, ok := workflow["workflow_id"]; ok {
		result.PaymentEngine.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
	}
	if runID, ok := workflow["run_id"]; ok {
		result.PaymentEngine.Workflow.RunID = fmt.Sprintf("%v", runID)
	}
	if attemptVal, ok := workflow["attempt"]; ok {
		if attemptFloat, ok := attemptVal.(float64); ok {
			result.PaymentEngine.Workflow.Attempt = int(attemptFloat)
		}
	}
	if state, ok := workflow["state"]; ok {
		if stateInt, ok := state.(float64); ok {
			result.PaymentEngine.Workflow.State = fmt.Sprintf("%d", int(stateInt))
		} else {
			result.PaymentEngine.Workflow.State = fmt.Sprintf("%v", state)
		}
	}

	// If we don't have created_at from transfer, use workflow timestamps
	if result.PaymentEngine.Transfers.CreatedAt == "" {
		if createdAt, ok := workflow["created_at"]; ok {
			result.PaymentEngine.Transfers.CreatedAt = fmt.Sprintf("%v", createdAt)
		}
	}
}

// populatePaymentCoreInfo populates payment core information
func (s *TransactionQueryService) populatePaymentCoreInfo(result *domain.TransactionResult) {
	// Query internal transactions
	if internalTxs, err := s.adapters.PaymentCore.QueryInternalTransactions(
		result.TransactionID, result.PaymentEngine.Transfers.CreatedAt); err == nil {
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
	if externalTxs, err := s.adapters.PaymentCore.QueryExternalTransactions(
		result.TransactionID, result.PaymentEngine.Transfers.CreatedAt); err == nil {
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

	// Query workflows for payment-core
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
		if workflows, err := s.adapters.PaymentCore.QueryWorkflows(runIDs); err == nil {
			for _, workflow := range workflows {
				workflowID := utils.GetStringValue(workflow, "workflow_id")
				runID := utils.GetStringValue(workflow, "run_id")
				state := workflow["state"]

				var stateNum int
				if stateInt, ok := state.(float64); ok {
					stateNum = int(stateInt)
				}

				var attempt int
				if attemptFloat, ok := workflow["attempt"].(float64); ok {
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

// QueryPaymentCoreTransactions queries Payment Core transactions for a given transaction ID within a time window
// This method can be used by adapters to populate Payment Core data
func (s *TransactionQueryService) QueryPaymentCoreTransactions(result *domain.TransactionResult, transactionID string, windowStart, windowEnd time.Time) {
	client := NewDoormanClient()

	// Query internal transactions
	if internalTxs, err := client.QueryPaymentCore(fmt.Sprintf(
		"SELECT tx_id, tx_type, status FROM internal_transaction WHERE tx_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
		transactionID,
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
	if externalTxs, err := client.QueryPaymentCore(fmt.Sprintf(
		"SELECT ref_id, tx_type, status FROM external_transaction WHERE ref_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
		transactionID,
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
		if pcWorkflows, err := client.QueryPaymentCore(pcWorkflowQuery); err == nil {
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
