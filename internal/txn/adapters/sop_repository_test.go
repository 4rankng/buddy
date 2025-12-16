package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"testing"
)

// TestEcotxnChargeFailedCaptureFailedTMError tests the specific case that's failing
func TestEcotxnChargeFailedCaptureFailedTMError(t *testing.T) {
	// Create the exact transaction data from the provided example
	transactionResult := &domain.TransactionResult{
		InputID:  "fd230a01dcd04282851b7b9dd6260c93",
		CaseType: domain.CaseNone, // Initialize to CaseNone
		PartnerpayEngine: &domain.PartnerpayEngineInfo{
			Charge: domain.PPEChargeInfo{
				Status:                  "FAILED",
				StatusReason:            "SYSTEM_ERROR",
				StatusReasonDescription: "error occurred in Thought Machine.",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_charge",
				State:      "502",
				Attempt:    0,
				RunID:      "fd230a01dcd04282851b7b9dd6260c93",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalCapture: domain.PCInternalInfo{
				TxID:      "3550ca0d10df4b0ab2dce80218cdf51f",
				GroupID:   "fd230a01dcd04282851b7b9dd6260c93",
				TxType:    "CAPTURE",
				TxStatus:  "FAILED",
				ErrorCode: "SYSTEM_ERROR",
				ErrorMsg:  "error occurred in Thought Machine.",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "500",
					Attempt:    0,
					RunID:      "3550ca0d10df4b0ab2dce80218cdf51f",
				},
			},
			InternalAuth: domain.PCInternalInfo{
				TxID:      "ce8c05866d134bb488038644c708740e",
				GroupID:   "fd230a01dcd04282851b7b9dd6260c93",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "ce8c05866d134bb488038644c708740e",
				},
			},
		},
	}

	// Create a new SOP repository
	sopRepo := NewSOPRepository()

	// Test case identification
	t.Run("WithoutDebug", func(t *testing.T) {
		result := sopRepo.IdentifyCase(transactionResult, "my")
		t.Logf("Identified case: %s", result)

		if result != domain.CaseEcotxnChargeFailedCaptureFailedTMError {
			t.Errorf("Expected case %s, got %s", domain.CaseEcotxnChargeFailedCaptureFailedTMError, result)
		}
	})
}

// TestIndividualConditions tests each condition individually to identify the failing one
func TestIndividualConditions(t *testing.T) {
	// Create the exact transaction data from the provided example
	transactionResult := &domain.TransactionResult{
		InputID:  "fd230a01dcd04282851b7b9dd6260c93",
		CaseType: domain.CaseNone, // Initialize to CaseNone
		PartnerpayEngine: &domain.PartnerpayEngineInfo{
			Charge: domain.PPEChargeInfo{
				Status:                  "FAILED",
				StatusReason:            "SYSTEM_ERROR",
				StatusReasonDescription: "error occurred in Thought Machine.",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_charge",
				State:      "502",
				Attempt:    0,
				RunID:      "fd230a01dcd04282851b7b9dd6260c93",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalCapture: domain.PCInternalInfo{
				TxID:      "3550ca0d10df4b0ab2dce80218cdf51f",
				GroupID:   "fd230a01dcd04282851b7b9dd6260c93",
				TxType:    "CAPTURE",
				TxStatus:  "FAILED",
				ErrorCode: "SYSTEM_ERROR",
				ErrorMsg:  "error occurred in Thought Machine.",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "500",
					Attempt:    0,
					RunID:      "3550ca0d10df4b0ab2dce80218cdf51f",
				},
			},
			InternalAuth: domain.PCInternalInfo{
				TxID:      "ce8c05866d134bb488038644c708740e",
				GroupID:   "fd230a01dcd04282851b7b9dd6260c93",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "ce8c05866d134bb488038644c708740e",
				},
			},
		},
	}

	// Get the rule we're testing
	var targetRule *CaseRule
	for _, rule := range getDefaultSOPRules() {
		if rule.CaseType == domain.CaseEcotxnChargeFailedCaptureFailedTMError {
			targetRule = &rule
			break
		}
	}

	if targetRule == nil {
		t.Fatal("Target rule not found")
	}

	sopRepo := NewSOPRepository()

	// Test each condition individually
	for i, condition := range targetRule.Conditions {
		t.Run(fmt.Sprintf("Condition_%d_%s", i, condition.FieldPath), func(t *testing.T) {
			fieldValue, ok := sopRepo.getFieldValue(condition.FieldPath, transactionResult)
			t.Logf("Field: %s", condition.FieldPath)
			t.Logf("  Expected: %v", condition.Value)
			t.Logf("  Actual: %v", fieldValue)
			t.Logf("  Found: %v", ok)

			if !ok {
				t.Errorf("Field not found: %s", condition.FieldPath)
				return
			}

			result := sopRepo.evaluateCondition(condition, transactionResult)
			t.Logf("  Condition result: %v", result)

			if !result {
				t.Errorf("Condition failed: %s %s %v", condition.FieldPath, condition.Operator, condition.Value)
			}
		})
	}
}
