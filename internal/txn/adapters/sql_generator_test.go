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
		ticket   DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "single run ID",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123"},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id IN (%s);",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id IN (%s);",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, rollback, "'run-123'")
			},
		},
		{
			name: "multiple run IDs",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123", "run-456", "run-789"},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id IN (%s);",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id IN (%s);",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123', 'run-456', 'run-789'")
				assert.Contains(t, rollback, "'run-123', 'run-456', 'run-789'")
			},
		},
		{
			name:    "empty run IDs",
			ticket:  DMLTicket{RunIDs: []string{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy, rollback, err := buildRunIDsOnlySQL(tt.ticket)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.validate(t, deploy, rollback)
			}
		})
	}
}

func TestBuildRunIDsWithWorkflowIDSQL(t *testing.T) {
	tests := []struct {
		name     string
		ticket   DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "valid ticket with workflow ID",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123"},
				WorkflowID:       "workflow-456",
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id IN (%s) AND workflow_id = '%s';",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id IN (%s) AND workflow_id = '%s';",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, deploy, "workflow-456")
				assert.Contains(t, rollback, "'run-123'")
				assert.Contains(t, rollback, "workflow-456")
			},
		},
		{
			name: "empty workflow ID",
			ticket: DMLTicket{
				RunIDs:     []string{"run-123"},
				WorkflowID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy, rollback, err := buildRunIDsWithWorkflowIDSQL(tt.ticket)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.validate(t, deploy, rollback)
			}
		})
	}
}

func TestBuildRunIDsWithWorkflowIDsSQL(t *testing.T) {
	tests := []struct {
		name     string
		ticket   DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "valid ticket with multiple workflow IDs",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123"},
				WorkflowIDs:      []string{"wf-1", "wf-2", "wf-3"},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id IN (%s) AND workflow_id IN (%s);",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id IN (%s) AND workflow_id IN (%s);",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, deploy, "wf-1, wf-2, wf-3")
				assert.Contains(t, rollback, "'run-123'")
				assert.Contains(t, rollback, "wf-1, wf-2, wf-3")
			},
		},
		{
			name: "empty workflow IDs",
			ticket: DMLTicket{
				RunIDs:      []string{"run-123"},
				WorkflowIDs: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy, rollback, err := buildRunIDsWithWorkflowIDsSQL(tt.ticket)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.validate(t, deploy, rollback)
			}
		})
	}
}

func TestBuildRunIDsOnlyWithPrevTransIDSQL(t *testing.T) {
	tests := []struct {
		name     string
		ticket   DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "valid ticket with prev_trans_id",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123"},
				PrevTransID:      "prev-456",
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = %s;",
				RollbackTemplate: "UPDATE workflow SET prev_trans_id = '%s' WHERE run_id = %s;",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, rollback, "prev-456")
				assert.Contains(t, rollback, "'run-123'")
			},
		},
		{
			name: "multiple run IDs (should fail for prev_trans_id)",
			ticket: DMLTicket{
				RunIDs:      []string{"run-123", "run-456"},
				PrevTransID: "prev-789",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy, rollback, err := buildRunIDsOnlyWithPrevTransIDSQL(tt.ticket)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.validate(t, deploy, rollback)
			}
		})
	}
}

func TestGenerateSQLFromTicket(t *testing.T) {
	tests := []struct {
		name        string
		ticket      DMLTicket
		wantErr     bool
		expectPC    bool
		expectPE    bool
		expectRPP   bool
		expectEmpty bool
	}{
		{
			name: "PC case with single run ID",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123"},
				DeployTemplate:   "UPDATE pc_workflow SET state = 201 WHERE run_id IN (%s);",
				RollbackTemplate: "UPDATE pc_workflow SET state = 200 WHERE run_id IN (%s);",
				TargetDB:         "PC",
				CaseType:         domain.CasePcExternalPaymentFlow200_11,
			},
			wantErr:  false,
			expectPC: true,
		},
		{
			name: "PE case with workflow ID",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123", "run-456"},
				WorkflowID:       "pe-workflow",
				DeployTemplate:   "UPDATE pe_workflow SET state = 221 WHERE run_id IN (%s) AND workflow_id = '%s';",
				RollbackTemplate: "UPDATE pe_workflow SET state = 210 WHERE run_id IN (%s) AND workflow_id = '%s';",
				TargetDB:         "PE",
				CaseType:         domain.CasePeTransferPayment210_0,
			},
			wantErr:  false,
			expectPE: true,
		},
		{
			name: "RPP case with workflow IDs",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123"},
				WorkflowIDs:      []string{"rpp-wf-1", "rpp-wf-2"},
				DeployTemplate:   "UPDATE rpp_workflow SET state = 210 WHERE run_id IN (%s) AND workflow_id IN (%s);",
				RollbackTemplate: "UPDATE rpp_workflow SET state = 200 WHERE run_id IN (%s) AND workflow_id IN (%s);",
				TargetDB:         "RPP",
				CaseType:         domain.CaseRppNoResponseResume,
			},
			wantErr:   false,
			expectRPP: true,
		},
		{
			name: "Thought Machine False Negative case (individual statements)",
			ticket: DMLTicket{
				RunIDs:           []string{"run-123", "run-456"},
				PrevTransID:      "original-prev-trans",
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = %s;",
				RollbackTemplate: "UPDATE workflow SET prev_trans_id = '%s' WHERE run_id = %s;",
				TargetDB:         "PC",
				CaseType:         domain.CaseThoughtMachineFalseNegative,
			},
			wantErr:  false,
			expectPC: true,
		},
		{
			name: "empty run IDs",
			ticket: DMLTicket{
				RunIDs:   []string{},
				TargetDB: "PC",
			},
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name: "invalid target DB",
			ticket: DMLTicket{
				RunIDs:   []string{"run-123"},
				TargetDB: "INVALID",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements, err := generateSQLFromTicket(tt.ticket)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				// Check that statements are added to the correct database
				if tt.expectPC {
					assert.NotEmpty(t, statements.PCDeployStatements)
					assert.NotEmpty(t, statements.PCRollbackStatements)
					assert.Empty(t, statements.PEDeployStatements)
					assert.Empty(t, statements.PERollbackStatements)
					assert.Empty(t, statements.RPPDeployStatements)
					assert.Empty(t, statements.RPPRollbackStatements)
				}

				if tt.expectPE {
					assert.Empty(t, statements.PCDeployStatements)
					assert.Empty(t, statements.PCRollbackStatements)
					assert.NotEmpty(t, statements.PEDeployStatements)
					assert.NotEmpty(t, statements.PERollbackStatements)
					assert.Empty(t, statements.RPPDeployStatements)
					assert.Empty(t, statements.RPPRollbackStatements)
				}

				if tt.expectRPP {
					assert.Empty(t, statements.PCDeployStatements)
					assert.Empty(t, statements.PCRollbackStatements)
					assert.Empty(t, statements.PEDeployStatements)
					assert.Empty(t, statements.PERollbackStatements)
					assert.NotEmpty(t, statements.RPPDeployStatements)
					assert.NotEmpty(t, statements.RPPRollbackStatements)
				}
			}
		})
	}
}

func TestFormatIDsForSQL(t *testing.T) {
	tests := []struct {
		name     string
		ids      []string
		expected string
	}{
		{
			name:     "single ID",
			ids:      []string{"test-123"},
			expected: "'test-123'",
		},
		{
			name:     "multiple IDs",
			ids:      []string{"test-1", "test-2", "test-3"},
			expected: "'test-1', 'test-2', 'test-3'",
		},
		{
			name:     "empty slice",
			ids:      []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatIDsForSQL(tt.ids)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		template string
		wantErr  bool
	}{
		{
			name:     "valid SQL with no placeholders",
			sql:      "UPDATE workflow SET state = 201 WHERE run_id = 'test';",
			template: "UPDATE workflow SET state = 201 WHERE run_id = 'test';",
			wantErr:  false,
		},
		{
			name:     "valid SQL with substituted placeholders",
			sql:      "UPDATE workflow SET state = 201 WHERE run_id IN ('test1', 'test2');",
			template: "UPDATE workflow SET state = 201 WHERE run_id IN (%s);",
			wantErr:  false,
		},
		{
			name:     "SQL with unsubstituted placeholders",
			sql:      "UPDATE workflow SET state = 201 WHERE run_id IN (%s);",
			template: "UPDATE workflow SET state = 201 WHERE run_id IN (%s);",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSQL(tt.sql, tt.template)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckIndividualStatementNeeded(t *testing.T) {
	tests := []struct {
		name     string
		caseType domain.Case
		expected bool
	}{
		{
			name:     "case with prev_trans_id",
			caseType: domain.CaseThoughtMachineFalseNegative,
			expected: true,
		},
		{
			name:     "case without prev_trans_id",
			caseType: domain.CasePcExternalPaymentFlow200_11,
			expected: false,
		},
		{
			name:     "unknown case",
			caseType: domain.Case("UnknownCase"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkIndividualStatementNeeded(tt.caseType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
