package service

import (
	"buddy/internal/clients/doorman"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	svcAdapters "buddy/internal/txn/service/adapters"
	"buddy/internal/txn/utils"
	"fmt"
	"strings"
	"sync"
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

var (
	txnSvc *TransactionQueryService
	once   sync.Once
)

// NewTransactionQueryService creates a new transaction query service singleton
func NewTransactionQueryService(env string) *TransactionQueryService {
	once.Do(func() {
		txnSvc = createTransactionService(env)
	})
	return txnSvc
}

// GetTransactionQueryService returns the singleton instance
func GetTransactionQueryService() *TransactionQueryService {
	if txnSvc == nil {
		panic("TransactionQueryService not initialized. Call NewTransactionQueryService(env) first.")
	}
	return txnSvc
}

// createTransactionService creates a new transaction query service for the given environment
func createTransactionService(env string) *TransactionQueryService {
	var adapterSet AdapterSet

	switch env {
	case "my":
		adapterSet = createMalaysiaAdapters()
	case "sg":
		adapterSet = createSingaporeAdapters()
	default:
		panic("unsupported environment: " + env)
	}

	return &TransactionQueryService{
		adapters: adapterSet,
		sopRepo:  adapters.SOPRepo,
		env:      env,
	}
}

// createMalaysiaAdapters creates adapters for Malaysia environment
func createMalaysiaAdapters() AdapterSet {
	client := doorman.Doorman
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
	client := doorman.Doorman
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
	if domain.IsRppE2EID(inputID) &&
		((env == "my" && s.adapters.RPPAdapter != nil) ||
			(env == "sg" && s.adapters.FastAdapter != nil)) {
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

// QueryEcoTransactionWithEnv retrieves ecological transaction information by run_id
func (s *TransactionQueryService) QueryEcoTransactionWithEnv(runID string, env string) *domain.TransactionResult {
	result := &domain.TransactionResult{
		InputID:  runID,
		CaseType: domain.CaseNone,
	}

	// Step 1: Fill PartnerpayEngine first for ecological transactions
	s.fillPartnerpayEngineFromRunID(result, runID)

	// Step 2: If we have partnerpay-engine data, query payment-core
	if result.PartnerpayEngine != nil {
		s.populatePaymentCoreEcoTxn(result)
	}

	// Step 3: Identify SOP case
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

// fillPartnerpayEngineFromRunID fills partnerpay-engine data from run_id
func (s *TransactionQueryService) fillPartnerpayEngineFromRunID(result *domain.TransactionResult, runID string) {
	if s.adapters.PartnerpayEngine == nil {
		return
	}

	if info, err := s.adapters.PartnerpayEngine.QueryCharge(runID); err == nil {
		result.PartnerpayEngine = &info
	}
}

// populatePaymentCoreEcoTxn queries payment-core using partnerpay-engine transaction info for ecological transactions
func (s *TransactionQueryService) populatePaymentCoreEcoTxn(result *domain.TransactionResult) {
	if result.PartnerpayEngine == nil || s.adapters.PaymentCore == nil {
		return
	}

	// Initialize PaymentCore if nil
	if result.PaymentCore == nil {
		result.PaymentCore = &domain.PaymentCoreInfo{}
	}

	// Use transaction_id as group_id to query payment-core
	groupID := result.PartnerpayEngine.Transfers.TransactionID
	createdAt := result.PartnerpayEngine.Transfers.CreatedAt

	if groupID != "" && createdAt != "" {
		// Query internal transactions
		if internalTxs, err := s.adapters.PaymentCore.QueryInternalTransactions(groupID, createdAt); err == nil {
			for _, internalTx := range internalTxs {
				txType := ""
				txStatus := ""
				txID := ""
				groupIDValue := ""
				createdAtValue := ""
				errorCode := ""
				errorMsg := ""

				if val, ok := internalTx["tx_type"].(string); ok {
					txType = val
				}
				if val, ok := internalTx["status"].(string); ok {
					txStatus = val
				}
				if val, ok := internalTx["tx_id"].(string); ok {
					txID = val
				}
				if val, ok := internalTx["group_id"].(string); ok {
					groupIDValue = val
				}
				if val, ok := internalTx["created_at"].(string); ok {
					createdAtValue = val
				}
				if val, ok := internalTx["error_code"].(string); ok {
					errorCode = val
				}
				if val, ok := internalTx["error_msg"].(string); ok {
					errorMsg = val
				}

				// Query workflow for this transaction
				workflowInfo := s.queryWorkflowInfo(txID)

				internalInfo := domain.PCInternalInfo{
					TxID:      txID,
					GroupID:   groupIDValue,
					TxType:    txType,
					TxStatus:  txStatus,
					ErrorCode: errorCode,
					ErrorMsg:  errorMsg,
					CreatedAt: createdAtValue,
					Workflow:  workflowInfo,
				}

				// Populate based on transaction type
				switch txType {
				case "AUTH":
					result.PaymentCore.InternalAuth = internalInfo
				case "CAPTURE":
					result.PaymentCore.InternalCapture = internalInfo
				}
			}
		}

		// Query external transactions
		if externalTxs, err := s.adapters.PaymentCore.QueryExternalTransactions(groupID, createdAt); err == nil {
			for _, externalTx := range externalTxs {
				txType := ""
				txStatus := ""
				refID := ""
				groupIDValue := ""
				createdAtValue := ""

				if val, ok := externalTx["tx_type"].(string); ok {
					txType = val
				}
				if val, ok := externalTx["status"].(string); ok {
					txStatus = val
				}
				if val, ok := externalTx["ref_id"].(string); ok {
					refID = val
				}
				if val, ok := externalTx["group_id"].(string); ok {
					groupIDValue = val
				}
				if val, ok := externalTx["created_at"].(string); ok {
					createdAtValue = val
				}

				// Query workflow for this transaction
				workflowInfo := s.queryWorkflowInfo(refID)

				externalInfo := domain.PCExternalInfo{
					RefID:     refID,
					GroupID:   groupIDValue,
					TxType:    txType,
					TxStatus:  txStatus,
					CreatedAt: createdAtValue,
					Workflow:  workflowInfo,
				}

				result.PaymentCore.ExternalTransfer = externalInfo
			}
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
	// Initialize PaymentCore if nil to prevent nil pointer dereference
	if result.PaymentCore == nil {
		result.PaymentCore = &domain.PaymentCoreInfo{}
	}

	// Defensive check: if PaymentEngine is nil, we can't proceed
	if result.PaymentEngine == nil {
		return
	}

	// Defensive check: if PaymentCore adapter is nil, we can't proceed
	if s.adapters.PaymentCore == nil {
		return
	}

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

			// Query workflow for this transaction
			txID := utils.GetStringValue(internalTx, "tx_id")
			workflowInfo := s.queryWorkflowInfo(txID)

			internalInfo := domain.PCInternalInfo{
				TxID:      txID,
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				ErrorCode: utils.GetStringValue(internalTx, "error_code"),
				ErrorMsg:  utils.GetStringValue(internalTx, "error_msg"),
				CreatedAt: utils.GetStringValue(internalTx, "created_at"),
				Workflow:  workflowInfo,
			}

			// Populate based on transaction type
			switch txType {
			case "AUTH":
				result.PaymentCore.InternalAuth = internalInfo
			case "CAPTURE":
				result.PaymentCore.InternalCapture = internalInfo
			}
		}
	}

	// Query external transactions
	if externalTxs, err := s.adapters.PaymentCore.QueryExternalTransactions(
		transactionID, result.PaymentEngine.Transfers.CreatedAt); err == nil {
		for _, externalTx := range externalTxs {
			txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(externalTx, "status")))

			// Query workflow for this transaction
			refID := utils.GetStringValue(externalTx, "ref_id")
			workflowInfo := s.queryWorkflowInfo(refID)

			externalInfo := domain.PCExternalInfo{
				RefID:     refID,
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				CreatedAt: utils.GetStringValue(externalTx, "created_at"),
				Workflow:  workflowInfo,
			}

			// Populate based on transaction type
			if txType == "TRANSFER" {
				result.PaymentCore.ExternalTransfer = externalInfo
			}
		}
	}
}

// queryWorkflowInfo queries workflow information for a given run ID
func (s *TransactionQueryService) queryWorkflowInfo(runID string) domain.WorkflowInfo {
	workflowInfo := domain.WorkflowInfo{}

	if runID == "" || s.adapters.PaymentCore == nil {
		return workflowInfo
	}

	// Query workflow for this run ID
	if workflows, err := s.adapters.PaymentCore.QueryWorkflows([]string{runID}); err == nil && len(workflows) > 0 {
		workflow := workflows[0] // Take the first match
		workflowID := utils.GetStringValue(workflow, "workflow_id")
		state := workflow["state"]

		var stateNum int
		if stateInt, ok := state.(float64); ok {
			stateNum = int(stateInt)
		}

		var attempt int
		if attemptFloat, ok := workflow["attempt"].(float64); ok {
			attempt = int(attemptFloat)
		}

		workflowInfo = domain.WorkflowInfo{
			WorkflowID:  workflowID,
			RunID:       runID,
			State:       fmt.Sprintf("%d", stateNum),
			Attempt:     attempt,
			PrevTransID: utils.GetStringValue(workflow, "prev_trans_id"),
		}
	}

	return workflowInfo
}

// QueryPaymentCoreTransactions queries Payment Core transactions for a given transaction ID within a time window
// This method can be used by adapters to populate Payment Core data
func (s *TransactionQueryService) QueryPaymentCoreTransactions(result *domain.TransactionResult, transactionID string, windowStart, windowEnd time.Time) {
	client := doorman.Doorman

	// Initialize PaymentCore if nil
	if result.PaymentCore == nil {
		result.PaymentCore = &domain.PaymentCoreInfo{}
	}

	// Query internal transactions
	if internalTxs, err := client.QueryPaymentCore(fmt.Sprintf(
		"SELECT tx_id, tx_type, status, error_code, error_msg, created_at FROM internal_transaction WHERE tx_id = '%s' AND created_at >= '%s' AND created_at <= '%s'",
		transactionID,
		windowStart.Format(time.RFC3339),
		windowEnd.Format(time.RFC3339),
	)); err == nil {
		for _, internalTx := range internalTxs {
			txType := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(utils.GetStringValue(internalTx, "status")))

			// Query workflow for this transaction
			txID := utils.GetStringValue(internalTx, "tx_id")
			workflowInfo := s.queryWorkflowInfoDirect(client, txID)

			internalInfo := domain.PCInternalInfo{
				TxID:      txID,
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				ErrorCode: utils.GetStringValue(internalTx, "error_code"),
				ErrorMsg:  utils.GetStringValue(internalTx, "error_msg"),
				CreatedAt: utils.GetStringValue(internalTx, "created_at"),
				Workflow:  workflowInfo,
			}

			// Populate based on transaction type
			switch txType {
			case "AUTH":
				result.PaymentCore.InternalAuth = internalInfo
			case "CAPTURE":
				result.PaymentCore.InternalCapture = internalInfo
			}
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

			// Query workflow for this transaction
			refID := utils.GetStringValue(externalTx, "ref_id")
			workflowInfo := s.queryWorkflowInfoDirect(client, refID)

			externalInfo := domain.PCExternalInfo{
				RefID:     refID,
				GroupID:   transactionID,
				TxType:    txType,
				TxStatus:  status,
				CreatedAt: utils.GetStringValue(externalTx, "created_at"),
				Workflow:  workflowInfo,
			}

			// Populate based on transaction type
			if txType == "TRANSFER" {
				result.PaymentCore.ExternalTransfer = externalInfo
			}
		}
	}
}

// queryWorkflowInfoDirect queries workflow information using a direct client connection
func (s *TransactionQueryService) queryWorkflowInfoDirect(client doorman.DoormanInterface, runID string) domain.WorkflowInfo {
	workflowInfo := domain.WorkflowInfo{}

	if runID == "" {
		return workflowInfo
	}

	// Query workflow for this run ID
	pcWorkflowQuery := fmt.Sprintf(
		"SELECT run_id, workflow_id, state, attempt, prev_trans_id FROM workflow_execution WHERE run_id = '%s' LIMIT 1",
		runID,
	)
	if pcWorkflows, err := client.QueryPaymentCore(pcWorkflowQuery); err == nil && len(pcWorkflows) > 0 {
		pcWorkflow := pcWorkflows[0]
		workflowID := utils.GetStringValue(pcWorkflow, "workflow_id")
		state := pcWorkflow["state"]

		var stateNum int
		if stateInt, ok := state.(float64); ok {
			stateNum = int(stateInt)
		}

		var attempt int
		if attemptFloat, ok := pcWorkflow["attempt"].(float64); ok {
			attempt = int(attemptFloat)
		}

		workflowInfo = domain.WorkflowInfo{
			WorkflowID:  workflowID,
			RunID:       runID,
			State:       fmt.Sprintf("%d", stateNum),
			Attempt:     attempt,
			PrevTransID: utils.GetStringValue(pcWorkflow, "prev_trans_id"),
		}
	}

	return workflowInfo
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
