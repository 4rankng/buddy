update /Users/frank.nguyen/Documents/buddy/docs/sops/MY_DML_SOP.md

for 
### `cash_in_stuck_100_update_mismatch`
- **Condition**: Cash in workflow stuck at state 100 with attempts. Update operation failing due to updatedAt mismatch.


you must check the updated_at from credit_transfer table
against the data->>$.CreditTransfer.UpdatedAt 
and classified as cash_in_stuck_100_update_mismatch only if the value mismatch
note that 
data->>$.CreditTransfer.UpdatedAt is in GMT+8
and credit_transfer.updated_at is in UTC

if there is no mismatch, we call the case as cash_in_stuck_100_retry
we can just set attempt=1 to restart the validation as shown here
-- rpp_cashin_stuck_100_0, retry
UPDATE workflow_execution
SET attempt = 1
WHERE run_id IN ('5c34e6ab0fea334f88b9b4cdb781902f')
AND workflow_id = 'wf_ct_cashin'
AND state = 100;
DML: https://doorman.infra.prd.g-bank.app/rds/dml/43211

if there is mismatch, we call it as cash_in_stuck_100_update_mismatch
UPDATE workflow_execution 
SET 
    attempt = 1,
    `data` = JSON_SET(`data`,
            '$.CreditTransfer.UpdatedAt','2025-11-24T22:27:21.103964+08:00' // convert from credit_transfer.updated_at
    )
WHERE 
	run_id = 'dd85dfadba453d969a635c49e3d87799' 
	AND workflow_id = 'wf_ct_cashin' 
	AND state = 100;