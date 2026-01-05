package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"os"
	"strings"
)

var autoChoices = make(map[domain.Case]int)

// ResetAutoChoices clears any saved user choices for "Apply to All"
func ResetAutoChoices() {
	autoChoices = make(map[domain.Case]int)
}

// GenerateSQLStatements generates SQL statements for all supported cases using templates.
func GenerateSQLStatements(results []domain.TransactionResult) domain.SQLStatements {
	statements := domain.SQLStatements{}

	// Reset auto-choices for "Apply to All" at the start of each batch
	ResetAutoChoices()

	// Group tickets by CaseType to allow cross-result consolidation
	groupedTickets := make(map[domain.Case]*domain.DMLTicket)
	caseErrors := make(map[domain.Case]string)

	for i := range results {
		// Populate index for use in interactive prompts
		results[i].Index = i + 1

		// SOP cases should already be identified by Identifydomain.Cases
		caseType := results[i].CaseType

		// Get the template function for this case
		if templateFunc, exists := sqlTemplates[caseType]; exists {
			ticket := templateFunc(results[i])
			if ticket != nil {
				if existing, exists := groupedTickets[caseType]; exists {
					// Merge templates from new ticket into existing one
					existing.Deploy = append(existing.Deploy, ticket.Deploy...)
					existing.Rollback = append(existing.Rollback, ticket.Rollback...)
				} else {
					groupedTickets[caseType] = ticket
				}
			} else if caseType == domain.CaseThoughtMachineFalseNegative {
				results[i].Error = "Cannot generate DMLs for thought_machine_false_negative case: prev_trans_id is required but not found in workflow data"
				caseErrors[caseType] = results[i].Error
			}
		}
	}

	// Generate SQL for each consolidated group of tickets
	for caseType, ticket := range groupedTickets {
		generatedSQL, err := generateSQLFromTicket(*ticket)
		if err != nil {
			// Store error globally if consolidation fails (unlikely given current logic)
			caseErrors[caseType] = err.Error()
			continue
		}
		appendStatements(&statements, generatedSQL)
	}

	// Generate transfer table UPDATE statements for transactions with payment-core internal_auth
	for _, result := range results {
		if shouldGenerateTransferUpdate(result) {
			transferUpdateSQL := generateTransferUpdateSQL(result)
			if transferUpdateSQL != "" {
				statements.PEDeployStatements = append(statements.PEDeployStatements, transferUpdateSQL)
			}
		}
	}

	// Re-assign errors back to relevant records if needed (optional, for visibility)
	for caseType, errStr := range caseErrors {
		for i := range results {
			if results[i].CaseType == caseType {
				results[i].Error = errStr
			}
		}
	}

	return statements
}

// shouldGenerateTransferUpdate checks if a transfer table UPDATE statement should be generated
func shouldGenerateTransferUpdate(result domain.TransactionResult) bool {
	// Check if PaymentCore has InternalAuth with SUCCESS status and TxID
	if result.PaymentCore == nil {
		return false
	}
	if result.PaymentCore.InternalAuth.TxStatus != "SUCCESS" {
		return false
	}
	if result.PaymentCore.InternalAuth.TxID == "" {
		return false
	}
	// Check if PaymentEngine has a valid transaction with updated_at timestamp
	if result.PaymentEngine == nil {
		return false
	}
	if result.PaymentEngine.Transfers.TransactionID == "" {
		return false
	}
	if result.PaymentEngine.Transfers.UpdatedAt == "" {
		return false
	}
	return true
}

// generateTransferUpdateSQL generates the SQL UPDATE statement for the transfer table
func generateTransferUpdateSQL(result domain.TransactionResult) string {
	// Format the AuthorisationID and created_at timestamp
	authorisationID := result.PaymentCore.InternalAuth.TxID
	transactionID := result.PaymentEngine.Transfers.TransactionID
	updatedAt := result.PaymentEngine.Transfers.UpdatedAt

	// Build the SQL UPDATE statement
	sql := fmt.Sprintf(
		"-- Update transfer table with AuthorisationID from payment-core internal_auth\n"+
			"UPDATE transfer\n"+
			"SET properties = JSON_SET(properties, '$.AuthorisationID', '%s'),\n"+
			"    updated_at = '%s'\n"+
			"WHERE transaction_id = '%s';",
		authorisationID, updatedAt, transactionID)

	return sql
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

	// Check if we already have an auto-choice for this case
	if choice, exists := autoChoices[result.CaseType]; exists {
		return handleChoice(choice, result)
	}

	// Display visual divider and transaction summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	WriteResult(os.Stdout, result, result.Index)

	// Prompt user for option
	fmt.Println("\nChoose an option:")
	fmt.Println("1. Resume to Success (Manual Success) - This once")
	fmt.Println("2. Reject/Fail (Manual Rejection) - This once")
	fmt.Println("3. Resume to Success (Manual Success) - Apply to all similar")
	fmt.Println("4. Reject/Fail (Manual Rejection) - Apply to all similar")
	fmt.Print("\nEnter your choice (1, 2, 3, or 4): ")

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read user choice: %v", err)
		return nil
	}

	// Handle "Apply to All" options
	switch choice {
	case 3:
		autoChoices[result.CaseType] = 1
		choice = 1
	case 4:
		autoChoices[result.CaseType] = 2
		choice = 2
	}

	return handleChoice(choice, result)
}

// handleChoice processes the selected option for a given result
func handleChoice(choice int, result domain.TransactionResult) *domain.DMLTicket {
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
		result.Error = fmt.Sprintf("invalid choice: %d (must be 1-4)", choice)
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
