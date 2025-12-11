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

// SOPCase represents the supported remediation cases from SOP.md
type SOPCase string

const (
	SOPCaseNone                             SOPCase = ""
	SOPCasePcExternalPaymentFlow200_11      SOPCase = "pc_external_payment_flow_200_11"
	SOPCasePcExternalPaymentFlow201_0RPP210 SOPCase = "pc_external_payment_flow_201_0_RPP_210"
	SOPCasePcExternalPaymentFlow201_0RPP900 SOPCase = "pc_external_payment_flow_201_0_RPP_900"
	SOPCasePeTransferPayment210_0           SOPCase = "pe_transfer_payment_210_0"
	SOPCaseRppCashoutReject101_19           SOPCase = "rpp_cashout_reject_101_19"
)
