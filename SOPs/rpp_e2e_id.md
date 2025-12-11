for command

mybuddy txn 20251209GXSPMYKL010ORB79174342

you can has a helper to classify
if the args is
- transaction id: eg ccc572052d6446a2b896fee381dcca3a
- file path: TS-4466.txt
- rpp e2d id: eg 20251209GXSPMYKL010ORB79174342

NOTE that e2d id is a fix format like
YYYYMMDDGXSPMY eg 20251209GXSPMY
and the overall length is fixed, I can provide list of example
20251209GXSPMYKL010ORB79174342
20251209GXSPMYKL030OQR15900197
20251209GXSPMYKL040OQR10829949
20251209GXSPMYKL040OQR41308688
20251209GXSPMYKL040OQR78229964

file path format is {name}.{ext} and you can test if the file path exist

and the rest is transaction id


if user call
mybuddy txn 20251209GXSPMYKL010ORB79174342

you query prd-payments-rpp-adapter-rds-mysql


select * from workflow_execution where run_id=(select partner_tx_id from credit_transfer where end_to_end_id='20251209GXSPMYKL010ORB79174342');

if

state=101 and attempt = 19 and workflow_id='wf_ct_cashout'
`
this case is called rpp_cashout_reject_101_19


RPP_DEPLOY.sql
-- rpp_cashout_reject_101_19, manual reject

UPDATE workflow_execution
SET state = 221, attempt = 1 ,  data = JSON_SET(data, '$.State', 221)
where run_id in
('2823f1ae2cc44331b49827bdffc44a16') and state = 101 and workflow_id = 'wf_ct_cashout';
`
RPP_Rollback.sql
UPDATE workflow_execution
SET state = 101, attempt = 0 ,  data = JSON_SET(data, '$.State', 101)
WHERE run_id IN (
	'33997a1f8dae4793a2e1bc711aa066af'
) and workflow_id = 'wf_ct_cashout';

we can keep adding into IN ()
if we have multiple cases


for stdout output (single txn) or write to file output (multiple txn)









if

state=210 and attempt = 0 and workflow_id='wf_ct_qr_payment'
`
this case is called rpp_qr_payment_reject_210_0


RPP_DEPLOY.sql

-- rpp_qr_payment_reject_210_0, manual reject
UPDATE workflow_execution
SET state = 221, attempt = 1 ,  data = JSON_SET(data, '$.State', 221)
where run_id in
('2823f1ae2cc44331b49827bdffc44a16') and state = 210 and workflow_id = 'wf_ct_qr_payment';

RPP_Rollback.sql
SET state = 210, attempt = 0 ,  data = JSON_SET(data, '$.State', 210)
where run_id in
('2823f1ae2cc44331b49827bdffc44a16') and state = 210 and workflow_id = 'wf_ct_qr_payment';

we can keep adding into IN ()
if we have multiple cases




if
mybuddy rpp resume XXXX
where XXX is single e2d id or file path eg TS-4468.txt

you query prd-payments-rpp-adapter-rds-mysql


select * from workflow_execution where run_id=(select partner_tx_id from credit_transfer where end_to_end_id='20251209GXSPMYKL010ORB79174342');

if state=210 and attempt = 0 and workflow_id in ('wf_ct_cashout', 'wf_ct_qr_payment')

RPP_Deploy.sql
```sql
-- rpp_no_response_resume_acsp
-- RPP did not respond in time, but status at Paynet is ACSP (Accepted Settlement in Process) or ACTC (Accepted Technical Validation)
UPDATE workflow_execution SET state = 222, attempt = 1,
  `data` = JSON_SET(`data`, '$.State', 222)
WHERE run_id IN (
  '663c03ec156e4046b283d58604a68f4f'
) AND state = 210 and workflow_id in ('wf_ct_cashout', 'wf_ct_qr_payment');
```

RPP_Rollback.sql
UPDATE workflow_execution SET state = 210, attempt = 0,
  `data` = JSON_SET(`data`, '$.State', 210)
WHERE run_id IN (
  '663c03ec156e4046b283d58604a68f4f'
);
