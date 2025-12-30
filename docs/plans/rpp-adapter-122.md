if this scenario

[rpp-adapter]
req_biz_msg_id: 20251228RPPEMYKL010HRM30218865
partner_tx_id: 6ec50daa4a373a2f9fae4a6aec670679
wf_ct_cashin:
   state=stFieldsValidationFailed(122) attempt=0
   run_id='6ec50daa4a373a2f9fae4a6aec670679'

then raise DML

RPP_Deploy.sql

UPDATE workflow_execution
SET state = 100, attempt = 1, `data` = JSON_SET(`data`, '$.State', 100)
WHERE run_id IN(
	'6ec50daa4a373a2f9fae4a6aec670679' // this value you should get from above wf_ct_cashin.run_id
) and workflow_id = 'wf_ct_cashin' and state = 122;

RPP_Rollback.sql
UPDATE workflow_execution
SET state = 122, attempt = 0, `data` = JSON_SET(`data`, '$.State', 122)
WHERE run_id IN(
	'6ec50daa4a373a2f9fae4a6aec670679' // this value you should get from above wf_ct_cashin.run_id
) and workflow_id = 'wf_ct_cashin';

also ensure that we always find related workflow in workflow_execution table for the rpp adapter record



fix the incorrect behavior below

frank.nguyen@DBSG-H4M0DVF2C7 buddy % mybuddy txn 20251228TNGDMYNB010ORM77048250
Initialize Doorman client for [my]...
Initialize Jira client for [my]...
[MY] Querying transaction: 20251228TNGDMYNB010ORM77048250
### [1] e2e_id: 20251228TNGDMYNB010ORM77048250
[rpp-adapter]
e2e_id: 20251228TNGDMYNB010ORM77048250
partner_tx_id: 6ec50daa4a373a2f9fae4a6aec670679
wf_ct_cashin:
   state=stFieldsValidationFailed(122) attempt=0
   run_id=6ec50daa4a373a2f9fae4a6aec670679

[Classification]
NOT_FOUND


here is clearly the classified scenario
wf_ct_cashin:
   state=stFieldsValidationFailed(122) attempt=0

we should have value for the classification and deploy and rollback sql