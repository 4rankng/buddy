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
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s';",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id = '%s';",
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
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s';",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id = '%s';",
			},
			wantErr: false, // buildSQLFromTemplate doesn't validate parameters
			validate: func(t *testing.T, deploy, rollback string) {
				// With empty parameters, %s will be replaced with !MISSING
				assert.Contains(t, deploy, "UPDATE workflow SET state = 201 WHERE run_id = '!MISSING';")
				assert.Contains(t, rollback, "UPDATE workflow SET state = 200 WHERE run_id = '!MISSING';")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deploy, rollback string
			var err error

			deploy, err = buildSQLFromTemplate(tt.ticket.DeployTemplate, tt.ticket.DeployParams)
			if err != nil && !tt.wantErr {
				t.Errorf("buildSQLFromTemplate for deploy failed: %v", err)
				return
			}

			if !tt.wantErr {
				rollback, err = buildSQLFromTemplate(tt.ticket.RollbackTemplate, tt.ticket.RollbackParams)
				if err != nil {
					t.Errorf("buildSQLFromTemplate for rollback failed: %v", err)
					return
				}
			}

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
		ticket   domain.DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "valid ticket with workflow ID",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
					{Name: "workflow_id", Value: "workflow-456", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
					{Name: "workflow_id", Value: "workflow-456", Type: "string"},
				},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s' AND workflow_id = '%s';",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id = '%s' AND workflow_id = '%s';",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, deploy, "'workflow-456'")
				assert.Contains(t, rollback, "'run-123'")
				assert.Contains(t, rollback, "'workflow-456'")
			},
		},
		{
			name: "missing workflow ID parameter",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s' AND workflow_id = '%s';",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id = '%s' AND workflow_id = '%s';",
			},
			wantErr: false, // buildSQLFromTemplate doesn't validate parameter count
			validate: func(t *testing.T, deploy, rollback string) {
				// With only run_id parameter, the second %s will be replaced with !MISSING
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, deploy, "!MISSING")
				assert.Contains(t, rollback, "'run-123'")
				assert.Contains(t, rollback, "!MISSING")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deploy, rollback string
			var err error

			deploy, err = buildSQLFromTemplate(tt.ticket.DeployTemplate, tt.ticket.DeployParams)
			if err != nil && !tt.wantErr {
				t.Errorf("buildSQLFromTemplate for deploy failed: %v", err)
				return
			}

			if !tt.wantErr {
				rollback, err = buildSQLFromTemplate(tt.ticket.RollbackTemplate, tt.ticket.RollbackParams)
				if err != nil {
					t.Errorf("buildSQLFromTemplate for rollback failed: %v", err)
					return
				}
			}

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
		ticket   domain.DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "valid ticket with workflow IDs hardcoded",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s' AND workflow_id IN ('wf-1', 'wf-2', 'wf-3');",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id = '%s' AND workflow_id IN ('wf-1', 'wf-2', 'wf-3');",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, deploy, "'wf-1', 'wf-2', 'wf-3'")
				assert.Contains(t, rollback, "'run-123'")
				assert.Contains(t, rollback, "'wf-1', 'wf-2', 'wf-3'")
			},
		},
		{
			name: "missing run ID parameter",
			ticket: domain.DMLTicket{
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s' AND workflow_id IN ('wf-1', 'wf-2');",
				RollbackTemplate: "UPDATE workflow SET state = 200 WHERE run_id = '%s' AND workflow_id IN ('wf-1', 'wf-2');",
			},
			wantErr: false, // buildSQLFromTemplate doesn't validate parameter count
			validate: func(t *testing.T, deploy, rollback string) {
				// With missing run_id parameter, %s will be replaced with !MISSING
				assert.Contains(t, deploy, "!MISSING")
				assert.Contains(t, rollback, "!MISSING")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deploy, rollback string
			var err error

			deploy, err = buildSQLFromTemplate(tt.ticket.DeployTemplate, tt.ticket.DeployParams)
			if err != nil && !tt.wantErr {
				t.Errorf("buildSQLFromTemplate for deploy failed: %v", err)
				return
			}

			if !tt.wantErr {
				rollback, err = buildSQLFromTemplate(tt.ticket.RollbackTemplate, tt.ticket.RollbackParams)
				if err != nil {
					t.Errorf("buildSQLFromTemplate for rollback failed: %v", err)
					return
				}
			}

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
		ticket   domain.DMLTicket
		wantErr  bool
		validate func(*testing.T, string, string)
	}{
		{
			name: "valid ticket with prev_trans_id",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "prev_trans_id", Value: "prev-456", Type: "string"},
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s';",
				RollbackTemplate: "UPDATE workflow SET prev_trans_id = '%s' WHERE run_id = '%s';",
			},
			wantErr: false,
			validate: func(t *testing.T, deploy, rollback string) {
				assert.Contains(t, deploy, "'run-123'")
				assert.Contains(t, rollback, "'prev-456'")
				assert.Contains(t, rollback, "'run-123'")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deploy, rollback string
			var err error

			deploy, err = buildSQLFromTemplate(tt.ticket.DeployTemplate, tt.ticket.DeployParams)
			if err != nil && !tt.wantErr {
				t.Errorf("buildSQLFromTemplate for deploy failed: %v", err)
				return
			}

			if !tt.wantErr {
				rollback, err = buildSQLFromTemplate(tt.ticket.RollbackTemplate, tt.ticket.RollbackParams)
				if err != nil {
					t.Errorf("buildSQLFromTemplate for rollback failed: %v", err)
					return
				}
			}

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
		ticket      domain.DMLTicket
		wantErr     bool
		expectPC    bool
		expectPE    bool
		expectRPP   bool
		expectEmpty bool
	}{
		{
			name: "PC case with single run ID",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE pc_workflow SET state = 201 WHERE run_id = '%s';",
				RollbackTemplate: "UPDATE pc_workflow SET state = 200 WHERE run_id = '%s';",
				TargetDB:         "PC",
				CaseType:         domain.CasePcExternalPaymentFlow200_11,
			},
			wantErr:  false,
			expectPC: true,
		},
		{
			name: "PE case with workflow ID",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE pe_workflow SET state = 221 WHERE run_id = '%s' AND workflow_id = 'transfer_payment';",
				RollbackTemplate: "UPDATE pe_workflow SET state = 210 WHERE run_id = '%s' AND workflow_id = 'transfer_payment';",
				TargetDB:         "PE",
				CaseType:         domain.CasePeTransferPayment210_0,
			},
			wantErr:  false,
			expectPE: true,
		},
		{
			name: "RPP case with workflow IDs hardcoded",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE rpp_workflow SET state = 210 WHERE run_id = '%s' AND workflow_id IN ('wf-1', 'wf-2');",
				RollbackTemplate: "UPDATE rpp_workflow SET state = 200 WHERE run_id = '%s' AND workflow_id IN ('wf-1', 'wf-2');",
				TargetDB:         "RPP",
				CaseType:         domain.CaseRppNoResponseResume,
			},
			wantErr:   false,
			expectRPP: true,
		},
		{
			name: "Thought Machine False Negative case",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "prev_trans_id", Value: "original-prev-trans", Type: "string"},
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				DeployTemplate:   "UPDATE workflow SET state = 201 WHERE run_id = '%s';",
				RollbackTemplate: "UPDATE workflow SET prev_trans_id = '%s' WHERE run_id = '%s';",
				TargetDB:         "PC",
				CaseType:         domain.CaseThoughtMachineFalseNegative,
			},
			wantErr:  false,
			expectPC: true,
		},
		{
			name: "empty run IDs",
			ticket: domain.DMLTicket{
				TargetDB: "PC",
			},
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name: "invalid target DB",
			ticket: domain.DMLTicket{
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: "run-123", Type: "string"},
				},
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

func TestFormatParameter(t *testing.T) {
	tests := []struct {
		name     string
		info     domain.ParamInfo
		expected string
	}{
		{
			name: "single string ID",
			info: domain.ParamInfo{
				Name:  "run_id",
				Value: "test-123",
				Type:  "string",
			},
			expected: "'test-123'",
		},
		{
			name: "int value",
			info: domain.ParamInfo{
				Name:  "count",
				Value: 42,
				Type:  "int",
			},
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatParameter(tt.info)
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
