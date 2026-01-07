# DML SOP: Payment Transaction Fix Protocols

This document outlines standard operating procedures (SOPs) for resolving stuck payment transactions across Payment Core (PC), Payment Engine (PE), RPP Adapters, and Partnerpay Engine (PPE).

---

## **Table of Content**

1. [Cashout Workflows](#cashout-workflows)
2. [Cashin Workflows](#cashin-workflows)
3. [QR Payment Workflows](#qr-payment-workflows)
4. [Partnerpay / Ecotxn Workflows](#partnerpay--ecotxn-workflows)
5. [Cross-Cutting / Infrastructure Issues](#cross-cutting--infrastructure-issues)
6. [General Safety Guidelines](#general-safety-guidelines)

---

## **Cashout Workflows**

This section covers issues related to cashout transactions involving PC external payment flow, PE transfer payment, and RPP cashout workflows.

### `pc_external_payment_flow_stuck_200_attempt_11`
- **Condition**: PC internal_payment_flow stuck at state 200 with max attempts (11).
- **Diagnosis**: Transaction has likely not reached PayNet/RPP yet.
- **Resolution**: Reject the transaction manually by moving state to 202 (Failed).
- **References**:
  - [DML 43008](https://doorman.infra.prd.g-bank.app/rds/dml/43008)
  - [DML 42990](https://doorman.infra.prd.g-bank.app/rds/dml/42990)
- **Sample Deploy Script** (TargetDB: PC):
  ```sql
  -- pc_external_payment_flow_200_11
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
  WHERE run_id = '{RUN_ID}'
  AND state = 200
  AND attempt = 11;
  ```

### `pc_external_payment_flow_201_0_RPP_210`
- **Condition**: PC 201/0, RPP 210. RPP did not respond in time, but status at Paynet is ACSP or ACTC.
- **Diagnosis**: RPP adapter timed out waiting for response, but transaction succeeded at Paynet side.
- **Resolution**: Move RPP adapter state to 222 to resume the workflow.
- **References**:
  - [DML 43011](https://doorman.infra.prd.g-bank.app/rds/dml/43011)
  - [DML 42921](https://doorman.infra.prd.g-bank.app/rds/dml/42921)
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- pc_external_payment_flow_201_0_RPP_210 - RPP 210, PE 220, PC 201. No response from RPP. Move to 222 to resume. ACSP
  UPDATE workflow_execution
  SET state = 222,
      attempt = 1,
      data = JSON_SET(data, '$.State', 222)
  WHERE run_id = '{RUN_ID}'
  AND state = 210;
  ```

### `pc_external_payment_flow_201_0_RPP_900`
- **Condition**: PC 201/0, RPP 900. RPP reached success state but workflow not completing.
- **Diagnosis**: RPP adapter in success state but needs to trigger downstream flows.
- **Resolution**: Republish from RPP by moving to state 301.
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- pc_external_payment_flow_201_0_RPP_900 - RPP 900, PE 220, PC 201. Republish from RPP to resume. ACSP
  UPDATE workflow_execution
  SET state = 301,
      attempt = 1,
      data = JSON_SET(data, '$.State', 301)
  WHERE run_id = '{RUN_ID}'
  AND state = 900;
  ```

### `pe_stuck_230_republish_pc`
- **Condition**: Payment Engine (PE) stuck at state 230.
- **Diagnosis**: PE waiting for confirmation but process stalled.
- **Resolution**: Republish PC (Payment Core) CAPTURE message to resume the workflow.
- **References**:
  - [DML 42624](https://doorman.infra.prd.g-bank.app/rds/dml/42624)
  - [DML 42784](https://doorman.infra.prd.g-bank.app/rds/dml/42784)
- **Sample Deploy Script** (TargetDB: PC):
  ```sql
  -- pe_stuck_230_republish_pc
  UPDATE workflow_execution
  SET state = 902,
      attempt = 1,
      data = JSON_SET(data, '$.State', 902)
  WHERE run_id = '{PC_RUN_ID}'
  AND workflow_id = 'internal_payment_flow'
  AND state = 900;
  ```

### `pe_transfer_payment_stuck_210_0`
- **Condition**: PE workflow_transfer_payment stuck at state 210 with attempt 0.
- **Diagnosis**: Transaction has not reached Paynet yet and can be safely rejected.
- **Resolution**: Manually reject by moving PE to state 221 with error details.
- **Sample Deploy Script** (TargetDB: PE):
  ```sql
  -- pe_transfer_payment_210_0 - Reject PE stuck 210. Reject transactions since it hasn't reached Paynet yet
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(
        data,
        '$.StreamMessage', JSON_OBJECT(
          'Status', 'FAILED',
          'ErrorCode', 'ADAPTER_ERROR',
          'ErrorMessage', 'Manual Rejected'),
        '$.State', 221)
  WHERE run_id = '{RUN_ID}'
  AND workflow_id = 'workflow_transfer_payment'
  AND state = 210;
  ```

### `pe_stuck_at_limit_check_102_4`
- **Condition**: PE workflow_transfer_payment stuck at state 102 (limit check) with multiple attempts.
- **Diagnosis**: Payment authorization succeeded but PE workflow stuck. Need to inject AuthorisationID.
- **Resolution**: Reject PE workflow and update transfer table with AuthorisationID from PC internal_auth.
- **Important**:
  - Retrieve AuthorisationID from PC internal_payment table (TxID where TxType='AUTH'). This maintains the critical link between Payment Engine and Payment Core.
  - Only update transfer table if AuthorisationID is missing in transfer table properties AND set AuthorisationID in workflow_execution data properties only if it is missing
- **Sample Deploy Script** (TargetDB: PE):
  ```sql
  -- Reject/Reset the Workflow Execution (cashout_pe102_reject)
  -- Note: Only set AuthorisationID if it's missing in workflow_execution data properties
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      `data` = JSON_SET(
          `data`,
          '$.StreamMessage',
          JSON_OBJECT(
              'Status', 'FAILED',
              'ErrorCode', "ADAPTER_ERROR",
              'ErrorMessage', 'Manual Rejected'
          ),
          '$.State', 221,
          '$.Properties.AuthorisationID', '{AUTHORISATION_ID}'
      )
  WHERE run_id IN ('{RUN_ID}')
    AND state = 102
    AND workflow_id = 'workflow_transfer_payment';

  -- Update transfer table with AuthorisationID from payment-core internal_auth
  -- Note: Only execute if AuthorisationID is missing in transfer table properties
  UPDATE transfer
  SET properties = JSON_SET(properties, '$.AuthorisationID', '{AUTHORISATION_ID}'),
      updated_at = '{UPDATED_AT}'
  WHERE transaction_id = '{TRANSACTION_ID}';
  ```

### `pe_stuck_300_rpp_not_found`
- **Condition**: PE stuck at state 300 (stAuthCompleted) with attempt=0. PC shows internal auth success (State 900). Internal capture is nil. RPP/PayNet is nil.
- **Diagnosis**: PE workflow progressed but RPP adapter never initialized. Transaction must be rejected.
- **Resolution**: Manually reject the transaction by moving PE state to 221 and injecting an error StreamMessage.
- **Sample Deploy Script** (TargetDB: PE):
  ```sql
  -- pe_stuck_300_rpp_not_found
  UPDATE workflow_execution
  SET  state = 221, attempt = 1, `data` = JSON_SET(
        `data`, '$.StreamMessage',
        JSON_OBJECT(
           'Status', 'FAILED',
           'ErrorCode', "ADAPTER_ERROR",
           'ErrorMessage', 'Manual Rejected'
        ),
     '$.State', 221)
  WHERE run_id = '{PE_RUN_ID}'
  AND state = 300
  AND workflow_id = 'workflow_transfer_payment';
  ```

### `pe220_pc201_rpp0_stuck_init`
- **Condition**: PE 220/0, PC 201/0, RPP wf_ct_qr_payment stuck at State 0 (stInit). RPP adapter never initialized properly.
- **Diagnosis**: RPP adapter stuck in initialization loop; transaction does not exist at PayNet side. PE workflow fails naturally; PC and RPP need manual cleanup.
- **Resolution**: Clean up PC and RPP to fail gracefully. PC moves to State 202 (Failed), RPP moves to terminal State 700. PE rejection is handled separately.
- **References**:
  - Related: `rpp_no_response_reject_not_found` (RPP State 210 variant)
  - Related: `rpp_no_response_reject_not_found_state_0` (RPP State 0 SOP-only variant)
- **Sample Deploy Script** (Multi-Database):
  ```sql
  -- PC Side - Manual PC rejection
  UPDATE workflow_execution
  SET state = 202, attempt = 1,
      `data` = JSON_SET(`data`,
        '$.StreamResp', JSON_OBJECT(
          'TxID', '',
          'Status', 'FAILED',
          'ErrorCode', 'ADAPTER_ERROR',
          'ExternalID', '',
          'ErrorMessage', 'Reject from adapter'
        ),
        '$.State', 202)
  WHERE run_id IN ('{PC_RUN_ID}') AND state = 201 AND attempt = 0;

  -- RPP Side - Move to terminal state
  UPDATE workflow_execution
  SET state = 700,
      `data` = JSON_SET(`data`, '$.State', 700)
  WHERE run_id IN ('{RPP_RUN_ID}') AND state = 0;
  ```
- **Note**: Supports batch operations using `WHERE run_id IN (...)` for multiple transactions.

### `cashout_pe220_pc201_reject`
- **Condition**: Cashout with PE 220, PC 201. Transaction needs manual rejection.
- **Diagnosis**: Cashout workflow stuck and needs to be rejected.
- **Resolution**: Reject PE workflow by moving to state 221 with error details.
- **Sample Deploy Script** (TargetDB: PE):
  ```sql
  -- cashout_pe220_pc201_reject
  UPDATE workflow_execution
  SET state = 221, attempt = 1, `data` = JSON_SET(
        `data`, '$.StreamMessage',
        JSON_OBJECT(
           'Status', 'FAILED',
           'ErrorCode', "ADAPTER_ERROR",
           'ErrorMessage', 'Manual Rejected'
        ),
     '$.State', 221)
  WHERE run_id IN ('{RUN_ID}') AND state = 220 AND workflow_id = 'workflow_transfer_payment';
  ```
- **Note**: Supports batch operations using `WHERE run_id IN (...)` for multiple transactions.

### `cashout_rpp210_pe220_pc201`
- **Condition**: RPP 210, PE 220, PC 201. RPP did not respond in time, but status at Paynet is ACSP or ACTC.
- **Diagnosis**: Transaction succeeded at Paynet but adapter timed out. User must choose between resume or reject.
- **Resolution**: Interactive choice - either resume to success (state 222) or reject (state 221).
- **Important**: Do NOT reject (400/221) if money has already moved (ACSP/ACTC). Resume to success instead.
- **Sample Deploy Scripts**:

  **Option A: Resume to Success** (TargetDB: RPP)
  ```sql
  -- cashout_rpp210_pe220_pc201_accept - RPP did not respond in time, ACSP status at Paynet. Move to 222 to resume.
  UPDATE workflow_execution
  SET state = 222,
        attempt = 1,
        data = JSON_SET(data, '$.State', 222)
  WHERE run_id = '{RUN_ID}'
  AND state = 210
  AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');
  ```

  **Option B: Manual Rejection** (TargetDB: PE)
  ```sql
  -- cashout_rpp210_pe220_pc201_reject
  UPDATE workflow_execution
  SET state = 221, attempt = 1, `data` = JSON_SET(
        `data`, '$.StreamMessage',
        JSON_OBJECT(
           'Status', 'FAILED',
           'ErrorCode', "ADAPTER_ERROR",
           'ErrorMessage', 'Manual Rejected'
        ),
     '$.State', 221)
  WHERE run_id IN ('{RUN_ID}') AND state = 220 AND workflow_id = 'workflow_transfer_payment';
  ```
  - **Note**: Supports batch operations using `WHERE run_id IN (...)` for multiple transactions.

### `rpp_cashout_reject_101_19`
- **Condition**: RPP wf_ct_cashout stuck at state 101 with attempt 19 (max retries).
- **Diagnosis**: RPP cashout workflow exhausted retries without success.
- **Resolution**: Manually reject by moving RPP state to 221.
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_cashout_reject_101_19, manual reject
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(data, '$.State', 221)
  WHERE run_id = '{RUN_ID}'
  AND state = 101
  AND workflow_id = 'wf_ct_cashout';
  ```

### `rpp_no_response_resume_acsp`
- **Condition**: RPP 210, PE 220, PC 201. RPP did not respond in time, but status at Paynet is ACSP or ACTC.
- **Diagnosis**: RPP adapter timed out, but transaction succeeded at Paynet.
- **Resolution**: Move RPP adapter state to 222 to resume the workflow.
- **References**:
  - [DML 43011](https://doorman.infra.prd.g-bank.app/rds/dml/43011)
  - [DML 42921](https://doorman.infra.prd.g-bank.app/rds/dml/42921)
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_no_response_resume_acsp
  -- RPP did not respond in time, but status at Paynet is ACSP (Accepted Settlement in Process) or ACTC (Accepted Technical Validation)
  UPDATE workflow_execution
  SET state = 222,
      attempt = 1,
      data = JSON_SET(data, '$.State', 222)
  WHERE run_id = '{RUN_ID}'
  AND state = 210
  AND workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment');
  ```

### `rpp_no_response_reject_not_found`
- **Condition**: RPP 210. No response from RPP and transaction does not exist at RPP/Paynet side.
- **Diagnosis**: RPP adapter sent request but Paynet has no record. Safe to reject.
- **Resolution**: Move RPP adapter state to 221 to reject (or manual reject PE stuck 210).
- **References**:
  - [DML 42997](https://doorman.infra.prd.g-bank.app/rds/dml/42997)
  - [DML 42648](https://doorman.infra.prd.g-bank.app/rds/dml/42648)
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_no_response_reject_not_found (Example for QR Payment)
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(data, '$.State', 221)
  WHERE run_id = '{RUN_ID}'
  AND state = 210
  AND workflow_id = 'wf_ct_qr_payment';
  ```

### `rpp_no_response_reject_not_found_state_0`
- **Condition**: RPP wf_ct_qr_payment stuck at state 0 (stInit) with any attempt count. PE 220, PC 201. Adapter never sent request to PayNet.
- **Diagnosis**: RPP adapter stuck in initialization loop; transaction does not exist at PayNet side.
- **Resolution**: Move RPP adapter state to 221 to reject the transaction manually.
- **References**:
  - Similar to: `rpp_no_response_reject_not_found` (State 210 variant)
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_no_response_reject_not_found_state_0 - RPP adapter stuck in initialization, never sent to Paynet
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(data, '$.State', 221)
  WHERE run_id = '{RUN_ID}'
  AND state = 0
  AND workflow_id = 'wf_ct_qr_payment';
  ```

### `rpp_adapter_publish_failure_311`
- **Condition**: Cash out RPP adapter stuck at 301 or 311 (stSuccessPublish/stPrepareFailurePublish) but failed to publish to Kafka.
- **Diagnosis**: RPP workflow reached terminal state but message publish to Kafka failed.
- **Resolution**: Resume publish failed stream on 311 or set attempt to 1 to resume.
- **References**:
  - [DML 42702](https://doorman.infra.prd.g-bank.app/rds/dml/42702)
  - [DML 42850](https://doorman.infra.prd.g-bank.app/rds/dml/42850)
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_adapter_publish_failure_311 - Resume failed publish
  UPDATE workflow_execution
  SET state = 311,
      attempt = 1,
      data = JSON_SET(data, '$.State', 311)
  WHERE run_id = '{RUN_ID}'
  AND state IN (301, 311)
  AND workflow_id = 'wf_ct_cashout';
  ```

---

## **Cashin Workflows**

This section covers issues related to cashin transactions including RPP cashin, RTP cashin, PE collection, and Fast adapter issues.

### `rpp_cashin_validation_failed_122_0`
- **Condition**: RPP Cashin workflow stuck at state 122 (stFieldsValidationFailed) with attempt 0.
- **Diagnosis**: Initial validation failed but should be retried.
- **Resolution**: Reset workflow to state 100 (stTransferPersisted) with attempt 1 to retry validation.
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_cashin_validation_failed_122_0, retry validation
  UPDATE workflow_execution
  SET state = 100,
    attempt = 1,
    data = JSON_SET(data, '$.State', 100)
  WHERE run_id = '{RUN_ID}'
  AND workflow_id = 'wf_ct_cashin'
  AND state = 122;
  ```

### `rpp_rtp_cashin_stuck_200_0`
- **Condition**: RPP RTP cashin stuck at state 200, attempt 0.
- **Diagnosis**: RTP cashin workflow stuck waiting for response from PPE.
- **Resolution**: Update PPE intent status to UPDATED and reset RPP workflow to state 110.
- **Sample Deploy Script** (Multi-Database: PPE + RPP):
  ```sql
  -- rpp_rtp_cashin_stuck_200_0 - Update PPE intent
  UPDATE intent SET status = 'UPDATED'
  WHERE intent_id = '{INTENT_ID}'
  AND status = 'CONFIRMED';

  -- rpp_rtp_cashin_stuck_200_0 - Reset RPP workflow
  UPDATE workflow_execution
  SET state = 110,
     attempt = 1,
     data = JSON_SET(data, '$.State', 110)
  WHERE run_id = '{RUN_ID}'
  AND state = 200
  AND workflow_id = 'wf_ct_rtp_cashin';
  ```

### `pe_220_0_fast_cashin_failed`
- **Condition**: PE workflow_transfer_collection stuck at state 220, attempt 0. Fast cashin failed.
- **Diagnosis**: Fast adapter cashin failed and workflow needs to be rejected.
- **Resolution**: Move PE workflow to state 221 with error details.
- **Sample Deploy Script** (TargetDB: PE):
  ```sql
  -- pe_220_0_fast_cashin_failed
  UPDATE workflow_execution
  SET attempt = 1,
      state = 221,
      data = JSON_SET(
        data,
        '$.State', 221,
        '$.StreamMessage.Status', 'FAILED',
        '$.StreamMessage.ErrorMessage', 'MANUAL REJECT')
  WHERE run_id = '{RUN_ID}'
  AND workflow_id = 'workflow_transfer_collection'
  AND state = 220
  AND attempt = 0;
  ```

### `cash_in_stuck_100_retry`
- **Condition**: Cash in workflow stuck at state 100 with attempts. Update operation failing, but timestamps are consistent.
- **Diagnosis**: Compare `credit_transfer.updated_at` (UTC) with `workflow_execution.data->>$.CreditTransfer.UpdatedAt` (GMT+8). If timestamps match after timezone conversion, this is a simple retry case.
- **Resolution**: Simply set `attempt=1` to restart validation without modifying timestamp data.
- **Reference**: [DML 43211](https://doorman.infra.prd.g-bank.app/rds/dml/43211)
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_cashin_stuck_100_0, retry (timestamps match after timezone conversion)
  UPDATE workflow_execution
  SET attempt = 1
  WHERE run_id IN ('{RUN_ID}')
  AND workflow_id = 'wf_ct_cashin'
  AND state = 100;
  ```

### `cash_in_stuck_100_update_mismatch`
- **Condition**: Cash in workflow stuck at state 100 with attempts. Update operation failing due to actual updatedAt mismatch.
- **Diagnosis**: Compare `credit_transfer.updated_at` (UTC) with `workflow_execution.data->>$.CreditTransfer.UpdatedAt` (GMT+8). If timestamps don't match even after timezone conversion, this is an update mismatch case.
- **Important**: `credit_transfer.updated_at` is in UTC, while `data->>$.CreditTransfer.UpdatedAt` is in GMT+8. Timezone conversion is critical.
- **Resolution**: Update both `attempt=1` and sync the `UpdatedAt` in workflow data by converting from UTC to GMT+8.
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- cash_in_stuck_100_update_mismatch (actual timestamp mismatch)
  UPDATE workflow_execution
  SET
      attempt = 1,
      `data` = JSON_SET(`data`,
          '$.CreditTransfer.UpdatedAt', '{CONVERTED_TIMESTAMP}'  -- Convert from credit_transfer.updated_at (UTC) to GMT+8
      )
  WHERE
      run_id = '{RUN_ID}'
      AND workflow_id = 'wf_ct_cashin'
      AND state = 100;
  ```
- **Important Notes**:
  - Always verify the mismatch exists before applying this fix
  - Timezone conversion is critical: UTC → GMT+8 (add 8 hours to UTC timestamp)
  - Try `cash_in_stuck_100_retry` first if unsure about the mismatch

---

## **QR Payment Workflows**

This section covers issues specific to QR payment transactions.

### `rpp_qr_payment_reject_210_0`
- **Condition**: RPP QR payment stuck at state 210, attempt 0.
- **Diagnosis**: QR payment workflow at state 210 with no response from Paynet.
- **Resolution**: Manually reject by moving RPP state to 221.
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_qr_payment_reject_210_0, manual reject
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(data, '$.State', 221)
  WHERE run_id = '{RUN_ID}'
  AND state = 210
  AND workflow_id = 'wf_ct_qr_payment';
  ```

### `pe_capture_processing_pc_capture_failed_rpp_success`
- **Condition**: PE capture processing, PC internal capture failed (state 500), but RPP shows success (state 900, attempt 0).
- **Diagnosis**: RPP transaction succeeded but PC capture failed. Need to retry PC capture.
- **Resolution**: Restart PC internal payment flow from state 0.
- **Sample Deploy Script** (TargetDB: PC):
  ```sql
  -- pe_capture_processing_pc_capture_failed_rpp_success (restart PC capture flow from 0)
  UPDATE workflow_execution
  SET state = 0,
      attempt = 1,
      data = JSON_SET(data, '$.State', 0)
  WHERE run_id = '{PC_RUN_ID}'
  AND workflow_id = 'internal_payment_flow'
  AND state = 500;
  ```

---

## **Partnerpay / Ecotxn Workflows**

This section covers issues related to Partnerpay Engine (PPE) and Ecotxn charge workflows.

### `ecotxn_ChargeFailed_CaptureFailed_TMError`
- **Condition**: PPE charge workflow stuck at state 502 with attempt 0. ThoughtMachine returning errors but charge may have succeeded.
- **Diagnosis**: Ecotxn charge failed with capture failed and TM error. Need to move to AuthCompleted state.
- **Resolution**: Move charge status to PROCESSING and workflow to state 300, wait for cron to cancel.
- **Sample Deploy Script** (TargetDB: PPE):
  ```sql
  -- ecotxn_ChargeFailed_CaptureFailed_TMError
  -- Move to AuthCompleted and wait for cron to cancel the transaction
  UPDATE charge SET
  status = 'PROCESSING',
  updated_at = '{UPDATED_AT}'
  WHERE transaction_id = '{TRANSACTION_ID}';

  UPDATE workflow_execution
  SET state = 300, data = JSON_SET(data, '$.State', 300,
  '$.ChargeStorage.Status', 'PROCESSING')
  WHERE run_id = '{RUN_ID}'
  AND workflow_id = 'workflow_charge'
  AND state = 502
  AND attempt = 0;
  ```

---

## **Cross-Cutting / Infrastructure Issues**

This section covers infrastructure-level issues and problems that span multiple workflow types.

### `thought_machine_false_negative`
- **Condition**: Thought Machine returning errors/false negatives, but transaction was successful. PE stuck or PC stuck 200.
- **Diagnosis**: TM returned error but operation actually succeeded at infrastructure level.
- **Resolution**: Patch data to retry flow; Move PE to 230 and retry PC capture.
- **References**:
  - [DML 42991](https://doorman.infra.prd.g-bank.app/rds/dml/42991)
  - [DML 42927](https://doorman.infra.prd.g-bank.app/rds/dml/42927)
- **Sample Deploy Script** (Multi-Database: PE + PC):
  ```sql
  -- thought_machine_false_negative - PE Deploy
  UPDATE workflow_execution
  SET state = 230,
      prev_trans_id = data->>'$.StreamMessage.ReferenceID',
      data = JSON_SET(data, '$.State', 230)
  WHERE run_id = '{PE_RUN_ID}'
  AND state = 701;

  -- thought_machine_false_negative (restart PC capture flow from 0)
  UPDATE workflow_execution
  SET state = 0,
      attempt = 1,
      data = JSON_SET(data, '$.State', 0)
  WHERE run_id = '{PC_RUN_ID}'
  AND workflow_id = 'internal_payment_flow'
  AND state = 500;
  ```

### `pe_stuck_223_hystrix_timeout`
- **Condition**: PE stuck at 223 (stTransferCompleted) or 220 due to Hystrix timeout during transition (Context not saved properly).
- **Diagnosis**: Hystrix timeout during state transition caused partial state update.
- **Resolution**: Reset state to previous known good state (e.g., 221) and reset attempt count to 1 to retry the transition.
- **References**:
  - [DML 42836](https://doorman.infra.prd.g-bank.app/rds/dml/42836)
  - [DML 42828](https://doorman.infra.prd.g-bank.app/rds/dml/42828)
- **Warning**: Do NOT cancel (400) if the money has already moved (ACSP/ACTC).
- **Sample Deploy Script** (TargetDB: PE):
  ```sql
  -- pe_stuck_223_hystrix_timeout - Reset to previous good state
  UPDATE workflow_execution
  SET state = 221,
      attempt = 1,
      data = JSON_SET(
        data,
        '$.State', 221,
        '$.StreamMessage', JSON_OBJECT()
      )
  WHERE run_id = '{RUN_ID}'
  AND workflow_id = 'workflow_transfer_payment'
  AND state IN (220, 223);
  ```

### `rpp_process_registry_stuck_init`
- **Condition**: RPP wf_process_registry stuck at state 0 (stInit) with any attempt count.
- **Diagnosis**: Process registry workflow initialization retry exhausted.
- **Resolution**: Set attempt to 1 to allow retry of initialization.
- **Important**: When a cashout or QR payment is stuck, check for related wf_process_registry issues. Resuming the registry workflow may be the cleaner fix before attempting other interventions.
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- rpp_process_registry_stuck_init, set attempt=1 to retry initialization
  UPDATE workflow_execution
  SET attempt = 1
  WHERE run_id = '{RUN_ID}'
  AND workflow_id = 'wf_process_registry'
  AND state = 0;
  ```

### `user_name_change_qr_invalidation`
- **Condition**: User changed name, old QR code needs to be invalidated to force generation of new one.
- **Diagnosis**: User profile change requires QR code refresh.
- **Resolution**: DML to mark specific QR entry as INACTIVE.
- **References**:
  - [DML 42999](https://doorman.infra.prd.g-bank.app/rds/dml/42999)
  - [DML 42917](https://doorman.infra.prd.g-bank.app/rds/dml/42917)
- **Sample Deploy Script** (TargetDB: RPP):
  ```sql
  -- user_name_change_qr_invalidation - Mark QR code as inactive
  UPDATE qr_code
  SET status = 'INACTIVE',
      updated_at = NOW()
  WHERE user_id = '{USER_ID}'
  AND status = 'ACTIVE';
  ```

---

## **General Safety Guidelines**

### 1. Safety Checks
When running DMLs, **always** include the current state in the `WHERE` clause (e.g., `WHERE workflow_id='...' AND state=223`) to avoid accidental state changes if the workflow moved while the ticket was pending.

### 2. ACSP/ACTC Rule
If RPP status is **ACSP** (Accepted Settlement in Process) or **ACTC** (Accepted Technical Validation), you **cannot** Cancel (400). You must Resume/Republish to ensure consistency.

### 3. Refunds
If automatic refund fails, use the "Retry Refund" flow (upload CSV to S3) before attempting manual credit.

### 4. Multi-Database Transactions
For cases involving multiple databases (PC, PE, RPP, PPE):
- Always execute scripts in the correct database order specified
- Verify each script's success before proceeding to the next
- Keep track of all run_ids for rollback purposes

### 5. Timestamp Consistency
When updating the transfer table or workflow_execution data, ensure the updated_at value used is consistent across related records. Use the same timestamp from the transfer record when updating workflow_execution to prevent audit discrepancies.

### 6. Rollback Capability
All SQL templates have corresponding rollback scripts auto-generated by the tooling. Test rollback procedures before executing deploy scripts in production.

### 7. Interactive Cases
Some cases like `cashout_rpp210_pe220_pc201` require user choice between resume and reject. Always verify the actual status at Paynet before choosing resume (ACSP/ACTC) vs reject (not found).
