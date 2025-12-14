package adapters

import (
	"buddy/internal/txn/domain"
)

// GenerateSQLStatements generates SQL statements for all supported cases using templates.
func GenerateSQLStatements(results []domain.TransactionResult) domain.SQLStatements {
	statements := domain.SQLStatements{}

	// Use map[domain.Case]domain.DMLTicket for automatic consolidation
	caseTickets := make(map[domain.Case]domain.DMLTicket)

	for i := range results {
		// SOP cases should already be identified by Identifydomain.Cases
		caseType := results[i].CaseType

		// Get the template function for this case
		if templateFunc, exists := sqlTemplates[caseType]; exists {
			newTicket := templateFunc(results[i])
			if newTicket != nil {
				if existingTicket, exists := caseTickets[caseType]; exists {
					// Merge parameters into existing ticket
					existingTicket.DeployParams = mergeParams(existingTicket.DeployParams, newTicket.DeployParams)
					existingTicket.RollbackParams = mergeParams(existingTicket.RollbackParams, newTicket.RollbackParams)
					caseTickets[caseType] = existingTicket
				} else {
					// Create new ticket
					caseTickets[caseType] = *newTicket
				}
			}
		}
	}

	// Process each consolidated ticket
	for _, ticket := range caseTickets {
		generatedSQL, err := generateSQLFromTicket(ticket)
		if err != nil {
			// For now, log the error and continue
			// In a production environment, you might want to handle this differently
			continue
		}
		appendStatements(&statements, generatedSQL)
	}

	return statements
}

// GenerateSQLFromTicket generates SQL statements from a DML ticket (exposed version)
func GenerateSQLFromTicket(ticket domain.DMLTicket) (domain.SQLStatements, error) {
	return generateSQLFromTicket(ticket)
}

// GetDMLTicketForRppResume returns a DML ticket for the RPP resume case only
func GetDMLTicketForRppResume(result domain.TransactionResult) *domain.DMLTicket {
	sopRepo := SOPRepo
	sopRepo.IdentifyCase(&result, "my") // Default to MY for backward compatibility
	if result.CaseType != domain.CaseRppNoResponseResume {
		return nil
	}

	return &domain.DMLTicket{
		DeployTemplate: `-- rpp_no_response_resume_acsp
-- RPP did not respond in time, but status at Paynet is ACSP (Accepted Settlement in Process) or ACTC (Accepted Technical Validation)
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 222)
WHERE run_id IN (%s)
AND state = 210
AND workflow_id IN (%s);`,
		RollbackTemplate: `-- RPP Rollback: Move workflows back to state 210
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 210)
WHERE run_id IN (%s)
AND workflow_id IN (%s);`,
		TargetDB: "RPP",
		DeployParams: []domain.ParamInfo{
			{Name: "run_ids", Value: []string{result.RPPAdapter.Workflow.RunID}, Type: "string_array"},
			{Name: "workflow_ids", Value: []string{"'wf_ct_cashout'", "'wf_ct_qr_payment'"}, Type: "string_array"},
		},
		RollbackParams: []domain.ParamInfo{
			{Name: "run_ids", Value: []string{result.RPPAdapter.Workflow.RunID}, Type: "string_array"},
			{Name: "workflow_ids", Value: []string{"'wf_ct_cashout'", "'wf_ct_qr_payment'"}, Type: "string_array"},
		},
		CaseType: domain.CaseRppNoResponseResume,
	}

}

// mergeParams merges parameter arrays, concatenating string_array values
func mergeParams(existing, new []domain.ParamInfo) []domain.ParamInfo {
	paramMap := make(map[string]domain.ParamInfo)

	// Add existing parameters
	for _, param := range existing {
		paramMap[param.Name] = param
	}

	// Merge new parameters
	for _, newParam := range new {
		if existingParam, exists := paramMap[newParam.Name]; exists {
			// For string_array types, concatenate the values
			if newParam.Type == "string_array" && existingParam.Type == "string_array" {
				if existingVals, ok := existingParam.Value.([]string); ok {
					if newVals, ok := newParam.Value.([]string); ok {
						mergedVals := append(existingVals, newVals...)
						paramMap[newParam.Name] = domain.ParamInfo{
							Name:  newParam.Name,
							Value: mergedVals,
							Type:  newParam.Type,
						}
					}
				}
			}
		} else {
			// Add new parameter
			paramMap[newParam.Name] = newParam
		}
	}

	// Convert back to slice
	result := make([]domain.ParamInfo, 0, len(paramMap))
	for _, param := range paramMap {
		result = append(result, param)
	}

	return result
}
