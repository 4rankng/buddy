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
				// With empty string value, it will be wrapped in single quotes in IN clause
				assert.Contains(t, deploy, "UPDATE workflow SET state = 201 WHERE run_id IN ('');")
				assert.Contains(t, rollback, "UPDATE workflow SET state = 200 WHERE run_id IN ('');")
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

	// Verify PE statements (now using IN clause)
	assert.Contains(t, statements.PEDeployStatements[0], "UPDATE pe_workflow SET state = 230 WHERE run_id IN ('pe-run-123');")
	assert.Contains(t, statements.PERollbackStatements[0], "UPDATE pe_workflow SET state = 701 WHERE run_id IN ('pe-run-123');")

	// Verify PC statements (now using IN clause)
	assert.Contains(t, statements.PCDeployStatements[0], "UPDATE pc_workflow SET state = 0 WHERE run_id IN ('pc-run-456');")
	assert.Contains(t, statements.PCRollbackStatements[0], "UPDATE pc_workflow SET state = 500 WHERE run_id IN ('pc-run-456');")

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
	assert.Contains(t, statements.PEDeployStatements[0], "WHERE run_id IN ('pe-test-run-id')")
	assert.Contains(t, statements.PERollbackStatements[0], "prev_trans_id = 'prev-trans-id'")
	assert.Contains(t, statements.PERollbackStatements[0], "WHERE run_id IN ('pe-test-run-id')")

	// Verify PC SQL content
	assert.Contains(t, statements.PCDeployStatements[0], "WHERE run_id IN ('pc-test-run-id')")
	assert.Contains(t, statements.PCDeployStatements[0], "workflow_id = 'internal_payment_flow'")
	assert.Contains(t, statements.PCRollbackStatements[0], "WHERE run_id IN ('pc-test-run-id')")
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
				CaseType: domain.CaseRppNoResponseResume,
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
				CaseType: domain.CaseRppNoResponseResume,
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

func TestGetDMLTicketForRppRtpCashinStuck200_0(t *testing.T) {
	tests := []struct {
		name    string
		result  domain.TransactionResult
		wantNil bool
	}{
		{
			name: "valid RTP cashin stuck at 200/0",
			result: domain.TransactionResult{
				InputID: "20251229BIMBMYKL070ORB53488076",
				RPPAdapter: &domain.RPPAdapterInfo{
					PartnerTxID: "0f20cdcbc8dd44a7915e6803c7542778",
					EndToEndID:  "20251229BIMBMYKL070ORB53488076",
					Workflow: []domain.WorkflowInfo{
						{
							WorkflowID: "wf_ct_rtp_cashin",
							State:      "200",
							Attempt:    0,
							RunID:      "52eaa330045138178bf0b0e6e33dde87",
						},
					},
				},
			},
			wantNil: false,
		},
		{
			name: "wrong state",
			result: domain.TransactionResult{
				InputID: "20251229BIMBMYKL070ORB53488076",
				RPPAdapter: &domain.RPPAdapterInfo{
					PartnerTxID: "0f20cdcbc8dd44a7915e6803c7542778",
					EndToEndID:  "20251229BIMBMYKL070ORB53488076",
					Workflow: []domain.WorkflowInfo{
						{
							WorkflowID: "wf_ct_rtp_cashin",
							State:      "210",
							Attempt:    0,
							RunID:      "52eaa330045138178bf0b0e6e33dde87",
						},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "nil RPP adapter",
			result: domain.TransactionResult{
				InputID:    "20251229BIMBMYKL070ORB53488076",
				RPPAdapter: nil,
			},
			wantNil: true,
		},
		{
			name: "empty partner tx id",
			result: domain.TransactionResult{
				InputID: "20251229BIMBMYKL070ORB53488076",
				RPPAdapter: &domain.RPPAdapterInfo{
					PartnerTxID: "",
					EndToEndID:  "20251229BIMBMYKL070ORB53488076",
					Workflow: []domain.WorkflowInfo{
						{
							WorkflowID: "wf_ct_rtp_cashin",
							State:      "200",
							Attempt:    0,
							RunID:      "52eaa330045138178bf0b0e6e33dde87",
						},
					},
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Debug: print case type before calling function
			t.Logf("Before: CaseType=%s", tt.result.CaseType)

			ticket := GetDMLTicketForRppRtpCashinStuck200_0(tt.result)

			// Debug: print case type after calling function
			t.Logf("After: CaseType=%s", tt.result.CaseType)
			t.Logf("Ticket: %v", ticket)

			if tt.wantNil {
				if ticket != nil {
					t.Errorf("expected nil ticket, got %v", ticket)
				}
				return
			}

			if ticket == nil {
				t.Fatal("expected non-nil ticket")
			}

			if ticket.CaseType != domain.CaseRppRtpCashinStuck200_0 {
				t.Errorf("expected case type %s, got %s", domain.CaseRppRtpCashinStuck200_0, ticket.CaseType)
			}

			if len(ticket.Deploy) < 2 {
				t.Error("expected at least 2 deploy statements (PPE and RPP)")
			}
			if len(ticket.Rollback) < 2 {
				t.Error("expected at least 2 rollback statements (PPE and RPP)")
			}
		})
	}
}

func TestGetDMLTicketForRppNoResponseRejectNotFound(t *testing.T) {
	tests := []struct {
		name          string
		result        domain.TransactionResult
		wantTicket    bool
		expectedRunID string
	}{
		{
			name: "wf_ct_qr_payment at state 0, attempt 20",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_qr_payment", RunID: "test-run-001", State: "0", Attempt: 20},
					},
				},
				CaseType: domain.CaseRppNoResponseRejectNotFound,
			},
			wantTicket:    true,
			expectedRunID: "test-run-001",
		},
		{
			name: "wf_ct_qr_payment at state 0, attempt 0",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_qr_payment", RunID: "test-run-002", State: "0", Attempt: 0},
					},
				},
				CaseType: domain.CaseRppNoResponseRejectNotFound,
			},
			wantTicket:    true,
			expectedRunID: "test-run-002",
		},
		{
			name: "wf_ct_qr_payment at state 0, attempt 5",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_qr_payment", RunID: "test-run-003", State: "0", Attempt: 5},
					},
				},
				CaseType: domain.CaseRppNoResponseRejectNotFound,
			},
			wantTicket:    true,
			expectedRunID: "test-run-003",
		},
		{
			name: "wrong workflow (wf_ct_cashout)",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_cashout", RunID: "test-run-004", State: "0", Attempt: 5},
					},
				},
				CaseType: domain.CaseRppNoResponseRejectNotFound,
			},
			wantTicket:    false,
			expectedRunID: "",
		},
		{
			name: "wrong state (210 instead of 0)",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_qr_payment", RunID: "test-run-005", State: "210", Attempt: 0},
					},
				},
				CaseType: domain.CaseRppNoResponseRejectNotFound,
			},
			wantTicket:    false,
			expectedRunID: "",
		},
		{
			name: "nil RPPAdapter",
			result: domain.TransactionResult{
				RPPAdapter: nil,
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
				CaseType: domain.CaseRppNoResponseRejectNotFound,
			},
			wantTicket:    false,
			expectedRunID: "",
		},
		{
			name: "first workflow matches, second doesn't",
			result: domain.TransactionResult{
				RPPAdapter: &domain.RPPAdapterInfo{
					Workflow: []domain.WorkflowInfo{
						{WorkflowID: "wf_ct_qr_payment", RunID: "test-run-006", State: "0", Attempt: 10},
						{WorkflowID: "wf_ct_cashout", RunID: "test-run-007", State: "0", Attempt: 5},
					},
				},
				CaseType: domain.CaseRppNoResponseRejectNotFound,
			},
			wantTicket:    true,
			expectedRunID: "test-run-006",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := GetDMLTicketForRppNoResponseRejectNotFound(tt.result)

			if tt.wantTicket {
				require.NotNil(t, ticket)
				require.Equal(t, domain.CaseRppNoResponseRejectNotFound, ticket.CaseType)
				require.Len(t, ticket.Deploy, 1)
				require.Len(t, ticket.Rollback, 1)
				assert.Equal(t, "RPP", ticket.Deploy[0].TargetDB)
				assert.Equal(t, "RPP", ticket.Rollback[0].TargetDB)
				assert.Equal(t, tt.expectedRunID, ticket.Deploy[0].Params[0].Value)
				assert.Equal(t, tt.expectedRunID, ticket.Rollback[0].Params[0].Value)

				// Verify SQL contains key elements
				deploySQL := ticket.Deploy[0].SQLTemplate
				assert.Contains(t, deploySQL, "state = 221")
				assert.Contains(t, deploySQL, "attempt = 1")
				assert.Contains(t, deploySQL, "workflow_id = 'wf_ct_qr_payment'")
				assert.Contains(t, deploySQL, "state = 0")

				// Verify rollback SQL
				rollbackSQL := ticket.Rollback[0].SQLTemplate
				assert.Contains(t, rollbackSQL, "state = 0")
				assert.Contains(t, rollbackSQL, "attempt = 0")
				assert.Contains(t, rollbackSQL, "workflow_id = 'wf_ct_qr_payment'")
			} else {
				assert.Nil(t, ticket)
			}
		})
	}
}
