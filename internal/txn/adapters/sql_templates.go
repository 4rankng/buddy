package adapters

import "buddy/internal/txn/domain"

// DMLTicket represents a SQL generation request with templates
type DMLTicket struct {
	RunIDs           []string // run_ids to update
	ReqBizMsgIDs     []string // optional req_biz_msg_ids for RPP cases
	PartnerTxIDs     []string // optional partner_tx_ids for RPP cases
	DeployTemplate   string   // SQL template for deploy
	RollbackTemplate string   // SQL template for rollback
	TargetDB         string   // "PC", "PE", or "RPP"
	WorkflowID       string   // optional workflow_id filter
	TargetState      int      // target state to check in WHERE clause
	TargetAttempt    int      // target attempt to check in WHERE clause
	StateField       string   // field name for state in WHERE clause (usually "state")
	WorkflowIDs      []string // multiple workflow_ids for IN clause

	// Consolidation metadata
	TransactionCount int // Number of transactions consolidated
}

// TemplateConfig defines the parameters required for a SQL template
type TemplateConfig struct {
	Parameters []string // List of parameter types: ["run_ids"], ["run_ids", "workflow_ids"]
}

// templateConfigs maps SOP cases to their template parameter configurations
var templateConfigs = map[domain.Case]TemplateConfig{
	domain.CasePcExternalPaymentFlow200_11:      {Parameters: []string{"run_ids"}},
	domain.CasePcExternalPaymentFlow201_0RPP210: {Parameters: []string{"run_ids"}},
	domain.CasePcExternalPaymentFlow201_0RPP900: {Parameters: []string{"run_ids"}},
	domain.CasePeTransferPayment210_0:           {Parameters: []string{"run_ids"}},
	domain.CasePeStuck230RepublishPC:            {Parameters: []string{"run_ids"}},
	domain.CasePe2200FastCashinFailed:           {Parameters: []string{"run_ids"}},
	domain.CaseRppCashoutReject101_19:           {Parameters: []string{"run_ids"}},
	domain.CaseRppQrPaymentReject210_0:          {Parameters: []string{"run_ids"}},
	domain.CaseRppNoResponseResume:              {Parameters: []string{"run_ids", "workflow_ids"}},
}

// sqlTemplates maps SOP cases to their DML tickets
var sqlTemplates = map[domain.Case]func(domain.TransactionResult) *DMLTicket{
	// ========================================
	// Payment Core (PC) Templates
	// ========================================
	domain.CasePcExternalPaymentFlow200_11: func(result domain.TransactionResult) *DMLTicket {
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
	domain.CasePeTransferPayment210_0: func(result domain.TransactionResult) *DMLTicket {
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
	domain.CasePe2200FastCashinFailed: func(result domain.TransactionResult) *DMLTicket {
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
	domain.CasePeStuck230RepublishPC: func(result domain.TransactionResult) *DMLTicket {
		if runIDs := getInternalPaymentFlowRunIDs(result); len(runIDs) > 0 {
			return &DMLTicket{
				RunIDs: runIDs,
				DeployTemplate: `-- pe_stuck_230_republish_pc
UPDATE workflow_execution
SET state = 902,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 902)
WHERE run_id IN (%s)
AND workflow_id = 'internal_payment_flow'
AND state = 900;`,
				RollbackTemplate: `UPDATE workflow_execution
SET state = 900,
    attempt = 1,
    ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 900)
WHERE run_id IN (%s)
AND workflow_id = 'internal_payment_flow'
AND state = 902;`,
				TargetDB:    "PC",
				WorkflowID:  "internal_payment_flow",
				TargetState: 900,
			}
		}
		return nil
	},

	// ========================================
	// RPP (Real-time Payment Processing) Templates
	// ========================================
	domain.CasePcExternalPaymentFlow201_0RPP210: func(result domain.TransactionResult) *DMLTicket {
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

	domain.CasePcExternalPaymentFlow201_0RPP900: func(result domain.TransactionResult) *DMLTicket {
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

	domain.CaseRppCashoutReject101_19: func(result domain.TransactionResult) *DMLTicket {
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

	domain.CaseRppQrPaymentReject210_0: func(result domain.TransactionResult) *DMLTicket {
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

	domain.CaseRppNoResponseResume: func(result domain.TransactionResult) *DMLTicket {
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
