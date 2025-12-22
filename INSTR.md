Update sgbuddy ecotxn publish ... 



we can find payment-core related record by searching group_id and created_at timestamp

select * from internal_transaction where group_id = '{charge table transction id}
and created_at >= {charge table created at  -  1 hour}
and created_at <= {charge table created at  + 1hour}

use tx_id to search workflow excution


Here I provide your previous solution to fix 2 stuck transaction id
1ed87447b552420790357c2d5abe5509 and ef2282dcdf00458fa309d7a9442232d6
(feel free to check charge table for more information)


############ DEPLOY ############

PPE_Deploy.sql
-- Transaction 1: 1ed87447b552420790357c2d5abe5509
UPDATE charge
SET 
    valued_at = '2025-10-24T15:30:01Z', // is taken from payment-core internal_transaction table
    updated_at = '2025-12-10T06:25:06Z' // existing updated_at in charge table, set to prevent override
WHERE transaction_id = '1ed87447b552420790357c2d5abe5509';

UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT( // all info here is taken from charge table
    'ID', 2489358,
    'Amount', 1300,
    'Status', 'COMPLETED',
    'Remarks', '',
    'TxnType', 'SPEND_MONEY',
    'Currency', 'SGD',
    'Metadata', JSON_OBJECT('service', 'Transport', 'featureCode', 'A-8HDXTHVGW4VFAV'),
    'ValuedAt', '2025-10-24T15:30:01Z',
    'CreatedAt', '2025-10-24T14:58:11Z',
    'PartnerID', 'b2da6c1e-b2e4-4162-82b7-ce43ebf8b211',
    'TxnDomain', 'DEPOSITS',
    'UpdatedAt', '2025-12-10T06:25:06Z', // same value with charge table updated_at
    'CustomerID', '6f1c366d-8f21-4fbd-ae3e-d52edb32b754',
    'ExternalID', '80adc50d977a4519912781ef034987e8',
    'Properties', JSON_OBJECT('VerdictID', 37536250, 'AuthorisationID', '81cdfddd213b48a9be168fea7368c999', 'NotificationFlags', JSON_OBJECT('Sms', 0, 'Push', 0, 'Email', 0), 'CancelIdempotencyKey', '', 'CaptureIdempotencyKey', '80c97e874e044d69bf53f47925027cd1'),
    'TxnSubtype', 'GRAB',
    'ReferenceID', '80adc50d977a4519912781ef034987e8',
    'BillingToken', '3bbee8f00440469295a34c57418cafd7',
    'StatusReason', '',
    'CaptureMethod', 'MANUAL',
    'SourceAccount', JSON_OBJECT('Number', '8885548902', 'DisplayName', ''),
    'TransactionID', '1ed87447b552420790357c2d5abe5509',
    'CapturedAmount', 1300,
    'DestinationAccount', JSON_OBJECT('Number', '209421001', 'DisplayName', ''),
    'TransactionPayLoad', JSON_OBJECT('ActivityID', 'A-8HDXTHVGW4VFAV', 'ActivityType', 'TRANSPORT'),
    'StatusReasonDescription', ''
))
WHERE
    run_id = '1ed87447b552420790357c2d5abe5509';


-- Transaction 2: ef2282dcdf00458fa309d7a9442232d6
UPDATE charge
SET 
    valued_at = '2025-10-24T06:00:02Z', // is taken from payment-core internal_transaction table
    updated_at = '2025-12-10T06:25:06Z'
WHERE transaction_id = 'ef2282dcdf00458fa309d7a9442232d6';

UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT( // all info here is taken from charge table
    'ID', 2487187,
    'Amount', 2310,
    'Status', 'COMPLETED',
    'Remarks', '',
    'TxnType', 'SPEND_MONEY',
    'Currency', 'SGD',
    'Metadata', JSON_OBJECT('service', 'GrabFood', 'featureCode', 'A-8HBDEUNGX53QAV'),
    'ValuedAt', '2025-10-24T06:00:02Z',
    'CreatedAt', '2025-10-24T05:25:19Z',
    'PartnerID', 'b2da6c1e-b2e4-4162-82b7-ce43ebf8b211',
    'TxnDomain', 'DEPOSITS',
    'UpdatedAt', '2025-12-10T06:25:06Z',
    'CustomerID', '37a5829d-8fe3-4e64-9558-4e8ec305955b',
    'ExternalID', '450102ad681a4d9ead7887014765719f',
    'Properties', JSON_OBJECT('VerdictID', 37515899, 'AuthorisationID', 'cc95f3a22a1f4604b3f7b3d780fe25da', 'NotificationFlags', JSON_OBJECT('Sms', 0, 'Push', 0, 'Email', 0), 'CancelIdempotencyKey', '', 'CaptureIdempotencyKey', '379b278f3fc84bf3896fe4b7b276f1c8'),
    'TxnSubtype', 'GRAB',
    'ReferenceID', '450102ad681a4d9ead7887014765719f',
    'BillingToken', 'f4a56eb61fda4dab89d7abdc5a436ecd',
    'StatusReason', '',
    'CaptureMethod', 'MANUAL',
    'SourceAccount', JSON_OBJECT('Number', '8884066567', 'DisplayName', ''),
    'TransactionID', 'ef2282dcdf00458fa309d7a9442232d6',
    'CapturedAmount', 2310,
    'DestinationAccount', JSON_OBJECT('Number', '209421001', 'DisplayName', ''),
    'TransactionPayLoad', JSON_OBJECT('ActivityID', 'A-8HBDEUNGX53QAV', 'ActivityType', 'FOOD'),
    'StatusReasonDescription', ''
))
WHERE
    run_id = 'ef2282dcdf00458fa309d7a9442232d6';

PC_Deploy.sql

UPDATE workflow_execution SET attempt=1, state=902
WHERE workflow_id='internal_payment_flow' AND run_id in ('d72d76ce474040c7970ab7438c9a234d', 'ccb3bb0f5c8845d1a0ac2bde67bbc4e9' ) AND state=900;

########### ROLLBACK #########

PPE_Rollback.sql
-- Transaction 1: 1ed87447b552420790357c2d5abe5509
UPDATE charge
SET 
    valued_at = '0000-00-00T00:00:00Z', // original value from charge table
    updated_at = '2025-12-10T06:25:06Z' // existing updated_at in charge table, set to prevent override
WHERE transaction_id = '1ed87447b552420790357c2d5abe5509';

UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT())
WHERE
    run_id = '1ed87447b552420790357c2d5abe5509';


-- Transaction 2: ef2282dcdf00458fa309d7a9442232d6
UPDATE charge
SET 
    valued_at = '0000-00-00T00:00:00Z',
    updated_at = '2025-12-10T06:25:06Z'
WHERE transaction_id = 'ef2282dcdf00458fa309d7a9442232d6';

UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT())
WHERE
    run_id = 'ef2282dcdf00458fa309d7a9442232d6';


PC_Rollback.sql

UPDATE workflow_execution SET attempt=0, state=900
WHERE workflow_id='internal_payment_flow' AND run_id in ('d72d76ce474040c7970ab7438c9a234d', 'ccb3bb0f5c8845d1a0ac2bde67bbc4e9' ); // group multiple txn together



