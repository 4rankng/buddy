package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"strings"
)

// templateGroupKey represents a unique key for grouping identical templates
type templateGroupKey struct {
	targetDB    string
	sqlTemplate string // SQL template without comments
	paramsKey   string // String representation of params excluding run_id
}

// groupedTemplate represents a group of templates that can be combined
type groupedTemplate struct {
	comment     string
	targetDB    string
	sqlTemplate string
	otherParams []domain.ParamInfo // All params except run_id
	runIDs      []string           // All run_ids to be combined
}

// appendStatements is a helper to merge results into main struct
func appendStatements(main *domain.SQLStatements, new domain.SQLStatements) {
	fmt.Printf("[DEBUG] appendStatements: PE deploy %d, PE rollback %d\n",
		len(new.PEDeployStatements), len(new.PERollbackStatements))
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

// extractComment extracts SQL comments from the beginning of a template
func extractComment(template string) (comment, sqlWithoutComment string) {
	lines := strings.Split(template, "\n")
	var commentLines []string
	var sqlLines []string
	inComment := true

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inComment && strings.HasPrefix(trimmed, "--") {
			commentLines = append(commentLines, line)
		} else {
			inComment = false
			sqlLines = append(sqlLines, line)
		}
	}

	if len(commentLines) > 0 {
		comment = strings.Join(commentLines, "\n")
	}
	if len(sqlLines) > 0 {
		sqlWithoutComment = strings.Join(sqlLines, "\n")
	}
	return
}

// buildParamsKey creates a string key from params excluding run_id
func buildParamsKey(params []domain.ParamInfo) string {
	var keyParts []string
	for _, p := range params {
		if p.Name != "run_id" {
			keyParts = append(keyParts, fmt.Sprintf("%s:%v", p.Name, p.Value))
		}
	}
	return strings.Join(keyParts, "|")
}

// extractRunID extracts the run_id parameter value from params
func extractRunID(params []domain.ParamInfo) string {
	for _, p := range params {
		if p.Name == "run_id" {
			if strVal, ok := p.Value.(string); ok {
				return strVal
			}
			return fmt.Sprintf("%v", p.Value)
		}
	}
	return ""
}

// removeRunIDParam removes the run_id parameter from params
func removeRunIDParam(params []domain.ParamInfo) []domain.ParamInfo {
	var result []domain.ParamInfo
	for _, p := range params {
		if p.Name != "run_id" {
			result = append(result, p)
		}
	}
	return result
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

// groupTemplates groups templates by their SQL template and params (excluding run_id)
func groupTemplates(templates []domain.TemplateInfo) map[templateGroupKey]*groupedTemplate {
	groups := make(map[templateGroupKey]*groupedTemplate)

	for _, tmpl := range templates {
		comment, sqlWithoutComment := extractComment(tmpl.SQLTemplate)
		paramsKey := buildParamsKey(tmpl.Params)
		runID := extractRunID(tmpl.Params)

		key := templateGroupKey{
			targetDB:    tmpl.TargetDB,
			sqlTemplate: sqlWithoutComment,
			paramsKey:   paramsKey,
		}

		if group, exists := groups[key]; exists {
			// Add run_id to existing group
			group.runIDs = append(group.runIDs, runID)
		} else {
			// Create new group
			groups[key] = &groupedTemplate{
				comment:     comment,
				targetDB:    tmpl.TargetDB,
				sqlTemplate: sqlWithoutComment,
				otherParams: removeRunIDParam(tmpl.Params),
				runIDs:      []string{runID},
			}
		}
	}

	return groups
}

// buildSQLFromGroupedTemplate builds SQL from a grouped template with run_id IN clause
func buildSQLFromGroupedTemplate(group *groupedTemplate) (string, error) {
	// Build IN clause for run_ids
	var runIDList []string
	for _, runID := range group.runIDs {
		runIDList = append(runIDList, fmt.Sprintf("'%s'", runID))
	}
	runIDClause := strings.Join(runIDList, ", ")

	// Format other parameters
	formattedParams := make([]interface{}, len(group.otherParams))
	for i, param := range group.otherParams {
		formattedParams[i] = formatParameter(param)
	}

	// Build SQL with run_id IN clause
	sql := group.sqlTemplate

	// Replace "run_id = %s" with "run_id IN (...)" clause
	if strings.Contains(sql, "run_id = %s") {
		sql = strings.Replace(sql, "run_id = %s", "run_id IN ("+runIDClause+")", 1)
	} else if strings.Contains(sql, "run_id IN (%s)") {
		// Template already has IN clause, just substitute the run_ids
		sql = strings.Replace(sql, "run_id IN (%s)", "run_id IN ("+runIDClause+")", 1)
	}

	// Substitute remaining parameters (AuthorisationID, etc.)
	if len(formattedParams) > 0 {
		sql = fmt.Sprintf(sql, formattedParams...)
	}

	// Add comment back
	if group.comment != "" {
		sql = group.comment + "\n" + sql
	}

	return sql, nil
}

// generateSQLFromTicket generates SQL statements from a DML ticket using TemplateInfo arrays
func generateSQLFromTicket(ticket domain.DMLTicket) (domain.SQLStatements, error) {
	// Validate input
	if len(ticket.Deploy) == 0 && len(ticket.Rollback) == 0 {
		return domain.SQLStatements{}, fmt.Errorf("ticket contains no templates")
	}

	// Debug logging
	fmt.Printf("[DEBUG] generateSQLFromTicket: case=%s, deploy=%d, rollback=%d\n",
		ticket.CaseType, len(ticket.Deploy), len(ticket.Rollback))

	// Define valid target databases
	validDatabases := map[string]struct{}{
		"PC":  {},
		"PE":  {},
		"RPP": {},
		"PPE": {},
	}

	statements := domain.SQLStatements{}

	// Group deploy templates
	deployGroups := groupTemplates(ticket.Deploy)

	// Process grouped deploy templates
	for _, group := range deployGroups {
		// Validate target DB
		if _, ok := validDatabases[group.targetDB]; !ok {
			return domain.SQLStatements{}, fmt.Errorf("unknown target database: %s", group.targetDB)
		}

		// Generate SQL for grouped template
		deploySQL, err := buildSQLFromGroupedTemplate(group)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate deploy SQL for case %s: %w", ticket.CaseType, err)
		}

		// Validate SQL (use original template for validation)
		if err := validateSQL(deploySQL, group.sqlTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("deploy SQL validation failed: %w", err)
		}

		addStatementToDatabase(&statements, group.targetDB, deploySQL, "")
	}

	// Group rollback templates
	rollbackGroups := groupTemplates(ticket.Rollback)

	// Debug logging
	fmt.Printf("[DEBUG] Processing rollback templates for case %s: %d groups\n", ticket.CaseType, len(rollbackGroups))

	// Process grouped rollback templates
	for _, group := range rollbackGroups {
		fmt.Printf("[DEBUG] Processing rollback group: targetDB=%s, runIDs=%d\n", group.targetDB, len(group.runIDs))

		// Validate target DB
		if _, ok := validDatabases[group.targetDB]; !ok {
			return domain.SQLStatements{}, fmt.Errorf("unknown target database: %s", group.targetDB)
		}

		// Generate SQL for grouped template
		rollbackSQL, err := buildSQLFromGroupedTemplate(group)
		if err != nil {
			return domain.SQLStatements{}, fmt.Errorf("failed to generate rollback SQL for case %s: %w", ticket.CaseType, err)
		}

		fmt.Printf("[DEBUG] Generated rollback SQL length: %d\n", len(rollbackSQL))

		// Validate SQL (use original template for validation)
		if err := validateSQL(rollbackSQL, group.sqlTemplate); err != nil {
			return domain.SQLStatements{}, fmt.Errorf("rollback SQL validation failed: %w", err)
		}

		addStatementToDatabase(&statements, group.targetDB, "", rollbackSQL)
		fmt.Printf("[DEBUG] Added rollback statement to %s\n", group.targetDB)
	}

	fmt.Printf("[DEBUG] Final statements: PEDeploy=%d, PERollback=%d\n",
		len(statements.PEDeployStatements), len(statements.PERollbackStatements))

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
