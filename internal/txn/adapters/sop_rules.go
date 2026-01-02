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

// getDefaultSOPRules returns the default SOP case rules
func getDefaultSOPRules() []CaseRule {
	return []CaseRule{
		{
			CaseType:    domain.CasePcExternalPaymentFlow200_11,
			Description: "PC External Payment Flow stuck at state 200 with attempt 11",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.State",
					Operator:  "eq",
					Value:     "200",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.Attempt",
					Operator:  "eq",
					Value:     11,
				},
			},
		},
		{
			CaseType:    domain.CasePeTransferPayment210_0,
			Description: "PE Transfer Payment stuck at state 210 with attempt 0",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "210",
				},
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePeStuckAtLimitCheck102_4,
			Description: "PE stuck at state 102 (stTransactionLimitChecked) with attempt 4",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "102",
				},
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     4,
				},
			},
		},
		{
			CaseType:    domain.CasePeStuck230RepublishPC,
			Description: "PE stuck at state 230 (capture) requires PC republish",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "230",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "900",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePe2200FastCashinFailed,
			Description: "PE Transfer Collection at state 220 with attempt 0 and Fast Adapter failed",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_collection",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "220",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "FastAdapter.Status",
					Operator:  "eq",
					Value:     "FAILED",
				},
			},
		},
		{
			CaseType:    domain.CaseRppCashoutReject101_19,
			Description: "RPP Cashout Reject at state 101 with attempt 19",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "wf_ct_cashout",
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "101",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     19,
				},
			},
		},
		{
			CaseType:    domain.CaseRppQrPaymentReject210_0,
			Description: "RPP QR Payment Reject at state 210 with attempt 0",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "wf_ct_qr_payment",
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "210",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CaseRppNoResponseResume,
			Description: "RPP No Response Resume (timeout scenario)",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "210",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "in",
					Value:     []string{"wf_ct_cashout", "wf_ct_qr_payment"},
				},
			},
		},
		{
			CaseType:    domain.CaseRppRtpCashinStuck200_0,
			Description: "RPP RTP Cashin stuck at state 200 with attempt 0",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "wf_ct_rtp_cashin",
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "200",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CaseRppCashinValidationFailed122_0,
			Description: "RPP Cashin Validation Failed at state 122 with attempt 0",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "wf_ct_cashin",
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "122",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePcExternalPaymentFlow201_0RPP900,
			Description: "PC External Payment Flow 201/0 with RPP 900 (completed)",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "220",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentEngine.Transfers.ExternalID",
					Operator:  "ne",
					Value:     "",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.State",
					Operator:  "eq",
					Value:     "201",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "RPPAdapter.Status",
					Operator:  "eq",
					Value:     "900",
				},
			},
		},
		{
			CaseType:    domain.CasePcExternalPaymentFlow201_0RPP210,
			Description: "PC External Payment Flow 201/0 with RPP not completed (stuck)",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "220",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentEngine.Transfers.ExternalID",
					Operator:  "ne",
					Value:     "",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.State",
					Operator:  "eq",
					Value:     "201",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "RPPAdapter.Status",
					Operator:  "ne",
					Value:     "900",
				},
			},
		},
		{
			CaseType:    domain.CaseThoughtMachineFalseNegative,
			Description: "Thought Machine returning errors/false negatives, but transaction was successful",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "701", // stCaptureFailed
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "500", // stFailed
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess,
			Description: "PE capture processing, PC capture failed, but RPP succeeded",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "230", // stCaptureProcessing
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "500", // stFailed
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalCapture.TxType",
					Operator:  "eq",
					Value:     "CAPTURE",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.TxStatus",
					Operator:  "eq",
					Value:     "FAILED",
				},
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "in",
					Value:     []string{"wf_ct_qr_payment", "wf_ct_cashout"},
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "900", // stSuccess
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CaseEcotxnChargeFailedCaptureFailedTMError,
			Description: "Ecotxn Charge Failed Capture Failed with TMError",
			Conditions: []RuleCondition{
				{
					FieldPath: "PartnerpayEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_charge",
				},
				{
					FieldPath: "PartnerpayEngine.Workflow.State",
					Operator:  "eq",
					Value:     "502",
				},
				{
					FieldPath: "PartnerpayEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PartnerpayEngine.Charge.StatusReason",
					Operator:  "eq",
					Value:     "SYSTEM_ERROR",
				},
				{
					FieldPath: "PartnerpayEngine.Charge.StatusReasonDescription",
					Operator:  "eq",
					Value:     "error occurred in Thought Machine.",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "500",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePeStuck300RppNotFound,
			Description: "PE stuck at state 300 with auth success, no capture, no RPP",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "300",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentCore.InternalAuth.Workflow.State",
					Operator:  "eq",
					Value:     "900",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "", // Checks for empty InternalCapture.Workflow.WorkflowID
				},
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "", // Checks for nil/empty RPPAdapter
				},
			},
		},
	}
}
