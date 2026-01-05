from /Users/frank.nguyen/Documents/buddy/TS-4558.txt_results.txt, there is only 2 classificatgion it mean we generate for 2 txns only
but /Users/frank.nguyen/Documents/buddy/PE_Deploy.sql contain more than 2 txn in it



[PE_Deploy.sql]
-- Reject/Reset the Workflow Execution (cashout_pe102_reject)
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    `data` = JSON_SET(
        `data`,
        '$.StreamMessage',
        JSON_OBJECT(
            'Status', 'FAILED',
            'ErrorCode', "ADAPTER_ERROR",
            'ErrorMessage', 'Manual Rejected'
        ),
        '$.State', 221,
        '$.Properties.AuthorisationID', 'e190d5d404274615a97b228dc0cdbdf5'
    )
WHERE run_id IN ('1b0d8214-8374-469a-a462-6ec5eb750658')
  AND state = 102
  AND workflow_id = 'workflow_transfer_payment';

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', 'e190d5d404274615a97b228dc0cdbdf5'),
    updated_at = '2025-12-15T03:48:44.683021Z'
WHERE transaction_id = '8d69bd2672a041c78d2c18784f83d8eb';

-- Reject/Reset the Workflow Execution (cashout_pe102_reject)
UPDATE workflow_execution
SET state = 221,
    attempt = 1,
    `data` = JSON_SET(
        `data`,
        '$.StreamMessage',
        JSON_OBJECT(
            'Status', 'FAILED',
            'ErrorCode', "ADAPTER_ERROR",
            'ErrorMessage', 'Manual Rejected'
        ),
        '$.State', 221,
        '$.Properties.AuthorisationID', 'ef8a3114ccab4c309cd7855270b5f221'
    )
WHERE run_id IN ('D060C5AD-C53F-4CEC-AC60-E3B04AB9DE46')
  AND state = 102
  AND workflow_id = 'workflow_transfer_payment';

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', 'ef8a3114ccab4c309cd7855270b5f221'),
    updated_at = '2025-12-26T13:32:25.422421Z'
WHERE transaction_id = '90a8976b531446be8e00d42f02ff2d0d';

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', '57cd93f7d6534e6ca2b19183c8d03538'),
    updated_at = '2026-01-05T10:43:02.56458Z'
WHERE transaction_id = '253c9e27c69f465bbeed564eb16a4f0e';

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', '757f9263d2394b3eb03bb66687092957'),
    updated_at = '2025-12-31T02:19:46.13746Z'
WHERE transaction_id = '198fe80766cb48b4aca3cf8a38f5baa5';


[TS-4558.txt_results.txt]
### [1] transaction_id: 253c9e27c69f465bbeed564eb16a4f0e
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: FAILED
created_at: 2025-12-29T23:38:56.474381Z
reference_id: 757f17d2-f034-457f-80a2-329cfad513f7
workflow_transfer_payment:
   state=workflow_transfer_payment:510 (stFailureNotified) attempt=0
   run_id=757f17d2-f034-457f-80a2-329cfad513f7

[payment-core]
internal_auth: SUCCESS
   tx_id=57cd93f7d6534e6ca2b19183c8d03538
   group_id=253c9e27c69f465bbeed564eb16a4f0e
   type=AUTH
   error_code='' error_msg=''
   workflow:
      workflow_id=internal_payment_flow
      state=internal_payment_flow:900 (stSuccess) attempt=0
      run_id=57cd93f7d6534e6ca2b19183c8d03538
external_transaction:
   ref_id=fa24d1a78f1e4e1a85948d80302ac0f0
   group_id=253c9e27c69f465bbeed564eb16a4f0e
   type=TRANSFER status=FAILED
   workflow:
      workflow_id=external_payment_flow
      state=external_payment_flow:500 (stFailed) attempt=0
      run_id=fa24d1a78f1e4e1a85948d80302ac0f0

[Classification]
NOT_FOUND

### [2] transaction_id: 8d69bd2672a041c78d2c18784f83d8eb
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-15T03:48:44.479991Z
reference_id: 1b0d8214-8374-469a-a462-6ec5eb750658
workflow_transfer_payment:
   state=workflow_transfer_payment:102 (stTransactionLimitChecked) attempt=4
   run_id=1b0d8214-8374-469a-a462-6ec5eb750658

[payment-core]
internal_auth: SUCCESS
   tx_id=e190d5d404274615a97b228dc0cdbdf5
   group_id=8d69bd2672a041c78d2c18784f83d8eb
   type=AUTH
   error_code='' error_msg=''
   workflow:
      workflow_id=internal_payment_flow
      state=internal_payment_flow:900 (stSuccess) attempt=0
      run_id=e190d5d404274615a97b228dc0cdbdf5

[Classification]
pe_stuck_at_limit_check_102_4

### [3] transaction_id: 198fe80766cb48b4aca3cf8a38f5baa5
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-31T02:19:45.585394Z
reference_id: DE633D80-C1BB-4369-9A32-134307E8A837
workflow_transfer_payment:
   state=workflow_transfer_payment:220 (stTransferProcessing) attempt=0
   run_id=DE633D80-C1BB-4369-9A32-134307E8A837

[payment-core]
internal_auth: SUCCESS
   tx_id=757f9263d2394b3eb03bb66687092957
   group_id=198fe80766cb48b4aca3cf8a38f5baa5
   type=AUTH
   error_code='' error_msg=''
   workflow:
      workflow_id=internal_payment_flow
      state=internal_payment_flow:900 (stSuccess) attempt=0
      run_id=757f9263d2394b3eb03bb66687092957
external_transaction:
   ref_id=1bf90ebf1608442594e684b71e8c91bf
   group_id=198fe80766cb48b4aca3cf8a38f5baa5
   type=TRANSFER status=PROCESSING
   workflow:
      workflow_id=external_payment_flow
      state=external_payment_flow:201 (stProcessing) attempt=0
      run_id=1bf90ebf1608442594e684b71e8c91bf

[rpp-adapter]
req_biz_msg_id: 20251231GXSPMYKL030OQR12076882
partner_tx_id: 198fe80766cb48b4aca3cf8a38f5baa5
wf_ct_qr_payment:
   state=wf_ct_qr_payment:0 (stInit) attempt=20
   run_id=198fe80766cb48b4aca3cf8a38f5baa5
info: RPP Status: PROCESSING

[Classification]
NOT_FOUND

### [4] transaction_id: 90a8976b531446be8e00d42f02ff2d0d
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-26T13:32:25.308547Z
reference_id: D060C5AD-C53F-4CEC-AC60-E3B04AB9DE46
workflow_transfer_payment:
   state=workflow_transfer_payment:102 (stTransactionLimitChecked) attempt=0
   run_id=D060C5AD-C53F-4CEC-AC60-E3B04AB9DE46

[payment-core]
internal_auth: SUCCESS
   tx_id=ef8a3114ccab4c309cd7855270b5f221
   group_id=90a8976b531446be8e00d42f02ff2d0d
   type=AUTH
   error_code='' error_msg=''
   workflow:
      workflow_id=internal_payment_flow
      state=internal_payment_flow:900 (stSuccess) attempt=0
      run_id=ef8a3114ccab4c309cd7855270b5f221

[Classification]
pe_stuck_at_limit_check_102_4

