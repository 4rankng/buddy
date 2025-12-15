the case below should be detected as

{
			CaseType:    domain.CaseThoughtMachineFalseNegative,
			Description: "Thought Machine returning errors/false negatives, but transaction was successful",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "701", // stCaptureFailed
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "500", // stFailed
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},

nitialize Jira client for [my]...
[MY] Querying transaction: 20251213GXSPMYKL010ORB33800412
### [1] e2e_id: 20251213GXSPMYKL010ORB33800412
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: FAILED
created_at: 2025-12-12T16:06:09.75018Z
reference_id: 37032302-27F5-4F60-95A4-D0100BA7775B
external_id: 20251213GXSPMYKL010ORB33800412
workflow_transfer_payment:
   state=stCaptureFailed(701) attempt=0
   run_id=37032302-27F5-4F60-95A4-D0100BA7775B

[payment-core]
internal_transaction:
   tx_id=3cebe07d53874174874fe97f61c5e80f
   group_id=721a751f39b64baf9fb1ef98525cf40a
   type=AUTH status=SUCCESS
   workflow:
      workflow_id=internal_payment_flow
      state=stSuccess(900) attempt=0
      run_id=3cebe07d53874174874fe97f61c5e80f
internal_transaction:
   tx_id=05749b4159464bbd882a728785ab68f5
   group_id=721a751f39b64baf9fb1ef98525cf40a
   type=CAPTURE status=FAILED
   workflow:
      workflow_id=internal_payment_flow
      state=stFailed(500) attempt=0
      run_id=05749b4159464bbd882a728785ab68f5
external_transaction:
   ref_id=08419e9921b94fec9b6e140bbb6c5b4d
   group_id=721a751f39b64baf9fb1ef98525cf40a
   type=TRANSFER status=SUCCESS
   workflow:
      workflow_id=external_payment_flow
      state=stPrepareSuccessPublish(900) attempt=0
      run_id=08419e9921b94fec9b6e140bbb6c5b4d

[rpp-adapter]
req_biz_msg_id: 20251213GXSPMYKL010ORB33800412
partner_tx_id: 721a751f39b64baf9fb1ef98525cf40a
wf_ct_cashout:
   state=stSuccess(900) attempt=0
   run_id=721a751f39b64baf9fb1ef98525cf40a
info: RPP Status: PROCESSING

[Classification]
NOT_FOUND


[MY] No SQL statements generated. Transaction may not require remediation or case conditions were not met.
frank.nguyen@DBSG-H4M0DVF2C7 buddy %