RPP-adapter rtp cash in stuck at 200 intent to UPDATED. replay confirm

in this scenario in rpp-adapter

credit_transfer has partner_tx_id='0f20cdcbc8dd44a7915e6803c7542778'
and end_to_end_id='20251229BIMBMYKL070ORB53488076'
created_at = '2025-12-29T01:23:57.299541Z'

then find the workflow by

select * from workflow_execution
WHERE created_at>='2025-12-29T01:20:57.299541Z'  // created_at - 3 min
and created_at <= '2025-12-29T01:26:57.299541Z' // created_at + 3 min
and data like '%20251229BIMBMYKL070ORB53488076%'; // end_to_end_id='20251229BIMBMYKL070ORB53488076'

give
state=200, attempt=0

and in partnerpay-engine

select * from intent where intent_id='0f20cdcbc8dd44a7915e6803c7542778'; // intent_id = partner_tx_id

give type='RTP_TOP_UP' and status='CONFIRMED'

then please raise DML

Partnerpay_Deploy.sql

UPDATE intent set status = 'UPDATED'
WHERE intent_id in (
'0f20cdcbc8dd44a7915e6803c7542778'
) and status = 'CONFIRMED';

Partnerpay_Rollback.sql

UPDATE intent set status = 'CONFIRMED'
WHERE intent_id in (
'0f20cdcbc8dd44a7915e6803c7542778'
) and status = 'CONFIRMED';

RPP_Deploy.sql
UPDATE workflow_execution set state = 110, attempt = 1,  `data` = JSON_SET(`data`, '$.State', 110)
WHERE run_id in (
'52eaa330045138178bf0b0e6e33dde87' // this run_id is taken from the rpp workflow info above
) and workflow_id = 'wf_ct_rtp_cashin' and state = 200;

RPP_Rollback.sql
UPDATE workflow_execution set state = 200, attempt = 0,  `data` = JSON_SET(`data`, '$.State', 200)
WHERE run_id in (
'52eaa330045138178bf0b0e6e33dde87'
) and workflow='wf_ct_rtp_cashin';