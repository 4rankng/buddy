package adapters

import (
	"buddy/internal/txn/domain"
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

// appendStatements is a helper to merge results into main struct
func appendStatements(main *SQLStatements, new SQLStatements) {
	main.PCDeployStatements = append(main.PCDeployStatements, new.PCDeployStatements...)
	main.PCRollbackStatements = append(main.PCRollbackStatements, new.PCRollbackStatements...)
	main.PEDeployStatements = append(main.PEDeployStatements, new.PEDeployStatements...)
	main.PERollbackStatements = append(main.PERollbackStatements, new.PERollbackStatements...)
	main.RPPDeployStatements = append(main.RPPDeployStatements, new.RPPDeployStatements...)
	main.RPPRollbackStatements = append(main.RPPRollbackStatements, new.RPPRollbackStatements...)
}

// getPcExtPayment200_11RunID extracts the relevant run_id for Case 1
func getPcExtPayment200_11RunID(result domain.TransactionResult) string {
	for _, w := range result.PaymentCore.Workflow {
		if w.WorkflowID == "external_payment_flow" && w.State == "200" && w.Attempt == 11 {
			return w.RunID
		}
	}
	return ""
}

// getInternalPaymentFlowRunIDs returns all internal_payment_flow run IDs for a transaction result
func getInternalPaymentFlowRunIDs(result domain.TransactionResult) []string {
	var runIDs []string
	for _, w := range result.PaymentCore.Workflow {
		if w.WorkflowID == "internal_payment_flow" && w.RunID != "" {
			runIDs = append(runIDs, w.RunID)
		}
	}
	return runIDs
}

// generateSQLFromTicket generates SQL statements from a DML ticket
func generateSQLFromTicket(ticket DMLTicket) SQLStatements {
	if len(ticket.RunIDs) == 0 {
		return SQLStatements{}
	}

	statements := SQLStatements{}

	// Use CaseType from ticket instead of a parameter
	config, exists := templateConfigs[ticket.CaseType]
	
	// Check if we need individual statements (for prev_trans_id cases)
	needsIndividualStatements := exists && contains(config.Parameters, "prev_trans_id")
	
	if needsIndividualStatements {
		// Generate individual statements for each run ID
		for _, runID := range ticket.RunIDs {
			quotedID := fmt.Sprintf("'%s'", runID)
			
			// Get the original prev_trans_id for this specific run_id
			// For thought_machine_false_negative, we need to query the current prev_trans_id
			// This would be done before creating the ticket, so we assume it's stored in the ticket
			var prevTransID string
			if ticket.CaseType == domain.CaseThoughtMachineFalseNegative {
				// For rollback, we need to use the original prev_trans_id value
				// This should be populated when the ticket is created
				prevTransID = ticket.PrevTransID
			}
			
			// Generate deploy and rollback statements
			var deploySQL, rollbackSQL string
			if !exists {
				// Default to run_ids only
				deploySQL = fmt.Sprintf(ticket.DeployTemplate, quotedID)
				rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, quotedID)
			} else {
				// Generate based on parameter types
				if len(config.Parameters) == 2 {
					// Check if it's workflow_id (single), workflow_ids (multiple), or prev_trans_id
					if contains(config.Parameters, "workflow_ids") {
						workflowIDsClause := strings.Join(ticket.WorkflowIDs, ", ")
						deploySQL = fmt.Sprintf(ticket.DeployTemplate, quotedID, workflowIDsClause)
						rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, quotedID, workflowIDsClause)
					} else if contains(config.Parameters, "workflow_id") {
						deploySQL = fmt.Sprintf(ticket.DeployTemplate, quotedID, ticket.WorkflowID)
						rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, quotedID, ticket.WorkflowID)
					} else if contains(config.Parameters, "prev_trans_id") {
						deploySQL = fmt.Sprintf(ticket.DeployTemplate, quotedID)
						rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, prevTransID, quotedID)
					} else {
						deploySQL = fmt.Sprintf(ticket.DeployTemplate, quotedID)
						rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, quotedID)
					}
				} else {
					deploySQL = fmt.Sprintf(ticket.DeployTemplate, quotedID)
					rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, quotedID)
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
		}
	} else {
		// Generate batch statement for all run IDs
		// Create comma-separated list of quoted IDs
		quotedIDs := make([]string, len(ticket.RunIDs))
		for i, id := range ticket.RunIDs {
			quotedIDs[i] = fmt.Sprintf("'%s'", id)
		}
		inClause := strings.Join(quotedIDs, ", ")

		// Generate deploy and rollback statements
		var deploySQL, rollbackSQL string
		if !exists {
			// Default to run_ids only
			deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause)
			rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause)
		} else {
			// Generate based on parameter types
			if len(config.Parameters) == 2 {
				// Check if it's workflow_id (single), workflow_ids (multiple), or prev_trans_id
				if contains(config.Parameters, "workflow_ids") {
					workflowIDsClause := strings.Join(ticket.WorkflowIDs, ", ")
					deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause, workflowIDsClause)
					rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause, workflowIDsClause)
				} else if contains(config.Parameters, "workflow_id") {
					deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause, ticket.WorkflowID)
					rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause, ticket.WorkflowID)
				} else if contains(config.Parameters, "prev_trans_id") {
					deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause)
					rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, ticket.PrevTransID, inClause)
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
	}

	return statements
}