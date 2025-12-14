package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"strings"
)

// SQLFuncBuilder generates SQL statements from a DML ticket
type SQLFuncBuilder func(ticket DMLTicket) (deploySQL string, rollbackSQL string, err error)

// SQLBuilderConfig maps case types to their SQL builders
var sqlBuilderConfigs = map[domain.Case]SQLFuncBuilder{
	domain.CasePcExternalPaymentFlow200_11:      buildRunIDsOnlySQL,
	domain.CasePcExternalPaymentFlow201_0RPP210: buildRunIDsOnlySQL,
	domain.CasePcExternalPaymentFlow201_0RPP900: buildRunIDsOnlySQL,
	domain.CasePeTransferPayment210_0:           buildRunIDsOnlySQL,
	domain.CasePeStuck230RepublishPC:            buildRunIDsOnlySQL,
	domain.CasePe2200FastCashinFailed:           buildRunIDsOnlySQL,
	domain.CaseRppCashoutReject101_19:           buildRunIDsOnlySQL,
	domain.CaseRppQrPaymentReject210_0:          buildRunIDsOnlySQL,
	domain.CaseRppNoResponseResume:              buildRunIDsWithWorkflowIDsSQL,
	domain.CaseThoughtMachineFalseNegative:      buildRunIDsOnlyWithPrevTransIDSQL,
}

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

// SQL Builder Functions

// buildRunIDsOnlySQL handles templates with only run_ids parameter (batch)
func buildRunIDsOnlySQL(ticket DMLTicket) (deploySQL, rollbackSQL string, err error) {
	// Validate input
	if len(ticket.RunIDs) == 0 {
		return "", "", fmt.Errorf("ticket contains no run IDs")
	}

	// Format IDs for SQL
	inClause := formatIDsForSQL(ticket.RunIDs)

	// Generate SQL using templates
	deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause)
	rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause)

	return deploySQL, rollbackSQL, nil
}

// buildRunIDsWithWorkflowIDSQL handles templates with run_ids and single workflow_id (batch)
func buildRunIDsWithWorkflowIDSQL(ticket DMLTicket) (deploySQL, rollbackSQL string, err error) {
	// Validate input
	if len(ticket.RunIDs) == 0 {
		return "", "", fmt.Errorf("ticket contains no run IDs")
	}
	if ticket.WorkflowID == "" {
		return "", "", fmt.Errorf("workflow ID required for case type %s", ticket.CaseType)
	}

	// Format IDs for SQL
	inClause := formatIDsForSQL(ticket.RunIDs)

	// Generate SQL using templates
	deploySQL = fmt.Sprintf(ticket.DeployTemplate, inClause, ticket.WorkflowID)
	rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, inClause, ticket.WorkflowID)

	return deploySQL, rollbackSQL, nil
}

// buildRunIDsWithWorkflowIDsSQL handles templates with run_ids and multiple workflow_ids (batch)
func buildRunIDsWithWorkflowIDsSQL(ticket DMLTicket) (deploySQL, rollbackSQL string, err error) {
	// Validate input
	if len(ticket.RunIDs) == 0 {
		return "", "", fmt.Errorf("ticket contains no run IDs")
	}
	if len(ticket.WorkflowIDs) == 0 {
		return "", "", fmt.Errorf("workflow IDs required for case type %s", ticket.CaseType)
	}

	// Format IDs for SQL
	runIDsClause := formatIDsForSQL(ticket.RunIDs)
	workflowIDsClause := strings.Join(ticket.WorkflowIDs, ", ")

	// Generate SQL using templates
	deploySQL = fmt.Sprintf(ticket.DeployTemplate, runIDsClause, workflowIDsClause)
	rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, runIDsClause, workflowIDsClause)

	return deploySQL, rollbackSQL, nil
}

// buildRunIDsOnlyWithPrevTransIDSQL handles templates with run_ids and prev_trans_id (individual)
func buildRunIDsOnlyWithPrevTransIDSQL(ticket DMLTicket) (deploySQL, rollbackSQL string, err error) {
	// Validate input
	if len(ticket.RunIDs) == 0 {
		return "", "", fmt.Errorf("ticket contains no run IDs")
	}

	// For prev_trans_id cases, we need to generate SQL for each run ID individually
	// We'll handle this in the generateSQLFromTicket function
	// This function will be called for each individual run ID
	if len(ticket.RunIDs) != 1 {
		return "", "", fmt.Errorf("prev_trans_id cases require individual run ID processing")
	}

	// Format the single run ID
	quotedID := fmt.Sprintf("'%s'", ticket.RunIDs[0])

	// Generate SQL using templates
	deploySQL = fmt.Sprintf(ticket.DeployTemplate, quotedID)
	rollbackSQL = fmt.Sprintf(ticket.RollbackTemplate, ticket.PrevTransID, quotedID)

	return deploySQL, rollbackSQL, nil
}

// Helper Functions for SQL Generation

// formatIDsForSQL converts a slice of IDs to a comma-separated, quoted string
func formatIDsForSQL(ids []string) string {
	quotedIDs := make([]string, len(ids))
	for i, id := range ids {
		quotedIDs[i] = fmt.Sprintf("'%s'", id)
	}
	return strings.Join(quotedIDs, ", ")
}

// validateSQL checks if the generated SQL matches expected template structure
func validateSQL(sql, template string) error {
	// Count placeholders in template
	placeholderCount := strings.Count(template, "%s")

	// Basic validation - ensure all placeholders are substituted
	if placeholderCount > 0 && strings.Contains(sql, "%s") {
		return fmt.Errorf("SQL contains unsubstituted placeholders")
	}

	return nil
}

// checkIndividualStatementNeeded determines if a case requires individual statements
func checkIndividualStatementNeeded(caseType domain.Case) bool {
	config, exists := templateConfigs[caseType]
	if !exists {
		return false
	}
	return contains(config.Parameters, "prev_trans_id")
}

// generateSQLFromTicket generates SQL statements from a DML ticket using configuration-driven approach
func generateSQLFromTicket(ticket DMLTicket) (SQLStatements, error) {
	// Validate input
	if len(ticket.RunIDs) == 0 {
		return SQLStatements{}, fmt.Errorf("ticket contains no run IDs")
	}

	// Validate target DB
	if ticket.TargetDB != "PC" && ticket.TargetDB != "PE" && ticket.TargetDB != "RPP" {
		return SQLStatements{}, fmt.Errorf("unknown target database: %s", ticket.TargetDB)
	}

	statements := SQLStatements{}

	// Check if we need individual statements (for prev_trans_id cases)
	if checkIndividualStatementNeeded(ticket.CaseType) {
		// Generate individual statements for each run ID
		for _, runID := range ticket.RunIDs {
			// Create a single-run ticket for individual processing
			singleTicket := ticket
			singleTicket.RunIDs = []string{runID}

			// Get the SQL builder for this case type
			builder, exists := sqlBuilderConfigs[ticket.CaseType]
			if !exists {
				// Default to simple run_ids only builder
				builder = buildRunIDsOnlySQL
			}

			// Generate SQL using the builder
			deploySQL, rollbackSQL, err := builder(singleTicket)
			if err != nil {
				return SQLStatements{}, fmt.Errorf("failed to generate SQL for case %s, run ID %s: %w", ticket.CaseType, runID, err)
			}

			// Validate generated SQL
			if err := validateSQL(deploySQL, ticket.DeployTemplate); err != nil {
				return SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
			}
			if err := validateSQL(rollbackSQL, ticket.RollbackTemplate); err != nil {
				return SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
			}

			// Add to appropriate statement list based on target DB
			addStatementToDatabase(&statements, ticket.TargetDB, deploySQL, rollbackSQL)
		}
	} else {
		// Generate batch statement for all run IDs
		// Get the SQL builder for this case type
		builder, exists := sqlBuilderConfigs[ticket.CaseType]
		if !exists {
			// Default to simple run_ids only builder
			builder = buildRunIDsOnlySQL
		}

		// Generate SQL using the builder
		deploySQL, rollbackSQL, err := builder(ticket)
		if err != nil {
			return SQLStatements{}, fmt.Errorf("failed to generate SQL for case %s: %w", ticket.CaseType, err)
		}

		// Validate generated SQL
		if err := validateSQL(deploySQL, ticket.DeployTemplate); err != nil {
			return SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
		}
		if err := validateSQL(rollbackSQL, ticket.RollbackTemplate); err != nil {
			return SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
		}

		// Add to appropriate statement list based on target DB
		addStatementToDatabase(&statements, ticket.TargetDB, deploySQL, rollbackSQL)
	}

	return statements, nil
}

// addStatementToDatabase adds SQL statements to the appropriate database section
func addStatementToDatabase(statements *SQLStatements, targetDB string, deploySQL, rollbackSQL string) {
	switch targetDB {
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
