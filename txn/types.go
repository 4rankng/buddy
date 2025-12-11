package txn

// WorkflowInfo contains information about a specific workflow execution
type WorkflowInfo struct {
	Type    string // workflow_charge, internal_transaction etc. it is workflow_execution.workflow_id
	RunID   string
	State   string
	Attempt int
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

// Workflow state mappings
var workflowStateMap = map[int]string{
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
}

// Internal payment flow state mappings
var internalPaymentFlowMap = map[int]string{
	100: "stPending",
	101: "stStreamPersisted",
	901: "stPrepareUpdateAuth",
	902: "stPrepareSuccessPublish",
	900: "stSuccess",
	501: "stPrepareFailurePublish",
	500: "stFailed",
}

// External payment flow state mappings
var externalPaymentFlowMap = map[int]string{
	200: "stSubmitted",
	201: "stProcessing",
	202: "stRespReceived",
	901: "stPrepareSuccessPublish",
	900: "stSuccess",
	501: "stPrepareFailurePublish",
	500: "stFailed",
}

// SOPCase represents the supported remediation cases from SOP.md
type SOPCase string

const (
	SOPCaseNone                             SOPCase = ""
	SOPCasePcExternalPaymentFlow200_11      SOPCase = "pc_external_payment_flow_200_11"
	SOPCasePcExternalPaymentFlow201_0RPP210 SOPCase = "pc_external_payment_flow_201_0_RPP_210"
	SOPCasePcExternalPaymentFlow201_0RPP900 SOPCase = "pc_external_payment_flow_201_0_RPP_900"
	SOPCasePeTransferPayment210_0           SOPCase = "pe_transfer_payment_210_0"
)
