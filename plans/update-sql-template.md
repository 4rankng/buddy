. Transaction [2]: e2e_id 20251231MBBEMYKL010ORM15134561
Diagnosis: This is a Cashin workflow stuck at state 100 (stTransferPersisted) with attempt=0.

Protocol: cash_in_stuck_100_update_mismatch.

Root Cause: Often caused by a concurrent update or optimistic lock failure where the updatedAt value mismatches.

Resolution: Update the updated_at timestamp and reset the attempt to 1 to allow the workflow to resume.

Execution Script (TargetDB: RPP)
SQL

UPDATE workflow_execution
SET state = 100,
    attempt = 1,
    updated_at = NOW(),
    data = JSON_SET(data, '$.State', 100)
WHERE run_id = '5c34e6ab0fea334f88b9b4cdb781902f'
AND workflow_id = 'wf_ct_cashin'
AND state = 100;