case pe_220_0_fast_cashin_failed
when 
[payment-engine]
workflow_transfer_collection: state=stAuthProcessing(220) attempt=0 
[fast-adapter]
status: FAILED (4)



sg-prd-m-payment-engine
PE_Deploy.sql

update workflow_execution
set attempt=1, state=221, data=json_set(data, 
'$.State', 221, 
'$.StreamMessage.Status', 'FAILED',
'$.StreamMessage.ErrorMessage', 'MANUAL REJECT')
where run_id = 'bc5f22c3fe474399adab3b9b0e6315a5'


PE_Rollback.sql
update workflow_execution
set attempt=0, state=220, data=json_set(data, 
'$.State', 220, 
'$.StreamMessage', JSON_OBJECT()
)
where run_id = 'bc5f22c3fe474399adab3b9b0e6315a5';


Test in stg:
Deploy
https://central-nonprod-doorman.sgbank.pr/rds/dml/10281
Rollback
https://central-nonprod-doorman.sgbank.pr/rds/dml/10282
