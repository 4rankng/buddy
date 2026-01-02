package adapters

import "buddy/internal/txn/domain"

// registerCrossDomainTemplates registers cross-domain templates
func registerCrossDomainTemplates(templates map[domain.Case]TemplateFunc) {
	templates[domain.CaseThoughtMachineFalseNegative] = thoughtMachineFalseNegative
	templates[domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess] = peCaptureProcessingPcCaptureFailedRppSuccess
	templates[domain.CaseCashoutRpp210Pe220Pc201] = GetDMLTicketForCashoutRpp210Pe220Pc201
}

// thoughtMachineFalseNegative handles ThoughtMachine false negative case (PE + PC)
func thoughtMachineFalseNegative(result domain.TransactionResult) *domain.DMLTicket {
	// Get PE and PC run IDs
	peRunID := result.PaymentEngine.Workflow.RunID
	var pcRunID string

	if result.PaymentCore != nil && result.PaymentCore.InternalCapture.Workflow.RunID != "" {
		pcRunID = result.PaymentCore.InternalCapture.Workflow.RunID
	}

	// Validate that we have both PE and PC data
	if peRunID == "" || pcRunID == "" {
		return nil // Missing required data
	}

	// Validate prev_trans_id for PE rollback
	if result.PaymentEngine.Workflow.PrevTransID == "" {
		return nil // Validation failure
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "PE",
				SQLTemplate: `-- thought_machine_false_negative - PE Deploy
UPDATE workflow_execution
SET state = 230,
    prev_trans_id = data->>'$.StreamMessage.ReferenceID',
    data = JSON_SET(data, '$.State', 230)
WHERE run_id = %s
AND state = 701;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: peRunID, Type: "string"},
				},
			},
			{
				TargetDB: "PC",
				SQLTemplate: `-- thought_machine_false_negative (restart PC capture flow from 0)
UPDATE workflow_execution
SET state = 0,
    attempt = 1,
    data = JSON_SET(data, '$.State', 0)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow'
AND state = 500;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: pcRunID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "PE",
				SQLTemplate: `-- thought_machine_false_negative - PE Rollback
UPDATE workflow_execution
SET state = 701,
    attempt = 0,
    prev_trans_id = %s,
    data = JSON_SET(data, '$.State', 701)
WHERE run_id = %s
AND state = 230;`,
				Params: []domain.ParamInfo{
					{Name: "prev_trans_id", Value: result.PaymentEngine.Workflow.PrevTransID, Type: "string"},
					{Name: "run_id", Value: peRunID, Type: "string"},
				},
			},
			{
				TargetDB: "PC",
				SQLTemplate: `-- thought_machine_false_negative - PC Rollback
UPDATE workflow_execution
SET state = 500,
    attempt = 0,
    data = JSON_SET(data, '$.State', 500)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: pcRunID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseThoughtMachineFalseNegative,
	}
}

// peCaptureProcessingPcCaptureFailedRppSuccess handles PE capture processing with PC capture failed but RPP success (PC + RPP)
func peCaptureProcessingPcCaptureFailedRppSuccess(result domain.TransactionResult) *domain.DMLTicket {
	// Get PC run ID
	var pcRunID string

	if result.PaymentCore != nil && result.PaymentCore.InternalCapture.Workflow.RunID != "" {
		pcRunID = result.PaymentCore.InternalCapture.Workflow.RunID
	}

	// Validate that we have PC data
	if pcRunID == "" {
		return nil // Missing required data
	}

	// Validate that we have RPP data with successful workflow
	if result.RPPAdapter == nil {
		return nil
	}

	// Find RPP workflow with state 900 (success) and attempt 0
	// This validates that RPP succeeded as required by the case
	rppRunID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"", // any workflow_id (wf_ct_qr_payment or wf_ct_cashout)
		"900",
		0,
	)

	if rppRunID == "" {
		return nil // No successful RPP workflow found
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "PC",
				SQLTemplate: `-- pe_capture_processing_pc_capture_failed_rpp_success (restart PC capture flow from 0)
UPDATE workflow_execution
SET state = 0,
    attempt = 1,
    data = JSON_SET(data, '$.State', 0)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow'
AND state = 500;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: pcRunID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "PC",
				SQLTemplate: `-- pe_capture_processing_pc_capture_failed_rpp_success - PC Rollback
UPDATE workflow_execution
SET state = 500,
    attempt = 0,
    data = JSON_SET(data, '$.State', 500)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: pcRunID, Type: "string"},
				},
			},
		},
		CaseType: domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess,
	}
}
