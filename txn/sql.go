package txn

import (
	"fmt"
	"os"
	"strings"
)

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SQLStatements contains the deploy and rollback SQL statements separated by database
type SQLStatements struct {
	PCDeployStatements    []string
	PCRollbackStatements  []string
	PEDeployStatements    []string
	PERollbackStatements  []string
	RPPDeployStatements   []string
	RPPRollbackStatements []string
}

// DMLTicket represents a SQL generation request with templates
type DMLTicket struct {
	RunIDs           []string // run_ids to update
	ReqBizMsgIDs     []string // optional req_biz_msg_ids for RPP cases
	PartnerTxIDs     []string // optional partner_tx_ids for RPP cases
	DeployTemplate   string   // SQL template for deploy
	RollbackTemplate string   // SQL template for rollback
	TargetDB         string   // "PC", "PE", or "RPP"
	WorkflowID       string   // optional workflow_id filter
	TargetState      int      // target state to check in WHERE clause
	TargetAttempt    int      // target attempt to check in WHERE clause
	StateField       string   // field name for state in WHERE clause (usually "state")
	WorkflowIDs      []string // multiple workflow_ids for IN clause

	// Consolidation metadata
	TransactionCount int // Number of transactions consolidated
}

// TemplateConfig defines the parameters required for a SQL template
type TemplateConfig struct {
	Parameters []string // List of parameter types: ["run_ids"], ["run_ids", "workflow_ids"]
}

// templateConfigs maps SOP cases to their template parameter configurations
var templateConfigs = map[SOPCase]TemplateConfig{
	SOPCasePcExternalPaymentFlow200_11:      {Parameters: []string{"run_ids"}},
	SOPCasePcExternalPaymentFlow201_0RPP210: {Parameters: []string{"run_ids"}},
	SOPCasePcExternalPaymentFlow201_0RPP900: {Parameters: []string{"run_ids"}},
	SOPCasePeTransferPayment210_0:           {Parameters: []string{"run_ids"}},
	SOPCaseRppCashoutReject101_19:           {Parameters: []string{"run_ids"}},
	SOPCaseRppQrPaymentReject210_0:          {Parameters: []string{"run_ids"}},
}

// sqlTemplates maps SOP cases to their DML tickets
var sqlTemplates = map[SOPCase]func(TransactionResult) *DMLTicket{
	SOPCasePcExternalPaymentFlow200_11: func(result TransactionResult) *DMLTicket {
		if runID := getPcExtPayment200_11RunID(result); runID != "" {
			return &DMLTicket{
				RunIDs: []string{runID},
				DeployTemplate: `-- pc_external_payment_flow_200_11
UPDATE workflow_execution
SET state = 202,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.StreamResp', JSON_OBJECT(
        'TxID', '',
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ExternalID', '',
        'ErrorMessage', 'Reject from adapter'),
      '$.State', 202)
WHERE run_id IN (%s)
AND state = 200
AND attempt = 11;`,
				RollbackTemplate: `UPDATE workflow_execution
SET state = 200,
    attempt = 11,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 200)
WHERE run_id IN (%s);`,
				TargetDB:      "PC",
				WorkflowID:    "pc_external_payment_flow",
				TargetState:   200,
				TargetAttempt: 11,
			}
		}
		return nil
	},
	SOPCasePcExternalPaymentFlow201_0RPP210: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPWorkflow.RunID},
			DeployTemplate: `-- RPP 210, PE 220, PC 201. No response from RPP. Move to 222 to resume. ACSP
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 222)
WHERE run_id IN (%s)
AND state = 210;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 201,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 201)
WHERE run_id IN (%s);`,
			TargetDB:      "RPP",
			TargetState:   210,
			TargetAttempt: 0,
		}
	},
	SOPCasePcExternalPaymentFlow201_0RPP900: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPWorkflow.RunID},
			DeployTemplate: `-- RPP 900, PE 220, PC 201. Republish from RPP to resume. ACSP
UPDATE workflow_execution
SET state = 301,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 301)
WHERE run_id IN (%s)
AND state = 900;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 900,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 900)
WHERE run_id IN (%s);`,
			TargetDB:      "RPP",
			TargetState:   900,
			TargetAttempt: 0,
		}
	},
	SOPCasePeTransferPayment210_0: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.PaymentEngineWorkflow.RunID},
			DeployTemplate: `-- Reject PE stuck 210. Reject transactions since it hasn't reached Paynet yet
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.StreamMessage', JSON_OBJECT(
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ErrorMessage', 'Manual Rejected'),
      '$.State', 221)
WHERE run_id IN (%s)
AND workflow_id = 'workflow_transfer_payment'
AND state = 210;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.StreamMessage', NULL,
      '$.State', 210)
WHERE run_id IN (%s)
AND workflow_id = 'workflow_transfer_payment';`,
			TargetDB:      "PE",
			WorkflowID:    "workflow_transfer_payment",
			TargetState:   210,
			TargetAttempt: 0,
		}
	},
	SOPCaseRppCashoutReject101_19: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPWorkflow.RunID},
			DeployTemplate: `-- rpp_cashout_reject_101_19, publish FAILED status
UPDATE workflow_execution
SET state = 311,
    attempt = 1,
    data = JSON_SET(data, '$.State', 311)
WHERE run_id IN (%s)
AND state = 101
AND workflow_id = 'wf_ct_cashout';`,
			RollbackTemplate: `-- RPP Rollback: Move workflows back to state 101
UPDATE workflow_execution
SET state = 101,
    attempt = 0,
    data = JSON_SET(data, '$.State', 101)
WHERE run_id IN (%s)
AND workflow_id = 'wf_ct_cashout';`,
			TargetDB:      "RPP",
			WorkflowID:    "'wf_ct_cashout'",
			TargetState:   101,
			TargetAttempt: 19,
		}
	},
	SOPCaseRppQrPaymentReject210_0: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPWorkflow.RunID},
			DeployTemplate: `-- rpp_qr_payment_reject_210_0, manual reject
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id IN (%s)
AND state = 210
AND workflow_id = 'wf_ct_qr_payment';`,
			RollbackTemplate: `-- RPP Rollback: Move qr_payment workflows back to state 210
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    data = JSON_SET(data, '$.State', 210)
WHERE run_id IN (%s)
AND workflow_id = 'wf_ct_qr_payment';`,
			TargetDB:      "RPP",
			WorkflowID:    "'wf_ct_qr_payment'",
			TargetState:   210,
			TargetAttempt: 0,
		}
	},
}

// getCaseTypeFromTicket determines the SOP case type based on ticket characteristics
func getCaseTypeFromTicket(ticket DMLTicket) SOPCase {
	// Match based on template content and other characteristics
	if strings.Contains(ticket.DeployTemplate, "pc_external_payment_flow_200_11") {
		return SOPCasePcExternalPaymentFlow200_11
	}
	if strings.Contains(ticket.DeployTemplate, "state = 222") {
		return SOPCasePcExternalPaymentFlow201_0RPP210
	}
	if strings.Contains(ticket.DeployTemplate, "state = 301") {
		return SOPCasePcExternalPaymentFlow201_0RPP900
	}
	if strings.Contains(ticket.DeployTemplate, "workflow_transfer_payment") {
		return SOPCasePeTransferPayment210_0
	}
	if strings.Contains(ticket.DeployTemplate, "state = 311") {
		return SOPCaseRppCashoutReject101_19
	}
	if strings.Contains(ticket.DeployTemplate, "state = 221") && strings.Contains(ticket.DeployTemplate, "wf_ct_qr_payment") {
		return SOPCaseRppQrPaymentReject210_0
	}
	return SOPCaseNone
}

// generateSQLFromTicket generates SQL statements from a DML ticket
func generateSQLFromTicket(ticket DMLTicket) SQLStatements {
	if len(ticket.RunIDs) == 0 {
		return SQLStatements{}
	}

	// Create comma-separated list of quoted IDs
	quotedIDs := make([]string, len(ticket.RunIDs))
	for i, id := range ticket.RunIDs {
		quotedIDs[i] = fmt.Sprintf("'%s'", id)
	}
	inClause := strings.Join(quotedIDs, ", ")

	statements := SQLStatements{}

	// Generate deploy and rollback statements
	var deploySQL, rollbackSQL string

	// Use explicit config instead of string counting
	caseType := getCaseTypeFromTicket(ticket)
	config, exists := templateConfigs[caseType]
	if !exists {
		// Default to run_ids only
		deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause)
		rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause)
	} else {
		// Generate based on parameter types
		if len(config.Parameters) == 2 {
			// Check if it's workflow_id (single) or workflow_ids (multiple)
			if contains(config.Parameters, "workflow_ids") {
				workflowIDsClause := strings.Join(ticket.WorkflowIDs, ", ")
				deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause, workflowIDsClause)
				rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause, workflowIDsClause)
			} else if contains(config.Parameters, "workflow_id") {
				deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause, ticket.WorkflowID)
				rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause, ticket.WorkflowID)
			} else {
				deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause)
				rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause)
			}
		} else {
			deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause)
			rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause)
		}
	}

	// Add to appropriate statement list based on target DB
	switch ticket.TargetDB {
	case "PC":
		statements.PCDeployStatements = append(statements.PCDeployStatements, deploySQL)
		statements.PCRollbackStatements = append(statements.PCRollbackStatements, rollbackSQL)
	case "PE":
		statements.PEDeployStatements = append(statements.PEDeployStatements, deploySQL)
		statements.PERollbackStatements = append(statements.PERollbackStatements, rollbackSQL)
	case "RPP":
		statements.RPPDeployStatements = append(statements.RPPDeployStatements, deploySQL)
		statements.RPPRollbackStatements = append(statements.RPPRollbackStatements, rollbackSQL)
	}

	return statements
}

// GenerateSQLStatements generates SQL statements for all supported cases using templates.
func GenerateSQLStatements(results []TransactionResult) SQLStatements {
	statements := SQLStatements{}

	fmt.Println("\n--- Generating SQL Statements ---")

	// Use map[SOPCase]DMLTicket for automatic consolidation
	caseTickets := make(map[SOPCase]DMLTicket)

	for i := range results {
		// identifySOPCase now takes a pointer and might trigger prompts
		caseType := identifySOPCase(&results[i])
		results[i].RPPInfo = string(caseType) // Store identified case for reference

		// Get the template function for this case
		if templateFunc, exists := sqlTemplates[caseType]; exists {
			newTicket := templateFunc(results[i])
			if newTicket != nil && len(newTicket.RunIDs) > 0 {
				if existingTicket, exists := caseTickets[caseType]; exists {
					// Merge IDs into existing ticket
					existingTicket.RunIDs = append(existingTicket.RunIDs, newTicket.RunIDs...)
					if len(newTicket.ReqBizMsgIDs) > 0 {
						existingTicket.ReqBizMsgIDs = append(existingTicket.ReqBizMsgIDs, newTicket.ReqBizMsgIDs...)
					}
					if len(newTicket.PartnerTxIDs) > 0 {
						existingTicket.PartnerTxIDs = append(existingTicket.PartnerTxIDs, newTicket.PartnerTxIDs...)
					}
					existingTicket.TransactionCount++
					caseTickets[caseType] = existingTicket
				} else {
					// Create new ticket with counter
					newTicket.TransactionCount = 1
					caseTickets[caseType] = *newTicket
				}
			}
		}
	}

	// Process each consolidated ticket
	for caseType, ticket := range caseTickets {
		if len(ticket.RunIDs) > 0 {
			fmt.Printf("Generating SQL for %s with %d transactions\n", caseType, ticket.TransactionCount)
			generatedSQL := generateSQLFromTicket(ticket)
			appendStatements(&statements, generatedSQL)
		}
	}

	return statements
}

// appendStatements is a helper to merge results into the main struct
func appendStatements(main *SQLStatements, new SQLStatements) {
	main.PCDeployStatements = append(main.PCDeployStatements, new.PCDeployStatements...)
	main.PCRollbackStatements = append(main.PCRollbackStatements, new.PCRollbackStatements...)
	main.PEDeployStatements = append(main.PEDeployStatements, new.PEDeployStatements...)
	main.PERollbackStatements = append(main.PERollbackStatements, new.PERollbackStatements...)
	main.RPPDeployStatements = append(main.RPPDeployStatements, new.RPPDeployStatements...)
	main.RPPRollbackStatements = append(main.RPPRollbackStatements, new.RPPRollbackStatements...)
}

// getPcExtPayment200_11RunID extracts the relevant run_id for Case 1
func getPcExtPayment200_11RunID(result TransactionResult) string {
	for _, w := range result.PaymentCoreWorkflows {
		if w.Type == "external_payment_flow" && w.State == "200" && w.Attempt == 11 {
			return w.RunID
		}
	}
	return ""
}

// WriteSQLFiles writes the SQL statements to database-specific Deploy.sql and Rollback.sql files
func WriteSQLFiles(statements SQLStatements, basePath string) error {
	// Write PC files
	if len(statements.PCDeployStatements) > 0 {
		deployPath := basePath + "_PC_Deploy.sql"
		if err := writeSQLFile(deployPath, statements.PCDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PCRollbackStatements) > 0 {
		rollbackPath := basePath + "_PC_Rollback.sql"
		if err := writeSQLFile(rollbackPath, statements.PCRollbackStatements); err != nil {
			return err
		}
	}

	// Write PE files
	if len(statements.PEDeployStatements) > 0 {
		deployPath := basePath + "_PE_Deploy.sql"
		if err := writeSQLFile(deployPath, statements.PEDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PERollbackStatements) > 0 {
		rollbackPath := basePath + "_PE_Rollback.sql"
		if err := writeSQLFile(rollbackPath, statements.PERollbackStatements); err != nil {
			return err
		}
	}

	// Write RPP files
	if len(statements.RPPDeployStatements) > 0 {
		deployPath := basePath + "_RPP_Deploy.sql"
		if err := writeSQLFile(deployPath, statements.RPPDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.RPPRollbackStatements) > 0 {
		rollbackPath := basePath + "_RPP_Rollback.sql"
		if err := writeSQLFile(rollbackPath, statements.RPPRollbackStatements); err != nil {
			return err
		}
	}

	return nil
}

// writeSQLFile writes SQL statements to a file
func writeSQLFile(filePath string, statements []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close SQL file %s: %v\n", filePath, err)
		}
	}()

	for _, stmt := range statements {
		if _, err := fmt.Fprintln(file, stmt); err != nil {
			fmt.Printf("Warning: failed to write SQL statement: %v\n", err)
		}
		if _, err := fmt.Fprintln(file); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}
	fmt.Printf("SQL statements written to %s\n", filePath)
	return nil
}
