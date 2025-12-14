package adapters

import "buddy/internal/txn/domain"

// sqlTemplates maps SOP cases to their DML tickets
var sqlTemplates = map[domain.Case]func(domain.TransactionResult) *domain.DMLTicket{
	// ========================================
	// Payment Core (PC) Templates
	// ========================================
	domain.CasePcExternalPaymentFlow200_11: func(result domain.TransactionResult) *domain.DMLTicket {
		// Since the case was identified, we know ExternalTransfer workflow matches
		if result.PaymentCore != nil && result.PaymentCore.ExternalTransfer.Workflow.RunID != "" {
			return &domain.DMLTicket{
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
WHERE run_id = %s
AND state = 200
AND attempt = 11;`,
				RollbackTemplate: `UPDATE workflow_execution
SET state = 200,
    attempt = 11,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 200)
WHERE run_id = %s;`,
				TargetDB: "PC",
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: result.PaymentCore.ExternalTransfer.Workflow.RunID, Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: result.PaymentCore.ExternalTransfer.Workflow.RunID, Type: "string"},
				},
				CaseType: domain.CasePcExternalPaymentFlow200_11,
			}
		}
		return nil
	},

	// ========================================
	// Payment Engine (PE) Templates
	// ========================================
	domain.CasePeTransferPayment210_0: func(result domain.TransactionResult) *domain.DMLTicket {
		return &domain.DMLTicket{
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
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_payment'
AND state = 210;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.StreamMessage', NULL,
      '$.State', 210)
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_payment';`,
			TargetDB: "PE",
			DeployParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.PaymentEngine.Workflow.RunID, Type: "string"},
			},
			RollbackParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.PaymentEngine.Workflow.RunID, Type: "string"},
			},
			CaseType: domain.CasePeTransferPayment210_0,
		}
	},
	domain.CasePe2200FastCashinFailed: func(result domain.TransactionResult) *domain.DMLTicket {
		if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
			return &domain.DMLTicket{
				DeployTemplate: `-- pe_220_0_fast_cashin_failed
UPDATE workflow_execution
SET attempt = 1,
    state = 221,
    ` + "`data`" + ` = JSON_SET(
      ` + "`data`" + `,
      '$.State', 221,
      '$.StreamMessage.Status', 'FAILED',
      '$.StreamMessage.ErrorMessage', 'MANUAL REJECT')
WHERE run_id = %s
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
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_collection';`,
				TargetDB: "PE",
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
				CaseType: domain.CasePe2200FastCashinFailed,
			}
		}
		return nil
	},
	domain.CasePeStuck230RepublishPC: func(result domain.TransactionResult) *domain.DMLTicket {
		if runID := getInternalPaymentFlowRunID(result); runID != "" {
			return &domain.DMLTicket{
				DeployTemplate: `-- pe_stuck_230_republish_pc
UPDATE workflow_execution
SET state = 902,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 902)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow'
AND state = 900;`,
				RollbackTemplate: `UPDATE workflow_execution
SET state = 900,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 900)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow'
AND state = 902;`,
				TargetDB: "PC",
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
				CaseType: domain.CasePeStuck230RepublishPC,
			}
		}
		return nil
	},
	domain.CaseThoughtMachineFalseNegative: func(result domain.TransactionResult) *domain.DMLTicket {
		if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
			// Check if prev_trans_id is available for rollback
			if result.PaymentEngine.Workflow.PrevTransID == "" {
				// Return nil to indicate validation failure
				return nil
			}

			return &domain.DMLTicket{
				DeployTemplate: `-- thought_machine_false_negative
UPDATE workflow_execution
SET state = 230,
	 prev_trans_id = JSON_EXTRACT(data, '$.StreamMessage.ReferenceID'),
	 ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 230)
WHERE run_id = %s
AND state = 701;`,
				RollbackTemplate: `UPDATE workflow_execution
SET
	 state = 701,
	 attempt = 0,
	 prev_trans_id = %s,
	 ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 701)
WHERE run_id = %s
AND state = 230;`,
				TargetDB: "PE",
				DeployParams: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
				RollbackParams: []domain.ParamInfo{
					{Name: "prev_trans_id", Value: result.PaymentEngine.Workflow.PrevTransID, Type: "string"},
					{Name: "run_id", Value: runID, Type: "string"},
				},
				CaseType: domain.CaseThoughtMachineFalseNegative,
			}
		}
		return nil
	},

	// ========================================
	// RPP (Real-time Payment Processing) Templates
	// ========================================
	domain.CasePcExternalPaymentFlow201_0RPP210: func(result domain.TransactionResult) *domain.DMLTicket {
		return &domain.DMLTicket{
			DeployTemplate: `-- RPP 210, PE 220, PC 201. No response from RPP. Move to 222 to resume. ACSP
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 222)
WHERE run_id = %s
AND state = 210;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 201,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 201)
WHERE run_id = %s;`,
			TargetDB: "RPP",
			DeployParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			RollbackParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			CaseType: domain.CasePcExternalPaymentFlow201_0RPP210,
		}
	},

	domain.CasePcExternalPaymentFlow201_0RPP900: func(result domain.TransactionResult) *domain.DMLTicket {
		return &domain.DMLTicket{
			DeployTemplate: `-- RPP 900, PE 220, PC 201. Republish from RPP to resume. ACSP
UPDATE workflow_execution
SET state = 301,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 301)
WHERE run_id = %s
AND state = 900;`,
			RollbackTemplate: `UPDATE workflow_execution
SET state = 900,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 900)
WHERE run_id = %s;`,
			TargetDB: "RPP",
			DeployParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			RollbackParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			CaseType: domain.CasePcExternalPaymentFlow201_0RPP900,
		}
	},

	domain.CaseRppCashoutReject101_19: func(result domain.TransactionResult) *domain.DMLTicket {
		return &domain.DMLTicket{
			DeployTemplate: `-- rpp_cashout_reject_101_19, manual reject
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id = %s
AND state = 101
AND workflow_id = 'wf_ct_cashout';`,
			RollbackTemplate: `-- RPP Rollback: Move workflows back to state 101
UPDATE workflow_execution
SET state = 101,
    attempt = 0,
    data = JSON_SET(data, '$.State', 101)
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashout';`,
			TargetDB: "RPP",
			DeployParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			RollbackParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			CaseType: domain.CaseRppCashoutReject101_19,
		}
	},

	domain.CaseRppQrPaymentReject210_0: func(result domain.TransactionResult) *domain.DMLTicket {
		return &domain.DMLTicket{
			DeployTemplate: `-- rpp_qr_payment_reject_210_0, manual reject
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(data, '$.State', 221)
WHERE run_id = %s
AND state = 210
AND workflow_id = 'wf_ct_qr_payment';`,
			RollbackTemplate: `-- RPP Rollback: Move qr_payment workflows back to state 210
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    data = JSON_SET(data, '$.State', 210)
WHERE run_id = %s
AND workflow_id = 'wf_ct_qr_payment';`,
			TargetDB: "RPP",
			DeployParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			RollbackParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			CaseType: domain.CaseRppQrPaymentReject210_0,
		}
	},

	domain.CaseRppNoResponseResume: func(result domain.TransactionResult) *domain.DMLTicket {
		return &domain.DMLTicket{
			DeployTemplate: `-- rpp_no_response_resume_acsp
-- RPP did not respond in time, but status at Paynet is ACSP (Accepted Settlement in Process) or ACTC (Accepted Technical Validation)
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 222)
WHERE run_id = %s
AND state = 210
AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');`,
			RollbackTemplate: `-- RPP Rollback: Move workflows back to state 210
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 210)
WHERE run_id = %s
AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');`,
			TargetDB: "RPP",
			DeployParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			RollbackParams: []domain.ParamInfo{
				{Name: "run_id", Value: result.RPPAdapter.Workflow.RunID, Type: "string"},
			},
			CaseType: domain.CaseRppNoResponseResume,
		}
	},
}
