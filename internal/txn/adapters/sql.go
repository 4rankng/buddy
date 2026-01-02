package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
)

// GenerateSQLStatements generates SQL statements for all supported cases using templates.
func GenerateSQLStatements(results []domain.TransactionResult) domain.SQLStatements {
	statements := domain.SQLStatements{}

	for i := range results {
		// SOP cases should already be identified by Identifydomain.Cases
		caseType := results[i].CaseType

		// Get the template function for this case
		if templateFunc, exists := sqlTemplates[caseType]; exists {
			ticket := templateFunc(results[i])
			if ticket != nil {
				// Generate SQL for each individual ticket (no consolidation)
				generatedSQL, err := generateSQLFromTicket(*ticket)
				if err != nil {
					// Store the error in the result for display
					results[i].Error = err.Error()
					continue
				}
				appendStatements(&statements, generatedSQL)
			} else if caseType == domain.CaseThoughtMachineFalseNegative {
				// Special handling for thought_machine_false_negative case when validation fails
				// Store the error in the result for display
				results[i].Error = "Cannot generate DMLs for thought_machine_false_negative case: prev_trans_id is required but not found in workflow data"
			}
		}
	}

	return statements
}

// GenerateSQLFromTicket generates SQL statements from a DML ticket (exposed version)
func GenerateSQLFromTicket(ticket domain.DMLTicket) (domain.SQLStatements, error) {
	return generateSQLFromTicket(ticket)
}

// GetDMLTicketForRppResume returns a DML ticket for the RPP resume case only
func GetDMLTicketForRppResume(result domain.TransactionResult) *domain.DMLTicket {
	// Only identify case if not already set
	if result.CaseType == "" {
		sopRepo := SOPRepo
		sopRepo.IdentifyCase(&result, "my") // Default to MY for backward compatibility
	}
	if result.CaseType != domain.CaseRppNoResponseResume {
		return nil
	}

	var runID string
	for _, wf := range result.RPPAdapter.Workflow {
		if wf.WorkflowID == "wf_ct_cashout" || wf.WorkflowID == "wf_ct_qr_payment" {
			runID = wf.RunID
			break
		}
	}

	if runID == "" {
		return nil
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_no_response_resume_acsp
-- RPP did not respond in time, but status at Paynet is ACSP (Accepted Settlement in Process) or ACTC (Accepted Technical Validation)
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 222)
WHERE run_id = %s
AND state = 210
AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_no_response_resume_acsp_rollback
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 210)
WHERE run_id = %s
AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseRppNoResponseResume,
	}

}

// GetDMLTicketForCashoutRpp210Pe220Pc201 returns a DML ticket for the cashout RPP 210, PE 220, PC 201 case
// Prompts the user to choose between accept (resume to success) or reject (manual rejection)
func GetDMLTicketForCashoutRpp210Pe220Pc201(result domain.TransactionResult) *domain.DMLTicket {
	sopRepo := SOPRepo
	sopRepo.IdentifyCase(&result, "my")
	if result.CaseType != domain.CaseCashoutRpp210Pe220Pc201 {
		return nil
	}

	// Prompt user for option
	fmt.Println("\nChoose an option:")
	fmt.Println("1. Resume to Success (Manual Success) - Move RPP to state 222 to resume")
	fmt.Println("2. Reject/Fail (Manual Rejection) - Move PE to state 221 with failure details")
	fmt.Print("\nEnter your choice (1 or 2): ")

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read user choice: %v", err)
		return nil
	}

	// Generate DML ticket based on user's choice
	switch choice {
	case 1:
		// Option 1: Resume to Success
		if templateFunc, exists := sqlTemplates[domain.CaseRpp210Pe220Pc201Accept]; exists {
			return templateFunc(result)
		}
	case 2:
		// Option 2: Reject/Fail
		if templateFunc, exists := sqlTemplates[domain.CaseRpp210Pe220Pc201Reject]; exists {
			return templateFunc(result)
		}
	default:
		result.Error = fmt.Sprintf("invalid choice: %d (must be 1 or 2)", choice)
		return nil
	}
	return nil
}

// GetDMLTicketForRppRtpCashinStuck200_0 returns a DML ticket for the RTP cashin stuck at 200 case
func GetDMLTicketForRppRtpCashinStuck200_0(result domain.TransactionResult) *domain.DMLTicket {
	sopRepo := SOPRepo
	sopRepo.IdentifyCase(&result, "my")
	if result.CaseType != domain.CaseRppRtpCashinStuck200_0 {
		return nil
	}

	// Use sqlTemplates map to generate ticket
	if templateFunc, exists := sqlTemplates[domain.CaseRppRtpCashinStuck200_0]; exists {
		return templateFunc(result)
	}
	return nil
}
