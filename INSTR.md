frank.nguyen@DBSG-H4M0DVF2C7 buddy % mybuddy  ecotxn fd230a01dcd04282851b7b9dd6260c93
Initialize Doorman client for [my]...
Initialize Jira client for [my]...
### [1] transaction_id: fd230a01dcd04282851b7b9dd6260c93
[partnerpay-engine]
charge.status: FAILED SYSTEM_ERROR error occurred in Thought Machine.
workflow_charge: stFailureNotified(502) Attempt=0 run_id=fd230a01dcd04282851b7b9dd6260c93


should also fill the payment-core with

SELECT * FROM internal_transaction
where group_id='fd230a01dcd04282851b7b9dd6260c93'
and created_at >= '2025-08-09T14:14:03.379002Z' // created_at - 30min
and created_at <= '2025-08-09T16:14:03.379002Z' // created_at + 30min

SELECT * FROM external_transaction
where group_id='fd230a01dcd04282851b7b9dd6260c93'
and created_at >= '2025-08-09T14:14:03.379002Z' // created_at - 30min
and created_at <= '2025-08-09T16:14:03.379002Z' // created_at + 30min

remember to display all these field in stdout and output text file if applicable
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
	InternalCapture  PCInternalInfo // when TxType is AUTH
	InternalAuth     PCInternalInfo // when TxType is CAPTURE
	ExternalTransfer PCExternalInfo // when TxType is TRANSFER
}

regardless of single txn or multiple txn, the output / deploy / rollback dml should be file not stdout . stdout is for summary only

WriteEcoTransactionInfo
should provide all available information from TransactionResult similar to the output of mybuddy txn abc
