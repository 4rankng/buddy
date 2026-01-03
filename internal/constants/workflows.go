package constants

// Workflow IDs
const (
	WorkflowTransferPayment = "workflow_transfer_payment"
	WorkflowInternalPayment = "internal_payment_flow"
	WorkflowExternalPayment = "external_payment_flow"
	WorkflowCashIn          = "wf_ct_cashin"
	WorkflowCashOut         = "wf_ct_cashout"
)

// Database Names
const (
	DBPaymentEngine    = "payment_engine"
	DBPaymentCore      = "payment_core"
	DBFastAdapter      = "fast_adapter"
	DBRPPAdapter       = "rpp_adapter"
	DBPartnerpayEngine = "partnerpay_engine"
)

// Transaction Types
const (
	TxTypeAuth     = "AUTH"
	TxTypeCapture  = "CAPTURE"
	TxTypeTransfer = "TRANSFER"
)

// Transaction Statuses
const (
	StatusCompleted = "Completed"
	StatusClosed    = "Closed"
	StatusDone      = "Done"
	StatusResolved  = "Resolved"
	StatusFailed    = "Failed"
	StatusPending   = "Pending"
)

// Fast Adapter Types
const (
	FastTypeCashIn  = "cashin"
	FastTypeCashOut = "cashout"
)

// Fast Adapter Statuses
const (
	FastStatusSuccess = "Success"
	FastStatusFailed  = "Failed"
	FastStatusPending = "Pending"
)

// Environment Names
const (
	EnvMalaysia  = "my"
	EnvSingapore = "sg"
)

// JIRA Project Keys
const (
	JiraProjectDefault = "BUDDY"
)

// SQL Template Placeholders
const (
	PlaceholderMissing = "!MISSING"
)

// Error Messages
const (
	ErrMsgHeaderNotFound    = "header row not found in CSV"
	ErrMsgNoCloseTransition = "no close transition found"
	ErrMsgInvalidDomain     = "invalid JIRA domain"
	ErrMsgTransitionFailed  = "transition failed"
	ErrMsgAttachmentFailed  = "failed to download attachment"
)
