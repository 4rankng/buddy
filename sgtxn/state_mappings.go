package sgtxn

// Workflow state mappings from PROMPT

// PaymentEngineWorkflowStates maps workflow_transfer_payment states
var PaymentEngineWorkflowStates = map[int]string{
	100: "stTransferPersisted",
	101: "stProcessingPublished",
	102: "stPreTransactionLimitCheck",
	103: "stPreRiskCheck",
	210: "stTransferProcessing",
	211: "stTransferStreamPersisted",
	220: "stAuthProcessing",
	221: "stAuthStreamPersisted",
	300: "stAuthSuccess",
	230: "stCapturePrepared",
	231: "stCaptureProcessing",
	232: "stCaptureStreamPersisted",
	240: "stCancelPrepared",
	241: "stCancelProcessing",
	242: "stCancelStreamPersisted",
	250: "stResumePrepared",
	501: "stPrepareFailureHandling",
	502: "stTransactionLimitCheckFailed",
	503: "stRiskCheckError",
	504: "stRiskCheckDeny",
	505: "stFailurePublished",
	510: "stFailureNotified",
	600: "stCanceled",
	610: "stCanceledPublished",
	701: "stCaptureFailed",
	702: "stCancelFailed",
	721: "stInvestigationRequiredPublished",
	722: "stInvestigationRequiredNotified",
	800: "stValidateSuccess",
	900: "stTransferCompleted",
	901: "stCaptureCompleted",
	902: "stTransferCompletedAutoPublish",
	905: "stCompletedPublished",
	910: "stCompletedNotified",
}

// PaymentCoreWorkflowStates maps payment core workflow states
var PaymentCoreWorkflowStates = map[int]string{
	200: "stSubmitted",
	900: "stPrepareUpdateAuth",
}

// FastAdapterCashinStates maps fast-adapter cashin status
var FastAdapterCashinStates = map[int]string{
	1:  "StAuthed",
	2:  "StAuthUnknown",
	3:  "StConfirmUnknown",
	4:  "StCancelUnknown",
	5:  "StErraneous",
	6:  "StConfirmed",
	7:  "StCanceled",
	8:  "StCanceledManual",
	9:  "StRejected",
	11: "StReversed",
	12: "StReversedUnknown",
	13: "StReversedManual",
	15: "StCancelAcknowledged",
	17: "StAuthedNoStatus",
	18: "StRejectedNoStatus",
}

// FastAdapterCashoutStates maps fast-adapter cashout status
var FastAdapterCashoutStates = map[int]string{
	1: "StPending",
	2: "StErraneous",
	3: "StAccepted",
	4: "StRejected",
}

// GetWorkflowStateName returns the state name for payment engine workflow
func GetPaymentEngineWorkflowStateName(state int) string {
	if name, ok := PaymentEngineWorkflowStates[state]; ok {
		return name
	}
	return "unknown"
}

// GetPaymentCoreWorkflowStateName returns the state name for payment core workflow
func GetPaymentCoreWorkflowStateName(state int) string {
	if name, ok := PaymentCoreWorkflowStates[state]; ok {
		return name
	}
	return "unknown"
}

// GetFastAdapterCashinStateName returns the state name for fast-adapter cashin
func GetFastAdapterCashinStateName(state int) string {
	if name, ok := FastAdapterCashinStates[state]; ok {
		return name
	}
	return "unknown"
}

// GetFastAdapterCashoutStateName returns the state name for fast-adapter cashout
func GetFastAdapterCashoutStateName(state int) string {
	if name, ok := FastAdapterCashoutStates[state]; ok {
		return name
	}
	return "unknown"
}
