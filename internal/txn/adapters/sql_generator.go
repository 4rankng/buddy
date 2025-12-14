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

// Helper Functions for SQL Generation

// formatParameter formats a parameter value based on its type for SQL usage
func formatParameter(info domain.ParamInfo) string {
	switch info.Type {
	case "string":
		return fmt.Sprintf("'%v'", info.Value)
	case "string_array":
		if arr, ok := info.Value.([]string); ok {
			quoted := make([]string, len(arr))
			for i, v := range arr {
				quoted[i] = fmt.Sprintf("'%s'", v)
			}
			return strings.Join(quoted, ", ")
		}
		// Fallback for unexpected types
		return fmt.Sprintf("'%v'", info.Value)
	case "int":
		return fmt.Sprintf("%v", info.Value)
	case "int_array":
		if arr, ok := info.Value.([]int); ok {
			strArr := make([]string, len(arr))
			for i, v := range arr {
				strArr[i] = fmt.Sprintf("%d", v)
			}
			return strings.Join(strArr, ", ")
		}
		// Fallback for unexpected types
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

	// Substitute parameters in template
	sql := fmt.Sprintf(template, formattedParams...)
	return sql, nil
}

// getParamValue finds and returns the value of a parameter by name
func getParamValue(params []domain.ParamInfo, name string) interface{} {
	for _, param := range params {
		if param.Name == name {
			return param.Value
		}
	}
	return nil
}

// updateParamValue creates a new parameter slice with updated value for the given parameter name
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
	// Only the thought machine false negative case needs individual statements due to prev_trans_id
	return caseType == domain.CaseThoughtMachineFalseNegative
}

// generateSQLFromTicket generates SQL statements from a DML ticket using the universal builder
func generateSQLFromTicket(ticket domain.DMLTicket) (domain.SQLStatements, error) {
	// Validate input
	if len(ticket.DeployParams) == 0 && len(ticket.RollbackParams) == 0 {
		return domain.SQLStatements{}, fmt.Errorf("ticket contains no parameters")
	}

	// Validate target DB
	if ticket.TargetDB != "PC" && ticket.TargetDB != "PE" && ticket.TargetDB != "RPP" {
		return domain.SQLStatements{}, fmt.Errorf("unknown target database: %s", ticket.TargetDB)
	}

	statements := domain.SQLStatements{}

	// Check if we need individual statements (for prev_trans_id cases)
	if checkIndividualStatementNeeded(ticket.CaseType) {
		// Generate individual statements for each run ID
		runIDsParam := getParamValue(ticket.DeployParams, "run_ids")
		if runIDsArray, ok := runIDsParam.([]string); ok {
			for _, runID := range runIDsArray {
				// Create a single-run ticket for individual processing
				singleTicket := ticket
				singleTicket.DeployParams = updateParamValue(ticket.DeployParams, "run_ids", []string{runID})
				singleTicket.RollbackParams = updateParamValue(ticket.RollbackParams, "run_ids", []string{runID})

				// Generate SQL using universal builder
				deploySQL, err := buildSQLFromTemplate(singleTicket.DeployTemplate, singleTicket.DeployParams)
				if err != nil {
					return domain.SQLStatements{}, fmt.Errorf("failed to generate deploy SQL for case %s, run ID %s: %w", ticket.CaseType, runID, err)
				}

				rollbackSQL, err := buildSQLFromTemplate(singleTicket.RollbackTemplate, singleTicket.RollbackParams)
				if err != nil {
					return domain.SQLStatements{}, fmt.Errorf("failed to generate rollback SQL for case %s, run ID %s: %w", ticket.CaseType, runID, err)
				}

				// Validate generated SQL
				if err := validateSQL(deploySQL, ticket.DeployTemplate); err != nil {
					return domain.SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
				}
				if err := validateSQL(rollbackSQL, ticket.RollbackTemplate); err != nil {
					return domain.SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
				}

				// Add to appropriate statement list based on target DB
				addStatementToDatabase(&statements, ticket.TargetDB, deploySQL, rollbackSQL)
			}
		}
	} else {
		// Generate batch statement for all parameters
		deploySQL, err := buildSQLFromTemplate(ticket.DeployTemplate, ticket.DeployParams)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate deploy SQL for case %s: %w", ticket.CaseType, err)
		}

		rollbackSQL, err := buildSQLFromTemplate(ticket.RollbackTemplate, ticket.RollbackParams)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate rollback SQL for case %s: %w", ticket.CaseType, err)
		}

		// Validate generated SQL
		if err := validateSQL(deploySQL, ticket.DeployTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
		}
		if err := validateSQL(rollbackSQL, ticket.RollbackTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
		}

		// Add to appropriate statement list based on target DB
		addStatementToDatabase(&statements, ticket.TargetDB, deploySQL, rollbackSQL)
	}

	return statements, nil
}

// addStatementToDatabase adds SQL statements to the appropriate database section
func addStatementToDatabase(statements *domain.SQLStatements, targetDB string, deploySQL, rollbackSQL string) {
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
