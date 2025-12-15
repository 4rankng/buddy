package adapters

// ChargeRecord represents a record from the charge table
type ChargeRecord struct {
	ID                      int     `json:"id"`
	TransactionID           string  `json:"transaction_id"`
	Amount                  float64 `json:"amount"`
	Status                  string  `json:"status"`
	Currency                string  `json:"currency"`
	PartnerID               string  `json:"partner_id"`
	CustomerID              string  `json:"customer_id"`
	ExternalID              string  `json:"external_id"`
	ReferenceID             string  `json:"reference_id"`
	TxnDomain               string  `json:"txn_domain"`
	TxnType                 string  `json:"txn_type"`
	TxnSubtype              string  `json:"txn_subtype"`
	Remarks                 string  `json:"remarks"`
	Metadata                string  `json:"metadata"`
	Properties              string  `json:"properties"`
	ValuedAt                string  `json:"valued_at"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
	BillingToken            string  `json:"billing_token"`
	StatusReason            string  `json:"status_reason"`
	CaptureMethod           string  `json:"capture_method"`
	SourceAccount           string  `json:"source_account"`
	DestinationAccount      string  `json:"destination_account"`
	CapturedAmount          float64 `json:"captured_amount"`
	TransactionPayLoad      string  `json:"transaction_payload"`
	StatusReasonDescription string  `json:"status_reason_description"`
}

// ChargeStorage represents the ChargeStorage object for workflow_execution
type ChargeStorage struct {
	ID                      int     `json:"ID"`
	Amount                  float64 `json:"Amount"`
	Status                  string  `json:"Status"`
	Remarks                 string  `json:"Remarks"`
	TxnType                 string  `json:"TxnType"`
	Currency                string  `json:"Currency"`
	Metadata                *string `json:"Metadata"`
	ValuedAt                string  `json:"ValuedAt"`
	CreatedAt               string  `json:"CreatedAt"`
	PartnerID               string  `json:"PartnerID"`
	TxnDomain               string  `json:"TxnDomain"`
	UpdatedAt               string  `json:"UpdatedAt"`
	CustomerID              string  `json:"CustomerID"`
	ExternalID              string  `json:"ExternalID"`
	Properties              *string `json:"Properties"`
	TxnSubtype              string  `json:"TxnSubtype"`
	ReferenceID             string  `json:"ReferenceID"`
	BillingToken            string  `json:"BillingToken"`
	StatusReason            string  `json:"StatusReason"`
	CaptureMethod           string  `json:"CaptureMethod"`
	SourceAccount           *string `json:"SourceAccount"`
	TransactionID           string  `json:"TransactionID"`
	CapturedAmount          float64 `json:"CapturedAmount"`
	DestinationAccount      *string `json:"DestinationAccount"`
	TransactionPayLoad      *string `json:"TransactionPayLoad"`
	StatusReasonDescription string  `json:"StatusReasonDescription"`
}

// ProcessResult holds the result of processing a single transaction
type ProcessResult struct {
	TransactionID string
	Success       bool
	ValueAt       string
	Error         error
}

// WorkflowExecution represents a record from the workflow_execution table
type WorkflowExecution struct {
	RunID   string `json:"run_id"`
	State   int    `json:"state"`
	Attempt int    `json:"attempt"`
	Data    string `json:"data"`
}
