package adapters

import "buddy/internal/txn/domain"

// registerPEBasicTemplates registers basic Payment Engine (PE) templates
func registerPEBasicTemplates(templates map[domain.Case]TemplateFunc) {
	templates[domain.CasePeTransferPayment210_0] = peTransferPayment210_0
	templates[domain.CasePeStuckAtLimitCheck102] = peStuckAtLimitCheck102
	templates[domain.CasePe2200FastCashinFailed] = pe2200FastCashinFailed
}

// peTransferPayment210_0 handles PE transfer payment stuck at state 210, attempt 0
func peTransferPayment210_0(result domain.TransactionResult) *domain.DMLTicket {
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
}

// peStuckAtLimitCheck102 handles PE stuck at limit check state 102
func peStuckAtLimitCheck102(result domain.TransactionResult) *domain.DMLTicket {
	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "PE",
				SQLTemplate: "-- Fix: Manually reject the transaction by moving PE state to 221 and injecting an error StreamMessage\n" +
					"UPDATE workflow_execution\n" +
					"SET  state = 221, attempt = 1, `data` = JSON_SET(\n" +
					"      `data`, '$.StreamMessage',\n" +
					"      JSON_OBJECT(\n" +
					"         'Status', 'FAILED',\n" +
					"         'ErrorCode', \"ADAPTER_ERROR\",\n" +
					"         'ErrorMessage', 'Manual Rejected'\n" +
					"      ),\n" +
					"   '$.State', 221)\n" +
					"WHERE run_id IN (%s) AND state = 102 AND workflow_id = 'workflow_transfer_payment';",
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: result.PaymentEngine.Workflow.RunID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "PE",
				SQLTemplate: "-- Rollback: Revert the transaction to its original state\n" +
					"UPDATE workflow_execution\n" +
					"SET  state = 102, attempt = 4, `data` = JSON_SET(\n" +
					"      `data`, '$.StreamMessage',\n" +
					"      JSON_OBJECT(),\n" +
					"   '$.State', 102)\n" +
					"WHERE run_id IN (\n" +
					"   %s\n" +
					");",
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: result.PaymentEngine.Workflow.RunID, Type: "string"},
				},
			},
		},
		CaseType: domain.CasePeStuckAtLimitCheck102,
	}
}

// pe2200FastCashinFailed handles PE 220, attempt 0 - fast cashin failed
func pe2200FastCashinFailed(result domain.TransactionResult) *domain.DMLTicket {
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
}
