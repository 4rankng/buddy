package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"strings"
)

// appendStatements is a helper to merge results into main struct
func appendStatements(main *domain.SQLStatements, new domain.SQLStatements) {
	main.PCDeployStatements = append(main.PCDeployStatements, new.PCDeployStatements...)
	main.PCRollbackStatements = append(main.PCRollbackStatements, new.PCRollbackStatements...)
	main.PEDeployStatements = append(main.PEDeployStatements, new.PEDeployStatements...)
	main.PERollbackStatements = append(main.PERollbackStatements, new.PERollbackStatements...)
	main.PPEDeployStatements = append(main.PPEDeployStatements, new.PPEDeployStatements...)
	main.PPERollbackStatements = append(main.PPERollbackStatements, new.PPERollbackStatements...)
	main.RPPDeployStatements = append(main.RPPDeployStatements, new.RPPDeployStatements...)
	main.RPPRollbackStatements = append(main.RPPRollbackStatements, new.RPPRollbackStatements...)
}

// Helper Functions for SQL Generation

// getInternalPaymentFlowRunID extracts a single run_id for internal_payment_flow
func getInternalPaymentFlowRunID(result domain.TransactionResult) string {
	// Check InternalCapture workflow
	if result.PaymentCore != nil && result.PaymentCore.InternalCapture.Workflow.WorkflowID == "internal_payment_flow" &&
		result.PaymentCore.InternalCapture.Workflow.RunID != "" {
		return result.PaymentCore.InternalCapture.Workflow.RunID
	}

	// Check InternalAuth workflow
	if result.PaymentCore != nil && result.PaymentCore.InternalAuth.Workflow.WorkflowID == "internal_payment_flow" &&
		result.PaymentCore.InternalAuth.Workflow.RunID != "" {
		return result.PaymentCore.InternalAuth.Workflow.RunID
	}

	// Check ExternalTransfer workflow
	if result.PaymentCore != nil && result.PaymentCore.ExternalTransfer.Workflow.WorkflowID == "internal_payment_flow" &&
		result.PaymentCore.ExternalTransfer.Workflow.RunID != "" {
		return result.PaymentCore.ExternalTransfer.Workflow.RunID
	}

	return ""
}

// countPlaceholders counts %s occurrences in a template
func countPlaceholders(template string) int {
	return strings.Count(template, "%s")
}

// formatParameter formats a parameter value based on its type for SQL usage
func formatParameter(info domain.ParamInfo) string {
	switch info.Type {
	case "string":
		return fmt.Sprintf("'%v'", info.Value)
	case "int":
		return fmt.Sprintf("%v", info.Value)
	default:
		// Default to string formatting for unknown types
		return fmt.Sprintf("'%v'", info.Value)
	}
}

// buildSQLFromTemplate builds SQL from a template and parameters using positional substitution
func buildSQLFromTemplate(template string, params []domain.ParamInfo) (string, error) {
	// Format all parameters
	formattedParams := make([]interface{}, len(params))
	for i, param := range params {
		formattedParams[i] = formatParameter(param)
	}

	// Count placeholders
	placeholderCount := strings.Count(template, "%s")

	// If we have fewer parameters than placeholders, add missing placeholders
	if len(formattedParams) < placeholderCount {
		for i := len(formattedParams); i < placeholderCount; i++ {
			formattedParams = append(formattedParams, "!MISSING")
		}
	}

	// Substitute parameters in template
	sql := fmt.Sprintf(template, formattedParams...)
	return sql, nil
}

// getParamValue finds and returns the value of a parameter by name
// DEPRECATED: No longer used with the new consolidation strategy
/*
func getParamValue(params []domain.ParamInfo, name string) interface{} {
	for _, param := range params {
		if param.Name == name {
			return param.Value
		}
	}
	return nil
}

// updateParamValue creates a new parameter slice with updated value for the given parameter name
// DEPRECATED: No longer used with the new consolidation strategy
func updateParamValue(params []domain.ParamInfo, name string, newValue interface{}) []domain.ParamInfo {
	newParams := make([]domain.ParamInfo, len(params))
	copy(newParams, params)

	for i, param := range newParams {
		if param.Name == name {
			newParams[i].Value = newValue
			break
		}
	}
	return newParams
}
*/

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

// generateSQLFromTicket generates SQL statements from a DML ticket using TemplateInfo arrays
func generateSQLFromTicket(ticket domain.DMLTicket) (domain.SQLStatements, error) {
	// Validate input
	if len(ticket.Deploy) == 0 && len(ticket.Rollback) == 0 {
		return domain.SQLStatements{}, fmt.Errorf("ticket contains no templates")
	}

	statements := domain.SQLStatements{}

	// Process all deploy templates
	for _, deploy := range ticket.Deploy {
		// Validate target DB
		if deploy.TargetDB != "PC" && deploy.TargetDB != "PE" && deploy.TargetDB != "RPP" {
			return domain.SQLStatements{}, fmt.Errorf("unknown target database: %s", deploy.TargetDB)
		}

		// Count placeholders and validate parameters
		deployPlaceholders := countPlaceholders(deploy.SQLTemplate)
		if len(deploy.Params) != deployPlaceholders {
			return domain.SQLStatements{}, fmt.Errorf("deploy parameters count mismatch: need %d, got %d", deployPlaceholders, len(deploy.Params))
		}

		// Generate SQL for deploy template
		deploySQL, err := buildSQLFromTemplate(deploy.SQLTemplate, deploy.Params)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate deploy SQL for case %s: %w", ticket.CaseType, err)
		}
		if err := validateSQL(deploySQL, deploy.SQLTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
		}
		addStatementToDatabase(&statements, deploy.TargetDB, deploySQL, "")
	}

	// Process all rollback templates
	for _, rollback := range ticket.Rollback {
		// Validate target DB
		if rollback.TargetDB != "PC" && rollback.TargetDB != "PE" && rollback.TargetDB != "RPP" {
			return domain.SQLStatements{}, fmt.Errorf("unknown target database: %s", rollback.TargetDB)
		}

		// Count placeholders and validate parameters
		rollbackPlaceholders := countPlaceholders(rollback.SQLTemplate)
		if len(rollback.Params) != rollbackPlaceholders {
			return domain.SQLStatements{}, fmt.Errorf("rollback parameters count mismatch: need %d, got %d", rollbackPlaceholders, len(rollback.Params))
		}

		// Generate SQL for rollback template
		rollbackSQL, err := buildSQLFromTemplate(rollback.SQLTemplate, rollback.Params)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate rollback SQL for case %s: %w", ticket.CaseType, err)
		}
		if err := validateSQL(rollbackSQL, rollback.SQLTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
		}
		addStatementToDatabase(&statements, rollback.TargetDB, "", rollbackSQL)
	}

	return statements, nil
}

// addStatementToDatabase adds SQL statements to the appropriate database section
func addStatementToDatabase(statements *domain.SQLStatements, targetDB string, deploySQL, rollbackSQL string) {
	switch targetDB {
	case "PC":
		if deploySQL != "" {
			statements.PCDeployStatements = append(statements.PCDeployStatements, deploySQL)
		}
		if rollbackSQL != "" {
			statements.PCRollbackStatements = append(statements.PCRollbackStatements, rollbackSQL)
		}
	case "PE":
		if deploySQL != "" {
			statements.PEDeployStatements = append(statements.PEDeployStatements, deploySQL)
		}
		if rollbackSQL != "" {
			statements.PERollbackStatements = append(statements.PERollbackStatements, rollbackSQL)
		}
	case "PPE":
		if deploySQL != "" {
			statements.PPEDeployStatements = append(statements.PPEDeployStatements, deploySQL)
		}
		if rollbackSQL != "" {
			statements.PPERollbackStatements = append(statements.PPERollbackStatements, rollbackSQL)
		}
	case "RPP":
		if deploySQL != "" {
			statements.RPPDeployStatements = append(statements.RPPDeployStatements, deploySQL)
		}
		if rollbackSQL != "" {
			statements.RPPRollbackStatements = append(statements.RPPRollbackStatements, rollbackSQL)
		}
	}
}
