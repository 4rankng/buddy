# DML SOP: Payment Transaction Fix Protocols

This document outlines standard operating procedures (SOPs) for resolving stuck payment transactions across Payment Core (PC), Payment Engine (PE), and RPP Adapters.

---

## **Table of Content**

1. [Payment Core (PC) Issues](#payment-core-pc-issues)
2. [Payment Engine (PE) Issues](#payment-engine-pe-issues)
3. [RPP Adapter Issues](#rpp-adapter-issues)
4. [Cross-Domain / Complex Issues](#cross-domain--complex-issues)
5. [General Safety Guidelines](#general-safety-guidelines)

---

## **Payment Core (PC) Issues**

### `pc_external_payment_flow_stuck_200_attempt_11`
- **Condition**: internal_payment_flow stuck at state 200 with max attempts (11).
- **Diagnosis**: Transaction has likely not reached PayNet/RPP yet.
- **Resolution**: Reject the transaction manually by moving state to 202 (Failed).
- **References**:
  - [DML 43008](https://doorman.infra.prd.g-bank.app/rds/dml/43008)
  - [DML 42990](https://doorman.infra.prd.g-bank.app/rds/dml/42990)
- **Sample Deploy Script**:
  ```sql
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
  WHERE run_id = {RUN_ID}
  AND state = 200
  AND attempt = 11;
  ```

---

## **Payment Engine (PE) Issues**

### `pe_stuck_230_republish_pc`
- **Condition**: Payment Engine (PE) stuck at state 230.
- **Diagnosis**: PE waiting for confirmation but process stalled.
- **Resolution**: Republish PC (Payment Core) CAPTURE message to resume the workflow.
- **References**:
  - [DML 42624](https://doorman.infra.prd.g-bank.app/rds/dml/42624)
  - [DML 42784](https://doorman.infra.prd.g-bank.app/rds/dml/42784)
- **Sample Deploy Script** (TargetDB: PC):
  ```sql
  UPDATE workflow_execution
  SET state = 902,
      attempt = 1,
      data = JSON_SET(data, '$.State', 902)
  WHERE run_id = {PC_RUN_ID}
  AND workflow_id = 'internal_payment_flow'
  AND state = 900;
  ```

### `pe_stuck_223_hystrix_timeout`
- **Condition**: PE stuck at 223 (stTransferCompleted) or 220 due to Hystrix timeout during transition (Context not saved properly).
- **Resolution**: Reset state to previous known good state (e.g., 221) and reset attempt count to 1 to retry the transition.
- **References**:
  - [DML 42836](https://doorman.infra.prd.g-bank.app/rds/dml/42836)
  - [DML 42828](https://doorman.infra.prd.g-bank.app/rds/dml/42828)
- **Warning**: Do NOT cancel (400) if the money has already moved (ACSP/ACTC).

### `pe_stuck_300_rpp_not_found`
- **Condition**: PE stuck at state 300 (stAuthCompleted) with attempt=0. PC shows internal auth success (State 900). Internal capture is nil. RPP/PayNet is nil.
- **Resolution**: Manually reject the transaction by moving PE state to 221 and injecting an error StreamMessage.
- **Sample Deploy Script**:
  ```sql
  UPDATE workflow_execution
  SET  state = 221, attempt = 1, `data` = JSON_SET(
        `data`, '$.StreamMessage',
        JSON_OBJECT(
           'Status', 'FAILED',
           'ErrorCode', "ADAPTER_ERROR",
           'ErrorMessage', 'Manual Rejected'
        ),
     '$.State', 221)
  WHERE run_id = {PE_RUN_ID}
  AND state = 300
  AND workflow_id = 'workflow_transfer_payment';
  ```

---

## **RPP Adapter Issues**

### `rpp_no_response_resume_acsp`
- **Condition**: RPP 210, PE 220, PC 201. RPP did not respond in time, but status at Paynet is ACSP or ACTC.
- **Resolution**: Move RPP adapter state to 222 to resume the workflow.
- **References**:
  - [DML 43011](https://doorman.infra.prd.g-bank.app/rds/dml/43011)
  - [DML 42921](https://doorman.infra.prd.g-bank.app/rds/dml/42921)
- **Sample Deploy Script**:
  ```sql
  UPDATE workflow_execution
  SET state = 222,
      attempt = 1,
      data = JSON_SET(data, '$.State', 222)
  WHERE run_id = {RUN_ID}
  AND state = 210
  AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');
  ```

### `rpp_no_response_reject_not_found`
- **Condition**: RPP 210. No response from RPP and transaction does not exist at RPP/Paynet side.
- **Resolution**: Move RPP adapter state to 221 to reject (or manual reject PE stuck 210).
- **References**:
  - [DML 42997](https://doorman.infra.prd.g-bank.app/rds/dml/42997)
  - [DML 42648](https://doorman.infra.prd.g-bank.app/rds/dml/42648)
- **Sample Deploy Script** (Example for QR Payment):
  ```sql
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(data, '$.State', 221)
  WHERE run_id = {RUN_ID}
  AND state = 210
  AND workflow_id = 'wf_ct_qr_payment';
  ```

### `rpp_no_response_reject_not_found_state_0`
- **Condition**: RPP wf_ct_qr_payment stuck at state 0 (stInit) with any attempt count. PE 220, PC 201. Adapter never sent request to PayNet.
- **Diagnosis**: RPP adapter stuck in initialization loop; transaction does not exist at PayNet side.
- **Resolution**: Move RPP adapter state to 221 to reject the transaction manually.
- **References**:
  - Similar to: `rpp_no_response_reject_not_found` (State 210 variant)
- **Sample Deploy Script**:
  ```sql
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(data, '$.State', 221)
  WHERE run_id = {RUN_ID}
  AND state = 0
  AND workflow_id = 'wf_ct_qr_payment';
  ```

### `rpp_adapter_publish_failure_311`
- **Condition**: Cash out RPP adapter stuck at 301 or 311 (stSuccessPublish/stPrepareFailurePublish) but failed to publish to Kafka.
- **Resolution**: Resume publish failed stream on 311 or set attempt to 1 to resume.
- **References**:
  - [DML 42702](https://doorman.infra.prd.g-bank.app/rds/dml/42702)
  - [DML 42850](https://doorman.infra.prd.g-bank.app/rds/dml/42850)

### `rpp_cashin_validation_failed_122_0`
- **Condition**: RPP Cashin workflow stuck at state 122 (stFieldsValidationFailed) with attempt 0.
- **Resolution**: Reset workflow to state 100 (stTransferPersisted) with attempt 1 to retry validation.
- **Sample Deploy Script**:
  ```sql
  UPDATE workflow_execution
  SET state = 100,
      attempt = 1,
      data = JSON_SET(data, '$.State', 100)
  WHERE run_id = {RUN_ID}
  AND workflow_id = 'wf_ct_cashin'
  AND state = 122;
  ```

---

## **Cross-Domain / Complex Issues**

### `thought_machine_false_negative`
- **Condition**: Thought Machine returning errors/false negatives, but transaction was successful. PE stuck or PC stuck 200.
- **Resolution**: Patch data to retry flow; Move PE to 230 and retry PC capture.
- **References**:
  - [DML 42991](https://doorman.infra.prd.g-bank.app/rds/dml/42991)
  - [DML 42927](https://doorman.infra.prd.g-bank.app/rds/dml/42927)
- **Sample Deploy Script**:
  ```sql
  -- PE Side
  UPDATE workflow_execution
  SET state = 230,
      prev_trans_id = data->>'$.StreamMessage.ReferenceID',
      data = JSON_SET(data, '$.State', 230)
  WHERE run_id = {PE_RUN_ID}
  AND state = 701;

  -- PC Side (Restart Capture)
  UPDATE workflow_execution
  SET state = 0,
      attempt = 1,
      data = JSON_SET(data, '$.State', 0)
  WHERE run_id = {PC_RUN_ID}
  AND workflow_id = 'internal_payment_flow'
  AND state = 500;
  ```

### `cash_in_stuck_100_update_mismatch`
- **Condition**: Cash in workflow stuck at state 100 with attempts. Update operation failing due to updatedAt mismatch.
- **Resolution**: Update updatedAt in workflow data and resume from state 100.
- **References**:
  - [DML 42880](https://doorman.infra.prd.g-bank.app/rds/dml/42880)
  - [DML 42697](https://doorman.infra.prd.g-bank.app/rds/dml/42697)

### `user_name_change_qr_invalidation`
- **Condition**: User changed name, old QR code needs to be invalidated to force generation of new one.
- **Resolution**: DML to mark specific QR entry as INACTIVE.
- **References**:
  - [DML 42999](https://doorman.infra.prd.g-bank.app/rds/dml/42999)
  - [DML 42917](https://doorman.infra.prd.g-bank.app/rds/dml/42917)

---

## **General Safety Guidelines**

### 1. Safety Checks
When running DMLs, **always** include the current state in the `WHERE` clause (e.g., `WHERE workflow_id='...' AND state=223`) to avoid accidental state changes if the workflow moved while the ticket was pending.

### 2. ACSP/ACTC Rule
If RPP status is **ACSP** (Accepted Settlement in Process) or **ACTC** (Accepted Technical Validation), you **cannot** Cancel (400). You must Resume/Republish to ensure consistency.

### 3. Refunds
If automatic refund fails, use the "Retry Refund" flow (upload CSV to S3) before attempting manual credit.
