package txn

// templateConfigs maps SOP cases to their template parameter configurations
var templateConfigs = map[SOPCase]TemplateConfig{
	SOPCasePcExternalPaymentFlow200_11:      {Parameters: []string{"run_ids"}},
	SOPCasePcExternalPaymentFlow201_0RPP210: {Parameters: []string{"run_ids"}},
	SOPCasePcExternalPaymentFlow201_0RPP900: {Parameters: []string{"run_ids"}},
	SOPCasePeTransferPayment210_0:           {Parameters: []string{"run_ids"}},
	SOPCasePe2200FastCashinFailed:           {Parameters: []string{"run_ids"}},
	SOPCaseRppCashoutReject101_19:           {Parameters: []string{"run_ids"}},
	SOPCaseRppQrPaymentReject210_0:          {Parameters: []string{"run_ids"}},
	SOPCaseRppNoResponseResume:              {Parameters: []string{"run_ids", "workflow_ids"}},
}

// sqlTemplates maps SOP cases to their DML tickets
var sqlTemplates = map[SOPCase]func(TransactionResult) *DMLTicket{
	// ========================================
	// Payment Core (PC) Templates
	// ========================================
	SOPCasePcExternalPaymentFlow200_11: func(result TransactionResult) *DMLTicket {
		if runID := getPcExtPayment200_11RunID(result); runID != "" {
			return &DMLTicket{
				RunIDs: []string{runID},
				DeployTemplate: `-- pc_external_payment_flow_200_11
UPDATE workflow_execution
SET state = 202,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.StreamResp', JSON_OBJECT(
        'TxID', '',
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ExternalID', '',
        'ErrorMessage', 'Reject from adapter'),
      '$.State', 202)
WHERE run_id IN (%s)
AND state = 200
AND attempt = 11;`,
				RollbackTemplate: `UPDATE workflow_execution
SET state = 200,
    attempt = 11,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 200)
WHERE run_id IN (%s);`,
				TargetDB:      "PC",
				WorkflowID:    "pc_external_payment_flow",
				TargetState:   200,
				TargetAttempt: 11,
			}
		}
		return nil
	},

	// ========================================
	// Payment Engine (PE) Templates
	// ========================================
	SOPCasePeTransferPayment210_0: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.PaymentEngine.Workflow.RunID},
			DeployTemplate: `-- Reject PE stuck 210. Reject transactions since it hasn't reached Paynet yet
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.StreamMessage', JSON_OBJECT(
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ErrorMessage', 'Manual Rejected'),
      '$.State', 221)
WHERE run_id IN (%s)
AND workflow_id = 'workflow_transfer_payment'
AND state = 210;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.StreamMessage', NULL,
      '$.State', 210)
WHERE run_id IN (%s)
AND workflow_id = 'workflow_transfer_payment';`,
			TargetDB:      "PE",
			WorkflowID:    "workflow_transfer_payment",
			TargetState:   210,
			TargetAttempt: 0,
		}
	},
	SOPCasePe2200FastCashinFailed: func(result TransactionResult) *DMLTicket {
		if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
			return &DMLTicket{
				RunIDs: []string{runID},
				DeployTemplate: `-- pe_220_0_fast_cashin_failed
UPDATE workflow_execution
SET attempt = 1,
    state = 221,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.State', 221,
      '$.StreamMessage.Status', 'FAILED',
      '$.StreamMessage.ErrorMessage', 'MANUAL REJECT')
WHERE run_id IN (%s)
AND workflow_id = 'workflow_transfer_collection'
AND state = 220
AND attempt = 0;`,
				RollbackTemplate: `UPDATE workflow_execution
SET attempt = 0,
    state = 220,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.State', 220,
      '$.StreamMessage', JSON_OBJECT())
WHERE run_id IN (%s)
AND workflow_id = 'workflow_transfer_collection';`,
				TargetDB:      "PE",
				WorkflowID:    "workflow_transfer_collection",
				TargetState:   220,
				TargetAttempt: 0,
			}
		}
		return nil
	},

	// ========================================
	// RPP (Real-time Payment Processing) Templates
	// ========================================
	SOPCasePcExternalPaymentFlow201_0RPP210: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPAdapter.Workflow.RunID},
			DeployTemplate: `-- RPP 210, PE 220, PC 201. No response from RPP. Move to 222 to resume. ACSP
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 222)
WHERE run_id IN (%s)
AND state = 210;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 201,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 201)
WHERE run_id IN (%s);`,
			TargetDB:      "RPP",
			TargetState:   210,
			TargetAttempt: 0,
		}
	},

	SOPCasePcExternalPaymentFlow201_0RPP900: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPAdapter.Workflow.RunID},
			DeployTemplate: `-- RPP 900, PE 220, PC 201. Republish from RPP to resume. ACSP
UPDATE workflow_execution
SET state = 301,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 301)
WHERE run_id IN (%s)
AND state = 900;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 900,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 900)
WHERE run_id IN (%s);`,
			TargetDB:      "RPP",
			TargetState:   900,
			TargetAttempt: 0,
		}
	},

	SOPCaseRppCashoutReject101_19: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPAdapter.Workflow.RunID},
			DeployTemplate: `-- rpp_cashout_reject_101_19, manual reject
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id IN (%s)
AND state = 101
AND workflow_id = 'wf_ct_cashout';`,
			RollbackTemplate: `-- RPP Rollback: Move workflows back to state 101
UPDATE workflow_execution
SET state = 101,
    attempt = 0,
    data = JSON_SET(data, '$.State', 101)
WHERE run_id IN (%s)
AND workflow_id = 'wf_ct_cashout';`,
			TargetDB:      "RPP",
			WorkflowID:    "'wf_ct_cashout'",
			TargetState:   101,
			TargetAttempt: 19,
		}
	},

	SOPCaseRppQrPaymentReject210_0: func(result TransactionResult) *DMLTicket {
		return &DMLTicket{
			RunIDs: []string{result.RPPAdapter.Workflow.RunID},
			DeployTemplate: `-- rpp_qr_payment_reject_210_0, manual reject
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id IN (%s)
AND state = 210
AND workflow_id = 'wf_ct_qr_payment';`,
			RollbackTemplate: `-- RPP Rollback: Move qr_payment workflows back to state 210
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    data = JSON_SET(data, '$.State', 210)
WHERE run_id IN (%s)
AND workflow_id = 'wf_ct_qr_payment';`,
			TargetDB:      "RPP",
			WorkflowID:    "'wf_ct_qr_payment'",
			TargetState:   210,
			TargetAttempt: 0,
		}
	},

	SOPCaseRppNoResponseResume: func(result TransactionResult) *DMLTicket {
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
		}
	},
}
