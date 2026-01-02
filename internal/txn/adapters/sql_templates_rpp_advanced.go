package adapters

import "buddy/internal/txn/domain"

// registerRPPAdvancedTemplates registers advanced RPP (Real-time Payment Processing) templates
func registerRPPAdvancedTemplates(templates map[domain.Case]TemplateFunc) {
	templates[domain.CasePcExternalPaymentFlow201_0RPP210] = pcExternalPaymentFlow201_0RPP210
	templates[domain.CasePcExternalPaymentFlow201_0RPP900] = pcExternalPaymentFlow201_0RPP900
	templates[domain.CaseRppRtpCashinStuck200_0] = rppRtpCashinStuck200_0
	templates[domain.CaseRpp210Pe220Pc201Accept] = rpp210Pe220Pc201Accept
}

// pcExternalPaymentFlow201_0RPP210 handles PC 201, RPP 210 - no response from RPP
// Note: This is cross-domain (PC + RPP) but registered here for organization
func pcExternalPaymentFlow201_0RPP210(result domain.TransactionResult) *domain.DMLTicket {
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
}

// pcExternalPaymentFlow201_0RPP900 handles PC 201, RPP 900 - republish from RPP
// Note: This is cross-domain (PC + RPP) but registered here for organization
func pcExternalPaymentFlow201_0RPP900(result domain.TransactionResult) *domain.DMLTicket {
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
}

// rppRtpCashinStuck200_0 handles RPP RTP cashin stuck at state 200, attempt 0
func rppRtpCashinStuck200_0(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find workflow matching case criteria
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"wf_ct_rtp_cashin",
		"200",
		0,
	)

	if runID == "" {
		return nil
	}

	// Get partner_tx_id for PPE SQL
	partnerTxID := result.RPPAdapter.PartnerTxID
	if partnerTxID == "" {
		return nil
	}

	return &domain.DMLTicket{
		Deploy: []domain.TemplateInfo{
			{
				TargetDB: "PPE",
				SQLTemplate: `-- rpp_rtp_cashin_stuck_200_0
UPDATE intent SET status = 'UPDATED'
WHERE intent_id = %s
AND status = 'CONFIRMED';`,
				Params: []domain.ParamInfo{
					{Name: "intent_id", Value: partnerTxID, Type: "string"},
				},
			},
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_rtp_cashin_stuck_200_0
UPDATE workflow_execution
SET state = 110,
   attempt = 1,
   data = JSON_SET(data, '$.State', 110)
WHERE run_id = %s
AND state = 200
AND workflow_id = 'wf_ct_rtp_cashin';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		Rollback: []domain.TemplateInfo{
			{
				TargetDB: "PPE",
				SQLTemplate: `-- rpp_rtp_cashin_stuck_200_0 Rollback
UPDATE intent SET status = 'CONFIRMED'
WHERE intent_id = %s;`,
				Params: []domain.ParamInfo{
					{Name: "intent_id", Value: partnerTxID, Type: "string"},
				},
			},
			{
				TargetDB: "RPP",
				SQLTemplate: `-- rpp_rtp_cashin_stuck_200_0 Rollback
UPDATE workflow_execution
SET state = 200,
   attempt = 0,
   data = JSON_SET(data, '$.State', 200)
WHERE run_id = %s
AND workflow_id = 'wf_ct_rtp_cashin';`,
				Params: []domain.ParamInfo{
					{Name: "run_id", Value: runID, Type: "string"},
				},
			},
		},
		CaseType: domain.CaseRppRtpCashinStuck200_0,
	}
}

// rpp210Pe220Pc201Accept handles RPP 210, PE 220, PC 201 - resume to success
// Updates RPP wf_ct_cashout to state 222 (stTransferManualResumeReceived)
func rpp210Pe220Pc201Accept(result domain.TransactionResult) *domain.DMLTicket {
	if result.RPPAdapter == nil {
		return nil
	}

	// Find workflow with state 210 and workflow_id in (wf_ct_cashout, wf_ct_qr_payment)
	runID := getRPPWorkflowRunIDByCriteria(
		result.RPPAdapter.Workflow,
		"", // allow any workflow_id, refined in SQL template
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
				SQLTemplate: `-- rpp210_pe220_pc201_accept - RPP did not respond in time, ACSP status at Paynet. Move to 222 to resume.
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
				SQLTemplate: `UPDATE workflow_execution
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
		CaseType: domain.CaseRpp210Pe220Pc201Accept,
	}
}
