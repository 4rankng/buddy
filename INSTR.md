why the transaction below does not match
	{
			CaseType:    domain.CaseEcotxnChargeFailedCaptureFailedTMError,
			Description: "Ecotxn Charge Failed Capture Failed with TMError",
			Conditions: []RuleCondition{
				{
					FieldPath: "PartnerpayEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_charge",
				},
				{
					FieldPath: "PartnerpayEngine.Workflow.State",
					Operator:  "eq",
					Value:     "502",
				},
				{
					FieldPath: "PartnerpayEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PartnerpayEngine.Charge.StatusReason",
					Operator:  "eq",
					Value:     "SYSTEM_ERROR",
				},
				{
					FieldPath: "PartnerpayEngine.Charge.StatusReasonDescription",
					Operator:  "eq",
					Value:     "error occurred in Thought Machine.",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "500",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},

### [1] transaction_id: fd230a01dcd04282851b7b9dd6260c93
[partnerpay-engine]
charge.status: FAILED SYSTEM_ERROR error occurred in Thought Machine.
workflow_charge: stFailureNotified(502) Attempt=0 run_id=fd230a01dcd04282851b7b9dd6260c93

[payment-core]
internal_capture: FAILED
   tx_id=3550ca0d10df4b0ab2dce80218cdf51f
   type=CAPTURE
   error_code='SYSTEM_ERROR' error_msg='error occurred in Thought Machine.'
   workflow:
      workflow_id=internal_payment_flow
      state=stFailed(500) attempt=0
      run_id=3550ca0d10df4b0ab2dce80218cdf51f
internal_auth: SUCCESS
   tx_id=ce8c05866d134bb488038644c708740e
   type=AUTH
   error_code='' error_msg=''
   workflow:
      workflow_id=internal_payment_flow
      state=stSuccess(900) attempt=0
      run_id=ce8c05866d134bb488038644c708740e



Also this is wrong

DML files written successfully for case: NOT_FOUND

We should nto create any DML files for case NOT_FOUND