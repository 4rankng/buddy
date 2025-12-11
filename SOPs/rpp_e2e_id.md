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
-- rpp_cashout_reject_101_19, publish FAILED status
SET state = 311, attempt = 1 ,  data = JSON_SET(data, '$.State', 311)
WHERE run_id IN (
	'33997a1f8dae4793a2e1bc711aa066af'
) AND state = 101 and workflow_id = 'wf_ct_cashout';
`
RPP_Rollback.sql
SET state = 101, attempt = 0 ,  data = JSON_SET(data, '$.State', 101)
WHERE run_id IN (
	'33997a1f8dae4793a2e1bc711aa066af'
) and workflow_id = 'wf_ct_cashout';

we can keep adding into IN ()
if we have multiple cases


for stdout output (single txn) or write to file output (multiple txn)

### [1] transaction_id: f4e858c9f47f4a469f09126f94f42ace
[payment-engine]
status: PROCESSING
created_at: 2025-12-08T18:15:31.543552Z
workflow_transfer_payment: state=stSuccess(220) attempt=0 run_id=A404BFA6-90CE-4219-B4D4-85F84D805171
[payment-core]
internal_transaction: AUTH SUCCESS
external_transaction: TRANSFER PROCESSING
payment_core_workflow_external_payment_flow: state=stSuccess(201) attempt=0 run_id=4b5069e464c54dbcaa4a470423677c35
payment_core_workflow_internal_payment_flow: state=stSuccess(900) attempt=0 run_id=64bdfe8b7cae409d8074289b102bca1e
[rpp-adapter]
credit_transfer.status: PROCESSING
wf_ct_cashout: state=stSuccess(101) attempt=19 run_id=64bdfe8b7cae409d8074289b102bca1e

Here is your reference for workflow state mappings
// Workflow state mappings
var workflowStateMap = map[int]string{
	100: "stTransferPersisted",
	101: "stProcessingPublished",
	102: "stTransactionLimitChecked",
	103: "stRedeemRewardRequired",
	104: "stResolveFeeRequired",
	105: "stFreeTransferFeeRewardRedeemed",
	106: "stFeeResolved",
	210: "stAuthProcessing",
	211: "stAuthStreamPersisted",
	300: "stAuthCompleted",
	220: "stTransferProcessing",
	221: "stTransferStreamPersisted",
	223: "stTransferCompleted",
	230: "stCaptureProcessing",
	231: "stCaptureStreamPersisted",
	235: "stCapturePrepared",
	400: "stTransferFailed",
	410: "stAutoCancelProcessing",
	412: "stAutoCancelStreamPersisted",
	501: "stPrepareFailureHandling",
	505: "stFailurePublished",
	510: "stFailureNotified",
	511: "stRewardRedeemVoidRequired",
	512: "stRewardRedeemVoided",
	701: "stCaptureFailed",
	702: "stCancelFailed",
	703: "stRewardRedeemInterventionRequired",
	900: "stCaptureCompleted",
	905: "stCompletedPublished",
	910: "stCompletedNotified",
	911: "stRewardRedeemCompletionRequired",
	912: "stRewardRedeemCompleted",
}

// Internal payment flow state mappings
var internalPaymentFlowMap = map[int]string{
	100: "stPending",
	101: "stStreamPersisted",
	901: "stPrepareUpdateAuth",
	902: "stPrepareSuccessPublish",
	900: "stSuccess",
	501: "stPrepareFailurePublish",
	500: "stFailed",
}

// External payment flow state mappings
var externalPaymentFlowMap = map[int]string{
	200: "stSubmitted",
	201: "stProcessing",
	202: "stRespReceived",
	901: "stPrepareSuccessPublish",
	900: "stSuccess",
	501: "stPrepareFailurePublish",
	500: "stFailed",
}

rpp-adapter workflow_id=wf_ct_cashout
	stCreditTransferPersisted = we.State(101)
	stCreditorDetailUpdated   = we.State(111)
	stPrepareCreditorInquiry  = we.State(120)

	stCreditorInquiryFailed  = we.State(121)
	stCreditorInquirySuccess = we.State(122)

	stTransferProcessing              = we.State(210)
	stTransferResponseReceived        = we.State(211)
	stTransferMessageRejectedReceived = we.State(212)
	stTransferManualRejectedReceived  = we.State(221)
	stTransferManualResumeReceived    = we.State(222)

	stPrepareSuccessPublish = we.State(301)
	stPrepareFailurePublish = we.State(311)
	stSkipPublish           = we.State(321)

	stTransferRetryAttemptExceeded = we.State(502)

	stFailed  = we.State(700)
	stSuccess = we.State(900)

rpp-adapter workflow_id='wf_ct_cashin'
	stTransferPersisted         = we.State(100)
	stRequestToPayUpdated       = we.State(110)
	stRequestToPayUpdateFailed  = we.State(111)
	stOriginalTransferValidated = we.State(121)
	stFieldsValidationFailed    = we.State(122)

	stTransferProcessing      = we.State(200)
	stTransferStreamPersisted = we.State(201)
	stTransferUpdated         = we.State(210)
	stTransferRespPrepared    = we.State(220)

	stCashInFailed   = we.State(700)
	stCashInToRefund = we.State(701)

	stCashInCompleted           = we.State(900)
	stCashInCompletedWithRefund = we.State(901)














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
