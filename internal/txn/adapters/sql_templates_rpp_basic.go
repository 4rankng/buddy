package adapters

import "buddy/internal/txn/domain"

// registerRPPBasicTemplates registers basic RPP (Real-time Payment Processing) templates
func registerRPPBasicTemplates(templates map[domain.Case]TemplateFunc) {
	templates[domain.CaseRppCashoutReject101_19] = rppCashoutReject101_19
	templates[domain.CaseRppQrPaymentReject210_0] = rppQrPaymentReject210_0
	templates[domain.CaseRppNoResponseRejectNotFound] = rppNoResponseRejectNotFound
	templates[domain.CaseRppNoResponseResume] = rppNoResponseResume
	templates[domain.CaseRppCashinValidationFailed122_0] = rppCashinValidationFailed122_0
	templates[domain.CaseRppProcessRegistryStuckInit] = rppProcessRegistryStuckInit

	templates[domain.CaseCashInStuck100Retry] = cashInStuck100Retry
	templates[domain.CaseCashInStuck100UpdateMismatch] = cashInStuck100UpdateMismatch
}

// rppCashoutReject101_19 handles RPP cashout reject at state 101, attempt 19
func rppCashoutReject101_19(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find specific workflow matching the case criteria
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_ct_cashout",
		"101",
		19,
	)

	if runID == "" {
		return nil
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_cashout_reject_101_19, manual reject
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id = %s
AND state = 101
AND workflow_id = 'wf_ct_cashout';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_cashout_reject_101_19_rollback
UPDATE workflow_execution
SET state = 101,
    attempt = 0,
    data = JSON_SET(data, '$.State', 101)
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashout';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseRppCashoutReject101_19,
	}
}

// rppQrPaymentReject210_0 handles RPP QR payment reject at state 210, attempt 0
func rppQrPaymentReject210_0(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find specific workflow matching the case criteria
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_ct_qr_payment",
		"210",
		0,
	)

	if runID == "" {
		return nil
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_qr_payment_reject_210_0, manual reject
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id = %s
AND state = 210
AND workflow_id = 'wf_ct_qr_payment';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_qr_payment_reject_210_0_rollback
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    data = JSON_SET(data, '$.State', 210)
WHERE run_id = %s
AND workflow_id = 'wf_ct_qr_payment';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseRppQrPaymentReject210_0,
	}
}

// rppNoResponseRejectNotFound handles RPP QR payment reject at state 0 (stInit), any attempt
// Use case: RPP adapter stuck in initialization loop and never sends to PayNet
func rppNoResponseRejectNotFound(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find workflow with state 0 (any attempt) for wf_ct_qr_payment
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_ct_qr_payment",
		"0",
		-1, // -1 means any attempt
	)

	if runID == "" {
		return nil
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_no_response_reject_not_found, manual reject for stuck initialization
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id = %s
AND state = 0
AND workflow_id = 'wf_ct_qr_payment';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_no_response_reject_not_found_rollback
UPDATE workflow_execution
SET state = 0,
    attempt = 0,
    data = JSON_SET(data, '$.State', 0)
WHERE run_id = %s
AND workflow_id = 'wf_ct_qr_payment';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseRppNoResponseRejectNotFound,
	}
}

// rppNoResponseResume handles RPP no response - resume transaction
func rppNoResponseResume(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find workflow with state 210 and attempt 0 (either wf_ct_cashout or wf_ct_qr_payment)
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"", // any workflow_id
		"210",
		0,
	)

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
    data = JSON_SET(data, '$.State', 222)
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
				SQLTemplate: `-- rpp_no_response_resume_rollback
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    data = JSON_SET(data, '$.State', 210)
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

// rppCashinValidationFailed122_0 handles RPP cashin validation failed at state 122, attempt 0
func rppCashinValidationFailed122_0(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find the specific workflow matching the case criteria
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_ct_cashin",
		"122",
		0,
	)

	if runID == "" {
		return nil // No matching workflow found
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_cashin_validation_failed_122_0, retry validation
UPDATE workflow_execution
SET state = 100,
	  attempt = 1,
	  data = JSON_SET(data, '$.State', 100)
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin'
AND state = 122;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_cashin_validation_failed_122_0_rollback
UPDATE workflow_execution
SET state = 122,
	  attempt = 0,
	  data = JSON_SET(data, '$.State', 122)
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseRppCashinValidationFailed122_0,
	}
}

// rppProcessRegistryStuckInit handles RPP wf_process_registry stuck at state 0 (stInit)
func rppProcessRegistryStuckInit(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find workflow with state 0 for wf_process_registry
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_process_registry",
		"0",
		-1, // any attempt
	)

	if runID == "" {
		return nil
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_process_registry_stuck_init, set attempt=1 to retry initialization
UPDATE workflow_execution
SET attempt = 1
WHERE run_id = %s
AND workflow_id = 'wf_process_registry'
AND state = 0;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_process_registry_stuck_init_rollback, reset attempt back to 0
UPDATE workflow_execution
SET attempt = 0
WHERE run_id = %s
AND workflow_id = 'wf_process_registry'
AND state = 0;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseRppProcessRegistryStuckInit,
	}
}

// cashInStuck100Retry handles cash-in workflows stuck at state 100 with simple retry case
// Use case: Timestamps match after timezone conversion, simple retry without timestamp modification
func cashInStuck100Retry(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find workflow with state 100 and attempts > 0 for wf_ct_cashin
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_ct_cashin",
		"100",
		-1, // any attempt > 0
	)

	if runID == "" {
		return nil
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- cash_in_stuck_100_retry, timestamps match after timezone conversion
UPDATE workflow_execution
SET attempt = 1
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin'
AND state = 100;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- cash_in_stuck_100_retry_rollback
UPDATE workflow_execution
SET attempt = 0
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin'
AND state = 100;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseCashInStuck100Retry,
	}
}

// cashInStuck100UpdateMismatch handles cash-in workflows stuck at state 100 with timestamp mismatch
// Use case: Timestamps don't match after timezone conversion, requires timestamp synchronization
func cashInStuck100UpdateMismatch(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find workflow with state 100 and attempts > 0 for wf_ct_cashin
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_ct_cashin",
		"100",
		-1, // any attempt > 0
	)

	if runID == "" {
		return nil
	}

	// For this template, we'll use a placeholder for the converted timestamp
	// In a real implementation, this would be populated by the case detection logic
	// that determines the converted timestamp from credit_transfer.updated_at (UTC) to GMT+8
	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- cash_in_stuck_100_update_mismatch, sync timestamp and retry
UPDATE workflow_execution
SET attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `,
        '$.CreditTransfer.UpdatedAt', %s)
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin'
AND state = 100;`,
				Params: []domain.ParamInfo{
					{Name: "converted_timestamp", Value: "{CONVERTED_TIMESTAMP}", Type: "string"},
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "RPP",
				SQLTemplate: `-- cash_in_stuck_100_update_mismatch_rollback
UPDATE workflow_execution
SET attempt = 0
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin'
AND state = 100;`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseCashInStuck100UpdateMismatch,
	}
}
