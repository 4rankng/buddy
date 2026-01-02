package adapters

import (
	"buddy/internal/txn/domain"
	"buddy/internal/utils"
)

// registerPEAdvancedTemplates registers advanced Payment Engine (PE) templates
func registerPEAdvancedTemplates(templates map[domain.Case]TemplateFunc) {
	templates[domain.CasePeStuck230RepublishPC] = peStuck230RepublishPC
	templates[domain.CasePeStuck300RppNotFound] = peStuck300RppNotFound
	templates[domain.CaseCashoutPe220Pc201Reject] = cashoutPe220Pc201Reject
	templates[domain.CaseRpp210Pe220Pc201Reject] = rpp210Pe220Pc201Reject
}

// peStuck230RepublishPC handles PE stuck at 230 - republish to PC
func peStuck230RepublishPC(result domain.TransactionResult) *domain.DMLTicket {
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
}

// peStuck300RppNotFound handles PE stuck at 300 when RPP is not found
func peStuck300RppNotFound(result domain.TransactionResult) *domain.DMLTicket {
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
}

// cashoutPe220Pc201Reject handles cashout PE 220, PC 201 reject
func cashoutPe220Pc201Reject(result domain.TransactionResult) *domain.DMLTicket {
	if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
		return &domain.DMLTicket{
			Deploy: []domain.TemplateInfo{
				{
					TargetDB: "PE",
					SQLTemplate: "-- cashout_pe220_pc201_reject\n" +
						"UPDATE workflow_execution\n" +
						"SET state = 221, attempt = 1, `data` = JSON_SET(\n" +
						"      `data`, '$.StreamMessage',\n" +
						"      JSON_OBJECT(\n" +
						"         'Status', 'FAILED',\n" +
						"         'ErrorCode', \"ADAPTER_ERROR\",\n" +
						"         'ErrorMessage', 'Manual Rejected'\n" +
						"      ),\n" +
						"   '$.State', 221)\n" +
						"WHERE run_id IN (%s) AND state = 220 AND workflow_id = 'workflow_transfer_payment';",
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			Rollback: []domain.TemplateInfo{
				{
					TargetDB: "PE",
					SQLTemplate: "UPDATE workflow_execution\n" +
						"SET state = 220, attempt = 1, `data` = JSON_SET(\n" +
						"      `data`, '$.StreamMessage',\n" +
						"      JSON_OBJECT(),\n" +
						"   '$.State', 220)\n" +
						"WHERE run_id IN (%s);",
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			CaseType: domain.CaseCashoutPe220Pc201Reject,
		}
	}
	return nil
}

// rpp210Pe220Pc201Reject handles RPP 210, PE 220, PC 201 - manual rejection
// Updates PE workflow_transfer_payment to state 221 with failure details
func rpp210Pe220Pc201Reject(result domain.TransactionResult) *domain.DMLTicket {
	if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
		return &domain.DMLTicket{
			Deploy: []domain.TemplateInfo{
				{
					TargetDB: "PE",
					SQLTemplate: "-- rpp210_pe220_pc201_reject\n" +
						"UPDATE workflow_execution\n" +
						"SET state = 221, attempt = 1, `data` = JSON_SET(\n" +
						"      `data`, '$.StreamMessage',\n" +
						"      JSON_OBJECT(\n" +
						"         'Status', 'FAILED',\n" +
						"         'ErrorCode', \"ADAPTER_ERROR\",\n" +
						"         'ErrorMessage', 'Manual Rejected'\n" +
						"      ),\n" +
						"   '$.State', 221)\n" +
						"WHERE run_id IN (%s) AND state = 220 AND workflow_id = 'workflow_transfer_payment';",
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			Rollback: []domain.TemplateInfo{
				{
					TargetDB: "PE",
					SQLTemplate: "UPDATE workflow_execution\n" +
						"SET state = 220, attempt = 1, `data` = JSON_SET(\n" +
						"      `data`, '$.StreamMessage',\n" +
						"      JSON_OBJECT(),\n" +
						"   '$.State', 220)\n" +
						"WHERE run_id IN (%s);",
					Params: []domain.ParamInfo{
						{Name: "run_id", Value: runID, Type: "string"},
					},
				},
			},
			CaseType: domain.CaseRpp210Pe220Pc201Reject,
		}
	}
	return nil
}
