pe220_pc201_rpp0_stuck_init
 the workflow is stuck at 220, PC external is 201, rpp adapter workflow is 0, can follow below,
https://doorman.infra.prd.g-bank.app/rds/dml/37220 pc reject prd-payments-payment-core-rds-mysql
UPDATE workflow_execution SET state = 202, attempt = 1,
                              `data` = JSON_SET(`data`, '$.StreamResp', JSON_OBJECT('TxID', '', 'Status', 'FAILED', 'ErrorCode', 'ADAPTER_ERROR', 'ExternalID', '', 'ErrorMessage', 'Reject from adapter'), '$.State', 202)
WHERE run_id in (
'9b52b3910abf4f288bb3b8e31d236378') and state = 201 and attempt = 0;


https://doorman.infra.prd.g-bank.app/rds/dml/37221 move wf to 700 prd-payments-rpp-adapter-rds-mysql
UPDATE workflow_execution 
SET state = 700,
         `data` = JSON_SET(`data`,'$.State', 700)
WHERE run_id in (
	'8c9f0b1f364a4d89a8150e8747406283'
)  and state = 0;

you can query payment-engine transfer table transaction_id=198fe80766cb48b4aca3cf8a38f5baa5
or cal mybuddy txn 198fe80766cb48b4aca3cf8a38f5baa5 to get list of related info