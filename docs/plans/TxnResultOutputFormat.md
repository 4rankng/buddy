ensure that if

user call sgbuddy txn TS-4475.txt
or mybuddy txn TS-4475.txt

you need to
[1] write to stdout the summary

frank.nguyen@DBSG-H4M0DVF2C7 buddy % mybuddy txn TS-4475.txt
[MY] Processing batch file: TS-4475.txt
Processing 6 transaction IDs from TS-4475.txt

--- Generating SQL Statements ---
Results written to TS-4475.txt-output.txt
Summary:
  Total: 6
  Unmatched: 6
  Matched: 0
Case Type Breakdown
  pc_external_payment_flow_200_11: 0
  pc_external_payment_flow_201_0_RPP_210: 0
  pc_external_payment_flow_201_0_RPP_900: 0
  pe_transfer_payment_210_0: 0
  pe_220_0_fast_cashin_failed: 0
  rpp_cashout_reject_101_19: 0
  rpp_qr_payment_reject_210_0: 0
  rpp_no_response_resume: 0

[2] write result to
TS-4475.txt-output.txt
that contains

### [1] transaction_id: 2fad693bedfb4453b7e48fc7d52ed1d3
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-09T17:58:09.656808Z
reference_id: cb651a63-0cb2-4711-9b85-f2c86f2e1808
workflow_transfer_payment:
   state=stTransferProcessing(220) attempt=0
   run_id=cb651a63-0cb2-4711-9b85-f2c86f2e1808
[payment-core]
internal_transaction:
   tx_id=8d0529b9f1e44167a18a1ea2de30210d
   group_id=2fad693bedfb4453b7e48fc7d52ed1d3
   type=AUTH status=SUCCESS
external_transaction:
   ref_id=e759487dcf364ed59a53606214136d67
   group_id=2fad693bedfb4453b7e48fc7d52ed1d3
   type=TRANSFER status=PROCESSING
internal_payment_flow:
   state=stSuccess(900) attempt=0
   run_id=8d0529b9f1e44167a18a1ea2de30210d
external_payment_flow:
   state=stSubmitted(200) attempt=11
   run_id=e759487dcf364ed59a53606214136d67
[Classification]
NOT_FOUND // depend on classification matching result



Currently why all txn not found

Processing 6 transaction IDs from TS-4475.txt

--- Generating SQL Statements ---
Results written to TS-4475.txt-output.txt
Summary: 
  Total: 6
  Unmatched: 0
  Matched: 6
Case Type Breakdown
  pc_external_payment_flow_200_11: 0
  pc_external_payment_flow_201_0_RPP_210: 0
  pc_external_payment_flow_201_0_RPP_900: 0
  pe_transfer_payment_210_0: 0
  pe_220_0_fast_cashin_failed: 0
  rpp_cashout_reject_101_19: 0
  rpp_qr_payment_reject_210_0: 0
  rpp_no_response_resume: 0

  and the output file is not correct

  ### [1] transaction_id: 2fad693bedfb4453b7e48fc7d52ed1d3
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-09T17:58:09.656808Z
reference_id: cb651a63-0cb2-4711-9b85-f2c86f2e1808
workflow_transfer_payment: state=stTransferProcessing(220) attempt=0 run_id=cb651a63-0cb2-4711-9b85-f2c86f2e1808
[payment-core]
internal_transaction: tx_id=8d0529b9f1e44167a18a1ea2de30210d group_id=2fad693bedfb4453b7e48fc7d52ed1d3 type=AUTH status=SUCCESS
external_transaction: ref_id=e759487dcf364ed59a53606214136d67 group_id=2fad693bedfb4453b7e48fc7d52ed1d3 type=TRANSFER status=PROCESSING
internal_payment_flow: state=stSuccess(900) attempt=0 run_id=8d0529b9f1e44167a18a1ea2de30210d
external_payment_flow: state=stSubmitted(200) attempt=11 run_id=e759487dcf364ed59a53606214136d67
[Classification]
NOT_FOUND

### [2] transaction_id: 7fed916dc6204fc2a3fc495dce974dcc
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-09T17:49:14.331456Z
reference_id: 6bbf4799-2b06-4f45-96c5-7d4f11bf4022
workflow_transfer_payment: state=stTransferProcessing(220) attempt=0 run_id=6bbf4799-2b06-4f45-96c5-7d4f11bf4022
[payment-core]
internal_transaction: tx_id=114822dacf12429d9e87668ad43e92eb group_id=7fed916dc6204fc2a3fc495dce974dcc type=AUTH status=SUCCESS
external_transaction: ref_id=bedfe904b928498c8ed5d58c957efe53 group_id=7fed916dc6204fc2a3fc495dce974dcc type=TRANSFER status=PROCESSING
internal_payment_flow: state=stSuccess(900) attempt=0 run_id=114822dacf12429d9e87668ad43e92eb
external_payment_flow: state=stSubmitted(200) attempt=11 run_id=bedfe904b928498c8ed5d58c957efe53
[Classification]
NOT_FOUND

### [3] transaction_id: d7713bcaeb5c4ebb8745e740fdb669f7
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-09T17:49:10.412542Z
reference_id: 8a8535aa-ed23-4e87-a66e-d68bc6da1a6d
workflow_transfer_payment: state=stTransferProcessing(220) attempt=0 run_id=8a8535aa-ed23-4e87-a66e-d68bc6da1a6d
[payment-core]
internal_transaction: tx_id=76b173b273004727ade6ef7018ac98cf group_id=d7713bcaeb5c4ebb8745e740fdb669f7 type=AUTH status=SUCCESS
external_transaction: ref_id=418a847b3a3d4bd692e7a4e835eb1614 group_id=d7713bcaeb5c4ebb8745e740fdb669f7 type=TRANSFER status=PROCESSING
external_payment_flow: state=stSubmitted(200) attempt=11 run_id=418a847b3a3d4bd692e7a4e835eb1614
internal_payment_flow: state=stSuccess(900) attempt=0 run_id=76b173b273004727ade6ef7018ac98cf
[Classification]
NOT_FOUND

### [4] transaction_id: 8f3cc2c8b0c6475b9b13d7bfd17d52f6
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-09T17:50:07.239339Z
reference_id: 43A04EAF-82BA-4BB6-A15B-ED30496900E8
workflow_transfer_payment: state=stTransferProcessing(220) attempt=0 run_id=43A04EAF-82BA-4BB6-A15B-ED30496900E8
[payment-core]
internal_transaction: tx_id=1913cb3a511a41e0b590ac60a20875ce group_id=8f3cc2c8b0c6475b9b13d7bfd17d52f6 type=AUTH status=SUCCESS
external_transaction: ref_id=b61082ef68e34e30980983b34c37356c group_id=8f3cc2c8b0c6475b9b13d7bfd17d52f6 type=TRANSFER status=PROCESSING
internal_payment_flow: state=stSuccess(900) attempt=0 run_id=1913cb3a511a41e0b590ac60a20875ce
external_payment_flow: state=stSubmitted(200) attempt=11 run_id=b61082ef68e34e30980983b34c37356c
[Classification]
NOT_FOUND

### [5] transaction_id: 428407c8e2484ab0856a94d0515dc7a8
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-09T17:49:19.093111Z
reference_id: C6D40BB6-3474-457E-9CFF-C2607E931003
workflow_transfer_payment: state=stTransferProcessing(220) attempt=0 run_id=C6D40BB6-3474-457E-9CFF-C2607E931003
[payment-core]
internal_transaction: tx_id=5f0601d2c353425992b19b397de3ec72 group_id=428407c8e2484ab0856a94d0515dc7a8 type=AUTH status=SUCCESS
external_transaction: ref_id=4d65d5b857e0400db744d0191bd65b01 group_id=428407c8e2484ab0856a94d0515dc7a8 type=TRANSFER status=PROCESSING
external_payment_flow: state=stSubmitted(200) attempt=11 run_id=4d65d5b857e0400db744d0191bd65b01
internal_payment_flow: state=stSuccess(900) attempt=0 run_id=5f0601d2c353425992b19b397de3ec72
[Classification]
NOT_FOUND

### [6] transaction_id: 9fd468b0b4c6474b9f5ff4a634527372
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2025-12-09T17:51:21.524232Z
reference_id: 348f9bf2-6005-4bda-9f17-30ed48cd6c03
workflow_transfer_payment: state=stTransferProcessing(220) attempt=0 run_id=348f9bf2-6005-4bda-9f17-30ed48cd6c03
[payment-core]
internal_transaction: tx_id=74684305c5fa497398786f64361b3d47 group_id=9fd468b0b4c6474b9f5ff4a634527372 type=AUTH status=SUCCESS
external_transaction: ref_id=80e807184e3c46c99e739fcad39c7e47 group_id=9fd468b0b4c6474b9f5ff4a634527372 type=TRANSFER status=PROCESSING
internal_payment_flow: state=stSuccess(900) attempt=0 run_id=74684305c5fa497398786f64361b3d47
external_payment_flow: state=stSubmitted(200) attempt=11 run_id=80e807184e3c46c99e739fcad39c7e47
[Classification]
NOT_FOUND

