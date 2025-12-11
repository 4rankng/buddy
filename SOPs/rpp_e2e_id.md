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

state=101 and attempt = 19 and workflow_id='wf_ct_cashout', 'wf_ct_qr_payment'
`
this case is called rpp_cashout_reject_101_19


RPP_DEPLOY.sql

UPDATE workflow_execution SET  attempt = 1
WHERE run_id IN (
	'33997a1f8dae4793a2e1bc711aa066af'
) AND state = 311 and workflow_id in (
'wf_ct_qr_payment',
	'wf_ct_cashout'
);

RPP_Rollback.sql
UPDATE workflow_execution SET attempt = 1
WHERE run_id IN (
	'33997a1f8dae4793a2e1bc711aa066af'
) and workflow_id in (
'wf_ct_qr_payment',
	'wf_ct_cashout'
);

we can keep adding into IN ()
if we have multiple cases


for stdout output (single txn) or file output (multiple txn)

### [1] transaction_id: f4e858c9f47f4a469f09126f94f42ace
[payment-engine]
status: PROCESSING
created_at: 2025-12-08T18:15:31.543552Z
workflow_transfer_payment: state=220 attempt=0 run_id=A404BFA6-90CE-4219-B4D4-85F84D805171
[payment-core]
internal_transaction: AUTH SUCCESS
external_transaction: TRANSFER PROCESSING
payment_core_workflow_external_payment_flow: state=201 attempt=0 run_id=4b5069e464c54dbcaa4a470423677c35
payment_core_workflow_internal_payment_flow: state=900 attempt=0 run_id=64bdfe8b7cae409d8074289b102bca1e
[rpp-adapter]
credit_transfer.status: PROCESSING
wf_ct_cashout: state=101 attempt=19 run_id=64bdfe8b7cae409d8074289b102bca1e