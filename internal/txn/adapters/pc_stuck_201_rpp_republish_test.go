package adapters

import (
	"buddy/internal/txn/domain"
	"encoding/json"
	"testing"
)

// TestPcStuck201WaitingRppRepublishFromRpp_CaseIdentification tests the case identification
// for PC stuck at 201/0 waiting for RPP message, RPP workflows at 900/0 (republish from RPP)
func TestPcStuck201WaitingRppRepublishFromRpp_CaseIdentification(t *testing.T) {
	// Create the transaction data matching the context provided
	// e2e_id: 20260204GXSPMYKL010ORB00010461
	transactionResult := &domain.TransactionResult{
		InputID:  "20260204GXSPMYKL010ORB00010461",
		CaseType: domain.CaseNone,
		PaymentEngine: &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				Type:        "PAYMENT",
				TxnSubtype:  "RPP_NETWORK",
				TxnDomain:   "DEPOSITS",
				Status:      "PROCESSING",
				ReferenceID: "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
				ExternalID:  "rpp-ext-id-123", // Non-empty external ID
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_transfer_payment",
				State:      "220",
				Attempt:    0,
				RunID:      "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalAuth: domain.PCInternalInfo{
				TxID:      "24fd946f809a4f4a9daca82819a8fe2e",
				GroupID:   "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "24fd946f809a4f4a9daca82819a8fe2e",
				},
			},
			ExternalTransfer: domain.PCExternalInfo{
				RefID:    "2da4581ac78b407598460f78f8cb74f4",
				GroupID:  "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:   "TRANSFER",
				TxStatus: "PROCESSING",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "external_payment_flow",
					State:      "201",
					Attempt:    0,
					RunID:      "2da4581ac78b407598460f78f8cb74f4",
				},
			},
		},
		RPPAdapter: &domain.RPPAdapterInfo{
			Workflow: []domain.WorkflowInfo{
				{
					WorkflowID: "wf_ct_cashout",
					State:      "900",
					Attempt:    0,
					RunID:      "5f7eff81c5c848e7b7ad03b1ab19e022",
					Data:       `{"State": 900}`,
				},
			},
		},
	}

	// Create a new SOP repository
	sopRepo := NewSOPRepository()

	// Test case identification
	result := sopRepo.IdentifyCase(transactionResult, "my")

	if result != domain.CasePcExternalPaymentFlow201_0RPP900 {
		t.Errorf("Expected case %s, got %s", domain.CasePcExternalPaymentFlow201_0RPP900, result)
	}
}

// TestPcStuck201WaitingRppRepublishFromRpp_SQLGeneration tests SQL generation for the case
func TestPcStuck201WaitingRppRepublishFromRpp_SQLGeneration(t *testing.T) {
	// Create the transaction data
	transactionResult := domain.TransactionResult{
		InputID:  "20260204GXSPMYKL010ORB00010461",
		CaseType: domain.CasePcExternalPaymentFlow201_0RPP900,
		PaymentEngine: &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				Type:        "PAYMENT",
				TxnSubtype:  "RPP_NETWORK",
				TxnDomain:   "DEPOSITS",
				Status:      "PROCESSING",
				ReferenceID: "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
				ExternalID:  "rpp-ext-id-123",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_transfer_payment",
				State:      "220",
				Attempt:    0,
				RunID:      "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalAuth: domain.PCInternalInfo{
				TxID:      "24fd946f809a4f4a9daca82819a8fe2e",
				GroupID:   "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "24fd946f809a4f4a9daca82819a8fe2e",
				},
			},
			ExternalTransfer: domain.PCExternalInfo{
				RefID:    "2da4581ac78b407598460f78f8cb74f4",
				GroupID:  "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:   "TRANSFER",
				TxStatus: "PROCESSING",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "external_payment_flow",
					State:      "201",
					Attempt:    0,
					RunID:      "2da4581ac78b407598460f78f8cb74f4",
				},
			},
		},
		RPPAdapter: &domain.RPPAdapterInfo{
			Workflow: []domain.WorkflowInfo{
				{
					WorkflowID: "wf_ct_cashout",
					State:      "900",
					Attempt:    0,
					RunID:      "5f7eff81c5c848e7b7ad03b1ab19e022",
					Data:       `{"State": 900}`,
				},
			},
		},
	}

	// Generate SQL using the template function
	results := []domain.TransactionResult{transactionResult}
	statements := GenerateSQLStatements(results)

	// Verify RPP deployment statements exist
	if len(statements.RPPDeployStatements) == 0 {
		t.Fatal("Expected RPP deploy statements")
	}

	// Check that the SQL contains expected keywords
	rppSQL := statements.RPPDeployStatements[0]

	expectedElements := []string{
		"UPDATE workflow_execution",
		"SET state = 301",
		"attempt = 1",
		"data = JSON_SET(data, '$.State', 301)",
		"WHERE workflow_id = 'wf_ct_cashout'",
		"AND run_id IN", // Uses IN clause for batch operations
		"AND attempt = 0",
		"AND state = 900",
	}

	for _, element := range expectedElements {
		if !containsString(rppSQL, element) {
			t.Errorf("Expected SQL to contain '%s', but got: %s", element, rppSQL)
		}
	}

	// Verify rollback statements exist
	if len(statements.RPPRollbackStatements) == 0 {
		t.Fatal("Expected RPP rollback statements")
	}

	// Check that rollback SQL contains expected keywords
	rollbackSQL := statements.RPPRollbackStatements[0]
	rollbackElements := []string{
		"UPDATE workflow_execution",
		"SET state = 900",
		"attempt = 0",
		"data = JSON_SET(data, '$.State', 900)",
		"WHERE run_id IN", // Uses IN clause for batch operations
		"AND workflow_id = 'wf_ct_cashout'",
	}

	for _, element := range rollbackElements {
		if !containsString(rollbackSQL, element) {
			t.Errorf("Expected rollback SQL to contain '%s', but got: %s", element, rollbackSQL)
		}
	}

	t.Logf("Deploy SQL:\n%s", rppSQL)
	t.Logf("Rollback SQL:\n%s", rollbackSQL)
}

// TestPcStuck201WaitingRppRepublishFromRpp_ConditionValidation tests each condition individually
func TestPcStuck201WaitingRppRepublishFromRpp_ConditionValidation(t *testing.T) {
	transactionResult := &domain.TransactionResult{
		InputID:  "20260204GXSPMYKL010ORB00010461",
		CaseType: domain.CaseNone,
		PaymentEngine: &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				Type:        "PAYMENT",
				TxnSubtype:  "RPP_NETWORK",
				TxnDomain:   "DEPOSITS",
				Status:      "PROCESSING",
				ReferenceID: "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
				ExternalID:  "rpp-ext-id-123",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_transfer_payment",
				State:      "220",
				Attempt:    0,
				RunID:      "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalAuth: domain.PCInternalInfo{
				TxID:      "24fd946f809a4f4a9daca82819a8fe2e",
				GroupID:   "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "24fd946f809a4f4a9daca82819a8fe2e",
				},
			},
			ExternalTransfer: domain.PCExternalInfo{
				RefID:    "2da4581ac78b407598460f78f8cb74f4",
				GroupID:  "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:   "TRANSFER",
				TxStatus: "PROCESSING",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "external_payment_flow",
					State:      "201",
					Attempt:    0,
					RunID:      "2da4581ac78b407598460f78f8cb74f4",
				},
			},
		},
		RPPAdapter: &domain.RPPAdapterInfo{
			Workflow: []domain.WorkflowInfo{
				{
					WorkflowID: "wf_ct_cashout",
					State:      "900",
					Attempt:    0,
					RunID:      "5f7eff81c5c848e7b7ad03b1ab19e022",
					Data:       `{"State": 900}`,
				},
			},
		},
	}

	// Get the rule we're testing
	var targetRule *CaseRule
	for _, rule := range getDefaultSOPRules() {
		if rule.CaseType == domain.CasePcExternalPaymentFlow201_0RPP900 {
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
		conditionName := condition.FieldPath + " " + condition.Operator + " " + toJSONString(condition.Value)
		t.Run(conditionName, func(t *testing.T) {
			fieldValue, ok := sopRepo.getFieldValue(condition.FieldPath, transactionResult)
			t.Logf("Field: %s", condition.FieldPath)
			t.Logf("  Expected: %v", condition.Value)
			t.Logf("  Actual: %v", fieldValue)
			t.Logf("  Found: %v", ok)

			if !ok && !(condition.Operator == "eq" && condition.Value == "") {
				t.Errorf("Field not found: %s", condition.FieldPath)
				return
			}

			result := sopRepo.evaluateCondition(condition, transactionResult)
			t.Logf("  Condition %d result: %v", i, result)

			if !result {
				t.Errorf("Condition %d failed: %s %s %v", i, condition.FieldPath, condition.Operator, condition.Value)
			}
		})
	}
}

// TestPcStuck201WaitingRppRepublishFromRpp_MissingRPPAdapterShouldNotMatch tests that missing RPP adapter doesn't match
func TestPcStuck201WaitingRppRepublishFromRpp_MissingRPPAdapterShouldNotMatch(t *testing.T) {
	transactionResult := &domain.TransactionResult{
		InputID:  "20260204GXSPMYKL010ORB00010461",
		CaseType: domain.CaseNone,
		PaymentEngine: &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				Type:        "PAYMENT",
				TxnSubtype:  "RPP_NETWORK",
				TxnDomain:   "DEPOSITS",
				Status:      "PROCESSING",
				ReferenceID: "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
				ExternalID:  "rpp-ext-id-123",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_transfer_payment",
				State:      "220",
				Attempt:    0,
				RunID:      "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalAuth: domain.PCInternalInfo{
				TxID:      "24fd946f809a4f4a9daca82819a8fe2e",
				GroupID:   "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "24fd946f809a4f4a9daca82819a8fe2e",
				},
			},
			ExternalTransfer: domain.PCExternalInfo{
				RefID:    "2da4581ac78b407598460f78f8cb74f4",
				GroupID:  "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:   "TRANSFER",
				TxStatus: "PROCESSING",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "external_payment_flow",
					State:      "201",
					Attempt:    0,
					RunID:      "2da4581ac78b407598460f78f8cb74f4",
				},
			},
		},
		RPPAdapter: nil, // No RPP adapter
	}

	sopRepo := NewSOPRepository()
	result := sopRepo.IdentifyCase(transactionResult, "my")

	if result == domain.CasePcExternalPaymentFlow201_0RPP900 {
		t.Errorf("Should not match when RPP adapter is missing")
	}
}

// TestPcStuck201WaitingRppRepublishFromRpp_WrongRPPStateShouldNotMatch tests that wrong RPP state doesn't match
func TestPcStuck201WaitingRppRepublishFromRpp_WrongRPPStateShouldNotMatch(t *testing.T) {
	transactionResult := &domain.TransactionResult{
		InputID:  "20260204GXSPMYKL010ORB00010461",
		CaseType: domain.CaseNone,
		PaymentEngine: &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				Type:        "PAYMENT",
				TxnSubtype:  "RPP_NETWORK",
				TxnDomain:   "DEPOSITS",
				Status:      "PROCESSING",
				ReferenceID: "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
				ExternalID:  "rpp-ext-id-123",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_transfer_payment",
				State:      "220",
				Attempt:    0,
				RunID:      "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalAuth: domain.PCInternalInfo{
				TxID:      "24fd946f809a4f4a9daca82819a8fe2e",
				GroupID:   "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "24fd946f809a4f4a9daca82819a8fe2e",
				},
			},
			ExternalTransfer: domain.PCExternalInfo{
				RefID:    "2da4581ac78b407598460f78f8cb74f4",
				GroupID:  "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:   "TRANSFER",
				TxStatus: "PROCESSING",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "external_payment_flow",
					State:      "201",
					Attempt:    0,
					RunID:      "2da4581ac78b407598460f78f8cb74f4",
				},
			},
		},
		RPPAdapter: &domain.RPPAdapterInfo{
			Workflow: []domain.WorkflowInfo{
				{
					WorkflowID: "wf_ct_cashout",
					State:      "210", // Wrong state (should be 900)
					Attempt:    0,
					RunID:      "5f7eff81c5c848e7b7ad03b1ab19e022",
					Data:       `{"State": 210}`,
				},
			},
		},
	}

	sopRepo := NewSOPRepository()
	result := sopRepo.IdentifyCase(transactionResult, "my")

	if result == domain.CasePcExternalPaymentFlow201_0RPP900 {
		t.Errorf("Should not match when RPP state is not 900")
	}
}

// TestPcStuck201WaitingRppRepublishFromRpp_NonZeroAttemptShouldNotMatch tests that non-zero attempt doesn't match
func TestPcStuck201WaitingRppRepublishFromRpp_NonZeroAttemptShouldNotMatch(t *testing.T) {
	transactionResult := &domain.TransactionResult{
		InputID:  "20260204GXSPMYKL010ORB00010461",
		CaseType: domain.CaseNone,
		PaymentEngine: &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				Type:        "PAYMENT",
				TxnSubtype:  "RPP_NETWORK",
				TxnDomain:   "DEPOSITS",
				Status:      "PROCESSING",
				ReferenceID: "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
				ExternalID:  "rpp-ext-id-123",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_transfer_payment",
				State:      "220",
				Attempt:    0,
				RunID:      "6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalAuth: domain.PCInternalInfo{
				TxID:      "24fd946f809a4f4a9daca82819a8fe2e",
				GroupID:   "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				ErrorCode: "",
				ErrorMsg:  "",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "24fd946f809a4f4a9daca82819a8fe2e",
				},
			},
			ExternalTransfer: domain.PCExternalInfo{
				RefID:    "2da4581ac78b407598460f78f8cb74f4",
				GroupID:  "5f7eff81c5c848e7b7ad03b1ab19e022",
				TxType:   "TRANSFER",
				TxStatus: "PROCESSING",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "external_payment_flow",
					State:      "201",
					Attempt:    0,
					RunID:      "2da4581ac78b407598460f78f8cb74f4",
				},
			},
		},
		RPPAdapter: &domain.RPPAdapterInfo{
			Workflow: []domain.WorkflowInfo{
				{
					WorkflowID: "wf_ct_cashout",
					State:      "900",
					Attempt:    1, // Non-zero attempt (should be 0)
					RunID:      "5f7eff81c5c848e7b7ad03b1ab19e022",
					Data:       `{"State": 900}`,
				},
			},
		},
	}

	sopRepo := NewSOPRepository()
	result := sopRepo.IdentifyCase(transactionResult, "my")

	if result == domain.CasePcExternalPaymentFlow201_0RPP900 {
		t.Errorf("Should not match when RPP attempt is not 0")
	}
}

// Helper function to check if a string contains a substring
func containsString(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 &&
		(str == substr ||
			(len(str) >= len(substr) &&
				contains(str, substr)))
}

func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to convert value to JSON string
func toJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "error"
	}
	return string(b)
}
