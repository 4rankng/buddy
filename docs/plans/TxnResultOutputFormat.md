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

