package adapters

import (
	"buddy/internal/txn/domain"
	"buddy/internal/utils"
)

// getRPPWorkflowRunIDByCriteria finds and returns the run_id of a workflow matching specific criteria.
// Parameters:
//   - workflows: slice of workflows to search
//   - workflowID: workflow_id to match (empty string means any workflow_id)
//   - state: state to match (empty string means any state)
//   - attempt: attempt number to match (-1 means any attempt)
//
// Returns empty string if no matching workflow is found.
func getRPPWorkflowRunIDByCriteria(workflows []domain.WorkflowInfo, workflowID, state string, attempt int) string {
	for _, wf := range workflows {
		// Check workflow_id if specified
		if workflowID != "" && wf.WorkflowID != workflowID {
			continue
		}
		// Check state if specified
		if state != "" && wf.State != state {
			continue
		}
		// Check attempt if specified (and not -1 which means any attempt)
		if attempt != -1 && wf.Attempt != attempt {
			continue
		}
		// All criteria matched
		return wf.RunID
	}
	// No matching workflow found
	return ""
}

// sqlTemplates maps SOP cases to their DML tickets
var sqlTemplates = map[domain.Case]func(domain.TransactionResult) *domain.DMLTicket{
	// ========================================
	// Payment Core (PC) Templates
	// ========================================
	domain.CasePcExternalPaymentFlow200_11: func(result domain.TransactionResult) *domain.DMLTicket {
		// Since the case was identified, we know ExternalTransfer workflow matches
		if result.PaymentCore != nil && result.PaymentCore.ExternalTransfer.Workflow.RunID != "" {
			return &domain.DMLTicket{
				Deploy: []domain.TemplateInfo{
					{
						TargetDB: "PC",
						SQLTemplate: `-- pc_external_payment_flow_200_11
UPDATE workflow_execution
SET state = 202,
    attempt = 1,
    data = JSON_SET(
      data,
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
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: result.PaymentCore.ExternalTransfer.Workflow.RunID, Type: "string"},
						},
					},
				},
				Rollback: []domain.TemplateInfo{
					{
						TargetDB: "PC",
						SQLTemplate: `UPDATE workflow_execution
SET state = 200,
    attempt = 11,
    data = JSON_SET(data, '$.State', 200)
WHERE run_id = %s;`,
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: result.PaymentCore.ExternalTransfer.Workflow.RunID, Type: "string"},
						},
					},
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
			Deploy: []domain.TemplateInfo{
				{
					TargetDB: "PE",
					SQLTemplate: `-- Reject PE stuck 210. Reject transactions since it hasn't reached Paynet yet
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(
      data,
      '$.StreamMessage', JSON_OBJECT(
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ErrorMessage', 'Manual Rejected'),
      '$.State', 221)
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_payment'
AND state = 210;`,
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: result.PaymentEngine.Workflow.RunID, Type: "string"},
					},
				},
			},
			Rollback: []domain.TemplateInfo{
				{
					TargetDB: "PE",
					SQLTemplate: `UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    data = JSON_SET(
      data,
      '$.StreamMessage', NULL,
      '$.State', 210)
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_payment';`,
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: result.PaymentEngine.Workflow.RunID, Type: "string"},
					},
				},
			},
			CaseType: domain.CasePeTransferPayment210_0,
		}
	},
	domain.CasePe2200FastCashinFailed: func(result domain.TransactionResult) *domain.DMLTicket {
		if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
			return &domain.DMLTicket{
				Deploy: []domain.TemplateInfo{
					{
						TargetDB: "PE",
						SQLTemplate: `-- pe_220_0_fast_cashin_failed
UPDATE workflow_execution
SET attempt = 1,
    state = 221,
    data = JSON_SET(
      data,
      '$.State', 221,
      '$.StreamMessage.Status', 'FAILED',
      '$.StreamMessage.ErrorMessage', 'MANUAL REJECT')
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_collection'
AND state = 220
AND attempt = 0;`,
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: runID, Type: "string"},
						},
					},
				},
				Rollback: []domain.TemplateInfo{
					{
						TargetDB: "PE",
						SQLTemplate: `UPDATE workflow_execution
SET attempt = 0,
    state = 220,
    data = JSON_SET(
      data,
      '$.State', 220,
      '$.StreamMessage', JSON_OBJECT())
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_collection';`,
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: runID, Type: "string"},
						},
					},
				},
				CaseType: domain.CasePe2200FastCashinFailed,
			}
		}
		return nil
	},
	domain.CasePeStuck230RepublishPC: func(result domain.TransactionResult) *domain.DMLTicket {
		if runID := getInternalPaymentFlowRunID(result); runID != "" {
			return &domain.DMLTicket{
				Deploy: []domain.TemplateInfo{
					{
						TargetDB: "PC",
						SQLTemplate: `-- pe_stuck_230_republish_pc
UPDATE workflow_execution
SET state = 902,
    attempt = 1,
    data = JSON_SET(data, '$.State', 902)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow'
AND state = 900;`,
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: runID, Type: "string"},
						},
					},
				},
				Rollback: []domain.TemplateInfo{
					{
						TargetDB: "PC",
						SQLTemplate: `UPDATE workflow_execution
SET state = 900,
    attempt = 1,
    data = JSON_SET(data, '$.State', 900)
WHERE run_id = %s
AND workflow_id = 'internal_payment_flow'
AND state = 902;`,
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: runID, Type: "string"},
						},
					},
				},
				CaseType: domain.CasePeStuck230RepublishPC,
			}
		}
		return nil
	},

	// ========================================
	// RPP (Real-time Payment Processing) Templates
	// ========================================
	domain.CasePcExternalPaymentFlow201_0RPP210: func(result domain.TransactionResult) *domain.DMLTicket {
		if result.RPPAdapter == nil {
			return nil
		}

		// Find workflow with state 210 (any workflow_id for this case)
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
					SQLTemplate: `-- RPP 210, PE 220, PC 201. No response from RPP. Move to 222 to resume. ACSP
	UPDATE workflow_execution
	SET state = 222,
	    attempt = 1,
	    data = JSON_SET(data, '$.State', 222)
	WHERE run_id = %s
	AND state = 210;`,
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			Rollback: []domain.TemplateInfo{
				{
					TargetDB: "RPP",
					SQLTemplate: `UPDATE workflow_execution
	SET state = 201,
	    attempt = 0,
	    data = JSON_SET(data, '$.State', 201)
	WHERE run_id = %s;`,
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			CaseType: domain.CasePcExternalPaymentFlow201_0RPP210,
		}
	},

	domain.CasePcExternalPaymentFlow201_0RPP900: func(result domain.TransactionResult) *domain.DMLTicket {
		if result.RPPAdapter == nil {
			return nil
		}

		// Find workflow with state 900 (any workflow_id for this case)
		runID := getRPPWorkflowRunIDByCriteria(
			result.RPPAdapter.Workflow,
			"", // any workflow_id
			"900",
			-1, // any attempt
		)

		if runID == "" {
			return nil
		}

		return &domain.DMLTicket{
			Deploy: []domain.TemplateInfo{
				{
					TargetDB: "RPP",
					SQLTemplate: `-- RPP 900, PE 220, PC 201. Republish from RPP to resume. ACSP
	UPDATE workflow_execution
	SET state = 301,
	    attempt = 1,
	    data = JSON_SET(data, '$.State', 301)
	WHERE run_id = %s
	AND state = 900;`,
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			Rollback: []domain.TemplateInfo{
				{
					TargetDB: "RPP",
					SQLTemplate: `UPDATE workflow_execution
	SET state = 900,
	    attempt = 0,
	    data = JSON_SET(data, '$.State', 900)
	WHERE run_id = %s;`,
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			CaseType: domain.CasePcExternalPaymentFlow201_0RPP900,
		}
	},

	domain.CaseRppCashoutReject101_19: func(result domain.TransactionResult) *domain.DMLTicket {
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
					SQLTemplate: `-- RPP Rollback: Move workflows back to state 101
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
	},

	domain.CaseRppQrPaymentReject210_0: func(result domain.TransactionResult) *domain.DMLTicket {
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
					SQLTemplate: `-- RPP Rollback: Move qr_payment workflows back to state 210
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
	},

	domain.CaseRppNoResponseResume: func(result domain.TransactionResult) *domain.DMLTicket {
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
					SQLTemplate: `-- RPP Rollback: Move workflows back to state 210
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
	},

	domain.CaseRppCashinValidationFailed122_0: func(result domain.TransactionResult) *domain.DMLTicket {
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
					SQLTemplate: `-- RPP Rollback: Move cashin workflow back to state 122
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
	},

	domain.CaseThoughtMachineFalseNegative: func(result domain.TransactionResult) *domain.DMLTicket {
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
    prev_trans_id = JSON_EXTRACT(data, '$.StreamMessage.ReferenceID'),
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
	},

	domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess: func(result domain.TransactionResult) *domain.DMLTicket {
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
	},
	domain.CaseEcotxnChargeFailedCaptureFailedTMError: func(result domain.TransactionResult) *domain.DMLTicket {
		if result.PartnerpayEngine != nil && result.PartnerpayEngine.Workflow.RunID != "" {
			// Get original updated_at from charge record
			originalUpdatedAt := result.PartnerpayEngine.Charge.UpdatedAt

			return &domain.DMLTicket{
				Deploy: []domain.TemplateInfo{
					{
						TargetDB: "PPE",
						SQLTemplate: `-- ecotxn_ChargeFailed_CaptureFailed_TMError
-- Move to AuthCompleted and wait for cron to cancel the transaction
UPDATE charge SET
status = 'PROCESSING',
updated_at = %s
WHERE transaction_id = %s;

UPDATE workflow_execution
SET state = 300, data = JSON_SET(data, '$.State', 300,
'$.ChargeStorage.Status', 'PROCESSING')
WHERE run_id = %s
AND workflow_id = 'workflow_charge'
AND state = 502
AND attempt = 0;`,
						Params: []domain.ParamInfo{
							{Name: "updated_at", Value: originalUpdatedAt, Type: "string"},
							{Name: "transaction_id", Value: result.PartnerpayEngine.Workflow.RunID, Type: "string"},
							{Name: "run_id", Value: result.PartnerpayEngine.Workflow.RunID, Type: "string"},
						},
					},
				},
				Rollback: []domain.TemplateInfo{
					{
						TargetDB: "PPE",
						SQLTemplate: `-- ecotxn_ChargeFailed_CaptureFailed_TMError Rollback
UPDATE charge SET
status = 'FAILED',
updated_at = %s
WHERE transaction_id = %s;

UPDATE workflow_execution
SET state = 502, data = JSON_SET(data, '$.State', 502,
'$.ChargeStorage.Status', 'FAILED')
WHERE run_id = %s
AND workflow_id = 'workflow_charge';`,
						Params: []domain.ParamInfo{
							{Name: "updated_at", Value: originalUpdatedAt, Type: "string"},
							{Name: "transaction_id", Value: result.PartnerpayEngine.Workflow.RunID, Type: "string"},
							{Name: "run_id", Value: result.PartnerpayEngine.Workflow.RunID, Type: "string"},
						},
					},
				},
				CaseType: domain.CaseEcotxnChargeFailedCaptureFailedTMError,
			}
		}
		return nil
	},
	domain.CasePeStuck300RppNotFound: func(result domain.TransactionResult) *domain.DMLTicket {
		if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
			// Get rollback stream message
			rollbackStreamMessage := utils.GetRollbackStreamMessage(result.PaymentEngine.Workflow.Data)

			return &domain.DMLTicket{
				Deploy: []domain.TemplateInfo{
					{
						TargetDB: "PE",
						SQLTemplate: `-- pe_stuck_300_rpp_not_found
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    data = JSON_SET(
      data,
      '$.StreamMessage', JSON_OBJECT(
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ErrorMessage', 'Manual Rejected'),
      '$.State', 221)
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_payment'
AND state = 300
AND attempt = 0;`,
						Params: []domain.ParamInfo{
							{Name: "run_id", Value: runID, Type: "string"},
						},
					},
				},
				Rollback: []domain.TemplateInfo{
					{
						TargetDB: "PE",
						SQLTemplate: `-- pe_stuck_300_rpp_not_found rollback
UPDATE workflow_execution
SET state = 300,
    attempt = 0,
    data = JSON_SET(
      data,
      '$.StreamMessage', %s,
      '$.State', 300)
WHERE run_id = %s
AND workflow_id = 'workflow_transfer_payment';`,
						Params: []domain.ParamInfo{
							{Name: "stream_message", Value: rollbackStreamMessage, Type: "string"},
							{Name: "run_id", Value: runID, Type: "string"},
						},
					},
				},
				CaseType: domain.CasePeStuck300RppNotFound,
			}
		}
		return nil
	},
}
