package txn

import (
	"fmt"
	"strconv"
)

// WorkflowInfo contains information about a specific workflow execution
type WorkflowInfo struct {
	Type    string // workflow_charge, internal_transaction etc. it is workflow_execution.workflow_id
	RunID   string
	State   string
	Attempt int
}

// GetFormattedState returns the formatted state with name and number
func (w *WorkflowInfo) GetFormattedState() string {
	return FormatWorkflowState(w.Type, w.State)
}

// TransactionResult represents the result of a transaction query
type TransactionResult struct {
	TransactionID  string // payment-engine transfers.transaction_id
	TransferStatus string // payment-engine transfers.status
	CreatedAt      string // payment-engine transfers.created_at
	Error          string

	// Workflow information from different systems
	PaymentEngineWorkflow WorkflowInfo
	PaymentCoreWorkflows  []WorkflowInfo
	RPPWorkflow           WorkflowInfo

	// Payment-core transaction statuses
	InternalTxStatus string
	ExternalTxStatus string

	// RPP / Interactive Data
	ReqBizMsgID string
	PartnerTxID string
	RPPStatus   string
	RPPInfo     string
}

// SOPCase represents the supported remediation cases from SOP.md
type SOPCase string

const (
	SOPCaseNone                             SOPCase = ""
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
		900: "stPrepareUpdateAuth",
		901: "stPrepareSuccessPublish",
		902: "stSuccess",
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
	Parameters []string // List of parameter types: [\"run_ids\"], [\"run_ids\", \"workflow_ids\"]
}
