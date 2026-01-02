package adapters

import "buddy/internal/txn/domain"

// registerPPETemplates registers all Partnerpay Engine (PPE) templates
func registerPPETemplates(templates map[domain.Case]TemplateFunc) {
	templates[domain.CaseEcotxnChargeFailedCaptureFailedTMError] = ecotxnChargeFailedCaptureFailedTMError
}

// ecotxnChargeFailedCaptureFailedTMError handles Ecotxn charge failed with capture failed and TM error
func ecotxnChargeFailedCaptureFailedTMError(result domain.TransactionResult) *domain.DMLTicket {
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
}
