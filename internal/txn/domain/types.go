package domain

// PETransfersInfo contains payment-engine transfer information
type PETransfersInfo struct {
	Type                 string  // payment-engine transfers.type
	TxnSubtype           string  // payment-engine transfers.txn_subtype
	TxnDomain            string  // payment-engine transfers.txn_domain
	TransactionID        string  // payment-engine transfers.transaction_id
	ReferenceID          string  // payment-engine transfers.reference_id
	Status               string  // payment-engine transfers.status
	ExternalID           string  // payment-engine transfers.external_id
	SourceAccountID      string  // payment-engine transfers.source_account_id
	DestinationAccountID string  // payment-engine transfers.destination_account_id
	Amount               float64 // payment-engine transfers.amount (in cents)
	CreatedAt            string  // payment-engine transfers.created_at
	UpdatedAt            string  // payment-engine transfers.updated_at
	Properties           string  // payment-engine transfers.properties (JSON string)
}

// WorkflowInfo contains information about a specific workflow execution
type WorkflowInfo struct {
	WorkflowID  string // workflow_execution.workflow_id
	Attempt     int    // workflow_execution.attempt
	State       string // workflow_execution.state
	RunID       string // workflow_execution.run_id
	PrevTransID string // workflow_execution.prev_trans_id
	Data        string // workflow_execution.data (full JSON data)
}

// GetFormattedState returns the formatted state with name and number
func (w *WorkflowInfo) GetFormattedState() string {
	return FormatWorkflowState(w.WorkflowID, w.State)
}

// PaymentEngineInfo contains payment-engine related information
type PaymentEngineInfo struct {
	Transfers PETransfersInfo
	Workflow  WorkflowInfo
}

// PCInternalTxnInfo contains payment-core internal transaction information
type PCInternalInfo struct {
	TxID      string // payment-core internal transaction ID
	GroupID   string // payment-core transaction group ID
	TxType    string // payment-core transaction type
	TxStatus  string // payment-core transaction status
	ErrorCode string // payment-core internal transaction error_code
	ErrorMsg  string // payment-core internal transaction error_msg
	CreatedAt string // created_at timestamp
	Workflow  WorkflowInfo
}

// PCExternalTxnInfo contains payment-core external transaction information
type PCExternalInfo struct {
	RefID     string // payment-core external transaction reference ID
	GroupID   string // payment-core transaction group ID
	TxType    string // payment-core transaction type
	TxStatus  string // payment-core transaction status
	CreatedAt string // created_at timestamp
	Workflow  WorkflowInfo
}

// PaymentCoreInfo contains payment-core related information
type PaymentCoreInfo struct {
	InternalAuth     PCInternalInfo // when TxType is AUTH
	InternalCapture  PCInternalInfo // when TxType is CAPTURE
	ExternalTransfer PCExternalInfo // when TxType is TRANSFER
}

// FastQueryParams contains parameters for querying Fast adapter
type FastQueryParams struct {
	InstructionID string
	Timestamp     string
}

// FastAdapterInfo contains fast adapter related information
type FastAdapterInfo struct {
	InstructionID    string // instruction_id or external_id
	Type             string // cashin, cashout, etc.
	Status           string // StErraneous, etc.
	StatusCode       int    // Numeric status code
	CancelReasonCode string // cancel_reason_code
	RejectReasonCode string // reject_reason_code
	CreatedAt        string // created_at timestamp
}

// RPPQueryParams contains parameters for querying RPP adapter
type RPPQueryParams struct {
	EndToEndID           string
	PartnerTxID          string
	SourceAccountID      string
	DestinationAccountID string
	Amount               float64
	Timestamp            string
}

// RPPAdapterInfo contains RPP adapter related information
type RPPAdapterInfo struct {
	ReqBizMsgID  string // RPP request business message ID
	PartnerMsgID string // RPP partner message ID
	PartnerTxID  string // RPP partner transaction ID
	EndToEndID   string // = payment-engine.transfers.external_id
	Status       string // RPP status
	CreatedAt    string // created_at timestamp
	Workflow     []WorkflowInfo
	Info         string // optional extra context (e.g. status reason description)
}

// PPEChargeInfo contains partnerpay-engine charge information
type PPEChargeInfo struct {
	TransactionID           string // partnerpay-engine charge.transaction_id
	Status                  string // partnerpay-engine charge.status
	StatusReason            string // partnerpay-engine charge.status_reason
	StatusReasonDescription string // partnerpay-engine charge.status_reason_description
	CreatedAt               string // created_at timestamp
	UpdatedAt               string // updated_at timestamp
	ErrorCode               string // partnerpay-engine charge error code (from workflow data)
	ErrorMsg                string // partnerpay-engine charge error message (from workflow data)
}

// PartnerpayEngineInfo contains partnerpay-engine related information
type PartnerpayEngineInfo struct {
	Charge   PPEChargeInfo
	Workflow WorkflowInfo
}

// TransactionResult represents the result of a transaction query
type TransactionResult struct {
	Index            int
	InputID          string
	PaymentEngine    *PaymentEngineInfo
	PartnerpayEngine *PartnerpayEngineInfo
	PaymentCore      *PaymentCoreInfo
	FastAdapter      *FastAdapterInfo
	RPPAdapter       *RPPAdapterInfo
	CaseType         Case // Store the identified SOP case to avoid re-identification
	Error            string
}

// Common status values
const NotFoundStatus = "NOT_FOUND"

// Case represents the supported remediation cases from SOP.md
type Case string

const (
	CaseNone                                         Case = NotFoundStatus
	CasePcExternalPaymentFlow200_11                  Case = "pc_external_payment_flow_200_11"
	CasePcExternalPaymentFlow201_0RPP210             Case = "pc_external_payment_flow_201_0_RPP_210"
	CasePcExternalPaymentFlow201_0RPP900             Case = "pc_external_payment_flow_201_0_RPP_900"
	CasePeTransferPayment210_0                       Case = "pe_transfer_payment_210_0"
	CasePeStuckAtLimitCheck102                       Case = "pe_stuck_at_limit_check_102_4"
	CasePeStuck230RepublishPC                        Case = "pe_stuck_230_republish_pc"
	CaseRppCashoutReject101_19                       Case = "rpp_cashout_reject_101_19"
	CaseRppQrPaymentReject210_0                      Case = "rpp_qr_payment_reject_210_0"
	CaseRppNoResponseRejectNotFound                  Case = "rpp_no_response_reject_not_found"
	CaseRppNoResponseResume                          Case = "rpp_no_response_resume"
	CaseRppCashinValidationFailed122_0               Case = "rpp_cashin_validation_failed_122_0"
	CaseRppRtpCashinStuck200_0                       Case = "rpp_rtp_cashin_stuck_200_0"
	CasePe2200FastCashinFailed                       Case = "pe_220_0_fast_cashin_failed"
	CaseThoughtMachineFalseNegative                  Case = "thought_machine_false_negative"
	CasePeCaptureProcessingPcCaptureFailedRppSuccess Case = "pe_capture_processing_pc_capture_failed_rpp_success"
	CaseEcotxnChargeFailedCaptureFailedTMError       Case = "ecotxn_ChargeFailed_CaptureFailed_TMError"
	CasePeStuck300RppNotFound                        Case = "pe_stuck_300_rpp_not_found"
	CaseCashoutPe220Pc201Reject                      Case = "cashout_pe220_pc201_reject"
	CaseCashoutRpp210Pe220Pc201                      Case = "cashout_rpp210_pe220_pc201"
	CaseRpp210Pe220Pc201Accept                       Case = "rpp210_pe220_pc201_accept"
	CaseRpp210Pe220Pc201Reject                       Case = "rpp210_pe220_pc201_reject"
	CasePe220Pc201Rpp0StuckInit                      Case = "pe220_pc201_rpp0_stuck_init"
	CaseRppProcessRegistryStuckInit                  Case = "rpp_process_registry_stuck_init"
	CaseCashInStuck100Retry                          Case = "cash_in_stuck_100_retry"
	CaseCashInStuck100UpdateMismatch                 Case = "cash_in_stuck_100_update_mismatch"
	CasePcStuck201WaitingRppRepublishFromRpp         Case = "pc_stuck_201_waiting_rpp_republish_from_rpp"
)

// GetCaseSummaryOrder returns the order in which SOP cases should be displayed in summaries
func GetCaseSummaryOrder() []Case {
	return []Case{
		CasePcExternalPaymentFlow200_11,
		CasePcExternalPaymentFlow201_0RPP210,
		CasePcExternalPaymentFlow201_0RPP900,
		CasePeTransferPayment210_0,
		CasePeStuckAtLimitCheck102,
		CasePeStuck230RepublishPC,
		CaseThoughtMachineFalseNegative,
		CasePeCaptureProcessingPcCaptureFailedRppSuccess,
		CasePe2200FastCashinFailed,
		CaseRppCashoutReject101_19,
		CaseRppQrPaymentReject210_0,
		CaseRppNoResponseRejectNotFound,
		CaseRppNoResponseResume,
		CaseRppCashinValidationFailed122_0,
		CaseRppRtpCashinStuck200_0,
		CaseEcotxnChargeFailedCaptureFailedTMError,
		CasePeStuck300RppNotFound,
		CaseCashoutPe220Pc201Reject,
		CaseCashoutRpp210Pe220Pc201,
		CaseRpp210Pe220Pc201Accept,
		CaseRpp210Pe220Pc201Reject,
		CasePe220Pc201Rpp0StuckInit,
		CaseRppProcessRegistryStuckInit,
		CaseCashInStuck100Retry,
		CaseCashInStuck100UpdateMismatch,
		CasePcStuck201WaitingRppRepublishFromRpp,
	}
}

// SQLStatements contains the deploy and rollback SQL statements separated by database
type SQLStatements struct {
	PCDeployStatements    []string
	PCRollbackStatements  []string
	PEDeployStatements    []string
	PERollbackStatements  []string
	PPEDeployStatements   []string
	PPERollbackStatements []string
	RPPDeployStatements   []string
	RPPRollbackStatements []string
}

// ParamInfo represents a parameter with its value and type for SQL generation
type ParamInfo struct {
	Name  string      // Parameter name (e.g., "run_id", "prev_trans_id")
	Value interface{} // Parameter value (use interface{} for type flexibility)
	Type  string      // Parameter type: "string", "int"
}

type TemplateInfo struct {
	TargetDB    string // "PC", "PE", or "RPP"
	SQLTemplate string
	Params      []ParamInfo
}

// DMLTicket represents a SQL generation request with templates
type DMLTicket struct {
	Deploy   []TemplateInfo // SQL template with %s placeholders
	Rollback []TemplateInfo // SQL template with %s placeholders
	CaseType Case           // SOP case type for this ticket
}
