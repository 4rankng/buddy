package txn

import (
	"testing"
)

func TestGetWorkflowStateName(t *testing.T) {
	tests := []struct {
		name         string
		workflowType string
		state        int
		expected     string
	}{
		{
			name:         "General workflow state 220",
			workflowType: "workflow_transfer_payment",
			state:        220,
			expected:     "stTransferProcessing",
		},
		{
			name:         "Internal payment flow state 900",
			workflowType: "payment_core_workflow_internal_payment_flow",
			state:        900,
			expected:     "stPrepareUpdateAuth",
		},
		{
			name:         "External payment flow state 201",
			workflowType: "payment_core_workflow_external_payment_flow",
			state:        201,
			expected:     "stProcessing",
		},
		{
			name:         "RPP cashout state 101",
			workflowType: "workflow_wf_ct_cashout",
			state:        101,
			expected:     "stCreditTransferPersisted",
		},
		{
			name:         "RPP cashin state 900",
			workflowType: "wf_ct_cashin",
			state:        900,
			expected:     "stCashInCompleted",
		},
		{
			name:         "Unknown workflow type",
			workflowType: "unknown_workflow",
			state:        100,
			expected:     "stTransferPersisted",
		},
		{
			name:         "Unknown state number",
			workflowType: "workflow_transfer_payment",
			state:        999,
			expected:     "stUnknown_999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWorkflowStateName(tt.workflowType, tt.state)
			if result != tt.expected {
				t.Errorf("GetWorkflowStateName(%q, %d) = %q; want %q", tt.workflowType, tt.state, result, tt.expected)
			}
		})
	}
}

func TestFormatWorkflowState(t *testing.T) {
	tests := []struct {
		name         string
		workflowType string
		stateStr     string
		expected     string
	}{
		{
			name:         "Valid state number",
			workflowType: "workflow_transfer_payment",
			stateStr:     "220",
			expected:     "stTransferProcessing(220)",
		},
		{
			name:         "External payment flow state",
			workflowType: "payment_core_workflow_external_payment_flow",
			stateStr:     "201",
			expected:     "stProcessing(201)",
		},
		{
			name:         "RPP cashout state",
			workflowType: "workflow_wf_ct_cashout",
			stateStr:     "210",
			expected:     "stTransferProcessing(210)",
		},
		{
			name:         "Invalid state string",
			workflowType: "workflow_transfer_payment",
			stateStr:     "invalid",
			expected:     "invalid",
		},
		{
			name:         "Empty state string",
			workflowType: "workflow_transfer_payment",
			stateStr:     "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatWorkflowState(tt.workflowType, tt.stateStr)
			if result != tt.expected {
				t.Errorf("FormatWorkflowState(%q, %q) = %q; want %q", tt.workflowType, tt.stateStr, result, tt.expected)
			}
		})
	}
}

func TestWorkflowInfoGetFormattedState(t *testing.T) {
	tests := []struct {
		name         string
		workflowInfo WorkflowInfo
		expected     string
	}{
		{
			name: "Payment engine workflow",
			workflowInfo: WorkflowInfo{
				Type:    "workflow_transfer_payment",
				State:   "220",
				Attempt: 0,
			},
			expected: "stTransferProcessing(220)",
		},
		{
			name: "Internal payment flow",
			workflowInfo: WorkflowInfo{
				Type:    "payment_core_workflow_internal_payment_flow",
				State:   "900",
				Attempt: 1,
			},
			expected: "stPrepareUpdateAuth(900)",
		},
		{
			name: "RPP cashout workflow",
			workflowInfo: WorkflowInfo{
				Type:    "wf_ct_cashout",
				State:   "101",
				Attempt: 19,
			},
			expected: "stCreditTransferPersisted(101)",
		},
		{
			name: "Invalid state",
			workflowInfo: WorkflowInfo{
				Type:    "workflow_transfer_payment",
				State:   "invalid",
				Attempt: 0,
			},
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.workflowInfo.GetFormattedState()
			if result != tt.expected {
				t.Errorf("WorkflowInfo.GetFormattedState() = %q; want %q", result, tt.expected)
			}
		})
	}
}
