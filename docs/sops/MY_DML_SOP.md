# DML SOP: Payment Transaction Fix Protocols

Based on the discussion log provided, here are the extracted case studies of stuck payment transactions (Payment Core - PC, Payment Engine - PE, RPP Adapter) and the corresponding fixes.

## Case Studies

### pc_external_payment_flow_stuck_200_attempt_11
- **Case**: Payment Core (PC) stuck at state 200 with max attempts (11)
- **Fix**: Reject the transaction manually as it has not reached PayNet/RPP yet
- **References**:
  - [DML 43008](https://doorman.infra.prd.g-bank.app/rds/dml/43008)
  - [DML 42990](https://doorman.infra.prd.g-bank.app/rds/dml/42990)

### rpp_no_response_resume_acsp
- **Case**: RPP 210, PE 220, PC 201. RPP did not respond in time, but status at Paynet is ACSP (Accepted Settlement in Process) or ACTC (Accepted Technical Validation)
- **Fix**: Move RPP adapter state to 222 to resume the workflow
- **References**:
  - [DML 43011](https://doorman.infra.prd.g-bank.app/rds/dml/43011)
  - [DML 42921](https://doorman.infra.prd.g-bank.app/rds/dml/42921)

### rpp_no_response_reject_not_found
- **Case**: RPP 210. No response from RPP and transaction does not exist at RPP/Paynet side
- **Fix**: Move RPP adapter state to 221 to reject (or manual reject PE stuck 210)
- **References**:
  - [DML 42997](https://doorman.infra.prd.g-bank.app/rds/dml/42997)
  - [DML 42648](https://doorman.infra.prd.g-bank.app/rds/dml/42648)

### pe_stuck_230_republish_pc
- **Case**: Payment Engine (PE) stuck at 230
- **Fix**: Republish PC (Payment Core) CAPTURE message to resume the workflow
- **References**:
  - [DML 42624](https://doorman.infra.prd.g-bank.app/rds/dml/42624)
  - [DML 42784](https://doorman.infra.prd.g-bank.app/rds/dml/42784)

### pe_stuck_223_hystrix_timeout
- **Case**: PE stuck at 223 (stTransferCompleted) or 220 due to Hystrix timeout during transition (Context not saved properly)
- **Fix**: Reset state to previous known good state (e.g., 221) and reset attempt count to 1 to retry the transition
- **Note**: Do NOT cancel (400) if the money has already moved (ACSP/ACTC)
- **References**:
  - [DML 42836](https://doorman.infra.prd.g-bank.app/rds/dml/42836)
  - [DML 42828](https://doorman.infra.prd.g-bank.app/rds/dml/42828)

### thought_machine_false_negative
- **Case**: Thought Machine returning errors/false negatives, but transaction was successful. PE stuck or PC stuck 200

payment engine workflow state = 701
payment core  internal_payment_flow state 500

- **Fix**: Patch data to retry flow; Move PE to 230 and retry PC capture
- **References**:
  - [DML 42991](https://doorman.infra.prd.g-bank.app/rds/dml/42991)
  - [DML 42927](https://doorman.infra.prd.g-bank.app/rds/dml/42927)

PE_Deploy.sql
# 20251202GXSPMYKL010ORB62198922
UPDATE workflow_execution SET state = 230,
  prev_trans_id = '6e0daa5cfcc24478a2c55097fe2d7cf8',
  `data` = JSON_SET(`data`, '$.State', 230)
WHERE run_id in (
  'DE9FD4A8-F738-407A-9E15-D0439CF87DAE'
) AND state = 701;

PE_Rollback.sql
UPDATE workflow_execution SET  state = 701,
  prev_trans_id = '24c37293816942c6bfcab8205ec81604',
  `data` = JSON_SET(`data`, '$.State', 701)
WHERE run_id in (
  'DE9FD4A8-F738-407A-9E15-D0439CF87DAE'
) AND state = 230;

PC_Deploy.sql
UPDATE workflow_execution SET state = 0, attempt = 1,
  `data` = JSON_SET(
	  `data`,
		'$.State', 0
	)
WHERE run_id in (
	'403b0708001a42868649a22ffbc4d7ae'
) AND workflow_id = 'internal_payment_flow' and state = 500;

PC_Rollback.sql
UPDATE workflow_execution SET state = 500, attempt = 0,
  `data` = JSON_SET(
	  `data`,
		'$.State', 500
	)
WHERE run_id in (
	'403b0708001a42868649a22ffbc4d7ae'
) AND workflow_id = 'internal_payment_flow';

### rpp_adapter_publish_failure_311`
- **Case**: Cash out RPP adapter stuck at 301 or 311 (stSuccessPublish/stPrepareFailurePublish) but failed to publish to Kafka
- **Fix**: Resume publish failed stream on 311 or set attempt to 1 to resume
- **References**:
  - [DML 42702](https://doorman.infra.prd.g-bank.app/rds/dml/42702)
  - [DML 42850](https://doorman.infra.prd.g-bank.app/rds/dml/42850)

### cash_in_stuck_100_update_mismatch
- **Case**: Cash in workflow stuck at state 100 with attempts. Update operation failing due to updatedAt mismatch between table and workflow execution
- **Fix**: Update updatedAt in workflow data and resume from state 100
- **References**:
  - [DML 42880](https://doorman.infra.prd.g-bank.app/rds/dml/42880)
  - [DML 42697](https://doorman.infra.prd.g-bank.app/rds/dml/42697)

### rpp_cashin_validation_failed_122_0
- **Case**: RPP Cashin workflow stuck at state 122 (stFieldsValidationFailed) with attempt 0
- **Fix**: Reset workflow to state 100 (stTransferPersisted) with attempt 1 to retry validation
- **References**:
  - [Add DML reference if available]

### user_name_change_qr_invalidation
- **Case**: User changed name, old QR code needs to be invalidated to force generation of new one
- **Fix**: DML to mark specific QR entry as INACTIVE
- **References**:
  - [DML 42999](https://doorman.infra.prd.g-bank.app/rds/dml/42999)
  - [DML 42917](https://doorman.infra.prd.g-bank.app/rds/dml/42917)

## Summary of Fix Protocols

### Safety Checks
When running DMLs, always include the current state in the WHERE clause (e.g., `WHERE workflow_id='...' AND state=223`) to avoid accidental state changes if the workflow moved while the ticket was pending.

### ACSP/ACTC Rule
If RPP status is ACSP or ACTC, you cannot Cancel (400). You must Resume/Republish.

### Refunds
If automatic refund fails, use the "Retry Refund" flow (upload CSV to S3) before attempting manual credit.





pe_stuck_300_rpp_not_found

Case: Payment Engine (PE) stuck at state 300 (stAuthCompleted) with attempt=0. Payment Core (PC) shows internal auth success (State 900). internal capture is nil
RPP/PayNet is nill

Fix: Manually reject the transaction by moving PE state to 221 and injecting an error StreamMessage.

SQL Deploy:

UPDATE workflow_execution
SET  state = 221, attempt = 1, `data` = JSON_SET(
      `data`, '$.StreamMessage',
      JSON_OBJECT(
         'Status', 'FAILED',
         'ErrorCode', "ADAPTER_ERROR",
         'ErrorMessage', 'Manual Rejected'
      ),
   '$.State', 221)
WHERE run_id IN (
   '72022562-BE9A-4CE9-B4EA-9D50668AEE59' -- Replace with actual run_id
) AND
   state = 300 AND
   workflow_id = 'workflow_transfer_payment';


SQL Rollback:

UPDATE workflow_execution
SET  state = 210, attempt =0, `data` = JSON_SET(
      `data`, '$.StreamMessage',
      JSON_OBJECT(
         'TxID', 'ebf23ba301874602af8782186e61658d', -- Replace with actual TxID if known, or restore original object
         'Status', 'SUCCESS',
         'ErrorCode', '',
         'ExternalID', '',
         'ReferenceID', 'f6d0b41d4ae54b4db9e1e40ce5d1ac38',
         'ErrorMessage', '',
         'ValueTimestamp', '2024-12-20T14:37:12.189499Z'
      ),
   '$.State', 210)
WHERE run_id IN (
   '72022562-BE9A-4CE9-B4EA-9D50668AEE59'
) AND workflow_id = 'workflow_transfer_payment';
