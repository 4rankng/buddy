package adapters

import (
	"buddy/internal/txn/domain"
)

// SQLStatements contains the deploy and rollback SQL statements separated by database
type SQLStatements struct {
	PCDeployStatements    []string
	PCRollbackStatements  []string
	PEDeployStatements    []string
	PERollbackStatements  []string
	RPPDeployStatements   []string
	RPPRollbackStatements []string
}

// GenerateSQLStatements generates SQL statements for all supported cases using templates.
func GenerateSQLStatements(results []domain.TransactionResult) SQLStatements {
	statements := SQLStatements{}

	// Use map[domain.Case]DMLTicket for automatic consolidation
	caseTickets := make(map[domain.Case]DMLTicket)

	for i := range results {
		// SOP cases should already be identified by Identifydomain.Cases
		caseType := results[i].CaseType

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
	for _, ticket := range caseTickets {
		if len(ticket.RunIDs) > 0 {
			generatedSQL := generateSQLFromTicket(ticket)
			appendStatements(&statements, generatedSQL)
		}
	}

	return statements
}

// GenerateSQLFromTicket generates SQL statements from a DML ticket (exposed version)
func GenerateSQLFromTicket(ticket DMLTicket) SQLStatements {
	return generateSQLFromTicket(ticket)
}

// GetDMLTicketForRppResume returns a DML ticket for the RPP resume case only
func GetDMLTicketForRppResume(result domain.TransactionResult) *DMLTicket {
	sopRepo := SOPRepo
	sopRepo.IdentifyCase(&result, "my") // Default to MY for backward compatibility
	if result.CaseType != domain.CaseRppNoResponseResume {
		return nil
	}

	return &DMLTicket{
		RunIDs:      []string{result.RPPAdapter.Workflow.RunID},
		WorkflowIDs: []string{"'wf_ct_cashout'", "'wf_ct_qr_payment'"},
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
		TargetDB:      "RPP",
		TargetState:   210,
		TargetAttempt: 0,
		CaseType:      domain.CaseRppNoResponseResume,
	}

}
