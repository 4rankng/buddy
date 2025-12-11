package txn

import (
	"fmt"
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
	if strings.Contains(ticket.DeployTemplate, "state = 222") && strings.Contains(ticket.DeployTemplate, "rpp_no_response_resume_acsp") {
		return SOPCaseRppNoResponseResume
	}
	return SOPCaseNone
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
