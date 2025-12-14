ADD this case




### thought_machine_false_negative


### [2] transaction_id: ced4efe76ea442ddbbca1f745ebe2386
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: FAILED
created_at: 2025-12-11T16:26:46.966221Z
reference_id: F58A163C-FEC9-42C2-B135-E8D97BB8C067
external_id: 20251212GXSPMYKL040OQR32194316
workflow_transfer_payment:
   state=stCaptureFailed(701) attempt=0
   run_id=F58A163C-FEC9-42C2-B135-E8D97BB8C067
[payment-core]
NOT_FOUND
[rpp-adapter]
req_biz_msg_id: 20251212GXSPMYKL040OQR32194316
partner_tx_id: ced4efe76ea442ddbbca1f745ebe2386
workflow_wf_ct_qr_payment: state= attempt=0 run_id=ced4efe76ea442ddbbca1f745ebe2386
info: RPP Status: PROCESSING
[Classification]


- **Case**: Thought Machine returning errors/false negatives, but transaction was successful. PE stuck or PC stuck 200
- **Fix**: Patch data to retry flow; Move PE to 230 and retry PC capture


so we can extract ReferenceID=JSON_EXTRACT(data, '$.StreamMessage.ReferenceID')
in payment-engine workflow_execution table
and store original OrigPrevTransID = workflow_execution.prev_trans_id

PE_Deploy.sql




# 20251202GXSPMYKL010ORB62198922
UPDATE workflow_execution SET state = 230,
  prev_trans_id = '{ReferenceID}',
  `data` = JSON_SET(`data`, '$.State', 230)
WHERE run_id in (
  '{run_id}'
) AND state = 701;

PE_Rollback.sql


UPDATE workflow_execution
SET
  state = 701,
  attempt=0,
  prev_trans_id = '{OrigPrevTransID}',
  `data` = JSON_SET(`data`, '$.State', 701)
WHERE run_id in (
  '{run_id}'
);



DEBUG



frank.nguyen@DBSG-H4M0DVF2C7 buddy % cd /Users/frank.nguyen/Documents/buddy && make deploy && mybuddy txn TS-4476
.txt
Building mybuddy with Malaysia environment...
mybuddy built successfully
Building sgbuddy with Singapore environment...
sgbuddy built successfully
Building and deploying binaries...
Building mybuddy with Malaysia environment...
mybuddy built successfully
Building sgbuddy with Singapore environment...
sgbuddy built successfully
Deployed to /Users/frank.nguyen/bin
You can now use 'mybuddy' and 'sgbuddy' commands from anywhere.
[MY] Processing batch file: TS-4476.txt
[MY] Processing batch file: TS-4476.txt
Processing 3 transaction IDs from TS-4476.txt
DEBUG: Evaluating condition - Field: PaymentCore.InternalTxns, Operator: eq, Value: <nil>, FieldValue: []
DEBUG: Nil comparison result for PaymentCore.InternalTxns: false
DEBUG: Evaluating condition - Field: PaymentCore.InternalTxns, Operator: eq, Value: <nil>, FieldValue: []
DEBUG: Nil comparison result for PaymentCore.InternalTxns: false
Results written to TS-4476.txt-output.txt
Summary: 
  Total: 3
  Unmatched: 2
  Matched: 1
Case Type Breakdown
  pc_external_payment_flow_200_11: 0
  pc_external_payment_flow_201_0_RPP_210: 0
  pc_external_payment_flow_201_0_RPP_900: 0
  pe_transfer_payment_210_0: 0
  pe_stuck_230_republish_pc: 1
  thought_machine_false_negative: 0
  pe_220_0_fast_cashin_failed: 0
  rpp_cashout_reject_101_19: 0
  rpp_qr_payment_reject_210_0: 0
  rpp_no_response_resume: 0


  why no match found
   thought_machine_false_negative: 0