package adapters

import (
	"testing"

	"buddy/internal/txn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRunIDsOnlySQL(t *testing.T) {
	tests := []struct {
		name     string
		ticket   domain.DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "single run ID",
			ticket: domain.DMLTicket{
				Deploy: []domain.TemplateInfo{
					{
						TargetDB:    "PC",
						SQLTemplate: "UPDATE workflow SET state = 201 WHERE run_id = %s;",
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: "run-123", Type: "string"},
						},
					},
				},
				Rollback: []domain.TemplateInfo{
					{
						TargetDB:    "PC",
						SQLTemplate: "UPDATE workflow SET state = 200 WHERE run_id = %s;",
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: "run-123", Type: "string"},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, rollback, "'run-123'")
			},
		},
		{
			name: "missing run ID",
			ticket: domain.DMLTicket{
				Deploy: []domain.TemplateInfo{
					{
						TargetDB:    "PC",
						SQLTemplate: "UPDATE workflow SET state = 201 WHERE run_id = %s;",
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: "", Type: "string"},
						},
					},
				},
				Rollback: []domain.TemplateInfo{
					{
						TargetDB:    "PC",
						SQLTemplate: "UPDATE workflow SET state = 200 WHERE run_id = %s;",
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: "", Type: "string"},
						},
					},
				},
			},
			wantErr: false, // Empty string is a valid parameter value
			validate: func(t *testing.T, deploy, rollback string) {
				// With empty string value, it will be wrapped in single quotes
				assert.Contains(t, deploy, "UPDATE workflow SET state = 201 WHERE run_id = '';")
				assert.Contains(t, rollback, "UPDATE workflow SET state = 200 WHERE run_id = '';")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements, err := GenerateSQLFromTicket(tt.ticket)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if len(statements.PCDeployStatements) > 0 && len(statements.PCRollbackStatements) > 0 {
					tt.validate(t, statements.PCDeployStatements[0], statements.PCRollbackStatements[0])
				}
			}
		})
	}
}

func TestMultipleDatabaseSupport(t *testing.T) {
	// Test that a single DMLTicket can generate SQL for multiple databases
	ticket := domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB:    "PE",
				SQLTemplate: "UPDATE pe_workflow SET state = 230 WHERE run_id = %s;",
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: "pe-run-123", Type: "string"},
				},
			},
			{
				TargetDB:    "PC",
				SQLTemplate: "UPDATE pc_workflow SET state = 0 WHERE run_id = %s;",
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: "pc-run-456", Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB:    "PE",
				SQLTemplate: "UPDATE pe_workflow SET state = 701 WHERE run_id = %s;",
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: "pe-run-123", Type: "string"},
				},
			},
			{
				TargetDB:    "PC",
				SQLTemplate: "UPDATE pc_workflow SET state = 500 WHERE run_id = %s;",
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: "pc-run-456", Type: "string"},
				},
			},
		},
	}

	statements, err := GenerateSQLFromTicket(ticket)
	require.NoError(t, err)

	// Verify PE statements
	assert.Contains(t, statements.PEDeployStatements[0], "UPDATE pe_workflow SET state = 230 WHERE run_id = 'pe-run-123';")
	assert.Contains(t, statements.PERollbackStatements[0], "UPDATE pe_workflow SET state = 701 WHERE run_id = 'pe-run-123';")

	// Verify PC statements
	assert.Contains(t, statements.PCDeployStatements[0], "UPDATE pc_workflow SET state = 0 WHERE run_id = 'pc-run-456';")
	assert.Contains(t, statements.PCRollbackStatements[0], "UPDATE pc_workflow SET state = 500 WHERE run_id = 'pc-run-456';")

	// Verify no statements in other databases
	assert.Empty(t, statements.RPPDeployStatements)
	assert.Empty(t, statements.RPPRollbackStatements)
}

func TestThoughtMachineFalseNegativeTemplate(t *testing.T) {
	// Create a mock transaction result
	result := domain.TransactionResult{
		PaymentEngine: &domain.PaymentEngineInfo{
			Workflow: domain.WorkflowInfo{
				RunID:       "pe-test-run-id",
				WorkflowID:  "workflow_transfer_payment",
				State:       "701",
				Attempt:     0,
				PrevTransID: "prev-trans-id",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalCapture: domain.PCInternalInfo{
				Workflow: domain.WorkflowInfo{
					RunID:      "pc-test-run-id",
					WorkflowID: "internal_payment_flow",
					State:      "500",
					Attempt:    0,
				},
			},
		},
		CaseType: domain.CaseThoughtMachineFalseNegative,
	}

	// Generate SQL statements
	statements := GenerateSQLStatements([]domain.TransactionResult{result})

	// Should not have any errors
	assert.Empty(t, result.Error)

	// Verify both PE and PC statements are generated
	assert.NotEmpty(t, statements.PEDeployStatements, "PE deploy statements should be generated")
	assert.NotEmpty(t, statements.PERollbackStatements, "PE rollback statements should be generated")
	assert.NotEmpty(t, statements.PCDeployStatements, "PC deploy statements should be generated")
	assert.NotEmpty(t, statements.PCRollbackStatements, "PC rollback statements should be generated")

	// Verify PE SQL content
	assert.Contains(t, statements.PEDeployStatements[0], "WHERE run_id = 'pe-test-run-id'")
	assert.Contains(t, statements.PERollbackStatements[0], "prev_trans_id = 'prev-trans-id'")
	assert.Contains(t, statements.PERollbackStatements[0], "WHERE run_id = 'pe-test-run-id'")

	// Verify PC SQL content
	assert.Contains(t, statements.PCDeployStatements[0], "WHERE run_id = 'pc-test-run-id'")
	assert.Contains(t, statements.PCDeployStatements[0], "workflow_id = 'internal_payment_flow'")
	assert.Contains(t, statements.PCRollbackStatements[0], "WHERE run_id = 'pc-test-run-id'")
	assert.Contains(t, statements.PCRollbackStatements[0], "workflow_id = 'internal_payment_flow'")
}

func TestGetDMLTicketForRppResume(t *testing.T) {
	tests := []struct {
		name          string
		result        domain.TransactionResult
		wantTicket    bool
		expectedRunID string
	}{
		{
			name: "first workflow matches wf_ct_cashout",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_cashout", RunID: "run-001", State: "210", Attempt: 0},
						{WorkflowID: "other_workflow", RunID: "run-002", State: "210", Attempt: 0},
					},
				},
			},
			wantTicket:    true,
			expectedRunID: "run-001",
		},
		{
			name: "second workflow matches wf_ct_qr_payment",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "other_workflow", RunID: "run-001", State: "210", Attempt: 0},
						{WorkflowID: "wf_ct_qr_payment", RunID: "run-002", State: "210", Attempt: 0},
					},
				},
			},
			wantTicket:    true,
			expectedRunID: "run-002",
		},
		{
			name: "no matching workflow found",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "other_workflow", RunID: "run-001", State: "210", Attempt: 0},
						{WorkflowID: "another_workflow", RunID: "run-002", State: "210", Attempt: 0},
					},
				},
			},
			wantTicket:    false,
			expectedRunID: "",
		},
		{
			name: "empty workflow slice",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{},
				},
			},
			wantTicket:    false,
			expectedRunID: "",
		},
		{
			name: "case type not matching rpp_no_response_resume",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_cashout", RunID: "run-001", State: "210", Attempt: 0},
					},
				},
				CaseType: domain.CaseNone,
			},
			wantTicket:    false,
			expectedRunID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := GetDMLTicketForRppResume(tt.result)

			if tt.wantTicket {
				require.NotNil(t, ticket)
				require.Equal(t, domain.CaseRppNoResponseResume, ticket.CaseType)
				require.Len(t, ticket.Deploy, 1)
				require.Len(t, ticket.Rollback, 1)
				assert.Equal(t, tt.expectedRunID, ticket.Deploy[0].Params[0].Value)
				assert.Equal(t, tt.expectedRunID, ticket.Rollback[0].Params[0].Value)
			} else {
				assert.Nil(t, ticket)
			}
		})
	}
}
