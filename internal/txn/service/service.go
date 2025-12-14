package service

import (
	"buddy/internal/clients/doorman"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DoormanClient implements the ports.ClientPort interface
type DoormanClient struct {
	client doorman.DoormanInterface
}

// NewDoormanClient creates a new DoormanClient
func NewDoormanClient() *DoormanClient {
	return &DoormanClient{
		client: doorman.Doorman,
	}
}

func (d *DoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return d.client.QueryPaymentEngine(query)
}

func (d *DoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	return d.client.QueryPaymentCore(query)
}

func (d *DoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	return d.client.QueryRppAdapter(query)
}

func (d *DoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	return d.client.QueryFastAdapter(query)
}

func (d *DoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return d.client.QueryPartnerpayEngine(query)
}

func (d *DoormanClient) ExecuteQuery(cluster, service, database, query string) ([]map[string]interface{}, error) {
	return d.client.ExecuteQuery(cluster, service, database, query)
}

// AdapterSet contains all adapters needed for transaction queries
type AdapterSet struct {
	PaymentEngine    ports.PaymentEnginePort
	PaymentCore      ports.PaymentCorePort
	RPPAdapter       ports.RPPAdapterPort
	FastAdapter      ports.FastAdapterPort
	PartnerpayEngine ports.PartnerpayEnginePort
}

// paymentEngineAdapter implements PaymentEnginePort inline
type paymentEngineAdapter struct {
	client ports.ClientPort
}

func (p *paymentEngineAdapter) QueryTransfer(transactionID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT transaction_id, status, reference_id, created_at, updated_at, type, txn_subtype, txn_domain, external_id FROM transfer WHERE transaction_id='%s'", transactionID)
	transfers, err := p.client.QueryPaymentEngine(query)
	if err != nil || len(transfers) == 0 {
		return nil, err
	}
	return transfers[0], nil
}

func (p *paymentEngineAdapter) QueryWorkflow(referenceID string) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt, created_at, updated_at FROM workflow_execution WHERE run_id='%s'", referenceID)
	workflows, err := p.client.QueryPaymentEngine(query)
	if err != nil || len(workflows) == 0 {
		return nil, err
	}
	return workflows[0], nil
}

// paymentCoreAdapter implements PaymentCorePort inline
type paymentCoreAdapter struct {
	client ports.ClientPort
}

func (p *paymentCoreAdapter) QueryInternalTransactions(transactionID string, createdAt string) ([]map[string]interface{}, error) {
	if createdAt == "" {
		return nil, nil
	}
	startTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	endTime := startTime.Add(1 * time.Hour)
	query := fmt.Sprintf("SELECT tx_id, tx_type, status FROM internal_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'", transactionID, createdAt, endTime.Format(time.RFC3339))
	return p.client.QueryPaymentCore(query)
}

func (p *paymentCoreAdapter) QueryExternalTransactions(transactionID string, createdAt string) ([]map[string]interface{}, error) {
	if createdAt == "" {
		return nil, nil
	}
	startTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	endTime := startTime.Add(1 * time.Hour)
	query := fmt.Sprintf("SELECT ref_id, tx_type, status FROM external_transaction WHERE group_id='%s' AND created_at >= '%s' AND created_at <= '%s'", transactionID, createdAt, endTime.Format(time.RFC3339))
	return p.client.QueryPaymentCore(query)
}

func (p *paymentCoreAdapter) QueryWorkflows(runIDs []string) ([]map[string]interface{}, error) {
	if len(runIDs) == 0 {
		return nil, nil
	}
	quotedRunIDs := make([]string, len(runIDs))
	for i, id := range runIDs {
		quotedRunIDs[i] = "'" + id + "'"
	}
	runIDsStr := strings.Join(quotedRunIDs, ", ")
	query := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id IN (%s)", runIDsStr)
	return p.client.QueryPaymentCore(query)
}

// fastAdapterAdapter implements FastAdapterPort inline
type fastAdapterAdapter struct {
	client ports.ClientPort
}

func (f *fastAdapterAdapter) QueryByInstructionID(instructionID, createdAt string) (*domain.FastAdapterInfo, error) {
	if instructionID == "" {
		return nil, nil
	}
	query := fmt.Sprintf("SELECT type, instruction_id, status, cancel_reason_code, reject_reason_code, created_at FROM transactions WHERE instruction_id = '%s'", instructionID)
	if createdAt != "" {
		startTime, err := time.Parse(time.RFC3339, createdAt)
		if err == nil {
			endTime := startTime.Add(1 * time.Hour)
			query += fmt.Sprintf(" AND created_at >= '%s' AND created_at <= '%s'", createdAt, endTime.Format(time.RFC3339))
		}
	}
	query += " LIMIT 1"
	results, err := f.client.QueryFastAdapter(query)
	if err != nil || len(results) == 0 {
		return nil, err
	}
	row := results[0]
	info := &domain.FastAdapterInfo{
		InstructionID:    getStringValue(row, "instruction_id"),
		Type:             getStringValue(row, "type"),
		CreatedAt:        getStringValue(row, "created_at"),
		CancelReasonCode: getStringValue(row, "cancel_reason_code"),
		RejectReasonCode: getStringValue(row, "reject_reason_code"),
	}
	if statusVal, ok := row["status"]; ok {
		var statusNum int
		if str, ok := statusVal.(string); ok {
			if num, err := strconv.Atoi(str); err == nil {
				statusNum = num
			}
		} else if num, ok := statusVal.(int); ok {
			statusNum = num
		} else if fnum, ok := statusVal.(float64); ok {
			statusNum = int(fnum)
		}

		// Look up status name from domain.FastAdapterStateMaps
		if stateMap, exists := domain.FastAdapterStateMaps[info.Type]; exists {
			if statusName, found := stateMap[statusNum]; found {
				info.Status = statusName
			} else {
				info.Status = "UNKNOWN"
			}
		} else {
			// Fallback for unknown adapter types
			switch statusNum {
			case 0:
				info.Status = "INITIATED"
			case 1:
				info.Status = "PENDING"
			case 2:
				info.Status = "PROCESSING"
			case 3:
				info.Status = "SUCCESS"
			case 4:
				info.Status = "FAILED"
			case 5:
				info.Status = "CANCELLED"
			case 6:
				info.Status = "REJECTED"
			case 7:
				info.Status = "TIMEOUT"
			case 8:
				info.Status = "ERROR"
			default:
				info.Status = "UNKNOWN"
			}
		}
		info.StatusCode = statusNum
	}
	return info, nil
}

// partnerpayEngineAdapter implements PartnerpayEnginePort inline
type partnerpayEngineAdapter struct {
	client ports.ClientPort
}

func (p *partnerpayEngineAdapter) QueryCharge(transactionID string) (domain.PartnerpayEngineInfo, error) {
	query := fmt.Sprintf("SELECT status, status_reason, status_reason_description, transaction_id FROM charge WHERE transaction_id='%s'", transactionID)
	charges, err := p.client.QueryPartnerpayEngine(query)
	if err != nil {
		return domain.PartnerpayEngineInfo{}, fmt.Errorf("failed to query charge table: %v", err)
	}
	if len(charges) == 0 {
		return domain.PartnerpayEngineInfo{Transfers: domain.PPEChargeInfo{TransactionID: transactionID, Status: domain.NotFoundStatus}}, nil
	}
	charge := charges[0]
	result := domain.PartnerpayEngineInfo{}
	if txID, ok := charge["transaction_id"].(string); ok && txID != "" {
		result.Transfers.TransactionID = txID
	} else {
		result.Transfers.TransactionID = transactionID
	}
	if status, ok := charge["status"].(string); ok {
		result.Transfers.Status = status
	}
	if statusReason, ok := charge["status_reason"].(string); ok {
		result.Transfers.StatusReason = statusReason
	}
	if statusReasonDesc, ok := charge["status_reason_description"].(string); ok {
		result.Transfers.StatusReasonDescription = statusReasonDesc
	}
	workflowQuery := fmt.Sprintf("SELECT run_id, workflow_id, state, attempt FROM workflow_execution WHERE run_id='%s' AND workflow_id='workflow_charge'", transactionID)
	if workflows, err := p.client.QueryPartnerpayEngine(workflowQuery); err == nil && len(workflows) > 0 {
		workflow := workflows[0]
		if workflowID, ok := workflow["workflow_id"]; ok {
			result.Workflow.WorkflowID = fmt.Sprintf("%v", workflowID)
		}
		result.Workflow.RunID = transactionID
		if attemptVal, ok := workflow["attempt"]; ok {
			if attemptFloat, ok := attemptVal.(float64); ok {
				result.Workflow.Attempt = int(attemptFloat)
			}
		}
		if state, ok := workflow["state"]; ok {
			if stateInt, ok := state.(float64); ok {
				stateNum := int(stateInt)
				result.Workflow.State = fmt.Sprintf("%d", stateNum)
			} else {
				result.Workflow.State = fmt.Sprintf("%v", state)
			}
		}
	}
	return result, nil
}

// rppAdapter implements RPPAdapterPort inline (Malaysia only)
type rppAdapter struct {
	client ports.ClientPort
}

func (r *rppAdapter) QueryByExternalID(externalID string) (*domain.RPPAdapterInfo, error) {
	query := fmt.Sprintf("SELECT req_biz_msg_id, partner_tx_id, partner_tx_sts AS status FROM credit_transfer WHERE end_to_end_id = '%s'", externalID)
	rppResults, err := r.client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil || len(rppResults) == 0 {
		return nil, err
	}
	row := rppResults[0]
	info := &domain.RPPAdapterInfo{
		ReqBizMsgID: getStringValue(row, "req_biz_msg_id"),
		PartnerTxID: getStringValue(row, "partner_tx_id"),
		EndToEndID:  externalID,
		Status:      getStringValue(row, "status"),
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

func (r *rppAdapter) QueryByE2EID(e2eID string) (*domain.TransactionResult, error) {
	// Validate E2E ID format
	if !r.IsRppE2EID(e2eID) {
		return &domain.TransactionResult{
			TransactionID: e2eID,
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
			RPPAdapter: domain.RPPAdapterInfo{
				EndToEndID: e2eID,
				Status:     domain.NotFoundStatus,
				Info:       "No record found in RPP adapter",
			},
		}, nil
	}

	// Extract RPP data
	row := rppResults[0]
	partnerTxID := getStringValue(row, "partner_tx_id")
	reqBizMsgID := getStringValue(row, "req_biz_msg_id")
	rppStatus := getStringValue(row, "status")
	createdAt := getStringValue(row, "created_at")

	// Initialize result
	result := &domain.TransactionResult{
		TransactionID: e2eID,
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
			result.RPPAdapter.Workflow.WorkflowID = getStringValue(workflow, "workflow_id")
			result.RPPAdapter.Workflow.State = getStringValue(workflow, "state")
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
			result.PaymentEngine.Transfers.TransactionID = getStringValue(transfer, "transaction_id")
			result.PaymentEngine.Transfers.Status = getStringValue(transfer, "status")
			result.PaymentEngine.Transfers.ReferenceID = getStringValue(transfer, "reference_id")
			result.PaymentEngine.Transfers.ExternalID = getStringValue(transfer, "external_id")
			result.PaymentEngine.Transfers.Type = getStringValue(transfer, "type")
			result.PaymentEngine.Transfers.TxnSubtype = getStringValue(transfer, "txn_subtype")
			result.PaymentEngine.Transfers.TxnDomain = getStringValue(transfer, "txn_domain")
			result.PaymentEngine.Transfers.CreatedAt = getStringValue(transfer, "created_at")

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
					result.PaymentEngine.Workflow.WorkflowID = getStringValue(peWorkflow, "workflow_id")
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
						txType := strings.TrimSpace(strings.ToUpper(getStringValue(internalTx, "tx_type")))
						status := strings.TrimSpace(strings.ToUpper(getStringValue(internalTx, "status")))

						result.PaymentCore.InternalTxns = append(result.PaymentCore.InternalTxns, domain.PCInternalTxnInfo{
							TxID:     getStringValue(internalTx, "tx_id"),
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
						txType := strings.TrimSpace(strings.ToUpper(getStringValue(externalTx, "tx_type")))
						status := strings.TrimSpace(strings.ToUpper(getStringValue(externalTx, "status")))

						result.PaymentCore.ExternalTxns = append(result.PaymentCore.ExternalTxns, domain.PCExternalTxnInfo{
							RefID:    getStringValue(externalTx, "ref_id"),
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
							workflowID := getStringValue(pcWorkflow, "workflow_id")
							runID := getStringValue(pcWorkflow, "run_id")
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

func (r *rppAdapter) IsRppE2EID(id string) bool {
	// RppE2EIDPattern from original code
	return len(id) == 30 && id[0:8] >= "20000101" && id[0:8] <= "20991231" && id[22:30] >= "00000000" && id[22:30] <= "99999999"
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
		sopRepo:  adapters.NewSOPRepository(),
		env:      env,
	}
}

// NewTransactionQueryServiceWithAdapters creates a new service with custom adapters
func NewTransactionQueryServiceWithAdapters(adapterSet AdapterSet, env string) *TransactionQueryService {
	return &TransactionQueryService{
		adapters: adapterSet,
		sopRepo:  adapters.NewSOPRepository(),
		env:      env,
	}
}

// createMalaysiaAdapters creates adapters for Malaysia environment
func createMalaysiaAdapters() AdapterSet {
	client := NewDoormanClient()
	return AdapterSet{
		PaymentEngine:    &paymentEngineAdapter{client: client},
		PaymentCore:      &paymentCoreAdapter{client: client},
		RPPAdapter:       &rppAdapter{client: client},
		FastAdapter:      &fastAdapterAdapter{client: client},
		PartnerpayEngine: &partnerpayEngineAdapter{client: client},
	}
}

// createSingaporeAdapters creates adapters for Singapore environment
func createSingaporeAdapters() AdapterSet {
	client := NewDoormanClient()
	return AdapterSet{
		PaymentEngine:    &paymentEngineAdapter{client: client},
		PaymentCore:      &paymentCoreAdapter{client: client},
		RPPAdapter:       nil, // Singapore doesn't use RPP
		FastAdapter:      &fastAdapterAdapter{client: client},
		PartnerpayEngine: &partnerpayEngineAdapter{client: client},
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
		CaseType:      domain.SOPCaseNone,
	}

	// Check if it's an RPP E2E ID and handle separately (Malaysia only)
	if s.env == "my" && s.adapters.RPPAdapter != nil {
		if rppAdapter, ok := s.adapters.RPPAdapter.(*rppAdapter); ok {
			if rppAdapter.IsRppE2EID(transactionID) {
				if rppResult, err := s.adapters.RPPAdapter.QueryByE2EID(transactionID); err == nil && rppResult != nil {
					return rppResult
				}
			}
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
	s.sopRepo.IdentifySOPCase(result, env)

	return result
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
	result.PaymentEngine.Transfers.ReferenceID = getStringValue(transfer, "reference_id")
	result.PaymentEngine.Transfers.CreatedAt = getStringValue(transfer, "created_at")
	result.PaymentEngine.Transfers.Type = getStringValue(transfer, "type")
	result.PaymentEngine.Transfers.TxnSubtype = getStringValue(transfer, "txn_subtype")
	result.PaymentEngine.Transfers.TxnDomain = getStringValue(transfer, "txn_domain")
	result.PaymentEngine.Transfers.ExternalID = getStringValue(transfer, "external_id")
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
			txType := strings.TrimSpace(strings.ToUpper(getStringValue(internalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(getStringValue(internalTx, "status")))

			result.PaymentCore.InternalTxns = append(result.PaymentCore.InternalTxns, domain.PCInternalTxnInfo{
				TxID:     getStringValue(internalTx, "tx_id"),
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
			txType := strings.TrimSpace(strings.ToUpper(getStringValue(externalTx, "tx_type")))
			status := strings.TrimSpace(strings.ToUpper(getStringValue(externalTx, "status")))

			result.PaymentCore.ExternalTxns = append(result.PaymentCore.ExternalTxns, domain.PCExternalTxnInfo{
				RefID:    getStringValue(externalTx, "ref_id"),
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
				workflowID := getStringValue(workflow, "workflow_id")
				runID := getStringValue(workflow, "run_id")
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

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func (s *TransactionQueryService) QueryPartnerpayEngine(runID string) (domain.PartnerpayEngineInfo, error) {
	return s.adapters.PartnerpayEngine.QueryCharge(runID)
}

// GetAdapters returns the adapter set for this service
func (s *TransactionQueryService) GetAdapters() AdapterSet {
	return s.adapters
}
