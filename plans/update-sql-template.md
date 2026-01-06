Based on the discussion in the DML Thread Jan 2026, there are several critical updates and refinements that should be applied to your MY_DML_SOP.md to ensure data consistency and reflect current best practices used by the team.

Here are the recommended changes:

1. Enhanced pe_stuck_at_limit_check_102_4 Protocol
In the thread, Khor Jeng Yong pointed out that for state 102 (limit check) issues, simply rejecting the workflow is insufficient. You must ensure the AuthorisationID is correctly mapped to maintain the link between Payment Engine and Payment Core.

Update: Add a specific instruction to retrieve the AuthorisationID from the PC internal_auth table.

Update: Explicitly include the requirement to update the transfer table alongside the workflow_execution table to prevent data mismatch.

2. New SOP: pe220_pc201_rpp0_stuck_init
The discussion highlighted a specific scenario where the RPP adapter is stuck at State 0 (stInit) while PC is at 201. This requires a multi-step cleanup across PC and RPP to "fail gracefully."

New Protocol:

PC Side: Manually reject the external payment flow (State 202).

RPP Side: Move the stuck stInit workflow to terminal State 700.

PE Side: Manual rejection (State 221).

3. Consolidation of RPP/PC Resumptions
Khor Jeng Yong suggested that instead of running separate DMLs for multiple transactions under the rpp210_pe220_pc201_accept case, they should be combined into a single update statement using WHERE run_id IN (...) to reduce manual overhead and risk.

4. Registry Workflow Awareness
The thread mentions that some stuck transactions are accompanied by inbound wf_process_registry issues.

Update: Add a note to the rpp_process_registry_stuck_init section to check for related inbound registry workflows whenever a cashout or QR payment is stuck, as resuming the registry may be the cleaner fix.

5. Refined Safety Guidelines: Timestamp Consistency
A new safety check should be added regarding the updated_at field.

Safety Check: When updating the transfer table or workflow_execution data, ensure the updated_at value used is consistent across related records. Discrepancies between the workflow's timestamp and the table's timestamp can cause audit or logic issues.

Proposed Document Additions
I have drafted the specific SQL and documentation updates for these findings:

Update to pe_stuck_at_limit_check_102_4
Note: You must retrieve the AuthorisationID from the PC internal_auth table for the corresponding transaction.

SQL

-- Update transfer table to maintain link with PC
UPDATE transfer 
SET properties = JSON_SET(properties, '$.AuthorisationID', '{AUTHORISATION_ID}'),
    updated_at = NOW() -- Ensure consistency with workflow update
WHERE transaction_id = '{TRANSACTION_ID}';
New Section: pe220_pc201_rpp0_stuck_init
Condition: PE 220, PC 201, RPP adapter stuck at 0 (stInit). Diagnosis: Adapter failed to initialize; transaction never reached PayNet. Resolution: Reject PC and move RPP to terminal state 700.

SQL

-- RPP Side: Move to terminal state
UPDATE workflow_execution 
SET state = 700, data = JSON_SET(data, '$.State', 700)
WHERE run_id = '{RPP_RUN_ID}' AND state = 0;