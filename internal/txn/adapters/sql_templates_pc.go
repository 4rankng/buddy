package adapters

import "buddy/internal/txn/domain"

// registerPCTemplates registers all Payment Core (PC) templates
func registerPCTemplates(templates map[domain.Case]TemplateFunc) {
	templates[domain.CasePcExternalPaymentFlow200_11] = pcExternalPaymentFlow200_11
}

// pcExternalPaymentFlow200_11 handles PC external payment flow stuck at state 200, attempt 11
func pcExternalPaymentFlow200_11(result domain.TransactionResult) *domain.DMLTicket {
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
}
