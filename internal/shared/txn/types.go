package txn

import (
	"fmt"
	"strconv"
)

// PETransfersInfo contains payment-engine transfer information
type PETransfersInfo struct {
	Type          string // payment-engine transfers.type
	TxnSubtype    string // payment-engine transfers.txn_subtype
	TxnDomain     string // payment-engine transfers.txn_domain
	TransactionID string // payment-engine transfers.transaction_id
	ReferenceID   string // payment-engine transfers.reference_id
	Status        string // payment-engine transfers.status
	ExternalID    string // payment-engine transfers.external_id
	CreatedAt     string // payment-engine transfers.created_at
}

// WorkflowInfo contains information about a specific workflow execution
type WorkflowInfo struct {
	WorkflowID string // workflow_execution.workflow_id
	Attempt    int    // workflow_execution.attempt
	State      string // workflow_execution.state
	RunID      string // workflow_execution.run_id
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
type PCInternalTxnInfo struct {
	TxID     string // payment-core internal transaction ID
	GroupID  string // payment-core transaction group ID
	TxType   string // payment-core transaction type
	TxStatus string // payment-core transaction status
}

// PCExternalTxnInfo contains payment-core external transaction information
type PCExternalTxnInfo struct {
	RefID    string // payment-core external transaction reference ID
	GroupID  string // payment-core transaction group ID
	TxType   string // payment-core transaction type
	TxStatus string // payment-core transaction status
}

// PaymentCoreInfo contains payment-core related information
type PaymentCoreInfo struct {
	InternalTxns []PCInternalTxnInfo
	ExternalTxns []PCExternalTxnInfo
	Workflow     []WorkflowInfo
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

// RPPAdapterInfo contains RPP adapter related information
type RPPAdapterInfo struct {
	ReqBizMsgID string // RPP request business message ID
	PartnerTxID string // RPP partner transaction ID
	Status      string // RPP status
	Workflow    WorkflowInfo
	Info        string // optional extra context (e.g. status reason description)
}

// PPEChargeInfo contains partnerpay-engine charge information
type PPEChargeInfo struct {
	TransactionID           string // partnerpay-engine charge.transaction_id
	Status                  string // partnerpay-engine charge.status
	StatusReason            string // partnerpay-engine charge.status_reason
	StatusReasonDescription string // partnerpay-engine charge.status_reason_description
}

// PartnerpayEngineInfo contains partnerpay-engine related information
type PartnerpayEngineInfo struct {
	Transfers PPEChargeInfo
	Workflow  WorkflowInfo
}

// TransactionResult represents the result of a transaction query
type TransactionResult struct {
	TransactionID    string
	PaymentEngine    PaymentEngineInfo
	PartnerpayEngine PartnerpayEngineInfo
	PaymentCore      PaymentCoreInfo
	FastAdapter      FastAdapterInfo
	RPPAdapter       RPPAdapterInfo
	CaseType         SOPCase // Store the identified SOP case to avoid re-identification
	Error            string
}

// Common status values
const NotFoundStatus = "NOT_FOUND"

// SOPCase represents the supported remediation cases from SOP.md
type SOPCase string

const (
	SOPCaseNone                             SOPCase = NotFoundStatus
	SOPCasePcExternalPaymentFlow200_11      SOPCase = "pc_external_payment_flow_200_11"
	SOPCasePcExternalPaymentFlow201_0RPP210 SOPCase = "pc_external_payment_flow_201_0_RPP_210"
	SOPCasePcExternalPaymentFlow201_0RPP900 SOPCase = "pc_external_payment_flow_201_0_RPP_900"
	SOPCasePeTransferPayment210_0           SOPCase = "pe_transfer_payment_210_0"
	SOPCaseRppCashoutReject101_19           SOPCase = "rpp_cashout_reject_101_19"
	SOPCaseRppQrPaymentReject210_0          SOPCase = "rpp_qr_payment_reject_210_0"
	SOPCaseRppNoResponseResume              SOPCase = "rpp_no_response_resume"
)

// WorkflowStateMaps contains state mappings for different workflow types
var WorkflowStateMaps = map[string]map[int]string{
	"workflow_transfer_payment": {
		100: "stTransferPersisted",
		101: "stProcessingPublished",
		102: "stTransactionLimitChecked",
		103: "stRedeemRewardRequired",
		104: "stResolveFeeRequired",
		105: "stFreeTransferFeeRewardRedeemed",
		106: "stFeeResolved",
		210: "stAuthProcessing",
		211: "stAuthStreamPersisted",
		300: "stAuthCompleted",
		220: "stTransferProcessing",
		221: "stTransferStreamPersisted",
		223: "stTransferCompleted",
		230: "stCaptureProcessing",
		231: "stCaptureStreamPersisted",
		235: "stCapturePrepared",
		400: "stTransferFailed",
		410: "stAutoCancelProcessing",
		412: "stAutoCancelStreamPersisted",
		501: "stPrepareFailureHandling",
		505: "stFailurePublished",
		510: "stFailureNotified",
		511: "stRewardRedeemVoidRequired",
		512: "stRewardRedeemVoided",
		701: "stCaptureFailed",
		702: "stCancelFailed",
		703: "stRewardRedeemInterventionRequired",
		900: "stCaptureCompleted",
		905: "stCompletedPublished",
		910: "stCompletedNotified",
		911: "stRewardRedeemCompletionRequired",
		912: "stRewardRedeemCompleted",
	},
	"internal_payment_flow": {
		100: "stPending",
		101: "stStreamPersisted",
		900: "stSuccess",
		901: "stPrepareUpdateAuth",
		902: "stPrepareSuccessPublish",
		500: "stFailed",
		501: "stPrepareFailurePublish",
	},
	"external_payment_flow": {
		200: "stSubmitted",
		201: "stProcessing",
		202: "stRespReceived",
		900: "stPrepareSuccessPublish",
		901: "stSuccess",
		500: "stFailed",
		501: "stPrepareFailurePublish",
	},
	"wf_ct_cashout": {
		101: "stCreditTransferPersisted",
		111: "stCreditorDetailUpdated",
		120: "stPrepareCreditorInquiry",
		121: "stCreditorInquiryFailed",
		122: "stCreditorInquirySuccess",
		210: "stTransferProcessing",
		211: "stTransferResponseReceived",
		212: "stTransferMessageRejectedReceived",
		221: "stTransferManualRejectedReceived",
		222: "stTransferManualResumeReceived",
		301: "stPrepareSuccessPublish",
		311: "stPrepareFailurePublish",
		321: "stSkipPublish",
		502: "stTransferRetryAttemptExceeded",
		700: "stFailed",
		900: "stSuccess",
	},
	"wf_ct_cashin": {
		100: "stTransferPersisted",
		110: "stRequestToPayUpdated",
		111: "stRequestToPayUpdateFailed",
		121: "stOriginalTransferValidated",
		122: "stFieldsValidationFailed",
		200: "stTransferProcessing",
		201: "stTransferStreamPersisted",
		210: "stTransferUpdated",
		220: "stTransferRespPrepared",
		700: "stCashInFailed",
		701: "stCashInToRefund",
		900: "stCashInCompleted",
		901: "stCashInCompletedWithRefund",
	},
	"workflow_charge": {
		100: "stProcessingPublished",
		101: "stPrepareAuth",
		102: "stPrepareTransfer",
		200: "stAuthProcessing",
		201: "stTriggerInAppAuth",
		210: "stTransferProcessing",
		300: "stAuthCompleted",
		301: "stPendingInAppAuth",
		302: "stInAppAuthStreamPersisted",
		310: "stPublishAuth",
		400: "stCaptureProcessing",
		600: "stCancelProcessing",
		700: "stCancelCompleted",
		777: "stCancelPublished",
		800: "stTransferCompleted",
		888: "stTransferPublished",
		871: "stTransferPublishedDirectNotify",
		890: "stNotified",
		900: "stPrepareFailureHandling",
		999: "stFailurePublished",
		500: "stCaptureFailed",
		501: "stCancelFailed",
		502: "stFailureNotified",
	},
}

// GetWorkflowStateName returns the human-readable state name for a given workflow type and state number
func GetWorkflowStateName(workflowType string, state int) string {
	// Determine which map to use based on workflow type
	var mapKey string
	switch workflowType {
	case "workflow_transfer_payment":
		mapKey = "workflow_transfer_payment"
	case "internal_payment_flow":
		mapKey = "internal_payment_flow"
	case "external_payment_flow":
		mapKey = "external_payment_flow"
	case "wf_ct_cashout":
		mapKey = "wf_ct_cashout"
	case "workflow_wf_ct_cashin", "wf_ct_cashin":
		mapKey = "wf_ct_cashin"
	case "workflow_charge":
		mapKey = "workflow_charge"
	default:
		mapKey = "workflow_transfer_payment"
	}

	// Get the appropriate state map
	if stateMap, exists := WorkflowStateMaps[mapKey]; exists {
		if stateName, exists := stateMap[state]; exists {
			return stateName
		}
	}

	// Fallback: return state as string if not found in map
	return fmt.Sprintf("stUnknown_%d", state)
}

// FormatWorkflowState formats a workflow state for display as "stateName(stateNumber)"
func FormatWorkflowState(workflowType string, stateStr string) string {
	if stateNum, err := strconv.Atoi(stateStr); err == nil {
		stateName := GetWorkflowStateName(workflowType, stateNum)
		return fmt.Sprintf("%s(%d)", stateName, stateNum)
	}
	// Return as-is if it's not a number
	return stateStr
}

// SQLStatements contains the deploy and rollback SQL statements separated by database
type SQLStatements struct {
	PCDeployStatements    []string
	PCRollbackStatements  []string
	PEDeployStatements    []string
	PERollbackStatements  []string
	RPPDeployStatements   []string
	RPPRollbackStatements []string
}

// DMLTicket represents a SQL generation request with templates
type DMLTicket struct {
	RunIDs           []string // run_ids to update
	ReqBizMsgIDs     []string // optional req_biz_msg_ids for RPP cases
	PartnerTxIDs     []string // optional partner_tx_ids for RPP cases
	DeployTemplate   string   // SQL template for deploy
	RollbackTemplate string   // SQL template for rollback
	TargetDB         string   // "PC", "PE", or "RPP"
	WorkflowID       string   // optional workflow_id filter
	TargetState      int      // target state to check in WHERE clause
	TargetAttempt    int      // target attempt to check in WHERE clause
	StateField       string   // field name for state in WHERE clause (usually "state")
	WorkflowIDs      []string // multiple workflow_ids for IN clause

	// Consolidation metadata
	TransactionCount int // Number of transactions consolidated
}

// TemplateConfig defines the parameters required for a SQL template
type TemplateConfig struct {
	Parameters []string // List of parameter types: ["run_ids"], ["run_ids", "workflow_ids"]
}
