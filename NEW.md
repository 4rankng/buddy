implement this in code base
 so that for this case where payment-core is stuck waiting for rpp-adapter message, we should republish message from rpp
 ## CONTEXT
 ### [1] e2e_id: 20260204GXSPMYKL010ORB00010461
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2026-02-03T16:04:45.013614Z
reference_id: 6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6
workflow_transfer_payment:
   state=workflow_transfer_payment:220 (stTransferProcessing) attempt=0
   run_id=6DEE383E-5ECD-4BFB-B4E1-9FAF444D7CE6

[payment-core]
internal_auth: SUCCESS
   tx_id=24fd946f809a4f4a9daca82819a8fe2e
   group_id=5f7eff81c5c848e7b7ad03b1ab19e022
   type=AUTH
   error_code='' error_msg=''
   workflow:
      workflow_id=internal_payment_flow
      state=internal_payment_flow:900 (stSuccess) attempt=0
      run_id=24fd946f809a4f4a9daca82819a8fe2e
external_transaction:
   ref_id=2da4581ac78b407598460f78f8cb74f4
   group_id=5f7eff81c5c848e7b7ad03b1ab19e022
   type=TRANSFER status=PROCESSING
   workflow:
      workflow_id=external_payment_flow
      state=external_payment_flow:201 (stProcessing) attempt=0
      run_id=2da4581ac78b407598460f78f8cb74f4
## DML

UPDATE workflow_execution 
SET attempt=1, state=301, data=JSON_SET(data, '$.State', 301)
WHERE workflow_id='wf_ct_cashout' AND run_id='5f7eff81c5c848e7b7ad03b1ab19e022' AND attempt=0 AND state=900;


Here is SOP

### `pc_stuck_201_waiting_rpp_republish_from_rpp`
- **Condition**: PC external_payment_flow stuck at state 201/0, PE workflow_transfer_payment at state 220/0, RPP wf_ct_cashout at state 900/0. Payment-core is waiting for RPP adapter message but RPP has already completed successfully.
- **Diagnosis**: RPP workflow reached success state (900) but failed to publish the success message to payment-core. Payment-core remains stuck in processing state (201) waiting for the response. PE workflow also stuck waiting for the callback.
- **Resolution**: Republish the success message from RPP by moving RPP workflow state to 301 (stSuccessPublish), which will trigger the message publication to downstream systems and unblock payment-core.
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- pc_stuck_201_waiting_rpp_republish_from_rpp - Republish success message from RPP to unblock PC
  UPDATE workflow_execution
  SET state = 301,
      attempt = 1,
      data = JSON_SET(data, '$.State', 301)
  WHERE workflow_id = 'wf_ct_cashout'
  AND run_id = '{RPP_RUN_ID}'
  AND attempt = 0
  AND state = 900;
  ```