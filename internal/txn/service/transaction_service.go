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
func (s *TransactionQueryService) QueryTransactionWithEnv(inputID string, env string) *domain.TransactionResult {
	result := &domain.TransactionResult{
		InputID:  inputID,
		CaseType: domain.CaseNone,
	}

	// Step 1: Determine input type and fill primary adapters
	isE2EID := domain.IsRppE2EID(inputID) &&
		((env == "my" && s.adapters.RPPAdapter != nil) ||
			(env == "sg" && s.adapters.FastAdapter != nil))

	if isE2EID {
		// E2E ID: Fill RPP/Fast adapter first
		s.fillAdapterFromE2EID(result, env)
	} else {
		// Transaction ID: Fill PaymentEngine first
		s.fillPaymentEngineFromTransactionID(result, inputID)
	}

	// Step 2: Ensure PaymentEngine is populated
	if result.PaymentEngine == nil {
		s.populatePaymentEngineFromAdapters(result, env)
	}

	// Step 3: Ensure PaymentCore is populated
	if result.PaymentCore == nil && result.PaymentEngine != nil {
		s.populatePaymentCoreInfo(result)
	}

	// Step 4: Ensure adapters are populated from PaymentEngine
	if result.PaymentEngine != nil {
		s.populateAdaptersFromPaymentEngine(result, env)
	}

	// Step 5: Identify SOP case
	s.sopRepo.IdentifyCase(result, env)

	return result
}

// fillAdapterFromE2EID fills RPP/Fast adapter from E2E ID
func (s *TransactionQueryService) fillAdapterFromE2EID(result *domain.TransactionResult, env string) {
	if env == "my" && s.adapters.RPPAdapter != nil {
		// Malaysia: Query RPP adapter
		if rppInfo, err := s.adapters.RPPAdapter.QueryByE2EID(result.InputID); err == nil && rppInfo != nil {
			result.RPPAdapter = rppInfo
		}
	} else if env == "sg" && s.adapters.FastAdapter != nil {
		// Singapore: Query Fast adapter
		if fastInfo, err := s.adapters.FastAdapter.QueryByInstructionID(result.InputID, ""); err == nil && fastInfo != nil {
			result.FastAdapter = fastInfo
		}
	}
}

// fillPaymentEngineFromTransactionID fills PaymentEngine from Transaction ID
func (s *TransactionQueryService) fillPaymentEngineFromTransactionID(result *domain.TransactionResult, transactionID string) {
	if transfer, err := s.adapters.PaymentEngine.QueryTransfer(transactionID); err == nil && transfer != nil {
		paymentEngineInfo := &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{},
			Workflow:  domain.WorkflowInfo{},
		}
		s.populatePaymentEngineInfoFromTransfer(paymentEngineInfo, transfer)
		result.PaymentEngine = paymentEngineInfo

		// Query workflow if we have reference_id
		if paymentEngineInfo.Transfers.ReferenceID != "" {
			if workflow, err := s.adapters.PaymentEngine.QueryWorkflow(
				paymentEngineInfo.Transfers.ReferenceID); err == nil && workflow != nil {
				s.populatePaymentEngineWorkflowFromWorkflow(paymentEngineInfo, workflow)
			}
		}
	} else {
		// For non-E2E IDs, if payment-engine query fails, return error
		if err != nil {
			result.Error = fmt.Sprintf("failed to query payment engine: %v", err)
		} else {
			result.Error = "No transaction found with the given ID"
		}
	}
}

// populatePaymentEngineFromAdapters populates PaymentEngine from RPP/Fast adapters
func (s *TransactionQueryService) populatePaymentEngineFromAdapters(result *domain.TransactionResult, env string) {
	if env == "my" && result.RPPAdapter != nil {
		// Use RPP adapter data
		if result.RPPAdapter.EndToEndID != "" && result.RPPAdapter.CreatedAt != "" {
			if transfer, err := s.adapters.PaymentEngine.QueryTransferByExternalID(
				result.RPPAdapter.EndToEndID, result.RPPAdapter.CreatedAt); err == nil && transfer != nil {
				paymentEngineInfo := &domain.PaymentEngineInfo{
					Transfers: domain.PETransfersInfo{},
					Workflow:  domain.WorkflowInfo{},
				}
				s.populatePaymentEngineInfoFromTransfer(paymentEngineInfo, transfer)
				result.PaymentEngine = paymentEngineInfo

				// Query workflow if we have reference_id
				if paymentEngineInfo.Transfers.ReferenceID != "" {
					if workflow, err := s.adapters.PaymentEngine.QueryWorkflow(
						paymentEngineInfo.Transfers.ReferenceID); err == nil && workflow != nil {
						s.populatePaymentEngineWorkflowFromWorkflow(paymentEngineInfo, workflow)
					}
				}
			}
		}
	} else if env == "sg" && result.FastAdapter != nil {
		// Use Fast adapter data
		if result.FastAdapter.InstructionID != "" && result.FastAdapter.CreatedAt != "" {
			if transfer, err := s.adapters.PaymentEngine.QueryTransferByExternalID(
				result.FastAdapter.InstructionID, result.FastAdapter.CreatedAt); err == nil && transfer != nil {
				paymentEngineInfo := &domain.PaymentEngineInfo{
					Transfers: domain.PETransfersInfo{},
					Workflow:  domain.WorkflowInfo{},
				}
				s.populatePaymentEngineInfoFromTransfer(paymentEngineInfo, transfer)
				result.PaymentEngine = paymentEngineInfo

				// Query workflow if we have reference_id
				if paymentEngineInfo.Transfers.ReferenceID != "" {
					if workflow, err := s.adapters.PaymentEngine.QueryWorkflow(
						paymentEngineInfo.Transfers.ReferenceID); err == nil && workflow != nil {
						s.populatePaymentEngineWorkflowFromWorkflow(paymentEngineInfo, workflow)
					}
				}
			}
		}
	}
}

// populateAdaptersFromPaymentEngine populates adapters from PaymentEngine
func (s *TransactionQueryService) populateAdaptersFromPaymentEngine(result *domain.TransactionResult, env string) {
	if result.PaymentEngine.Transfers.ExternalID == "" {
		return
	}

	if env == "my" && result.RPPAdapter == nil && s.adapters.RPPAdapter != nil {
		// Query RPP adapter using external_id
		if rppInfo, err := s.adapters.RPPAdapter.QueryByE2EID(
			result.PaymentEngine.Transfers.ExternalID); err == nil && rppInfo != nil {
			result.RPPAdapter = rppInfo
		}
	} else if env == "sg" && result.FastAdapter == nil && s.adapters.FastAdapter != nil {
		// Query Fast adapter using external_id
		if fastInfo, err := s.adapters.FastAdapter.QueryByInstructionID(
			result.PaymentEngine.Transfers.ExternalID,
			result.PaymentEngine.Transfers.CreatedAt); err == nil && fastInfo != nil {
			result.FastAdapter = fastInfo
		}
	}
}

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func (s *TransactionQueryService) QueryPartnerpayEngine(runID string) (domain.PartnerpayEngineInfo, error) {
	return s.adapters.PartnerpayEngine.QueryCharge(runID)
}

// GetAdapters returns the adapter set for this service
func (s *TransactionQueryService) GetAdapters() AdapterSet {
	return s.adapters
}

// populatePaymentCoreInfo populates payment core information
func (s *TransactionQueryService) populatePaymentCoreInfo(result *domain.TransactionResult) {
	// Use the transaction_id from payment engine if available, otherwise use InputID
	transactionID := result.PaymentEngine.Transfers.TransactionID
	if transactionID == "" {
		transactionID = result.InputID
	}

	// Query internal transactions
	if internalTxs, err := s.adapters.PaymentCore.QueryInternalTransactions(
		transactionID, result.PaymentEngine.Transfers.CreatedAt); err == nil {
		for _, internalTx := range internalTxs {
			txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "status")))

			result.PaymentCore.InternalTxns = append(result.PaymentCore.InternalTxns, domain.PCInternalTxnInfo{
				TxID:      utils.GetStringValue(internalTx, "tx_id"),
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				CreatedAt: utils.GetStringValue(internalTx, "created_at"),
			})
		}
	}

	// Query external transactions
	if externalTxs, err := s.adapters.PaymentCore.QueryExternalTransactions(
		transactionID, result.PaymentEngine.Transfers.CreatedAt); err == nil {
		for _, externalTx := range externalTxs {
			txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "status")))

			result.PaymentCore.ExternalTxns = append(result.PaymentCore.ExternalTxns, domain.PCExternalTxnInfo{
				RefID:     utils.GetStringValue(externalTx, "ref_id"),
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				CreatedAt: utils.GetStringValue(externalTx, "created_at"),
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
		"SELECT tx_id, tx_type, status, created_at FROM internal_transaction WHERE tx_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
		transactionID,
		windowStart.Format(time.RFC3339),
		windowEnd.Format(time.RFC3339),
	)); err == nil {
		for _, internalTx := range internalTxs {
			txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "status")))

			result.PaymentCore.InternalTxns = append(result.PaymentCore.InternalTxns, domain.PCInternalTxnInfo{
				TxID:      utils.GetStringValue(internalTx, "tx_id"),
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				CreatedAt: utils.GetStringValue(internalTx, "created_at"),
			})
		}
	}

	// Query external transactions
	if externalTxs, err := client.QueryPaymentCore(fmt.Sprintf(
		"SELECT ref_id, tx_type, status, created_at FROM external_transaction WHERE ref_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
		transactionID,
		windowStart.Format(time.RFC3339),
		windowEnd.Format(time.RFC3339),
	)); err == nil {
		for _, externalTx := range externalTxs {
			txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "status")))

			result.PaymentCore.ExternalTxns = append(result.PaymentCore.ExternalTxns, domain.PCExternalTxnInfo{
				RefID:     utils.GetStringValue(externalTx, "ref_id"),
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				CreatedAt: utils.GetStringValue(externalTx, "created_at"),
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

// populatePaymentEngineInfoFromTransfer populates payment engine transfer information from transfer map to PaymentEngineInfo
func (s *TransactionQueryService) populatePaymentEngineInfoFromTransfer(paymentEngineInfo *domain.PaymentEngineInfo, transfer map[string]interface{}) {
	// Extract status
	if status, ok := transfer["status"].(string); ok {
		paymentEngineInfo.Transfers.Status = status
	}

	// Extract transaction_id
	if txID, ok := transfer["transaction_id"].(string); ok && txID != "" {
		paymentEngineInfo.Transfers.TransactionID = txID
	}

	// Extract other fields
	paymentEngineInfo.Transfers.ReferenceID = utils.GetStringValue(transfer, "reference_id")
	paymentEngineInfo.Transfers.CreatedAt = utils.GetStringValue(transfer, "created_at")
	paymentEngineInfo.Transfers.Type = utils.GetStringValue(transfer, "type")
	paymentEngineInfo.Transfers.TxnSubtype = utils.GetStringValue(transfer, "txn_subtype")
	paymentEngineInfo.Transfers.TxnDomain = utils.GetStringValue(transfer, "txn_domain")
	paymentEngineInfo.Transfers.ExternalID = utils.GetStringValue(transfer, "external_id")
}

// populatePaymentEngineWorkflowFromWorkflow populates payment engine workflow information from workflow map to PaymentEngineInfo
func (s *TransactionQueryService) populatePaymentEngineWorkflowFromWorkflow(paymentEngineInfo *domain.PaymentEngineInfo, workflow map[string]interface{}) {
	if workflowID, ok := workflow["workflow_id"]; ok {
		paymentEngineInfo.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
	}
	if runID, ok := workflow["run_id"]; ok {
		paymentEngineInfo.Workflow.RunID = fmt.Sprintf("%v", runID)
	}
	if attemptVal, ok := workflow["attempt"]; ok {
		if attemptFloat, ok := attemptVal.(float64); ok {
			paymentEngineInfo.Workflow.Attempt = int(attemptFloat)
		}
	}
	if state, ok := workflow["state"]; ok {
		if stateInt, ok := state.(float64); ok {
			paymentEngineInfo.Workflow.State = fmt.Sprintf("%d", int(stateInt))
		} else {
			paymentEngineInfo.Workflow.State = fmt.Sprintf("%v", state)
		}
	}

	// Populate prev_trans_id field
	if prevTransID, ok := workflow["prev_trans_id"]; ok {
		paymentEngineInfo.Workflow.PrevTransID = fmt.Sprintf("%v", prevTransID)
	}

	// If we do not have created_at from transfer, use workflow timestamps
	if paymentEngineInfo.Transfers.CreatedAt == "" {
		if createdAt, ok := workflow["created_at"]; ok {
			paymentEngineInfo.Transfers.CreatedAt = fmt.Sprintf("%v", createdAt)
		}
	}
}
