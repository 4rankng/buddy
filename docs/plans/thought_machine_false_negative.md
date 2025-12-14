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
  'DE9FD4A8-F738-407A-9E15-D0439CF87DAE'
) AND state = 701;

PE_Rollback.sql


UPDATE workflow_execution SET  state = 701,
  prev_trans_id = '{OrigPrevTransID}',
  `data` = JSON_SET(`data`, '$.State', 701)
WHERE run_id in (
  'DE9FD4A8-F738-407A-9E15-D0439CF87DAE'
) AND state = 230;



