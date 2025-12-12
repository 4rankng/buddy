package sgtxn

// SGTransactionResult represents the aggregated result from Singapore databases
type SGTransactionResult struct {
	TransactionID  string
	TransferStatus string
	CreatedAt      string
	ExternalID     string
	ReferenceID    string
	Error          string

	// Payment Engine workflow
	PaymentEngineWorkflow SGWorkflowInfo

	// Payment Core data
	InternalTxStatuses   []string
	ExternalTxStatuses   []string
	PaymentCoreWorkflows []SGWorkflowInfo

	// Fast Adapter data
	FastAdapterType       string
	FastAdapterStatus     string
	FastAdapterCancelCode string
	FastAdapterRejectCode string
}

// SGWorkflowInfo represents workflow execution information
type SGWorkflowInfo struct {
	Type    string
	RunID   string
	State   string
	Attempt int
}
