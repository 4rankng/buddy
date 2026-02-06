package adapters

import "buddy/internal/txn/domain"

// CaseRule defines a rule for identifying SOP cases
type CaseRule struct {
	CaseType    domain.Case
	Description string
	Country     string // optional: "", "my", "sg" for country-specific rules
	Conditions  []RuleCondition
}

// RuleCondition defines a single condition in a rule
type RuleCondition struct {
	FieldPath string      // e.g., "PaymentEngine.Workflow.State"
	Operator  string      // eq, ne, lt, gt, in, not_in, regex, contains
	Value     interface{} // Expected value(s)
	Country   string      // optional: "", "my", "sg" for country-specific rules
}

// Helper constants for common field paths
const (
	pathPEWorkflowID         = "PaymentEngine.Workflow.WorkflowID"
	pathPEWorkflowState      = "PaymentEngine.Workflow.State"
	pathPEWorkflowAttempt    = "PaymentEngine.Workflow.Attempt"
	pathPETransfersExtID     = "PaymentEngine.Transfers.ExternalID"
	pathPCExtTransfersWfID   = "PaymentCore.ExternalTransfer.Workflow.WorkflowID"
	pathPCExtTransfersState  = "PaymentCore.ExternalTransfer.Workflow.State"
	pathPCExtTransfersAttempt = "PaymentCore.ExternalTransfer.Workflow.Attempt"
	pathPCIntAuthState       = "PaymentCore.InternalAuth.Workflow.State"
	pathPCIntCaptureWfID     = "PaymentCore.InternalCapture.Workflow.WorkflowID"
	pathPCIntCaptureState    = "PaymentCore.InternalCapture.Workflow.State"
	pathPCIntCaptureAttempt  = "PaymentCore.InternalCapture.Workflow.Attempt"
	pathPCIntCaptureTxType   = "PaymentCore.InternalCapture.TxType"
	pathPCIntCaptureTxStatus = "PaymentCore.InternalCapture.TxStatus"
	pathRPPAdapterWfID       = "RPPAdapter.Workflow.WorkflowID"
	pathRPPAdapterState      = "RPPAdapter.Workflow.State"
	pathRPPAdapterAttempt    = "RPPAdapter.Workflow.Attempt"
	pathRPPAdapterStatus     = "RPPAdapter.Status"
	pathPartnerpayWfID       = "PartnerpayEngine.Workflow.WorkflowID"
	pathPartnerpayState      = "PartnerpayEngine.Workflow.State"
	pathPartnerpayAttempt    = "PartnerpayEngine.Workflow.Attempt"
	pathFastAdapterStatus    = "FastAdapter.Status"
)

// Common state values
const (
	stateInit              = "0"
	stateSuccess           = "900"
	stateFailed            = "500"
	stateProcessing        = "201"
	stateTransferProcessing = "220"
	stateCaptureFailed      = "701"
	stateCaptureFailedCond  = "701"
	stateCaptureProcessing = "230"
	stateTransactionFailed = "500"
	stateRppWaitingResponse = "210"
	stateRppReject         = "101"
	stateRppReceived       = "200"
	stateValidationFailed  = "122"
)

// Common workflow IDs
const (
	wfTransferPayment = "workflow_transfer_payment"
	wfTransferCollection = "workflow_transfer_collection"
	wfExternalPayment = "external_payment_flow"
	wfInternalPayment = "internal_payment_flow"
	wfInternalCapture = "internal_payment_flow"
	wfCashout         = "wf_ct_cashout"
	wfQrPayment       = "wf_ct_qr_payment"
	wfCashin          = "wf_ct_cashin"
	wfRtpCashin       = "wf_ct_rtp_cashin"
	wfProcessRegistry = "wf_process_registry"
)

// Helper function to create an equality condition
func cond(fieldPath string, value interface{}) RuleCondition {
	return RuleCondition{FieldPath: fieldPath, Operator: "eq", Value: value}
}

// Helper function to create a not-equal condition
func condNe(fieldPath string, value interface{}) RuleCondition {
	return RuleCondition{FieldPath: fieldPath, Operator: "ne", Value: value}
}

// Helper function to create an "in" condition
func condIn(fieldPath string, values []string) RuleCondition {
	return RuleCondition{FieldPath: fieldPath, Operator: "in", Value: values}
}

// Helper function to create PE workflow conditions (state, attempt)
func peWorkflowConds(state string, attempt int) []RuleCondition {
	return []RuleCondition{
		cond(pathPEWorkflowState, state),
		cond(pathPEWorkflowAttempt, attempt),
	}
}

// Helper function to create PC external transfer conditions
func pcExtTransferConds(state string, attempt int) []RuleCondition {
	return []RuleCondition{
		cond(pathPCExtTransfersWfID, wfExternalPayment),
		cond(pathPCExtTransfersState, state),
		cond(pathPCExtTransfersAttempt, attempt),
	}
}

// Helper function to create RPP adapter conditions
func rppAdapterConds(workflowID, state string, attempt int) []RuleCondition {
	return []RuleCondition{
		cond(pathRPPAdapterWfID, workflowID),
		cond(pathRPPAdapterState, state),
		cond(pathRPPAdapterAttempt, attempt),
	}
}

// getDefaultSOPRules returns the default SOP case rules
func getDefaultSOPRules() []CaseRule {
	return []CaseRule{
		// 1. Complex Rules (PE + PC + RPP)
		// Most specific rule first: PC 201/0, PE 220/0, RPP wf_ct_cashout at 900/0
		{
			CaseType:    domain.CasePcStuck201WaitingRppRepublishFromRpp,
			Description: "PC stuck at 201/0, PE at 220/0, RPP wf_ct_cashout at 900/0 - republish RPP success message",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPCExtTransfersWfID, wfExternalPayment),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathPCExtTransfersAttempt, 0),
				cond(pathRPPAdapterWfID, wfCashout),
				cond(pathRPPAdapterState, stateSuccess),
				cond(pathRPPAdapterAttempt, 0),
			},
		},
		{
			CaseType:    domain.CasePcExternalPaymentFlow201_0RPP900,
			Description: "PC External Payment Flow 201/0 with RPP 900 (completed)",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPEWorkflowAttempt, 0),
				condNe(pathPETransfersExtID, ""),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathPCExtTransfersAttempt, 0),
				cond(pathRPPAdapterStatus, stateSuccess),
			},
		},
		{
			CaseType:    domain.CasePcExternalPaymentFlow201_0RPP210,
			Description: "PC External Payment Flow 201/0 with RPP not completed (stuck)",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPEWorkflowAttempt, 0),
				condNe(pathPETransfersExtID, ""),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathPCExtTransfersAttempt, 0),
				condNe(pathRPPAdapterStatus, stateSuccess),
			},
		},
		{
			CaseType:    domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess,
			Description: "PE capture processing, PC capture failed, but RPP succeeded",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowState, stateCaptureProcessing),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPCIntCaptureWfID, wfInternalCapture),
				cond(pathPCIntCaptureState, stateFailed),
				cond(pathPCIntCaptureAttempt, 0),
				cond(pathPCIntCaptureTxType, "CAPTURE"),
				cond(pathPCIntCaptureTxStatus, "FAILED"),
				condIn(pathRPPAdapterWfID, []string{wfQrPayment, wfCashout}),
				cond(pathRPPAdapterState, stateSuccess),
				cond(pathRPPAdapterAttempt, 0),
			},
		},
		{
			CaseType:    domain.CasePeStuck300RppNotFound,
			Description: "PE stuck at state 300 with auth success, no capture, no RPP",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowState, "300"),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPEWorkflowID, wfTransferPayment),
				cond("PaymentCore.InternalAuth.Workflow.State", stateSuccess),
				cond(pathPCIntCaptureWfID, ""), // Checks for empty InternalCapture.Workflow.WorkflowID
				cond(pathRPPAdapterWfID, ""),   // Checks for nil/empty RPPAdapter
			},
		},
		{
			CaseType:    domain.CaseCashoutPe220Pc201Reject,
			Description: "Cashout PE 220/0, PC 201/0, RPP PROCESSING - manual reject",
			Country:     "sg",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPCExtTransfersWfID, wfExternalPayment),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathPCExtTransfersAttempt, 0),
				cond(pathRPPAdapterStatus, "PROCESSING"),
			},
		},
		{
			CaseType:    domain.CaseCashoutRpp210Pe220Pc201,
			Description: "Cashout PE 220/0, PC 201/0, RPP process registry 0/0, RPP cashout 210/0 - manual intervention required",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPCExtTransfersWfID, wfExternalPayment),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathPCExtTransfersAttempt, 0),
				condIn(pathRPPAdapterWfID, []string{wfProcessRegistry, wfCashout, wfQrPayment}),
				condIn(pathRPPAdapterState, []string{stateInit, "210"}),
				cond(pathRPPAdapterAttempt, 0),
			},
		},
		{
			CaseType:    domain.CaseRpp210Pe220Pc201Accept,
			Description: "RPP 210, PE 220, PC 201 - Manual Accept",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathRPPAdapterState, "210"),
				condIn(pathRPPAdapterWfID, []string{wfCashout, wfQrPayment}),
			},
		},
		{
			CaseType:    domain.CaseRpp210Pe220Pc201Reject,
			Description: "RPP 210, PE 220, PC 201 - Manual Reject",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathRPPAdapterState, "210"),
				condIn(pathRPPAdapterWfID, []string{wfCashout, wfQrPayment}),
			},
		},
		{
			CaseType:    domain.CasePe220Pc201Rpp0StuckInit,
			Description: "PE 220/0, PC 201/0, RPP wf_ct_qr_payment stuck at State 0 - manual PE rejection required",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPCExtTransfersWfID, wfExternalPayment),
				cond(pathPCExtTransfersState, stateProcessing),
				cond(pathPCExtTransfersAttempt, 0),
				cond(pathRPPAdapterWfID, wfQrPayment),
				cond(pathRPPAdapterState, stateInit),
			},
		},

		// 2. Medium Complexity (PE + PC or Partnerpay + PC)
		{
			CaseType:    domain.CaseThoughtMachineFalseNegative,
			Description: "Thought Machine returning errors/false negatives, but transaction was successful",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowState, stateCaptureFailedCond),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPCIntCaptureWfID, wfInternalCapture),
				cond(pathPCIntCaptureState, stateFailed),
				cond(pathPCIntCaptureAttempt, 0),
			},
		},
		{
			CaseType:    domain.CaseEcotxnChargeFailedCaptureFailedTMError,
			Description: "Ecotxn Charge Failed Capture Failed with TMError",
			Conditions: []RuleCondition{
				cond(pathPartnerpayWfID, "workflow_charge"),
				cond(pathPartnerpayState, "502"),
				cond(pathPartnerpayAttempt, 0),
				cond("PartnerpayEngine.Charge.StatusReason", "SYSTEM_ERROR"),
				cond("PartnerpayEngine.Charge.StatusReasonDescription", "error occurred in Thought Machine."),
				cond(pathPCIntCaptureWfID, wfInternalCapture),
				cond(pathPCIntCaptureState, stateFailed),
				cond(pathPCIntCaptureAttempt, 0),
			},
		},
		{
			CaseType:    domain.CasePeStuck230RepublishPC,
			Description: "PE stuck at state 230 (capture) requires PC republish",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowState, stateCaptureProcessing),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathPCIntCaptureState, stateSuccess),
				cond(pathPCIntCaptureAttempt, 0),
			},
		},
		{
			CaseType:    domain.CasePe2200FastCashinFailed,
			Description: "PE Transfer Collection at state 220 with attempt 0 and Fast Adapter failed",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowID, wfTransferCollection),
				cond(pathPEWorkflowState, stateTransferProcessing),
				cond(pathPEWorkflowAttempt, 0),
				cond(pathFastAdapterStatus, "FAILED"),
			},
		},

		// 3. Simple Rules (Single Domain)
		{
			CaseType:    domain.CasePeTransferPayment210_0,
			Description: "PE Transfer Payment stuck at state 210 with attempt 0",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowState, "210"),
				cond(pathPEWorkflowID, wfTransferPayment),
				cond(pathPEWorkflowAttempt, 0),
			},
		},
		{
			CaseType:    domain.CasePcExternalPaymentFlow200_11,
			Description: "PC External Payment Flow stuck at state 200 with attempt 11",
			Conditions: []RuleCondition{
				cond(pathPCExtTransfersState, "200"),
				cond(pathPCExtTransfersAttempt, 11),
			},
		},
		{
			CaseType:    domain.CasePeStuckAtLimitCheck102,
			Description: "PE stuck at state 102 (stTransactionLimitChecked)",
			Conditions: []RuleCondition{
				cond(pathPEWorkflowState, "102"),
				cond(pathPEWorkflowID, wfTransferPayment),
			},
		},
		{
			CaseType:    domain.CaseRppNoResponseResume,
			Description: "RPP No Response Resume (timeout scenario)",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathRPPAdapterState, stateRppWaitingResponse),
				cond(pathRPPAdapterAttempt, 0),
				condIn(pathRPPAdapterWfID, []string{wfCashout, wfQrPayment}),
			},
		},
		{
			CaseType:    domain.CaseRppCashoutReject101_19,
			Description: "RPP Cashout Reject at state 101 with attempt 19",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathRPPAdapterWfID, wfCashout),
				cond(pathRPPAdapterState, stateRppReject),
				cond(pathRPPAdapterAttempt, 19),
			},
		},
		{
			CaseType:    domain.CaseRppRtpCashinStuck200_0,
			Description: "RPP RTP Cashin stuck at state 200 with attempt 0",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathRPPAdapterWfID, wfRtpCashin),
				cond(pathRPPAdapterState, stateRppReceived),
				cond(pathRPPAdapterAttempt, 0),
			},
		},
		{
			CaseType:    domain.CaseRppCashinValidationFailed122_0,
			Description: "RPP Cashin Validation Failed at state 122 with attempt 0",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathRPPAdapterWfID, wfCashin),
				cond(pathRPPAdapterState, stateValidationFailed),
				cond(pathRPPAdapterAttempt, 0),
			},
		},
		{
			CaseType:    domain.CaseRppProcessRegistryStuckInit,
			Description: "RPP Process Registry stuck at state 0 (stInit)",
			Country:     "my",
			Conditions: []RuleCondition{
				cond(pathRPPAdapterWfID, wfProcessRegistry),
				cond(pathRPPAdapterState, stateInit),
			},
		},
	}
}