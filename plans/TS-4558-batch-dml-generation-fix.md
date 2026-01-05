# TS-4558: Batch Processing Missing DML Generation (and incomplete PE deploy/rollback)

## Context

When running batch mode with `mybuddy txn <file>`, the tool correctly:
- reads transaction IDs from the input file, and
- queries transaction statuses, then
- writes an analysis file `<input>_results.txt`.

However, the tool does **not** reliably generate the corresponding SQL DML scripts (`*.sql`) needed to apply fixes for the detected issues. Separately, for the SOP logic `pe_stuck_at_limit_check_102_4`, SQL generation is currently incomplete: it updates the `transfer` table but does not update `workflow_execution`, and it also fails to produce the required rollback file.

---

## Reproduction

### Command

```bash
make deploy && mybuddy txn TS-4558.txt
```

### Console output (key lines)

```text
[MY] Processing batch file: TS-4558.txt
[MY] Found 4 transaction IDs to process
...
[MY] Writing batch results to: TS-4558.txt_results.txt
[MY] Batch processing completed. Results written to TS-4558.txt_results.txt

[MY] SQL Generation Summary:
[MY]   Generated 4 SQL statements:
[MY]     PE Deploy: 4 statements

SQL statements written to PE_Deploy.sql
[MY] SQL DML files generated: [PE_Deploy.sql]
```

### Observed outputs

- ✅ `TS-4558.txt_results.txt` is created successfully.
- ⚠️ Only **one** SQL file is created: `PE_Deploy.sql`
- ❌ `PE_Rollback.sql` is **not** created.
- ❌ For `pe_stuck_at_limit_check_102_4`, the deploy script currently **only updates `transfer`**, but it **must update `workflow_execution` too**.

---

## SOP logic: `pe_stuck_at_limit_check_102_4`

### Meaning

This pattern indicates:
- Payment Engine (PE) workflow is stuck at **state 102** ("Limit Checked"),
- Payment Core has successfully authorized the transaction,
- so we must manually link the authorization ID and then reject/reset the stuck workflow.

### Required fix (two-step deploy)

1) **Update workflow execution**: force PE workflow from 102 → **221 (Manual Reject)** and include the stream message + state update.

2) **Update transfer**: inject the **AuthorisationID** into `transfer.properties` using `JSON_SET`, and **preserve the original `updated_at`** timestamp (explicitly set it).

---

## Expected SQL output for a single example

### Deploy (PE_Deploy.sql)

The deploy script must contain **both** statements (workflow + transfer):

```sql
-- 1) Reject/Reset the Workflow Execution (cashout_pe102_reject)
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
        '$.Properties.AuthorisationID', 'ef8a3114ccab4c309cd7855270b5f221'
    )
WHERE run_id IN ('D060C5AD-C53F-4CEC-AC60-E3B04AB9DE46')
  AND state = 102
  AND workflow_id = 'workflow_transfer_payment';

-- 2) Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', 'ef8a3114ccab4c309cd7855270b5f221'),
    updated_at = '2025-12-26T13:32:25.308547Z'  -- PRESERVE ORIGINAL TIMESTAMP
WHERE transaction_id = 'TX123456';
```

### Rollback (PE_Rollback.sql)

Rollback must revert `workflow_execution` back to the original state (example shown):

```sql
-- cashout_pe102_reject_rollback
UPDATE workflow_execution
SET state = 102,
    attempt = 4,
    `data` = JSON_SET(
        `data`,
        '$.StreamMessage', JSON_OBJECT(),
        '$.State', 102,
        '$.Properties.AuthorisationID', NULL
    )
WHERE run_id IN ('D060C5AD-C53F-4CEC-AC60-E3B04AB9DE46');
```

---

## Problem 1: Why is `PE_Rollback.sql` not created?

### Current observation

Even though the deploy output is written, the corresponding rollback file is missing:
- `PE_Deploy.sql` exists
- `PE_Rollback.sql` does **not**

### Required behavior

Every SOP fix must produce a **pair**:
- `PE_Deploy.sql`
- `PE_Rollback.sql`

If a deploy script is emitted for an SOP, the rollback must be emitted too (unless explicitly impossible, which this case is not).

---

## Problem 2: Deploy script is incomplete for `pe_stuck_at_limit_check_102_4`

### Current behavior

Deploy script only updates:
- `transfer`

### Required behavior

Deploy script must update:
- `workflow_execution` (102 → 221) **and**
- `transfer` (inject AuthorisationID + preserve updated_at)

---

## Suspected code location(s)

- SOP template / generator:
  - `/Users/frank.nguyen/Documents/buddy/internal/txn/adapters/sql_templates_pe_basic.go`

If rollback generation is missing, the likely issue is:
- rollback template is not registered / not returned for this SOP, or
- the writer only writes deploy files in batch mode (or only for PE deploy), or
- rollback statements are generated but discarded due to filtering / empty-check / file naming.

---

## Implementation plan (high level)

### A) Ensure batch processing writes SQL files (deploy + rollback)

In the batch processing flow, after writing `<input>_results.txt`, ensure it also:
1. generates SQL statements from results (deploy and rollback),
2. clears any previous SQL output for a clean run,
3. writes all SQL files to disk.

Conceptual flow:

```go
// After WriteBatchResults(results, outputFilename)

// 1) Generate the DML objects (deploy + rollback)
sqlStatements := adapters.GenerateSQLStatements(results)

// 2) Clear previous SQL files to avoid appending to old runs
adapters.ClearSQLFiles()

// 3) Write the new SQL files (PE deploy + PE rollback + any others)
filesCreated, err := adapters.WriteSQLFiles(sqlStatements, ".")
if err != nil {
    fmt.Printf("Error writing SQL files: %v\n", err)
} else {
    if len(filesCreated) > 0 {
        fmt.Printf("SQL DML files generated: %v\n", filesCreated)
    } else {
        fmt.Println("No SQL fixes required for these transactions.")
    }
}
```

### B) Fix SOP adapter: `pe_stuck_at_limit_check_102_4`

The adapter must:
- verify `PaymentCore.InternalAuth.Status == "SUCCESS"`,
- extract `PaymentCore.InternalAuth.TxID` (AuthorisationID),
- generate **two deploy statements** (workflow + transfer),
- generate **rollback** for workflow update (at minimum),
- ensure `transfer.updated_at` is explicitly set to the current DB value to preserve timestamp.

---

## Verification checklist

1. Rebuild and deploy:
   ```bash
   make deploy
   ```

2. Run batch:
   ```bash
   mybuddy txn TS-4558.txt
   ```

3. Confirm console includes:
   - `SQL DML files generated: [...]`

4. Confirm files exist:
   ```bash
   ls -la *.sql
   ```

5. Confirm **both** exist:
   - `PE_Deploy.sql`
   - `PE_Rollback.sql`

6. Confirm content:
   ```bash
   cat PE_Deploy.sql
   cat PE_Rollback.sql
   ```

7. Confirm `PE_Deploy.sql` includes:
   - `UPDATE workflow_execution ... state = 221 ...`
   - `UPDATE transfer ... AuthorisationID ... updated_at = '<original>'`
