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













for ecotxn in partnerpay-engine
when we call sgbuddy ecotxn publish 7eba1b67c9174d21bb66bb089ebd6fd3

we can generate this 

PPE_Deploy.sql
UPDATE charge
SET 
status = 'COMPLETED', 
valued_at = '2025-10-24T15:30:01Z', // this value from payment-core
updated_at = '2025-12-16T07:06:07Z' // this is charge.updated_at, we need to set to preserve, otherwise it will be autodate to current timestamp
WHERE transaction_id = '7eba1b67c9174d21bb66bb089ebd6fd3';


-- Update workflow_execution for transaction_id: 7eba1b67c9174d21bb66bb089ebd6fd3
UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT( // these value from charge table
    'ID', 786874,
    'Amount', 200,
    'Status', 'COMPLETED',
    'Remarks', '',
    'TxnType', 'SPEND_MONEY',
    'Currency', 'SGD',
    'Metadata', JSON_OBJECT('featureCode', 'A-8NCF3OMGWQPOD9', 'service', 'Goblin'),
    'ValuedAt', '2025-10-24T15:30:01Z', // this value from payment-core
    'CreatedAt', '2025-12-10T02:02:51Z',
    'PartnerID', '28a61200-7d6d-4419-947a-173ff81cf7db',
    'TxnDomain', 'DEPOSITS',
    'UpdatedAt', '2025-12-16T07:06:07Z',
    'CustomerID', '0948a27e-bd0d-4aff-8071-0bf3fc9469fd',
    'ExternalID', 'be618a3ad75648a79ae8be1ee4fa0d43',
    'Properties', JSON_OBJECT('AuthorisationID', 'abc654828dc047bca0898989a574a41a', 'CancelIdempotencyKey', '', 'CaptureIdempotencyKey', '', 'NotificationFlags', JSON_OBJECT('Email', 1, 'Push', 1, 'Sms', 0), 'VerdictID', 4032259),
    'TxnSubtype', 'GRAB',
    'ReferenceID', 'be618a3ad75648a79ae8be1ee4fa0d43',
    'BillingToken', '1e0b4d582c204a55af42f6ceb84a0d73',
    'StatusReason', '',
    'CaptureMethod', 'AUTOMATIC',
    'SourceAccount', JSON_OBJECT('DisplayName', '', 'Number', '8880261519'),
    'TransactionID', '7eba1b67c9174d21bb66bb089ebd6fd3',
    'CapturedAmount', 200,
    'DestinationAccount', JSON_OBJECT('DisplayName', '', 'Number', '209421001'),
    'TransactionPayLoad', JSON_OBJECT('ActivityID', 'A-8NCF3OMGWQPOD9', 'ActivityType', 'DEFAULT'),
    'StatusReasonDescription', ''
))
WHERE
    run_id = '7eba1b67c9174d21bb66bb089ebd6fd3';

PPE_Rollback.sql

UPDATE charge
SET 
status = 'COMPLETED', 
valued_at = '000-00-00T00:00:00Z', // use original charge.valued_at
updated_at = '2025-12-16T07:06:07Z' // this is charge.updated_at, we need to set to preserve, otherwise it will be autodate to current timestamp
WHERE transaction_id = '7eba1b67c9174d21bb66bb089ebd6fd3';


-- Update workflow_execution for transaction_id: 7eba1b67c9174d21bb66bb089ebd6fd3
UPDATE workflow_execution
SET
    state = 888, // original workflow_execution.state value
    attempt = 0, // original workflow_execution.attempt value
    data = JSON_SET(data,
            '$.State', 888, // original workflow_execution.state value
            '$.ChargeStorage', JSON_OBJECT( // original value from workflow_execution.data.ChargeStorage
    ))
WHERE
    run_id = '7eba1b67c9174d21bb66bb089ebd6fd3';