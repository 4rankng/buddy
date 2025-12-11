# Workflow Remediation Playbook

### Before You Run Anything
- Get prod access to `prd-payments-payment-core-rds-mysql`, `prd-payments-rpp-adapter-rds-mysql`, and `prd-payments-payment-engine-rds-mysql`.
- Collect every impacted `run_id` (and `req_biz_msg_id`/`partner_tx_id` when RPP is involved).
- Copy the matching deploy + rollback blocks into the ticket-specific SQL bundles and have a second engineer skim them.
- For each case: confirm current `state`/`attempt`, execute deploy SQL, capture row counts, re-query to verify, and keep the rollback block ready.

---

## Case: `pc_external_payment_flow_200_11` (Payment Core)
- **Check:** `external_payment_flow = stSubmitted`, `state = 200`, `attempt = 11` on the provided `run_id`s.
- **Deploy SQL → `PC_Deploy.sql`:**
```sql
-- pc_external_payment_flow_200_11
UPDATE workflow_execution
SET state = 202,
    attempt = 1,
    `data` = JSON_SET(
      `data`,
      '$.StreamResp', JSON_OBJECT(
        'TxID', '',
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ExternalID', '',
        'ErrorMessage', 'Reject from adapter'),
      '$.State', 202)
WHERE run_id IN (
  'ccc572052d6446a2b896fee381dcca3a'
)
AND state = 200
AND attempt = 11;
```
- **Rollback SQL → `PC_Rollback.sql`:**
```sql
UPDATE workflow_execution
SET state = 200,
    attempt = 11,
    `data` = JSON_SET(`data`, '$.State', 200)
WHERE run_id IN (
  'ccc572052d6446a2b896fee381dcca3a'
);
```

---

## Case: `pc_external_payment_flow_201_0_RPP_210` (RPP Adapter)
- **Collect IDs:** Ask requester for `req_biz_msg_id` (ex `20251209GXSPMYKL040OQR78229964`). Fetch `partner_tx_id`:
```sql prd-payments-rpp-adapter-rds-mysql
SELECT partner_tx_id FROM credit_transfer WHERE req_biz_msg_id = '20251209GXSPMYKL040OQR78229964';
```
- **Check:**
```sql prd-payments-rpp-adapter-rds-mysql
SELECT run_id, attempt, state FROM workflow_execution WHERE run_id = 'f4e858c9f47f4a469f09126f94f42ace';
```
Proceed only if `attempt = 0` and `state = 210`.
- **Deploy SQL → `RPP_Deploy.sql`:**
```sql
-- RPP 210, PE 220, PC 201. No response from RPP. Move to 222 to resume. ACSP
UPDATE workflow_execution
SET state = 222,
    attempt = 1,
    `data` = JSON_SET(`data`, '$.State', 222)
WHERE run_id IN (
  '2d8facb6f14846ac907325362ffa99cc'
)
AND state = 210;
```
- **Rollback SQL → `RPP_Rollback.sql`:**
```sql
UPDATE workflow_execution
SET state = 201,
    attempt = 0,
    `data` = JSON_SET(`data`, '$.State', 201)
WHERE run_id IN (
  '2d8facb6f14846ac907325362ffa99cc'
);
```

---

## Case: `pc_external_payment_flow_201_0_RPP_900` (RPP Adapter)
- **Collect IDs:** Same `req_biz_msg_id` workflow as above → confirm `partner_tx_id` and `run_id` targeting.
- **Check:** `state = 900`, `attempt = 0` on the target `run_id`:
```sql prd-payments-rpp-adapter-rds-mysql
SELECT run_id, attempt, state FROM workflow_execution WHERE run_id = 'f4e858c9f47f4a469f09126f94f42ace';
```
- **Deploy SQL → `RPP_Deploy.sql`:**
```sql
-- RPP 900, PE 220, PC 201. Republish from RPP to resume. ACSP
UPDATE workflow_execution
SET state = 301,
    attempt = 1,
    `data` = JSON_SET(`data`, '$.State', 301)
WHERE run_id IN (
  '5af00b2de3f3488088699ec8256b8ce6'
)
AND state = 900;
```
- **Rollback SQL → `RPP_Rollback.sql`:**
```sql
UPDATE workflow_execution
SET state = 900,
    attempt = 0,
    `data` = JSON_SET(`data`, '$.State', 900)
WHERE run_id IN (
  '5af00b2de3f3488088699ec8256b8ce6'
);
```

---

## Case: `pe_transfer_payment_210_0` (Payment Engine)
- **Check:** `workflow_id = 'workflow_transfer_payment'`, `state = 210`, `attempt = 0`.
- **Deploy SQL → `PE_Deploy.sql`** (note the ticket uses function name `FixPeTransferPayment210_0`).
```sql
-- Reject PE stuck 210. Reject transactions since it hasn't reached Paynet yet
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    `data` = JSON_SET(
      `data`,
      '$.StreamMessage', JSON_OBJECT(
        'Status', 'FAILED',
        'ErrorCode', 'ADAPTER_ERROR',
        'ErrorMessage', 'Manual Rejected'),
      '$.State', 221)
WHERE run_id IN (
  '641f4202-1931-4f49-ab8d-a2716ca80e19'
)
AND workflow_id = 'workflow_transfer_payment'
AND state = 210;
```
- **Rollback SQL → `PE_Rollback.sql`:**
```sql
UPDATE workflow_execution
SET state = 210,
    attempt = 0,
    `data` = JSON_SET(
      `data`,
      '$.StreamMessage', NULL,
      '$.State', 210)
WHERE run_id IN (
  '641f4202-1931-4f49-ab8d-a2716ca80e19'
)
AND workflow_id = 'workflow_transfer_payment';
```

---

### Evidence & Tickets
- Attach the `SELECT`/`UPDATE` screenshots, SQL files, and timestamped notes to the incident or change ticket right after execution.
