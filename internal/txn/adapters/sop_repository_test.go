package adapters

import (
	"buddy/internal/txn/domain"
	"testing"
)

func TestThoughtMachineFalseNegativeRule(t *testing.T) {
	// Test case: Thought Machine false negative with PE state 701 but RPP indicates success
	result := domain.TransactionResult{
		PaymentEngine: domain.WorkflowInfo{
			Workflow: domain.WorkflowDetails{
				RunID:     "ced4efe76ea442ddbbca1f745ebe2386",
				WorkflowID: "workflow_transfer_payment",
				State:     "701", // stCaptureFailed
				Attempt:   0,
			},
		},
		RPPAdapter: domain.WorkflowInfo{
			Status: "PROCESSING", // Not 900, indicating RPP thinks it's still processing
		},
	}

	// Get the SOP repository
	repo := NewSOPRepository()

	// Identify the case
	caseType, err := repo.IdentifyCase(result)
	if err != nil {
		t.Fatalf("Failed to identify case: %v", err)
	}

	// Verify the case is identified correctly
	if caseType != domain.CaseThoughtMachineFalseNegative {
		t.Errorf("Expected case type %s, got %s", domain.CaseThoughtMachineFalseNegative, caseType)
	}
}

func TestThoughtMachineFalseNegativeRuleVariations(t *testing.T) {
	repo := NewSOPRepository()

	tests := []struct {
		name            string
		result          domain.TransactionResult
		expectedCase    domain.Case
		shouldMatch     bool
	}{
		{
			name: "PE state 701 with workflow_transfer_payment should match",
			result: domain.TransactionResult{
				PaymentEngine: domain.WorkflowInfo{
					Workflow: domain.WorkflowDetails{
						RunID:     "test-run-id-1",
						WorkflowID: "workflow_transfer_payment",
						State:     "701",
						Attempt:   0,
					},
				},
				RPPAdapter: domain.WorkflowInfo{
					Status: "PROCESSING",
				},
			},
			expectedCase: domain.CaseThoughtMachineFalseNegative,
			shouldMatch:  true,
		},
		{
			name: "PE state 701 with different workflow should not match",
			result: domain.TransactionResult{
				PaymentEngine: domain.WorkflowInfo{
					Workflow: domain.WorkflowDetails{
						RunID:     "test-run-id-2",
						WorkflowID: "different_workflow",
						State:     "701",
						Attempt:   0,
					},
				},
				RPPAdapter: domain.WorkflowInfo{
					Status: "PROCESSING",
				},
			},
			expectedCase: "",
			shouldMatch:  false,
		},
		{
			name: "PE state 500 (stCaptureFailed but not 701) should not match",
			result: domain.TransactionResult{
				PaymentEngine: domain.WorkflowInfo{
					Workflow: domain.WorkflowDetails{
						RunID:     "test-run-id-3",
						WorkflowID: "workflow_transfer_payment",
						State:     "500", // Different failure state
						Attempt:   0,
					},
				},
				RPPAdapter: domain.WorkflowInfo{
					Status: "PROCESSING",
				},
			},
			expectedCase: "",
			shouldMatch:  false,
		},
		{
			name: "PE state 701 with RPP status 900 should not match",
			result: domain.TransactionResult{
				PaymentEngine: domain.WorkflowInfo{
					Workflow: domain.WorkflowDetails{
						RunID:     "test-run-id-4",
						WorkflowID: "workflow_transfer_payment",
						State:     "701",
						Attempt:   0,
					},
				},
				RPPAdapter: domain.WorkflowInfo{
					Status: "900", // RPP indicates completion
				},
			},
			expectedCase: "",
			shouldMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caseType, err := repo.IdentifyCase(tt.result)
			if tt.shouldMatch {
				if err != nil {
					t.Errorf("Unexpected error identifying case: %v", err)
				}
				if caseType != tt.expectedCase {
					t.Errorf("Expected case type %s, got %s", tt.expectedCase, caseType)
				}
			} else {
				// Should not match the Thought Machine false negative rule
				if err == nil && caseType == domain.CaseThoughtMachineFalseNegative {
					t.Error("Unexpectedly matched Thought Machine false negative case")
				}
			}
		})
	}
}

func TestSOPRepositoryContainsAllCases(t *testing.T) {
	repo := NewSOPRepository()

	// Get all rules to verify Thought Machine false negative is included
	rules := repo.GetRules()

	found := false
	for _, rule := range rules {
		if rule.CaseType == domain.CaseThoughtMachineFalseNegative {
			found = true

			// Verify the rule has correct conditions
			if len(rule.Conditions) != 3 {
				t.Errorf("Expected 3 conditions for Thought Machine false negative rule, got %d", len(rule.Conditions))
			}

			// Check for expected conditions
			hasPEState := false
			hasWorkflowID := false
			hasRPPStatus := false

			for _, condition := range rule.Conditions {
				if condition.FieldPath == "PaymentEngine.Workflow.State" && condition.Operator == "eq" && condition.Value == "701" {
					hasPEState = true
				}
				if condition.FieldPath == "PaymentEngine.Workflow.WorkflowID" && condition.Operator == "eq" && condition.Value == "workflow_transfer_payment" {
					hasWorkflowID = true
				}
				if condition.FieldPath == "RPPAdapter.Status" && condition.Operator == "ne" && condition.Value == "900" {
					hasRPPStatus = true
				}
			}

			if !hasPEState {
				t.Error("Thought Machine false negative rule missing PE state 701 condition")
			}
			if !hasWorkflowID {
				t.Error("Thought Machine false negative rule missing workflow_transfer_payment condition")
			}
			if !hasRPPStatus {
				t.Error("Thought Machine false negative rule missing RPP status != 900 condition")
			}

			break
		}
	}

	if !found {
		t.Error("Thought Machine false negative rule not found in SOP repository")
	}
}