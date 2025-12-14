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

// countPlaceholders counts %s occurrences in a template
func countPlaceholders(template string) int {
	return strings.Count(template, "%s")
}

// flattenParameters converts array parameters into individual parameters
// It groups array parameters by index to maintain pairing
func flattenParameters(params []domain.ParamInfo) []domain.ParamInfo {
	// First, find the maximum array length
	maxArrayLen := 0
	for _, param := range params {
		switch param.Type {
		case "string_array":
			if arr, ok := param.Value.([]string); ok && len(arr) > maxArrayLen {
				maxArrayLen = len(arr)
			}
		case "int_array":
			if arr, ok := param.Value.([]int); ok && len(arr) > maxArrayLen {
				maxArrayLen = len(arr)
			}
		}
	}

	// If no arrays, return params as-is
	if maxArrayLen == 0 {
		return params
	}

	// Create flattened parameters by grouping array elements
	var flattened []domain.ParamInfo
	for i := 0; i < maxArrayLen; i++ {
		for _, param := range params {
			switch param.Type {
			case "string_array":
				if arr, ok := param.Value.([]string); ok {
					if i < len(arr) {
						flattened = append(flattened, domain.ParamInfo{
							Name:  param.Name,
							Value: arr[i],
							Type:  "string",
						})
					}
				}
			case "int_array":
				if arr, ok := param.Value.([]int); ok {
					if i < len(arr) {
						flattened = append(flattened, domain.ParamInfo{
							Name:  param.Name,
							Value: arr[i],
							Type:  "int",
						})
					}
				}
			default:
				// For non-array parameters, repeat them for each group
				if i == 0 {
					flattened = append(flattened, param)
				}
			}
		}
	}

	return flattened
}

// groupParameters groups flattened parameters by statement count
func groupParameters(params []domain.ParamInfo, placeholders int) [][]domain.ParamInfo {
	if placeholders == 0 {
		return [][]domain.ParamInfo{params}
	}

	totalParams := len(params)

	// Check if we have enough parameters
	if totalParams%placeholders != 0 {
		return nil // Not enough parameters to form complete statements
	}

	statementsCount := totalParams / placeholders
	groups := make([][]domain.ParamInfo, statementsCount)

	for i := 0; i < statementsCount; i++ {
		start := i * placeholders
		end := start + placeholders
		group := make([]domain.ParamInfo, placeholders)
		copy(group, params[start:end])
		groups[i] = group
	}

	return groups
}

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

// generateSQLFromTicket generates SQL statements from a DML ticket using parameter-based logic
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

	// DEPLOY: Use parameter-based logic
	deployPlaceholders := countPlaceholders(ticket.DeployTemplate)
	deployFlattened := flattenParameters(ticket.DeployParams)

	if len(deployFlattened) < deployPlaceholders {
		return domain.SQLStatements{}, fmt.Errorf("insufficient deploy parameters: need %d, got %d", deployPlaceholders, len(deployFlattened))
	}

	if len(deployFlattened) == deployPlaceholders {
		// Generate single statement
		deploySQL, err := buildSQLFromTemplate(ticket.DeployTemplate, deployFlattened)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate deploy SQL for case %s: %w", ticket.CaseType, err)
		}
		if err := validateSQL(deploySQL, ticket.DeployTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
		}
		addStatementToDatabase(&statements, ticket.TargetDB, deploySQL, "")
	} else {
		// Generate multiple statements
		deployGroups := groupParameters(deployFlattened, deployPlaceholders)
		if deployGroups == nil {
			return domain.SQLStatements{}, fmt.Errorf("deploy parameters count (%d) is not divisible by placeholder count (%d)", len(deployFlattened), deployPlaceholders)
		}
		for _, group := range deployGroups {
			deploySQL, err := buildSQLFromTemplate(ticket.DeployTemplate, group)
			if err != nil {
				return domain.SQLStatements{}, fmt.Errorf("failed to generate deploy SQL for case %s: %w", ticket.CaseType, err)
			}
			if err := validateSQL(deploySQL, ticket.DeployTemplate); err != nil {
				return domain.SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
			}
			addStatementToDatabase(&statements, ticket.TargetDB, deploySQL, "")
		}
	}

	// ROLLBACK: Use parameter-based logic
	rollbackPlaceholders := countPlaceholders(ticket.RollbackTemplate)
	rollbackFlattened := flattenParameters(ticket.RollbackParams)

	if len(rollbackFlattened) < rollbackPlaceholders {
		return domain.SQLStatements{}, fmt.Errorf("insufficient rollback parameters: need %d, got %d", rollbackPlaceholders, len(rollbackFlattened))
	}

	if len(rollbackFlattened) == rollbackPlaceholders {
		// Generate single statement
		rollbackSQL, err := buildSQLFromTemplate(ticket.RollbackTemplate, rollbackFlattened)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate rollback SQL for case %s: %w", ticket.CaseType, err)
		}
		if err := validateSQL(rollbackSQL, ticket.RollbackTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
		}
		addStatementToDatabase(&statements, ticket.TargetDB, "", rollbackSQL)
	} else {
		// Generate multiple statements
		rollbackGroups := groupParameters(rollbackFlattened, rollbackPlaceholders)
		if rollbackGroups == nil {
			return domain.SQLStatements{}, fmt.Errorf("rollback parameters count (%d) is not divisible by placeholder count (%d)", len(rollbackFlattened), rollbackPlaceholders)
		}
		for _, group := range rollbackGroups {
			rollbackSQL, err := buildSQLFromTemplate(ticket.RollbackTemplate, group)
			if err != nil {
				return domain.SQLStatements{}, fmt.Errorf("failed to generate rollback SQL for case %s: %w", ticket.CaseType, err)
			}
			if err := validateSQL(rollbackSQL, ticket.RollbackTemplate); err != nil {
				return domain.SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
			}
			addStatementToDatabase(&statements, ticket.TargetDB, "", rollbackSQL)
		}
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
	case "RPP":
		if deploySQL != "" {
			statements.RPPDeployStatements = append(statements.RPPDeployStatements, deploySQL)
		}
		if rollbackSQL != "" {
			statements.RPPRollbackStatements = append(statements.RPPRollbackStatements, rollbackSQL)
		}
	}
}
