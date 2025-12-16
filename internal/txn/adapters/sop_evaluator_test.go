package adapters

import (
	"buddy/internal/txn/domain"
	"testing"
)

func TestEvaluateRule_PeStuck300RppNotFound(t *testing.T) {
	tests := []struct {
		name     string
		result   *domain.TransactionResult
		expected bool
	}{
		{
			name: "Should match when PaymentCore is empty struct and RPPAdapter is nil",
			result: &domain.TransactionResult{
				PaymentEngine: &domain.PaymentEngineInfo{
					Workflow: domain.WorkflowInfo{
						WorkflowID: "workflow_transfer_payment",
						State:     "300",
						Attempt:    0,
					},
				},
				PaymentCore: &domain.PaymentCoreInfo{
					InternalAuth: domain.PCInternalInfo{
						Workflow: domain.WorkflowInfo{
							WorkflowID: "internal_payment_flow",
							State:     "900",
							Attempt:    0,
						},
					},
					// InternalCapture is empty (zero value)
					// ExternalTransfer is empty (zero value)
				},
				// RPPAdapter is nil
				RPPAdapter: nil,
			},
			expected: true,
		},
		{
			name: "Should not match when RPPAdapter has data",
			result: &domain.TransactionResult{
				PaymentEngine: &domain.PaymentEngineInfo{
					Workflow: domain.WorkflowInfo{
						WorkflowID: "workflow_transfer_payment",
						State:     "300",
						Attempt:    0,
					},
				},
				PaymentCore: &domain.PaymentCoreInfo{
					InternalAuth: domain.PCInternalInfo{
						Workflow: domain.WorkflowInfo{
							WorkflowID: "internal_payment_flow",
							State:     "900",
							Attempt:    0,
						},
					},
				},
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: domain.WorkflowInfo{
						WorkflowID: "wf_ct_qr_payment",
						State:     "210",
						Attempt:    0,
					},
				},
			},
			expected: false,
		},
		{
			name: "Should not match when InternalCapture has data",
			result: &domain.TransactionResult{
				PaymentEngine: &domain.PaymentEngineInfo{
					Workflow: domain.WorkflowInfo{
						WorkflowID: "workflow_transfer_payment",
						State:     "300",
						Attempt:    0,
					},
				},
				PaymentCore: &domain.PaymentCoreInfo{
					InternalAuth: domain.PCInternalInfo{
						Workflow: domain.WorkflowInfo{
							WorkflowID: "internal_payment_flow",
							State:     "900",
							Attempt:    0,
						},
					},
					InternalCapture: domain.PCInternalInfo{
						Workflow: domain.WorkflowInfo{
							WorkflowID: "internal_payment_flow",
							State:     "500",
							Attempt:    0,
						},
					},
				},
				RPPAdapter: nil,
			},
			expected: false,
		},
	}

	repo := NewSOPRepository()
	rule := CaseRule{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.evaluateRule(rule, tt.result)
			if result != tt.expected {
				t.Errorf("evaluateRule() = %v, want %v", result, tt.expected)
			}
		})
	}
}