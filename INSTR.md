we should have case type classification for ecotxn 

case 'ecotxn_ChargeFailed_CaptureFailed_TMError

when
[partnerpay-engine]
workflow_charge state=502 attempt=0
charge error code = 'SYSTEM_ERROR' error msg 'error occurred in Thought Machine.'


[payment-core]
      workflow_id=internal_payment_flow
      state=stFailed(500) attempt=0

then we need to generate

PPE_Deploy.sql

update charge set
status = 'PROCESSING',
updated_at = {the original updated_at from charge record}
where transaction_id = {transaction_id}

update workflow_execution
set state=300, data=JSON_SET(data, '$.State', 300,
'$.ChargeStorage.Status', 'PROCESSING'
)
WHERE run_id = {transaction_id} AND workflow_id='workflow_charge' AND state=502 AND attempt=0;

PPE_Rollback.sql

update charge set
status = 'FAILED',
updated_at = {the original updated_at from charge record}
where transaction_id = {transaction_id}

update workflow_execution
set state=502, data=JSON_SET(data, '$.State', 502,
'$.ChargeStorage.Status', 'FAILED'
)
WHERE run_id = {transaction_id} AND workflow_id='workflow_charge';