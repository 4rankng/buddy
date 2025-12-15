build command

sgbuddy ecotxn publish abc // single transaction_id
or
sgbuddy ecotxn publish TSE-833.txt // file path contains multile transactions

PPE_Deploy.sql

-- Update charge table for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE charge
SET status = 'COMPLETED', valued_at = '2025-10-24T15:30:01.311411Z'
WHERE transaction_id = '1ed87447b552420790357c2d5abe5509';

-- Update workflow_execution for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT(
	'ID', 2489358,
	'Amount', 1300,  // amount in integer
	'Status', 'COMPLETED',
	'Remarks', '',
	'TxnType', 'SPEND_MONEY',
	'Currency', 'SGD',
	'Metadata', JSON_OBJECT('featureCode', 'A-8HDXTHVGW4VFAV', 'service', 'Transport'),
	'ValuedAt', '2025-10-24T15:30:01.311411Z',
	'CreatedAt', '2025-10-24T14:58:11Z',
	'PartnerID', 'b2da6c1e-b2e4-4162-82b7-ce43ebf8b211',
	'TxnDomain', 'DEPOSITS',
	'UpdatedAt', '2025-12-10 15:40:24',
	'CustomerID', '6f1c366d-8f21-4fbd-ae3e-d52edb32b754',
	'ExternalID', '80adc50d977a4519912781ef034987e8',
	'Properties', JSON_OBJECT('AuthorisationID', '81cdfddd213b48a9be168fea7368c999', 'CancelIdempotencyKey', '', 'CaptureIdempotencyKey', '80c97e874e044d69bf53f47925027cd1', 'NotificationFlags', JSON_OBJECT('Email', 0, 'Push', 0, 'Sms', 0), 'VerdictID', 37536250),
	'TxnSubtype', 'GRAB',
	'ReferenceID', '80adc50d977a4519912781ef034987e8',
	'BillingToken', '3bbee8f00440469295a34c57418cafd7',
	'StatusReason', '',
	'CaptureMethod', 'MANUAL',
	'SourceAccount', JSON_OBJECT('DisplayName', '', 'Number', '8885548902'),
	'TransactionID', '1ed87447b552420790357c2d5abe5509',
	'CapturedAmount', 1300, // in integer
	'DestinationAccount', JSON_OBJECT('DisplayName', '', 'Number', '209421001'),
	'TransactionPayLoad', JSON_OBJECT('ActivityID', 'A-8HDXTHVGW4VFAV', 'ActivityType', 'TRANSPORT'),
	'StatusReasonDescription', ''
))
WHERE
    run_id = '1ed87447b552420790357c2d5abe5509'
	and state = {original state}
	;

PPE_Rollback.sql


-- Rollback charge table for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE charge
SET status = 'COMPLETED', valued_at = '0000-00-00T00:00:00.00Z'
WHERE transaction_id = '1ed87447b552420790357c2d5abe5509';

-- Rollback workflow_execution for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE workflow_execution
SET
	state = {original state},
	attempt = 0,
    data = JSON_SET(data,
            '$.ChargeStorage', JSON_OBJECT(/* original value */),
			'$.State', {original state})
WHERE
    run_id = '1ed87447b552420790357c2d5abe5509';

