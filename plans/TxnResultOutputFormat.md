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
workflow_transfer_payment: state=stTransferProcessing(220) attempt=0 run_id=cb651a63-0cb2-4711-9b85-f2c86f2e1808
[payment-core]
internal_transaction: tx_id=8d0529b9f1e44167a18a1ea2de30210d group_id=2fad693bedfb4453b7e48fc7d52ed1d3 type=AUTH status=SUCCESS
external_transaction: ref_id=e759487dcf364ed59a53606214136d67 group_id=2fad693bedfb4453b7e48fc7d52ed1d3 type=TRANSFER status=PROCESSING
internal_payment_flow: state=stSuccess(900) attempt=0 run_id=8d0529b9f1e44167a18a1ea2de30210d
external_payment_flow: state=stSubmitted(200) attempt=11 run_id=e759487dcf364ed59a53606214136d67
[Classification]
NOT_FOUND // depend on classification matching result

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
NOT_FOUND // depend on classification matching result








frank.nguyen@DBSG-H4M0DVF2C7 buddy % ./bin/mybuddy txn TS-4475.txt
[MY] Processing batch file: TS-4475.txt
[MY] Processing batch file: TS-4475.txt
Processing 6 transaction IDs from TS-4475.txt
[DEBUG] Identifying SOP case for transaction 2fad693bedfb4453b7e48fc7d52ed1d3 in environment my
[DEBUG] PaymentCore.Workflow count: 2
[DEBUG] PaymentEngine.Workflow.State: 220
[DEBUG] PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment
[DEBUG] PaymentEngine.Workflow.Attempt: 0
[DEBUG] RPPAdapter.Status: 
[DEBUG] RPPAdapter.Workflow.State: 
[DEBUG] RPPAdapter.Workflow.WorkflowID: 
[DEBUG] RPPAdapter.Workflow.Attempt: 0
[DEBUG] FastAdapter.Status: 
[DEBUG] Evaluating rule: pc_external_payment_flow_200_11
[DEBUG] Evaluating condition: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Getting field value for path: PaymentCore.Workflow.WorkflowID (parts: [PaymentCore Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentCore, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: slice
[DEBUG] Encountered slice with 2 elements
[DEBUG] Returning whole slice for path PaymentCore.Workflow.WorkflowID
[DEBUG] Field value for PaymentCore.Workflow.WorkflowID: [{internal_payment_flow 0 900 8d0529b9f1e44167a18a1ea2de30210d} {external_payment_flow 11 200 e759487dcf364ed59a53606214136d67}] (type: []domain.WorkflowInfo)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Rule did not match: pc_external_payment_flow_200_11
[DEBUG] Evaluating rule: pe_transfer_payment_210_0
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 210
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.State eq 210
[DEBUG] Rule did not match: pe_transfer_payment_210_0
[DEBUG] Evaluating rule: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating condition: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Getting field value for path: PaymentEngine.Workflow.WorkflowID (parts: [PaymentEngine Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Rule did not match: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating rule: rpp_cashout_reject_101_19
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Rule did not match: rpp_cashout_reject_101_19
[DEBUG] Evaluating rule: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Rule did not match: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating rule: rpp_no_response_resume
[DEBUG] Evaluating condition: RPPAdapter.Workflow.State eq 210
[DEBUG] Getting field value for path: RPPAdapter.Workflow.State (parts: [RPPAdapter Workflow State])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.State eq 210
[DEBUG] Rule did not match: rpp_no_response_resume
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_210
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_210
[DEBUG] No rules matched for transaction 2fad693bedfb4453b7e48fc7d52ed1d3
[DEBUG] Identifying SOP case for transaction 7fed916dc6204fc2a3fc495dce974dcc in environment my
[DEBUG] PaymentCore.Workflow count: 2
[DEBUG] PaymentEngine.Workflow.State: 220
[DEBUG] PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment
[DEBUG] PaymentEngine.Workflow.Attempt: 0
[DEBUG] RPPAdapter.Status: 
[DEBUG] RPPAdapter.Workflow.State: 
[DEBUG] RPPAdapter.Workflow.WorkflowID: 
[DEBUG] RPPAdapter.Workflow.Attempt: 0
[DEBUG] FastAdapter.Status: 
[DEBUG] Evaluating rule: pc_external_payment_flow_200_11
[DEBUG] Evaluating condition: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Getting field value for path: PaymentCore.Workflow.WorkflowID (parts: [PaymentCore Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentCore, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: slice
[DEBUG] Encountered slice with 2 elements
[DEBUG] Returning whole slice for path PaymentCore.Workflow.WorkflowID
[DEBUG] Field value for PaymentCore.Workflow.WorkflowID: [{internal_payment_flow 0 900 114822dacf12429d9e87668ad43e92eb} {external_payment_flow 11 200 bedfe904b928498c8ed5d58c957efe53}] (type: []domain.WorkflowInfo)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Rule did not match: pc_external_payment_flow_200_11
[DEBUG] Evaluating rule: pe_transfer_payment_210_0
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 210
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.State eq 210
[DEBUG] Rule did not match: pe_transfer_payment_210_0
[DEBUG] Evaluating rule: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating condition: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Getting field value for path: PaymentEngine.Workflow.WorkflowID (parts: [PaymentEngine Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Rule did not match: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating rule: rpp_cashout_reject_101_19
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Rule did not match: rpp_cashout_reject_101_19
[DEBUG] Evaluating rule: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Rule did not match: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating rule: rpp_no_response_resume
[DEBUG] Evaluating condition: RPPAdapter.Workflow.State eq 210
[DEBUG] Getting field value for path: RPPAdapter.Workflow.State (parts: [RPPAdapter Workflow State])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.State eq 210
[DEBUG] Rule did not match: rpp_no_response_resume
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_210
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_210
[DEBUG] No rules matched for transaction 7fed916dc6204fc2a3fc495dce974dcc
[DEBUG] Identifying SOP case for transaction d7713bcaeb5c4ebb8745e740fdb669f7 in environment my
[DEBUG] PaymentCore.Workflow count: 2
[DEBUG] PaymentEngine.Workflow.State: 220
[DEBUG] PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment
[DEBUG] PaymentEngine.Workflow.Attempt: 0
[DEBUG] RPPAdapter.Status: 
[DEBUG] RPPAdapter.Workflow.State: 
[DEBUG] RPPAdapter.Workflow.WorkflowID: 
[DEBUG] RPPAdapter.Workflow.Attempt: 0
[DEBUG] FastAdapter.Status: 
[DEBUG] Evaluating rule: pc_external_payment_flow_200_11
[DEBUG] Evaluating condition: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Getting field value for path: PaymentCore.Workflow.WorkflowID (parts: [PaymentCore Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentCore, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: slice
[DEBUG] Encountered slice with 2 elements
[DEBUG] Returning whole slice for path PaymentCore.Workflow.WorkflowID
[DEBUG] Field value for PaymentCore.Workflow.WorkflowID: [{external_payment_flow 11 200 418a847b3a3d4bd692e7a4e835eb1614} {internal_payment_flow 0 900 76b173b273004727ade6ef7018ac98cf}] (type: []domain.WorkflowInfo)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Rule did not match: pc_external_payment_flow_200_11
[DEBUG] Evaluating rule: pe_transfer_payment_210_0
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 210
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.State eq 210
[DEBUG] Rule did not match: pe_transfer_payment_210_0
[DEBUG] Evaluating rule: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating condition: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Getting field value for path: PaymentEngine.Workflow.WorkflowID (parts: [PaymentEngine Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Rule did not match: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating rule: rpp_cashout_reject_101_19
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Rule did not match: rpp_cashout_reject_101_19
[DEBUG] Evaluating rule: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Rule did not match: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating rule: rpp_no_response_resume
[DEBUG] Evaluating condition: RPPAdapter.Workflow.State eq 210
[DEBUG] Getting field value for path: RPPAdapter.Workflow.State (parts: [RPPAdapter Workflow State])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.State eq 210
[DEBUG] Rule did not match: rpp_no_response_resume
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_210
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_210
[DEBUG] No rules matched for transaction d7713bcaeb5c4ebb8745e740fdb669f7
[DEBUG] Identifying SOP case for transaction 8f3cc2c8b0c6475b9b13d7bfd17d52f6 in environment my
[DEBUG] PaymentCore.Workflow count: 2
[DEBUG] PaymentEngine.Workflow.State: 220
[DEBUG] PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment
[DEBUG] PaymentEngine.Workflow.Attempt: 0
[DEBUG] RPPAdapter.Status: 
[DEBUG] RPPAdapter.Workflow.State: 
[DEBUG] RPPAdapter.Workflow.WorkflowID: 
[DEBUG] RPPAdapter.Workflow.Attempt: 0
[DEBUG] FastAdapter.Status: 
[DEBUG] Evaluating rule: pc_external_payment_flow_200_11
[DEBUG] Evaluating condition: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Getting field value for path: PaymentCore.Workflow.WorkflowID (parts: [PaymentCore Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentCore, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: slice
[DEBUG] Encountered slice with 2 elements
[DEBUG] Returning whole slice for path PaymentCore.Workflow.WorkflowID
[DEBUG] Field value for PaymentCore.Workflow.WorkflowID: [{internal_payment_flow 0 900 1913cb3a511a41e0b590ac60a20875ce} {external_payment_flow 11 200 b61082ef68e34e30980983b34c37356c}] (type: []domain.WorkflowInfo)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Rule did not match: pc_external_payment_flow_200_11
[DEBUG] Evaluating rule: pe_transfer_payment_210_0
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 210
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.State eq 210
[DEBUG] Rule did not match: pe_transfer_payment_210_0
[DEBUG] Evaluating rule: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating condition: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Getting field value for path: PaymentEngine.Workflow.WorkflowID (parts: [PaymentEngine Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Rule did not match: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating rule: rpp_cashout_reject_101_19
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Rule did not match: rpp_cashout_reject_101_19
[DEBUG] Evaluating rule: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Rule did not match: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating rule: rpp_no_response_resume
[DEBUG] Evaluating condition: RPPAdapter.Workflow.State eq 210
[DEBUG] Getting field value for path: RPPAdapter.Workflow.State (parts: [RPPAdapter Workflow State])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.State eq 210
[DEBUG] Rule did not match: rpp_no_response_resume
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_210
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_210
[DEBUG] No rules matched for transaction 8f3cc2c8b0c6475b9b13d7bfd17d52f6
[DEBUG] Identifying SOP case for transaction 428407c8e2484ab0856a94d0515dc7a8 in environment my
[DEBUG] PaymentCore.Workflow count: 2
[DEBUG] PaymentEngine.Workflow.State: 220
[DEBUG] PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment
[DEBUG] PaymentEngine.Workflow.Attempt: 0
[DEBUG] RPPAdapter.Status: 
[DEBUG] RPPAdapter.Workflow.State: 
[DEBUG] RPPAdapter.Workflow.WorkflowID: 
[DEBUG] RPPAdapter.Workflow.Attempt: 0
[DEBUG] FastAdapter.Status: 
[DEBUG] Evaluating rule: pc_external_payment_flow_200_11
[DEBUG] Evaluating condition: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Getting field value for path: PaymentCore.Workflow.WorkflowID (parts: [PaymentCore Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentCore, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: slice
[DEBUG] Encountered slice with 2 elements
[DEBUG] Returning whole slice for path PaymentCore.Workflow.WorkflowID
[DEBUG] Field value for PaymentCore.Workflow.WorkflowID: [{external_payment_flow 11 200 4d65d5b857e0400db744d0191bd65b01} {internal_payment_flow 0 900 5f0601d2c353425992b19b397de3ec72}] (type: []domain.WorkflowInfo)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Rule did not match: pc_external_payment_flow_200_11
[DEBUG] Evaluating rule: pe_transfer_payment_210_0
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 210
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.State eq 210
[DEBUG] Rule did not match: pe_transfer_payment_210_0
[DEBUG] Evaluating rule: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating condition: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Getting field value for path: PaymentEngine.Workflow.WorkflowID (parts: [PaymentEngine Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Rule did not match: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating rule: rpp_cashout_reject_101_19
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Rule did not match: rpp_cashout_reject_101_19
[DEBUG] Evaluating rule: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Rule did not match: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating rule: rpp_no_response_resume
[DEBUG] Evaluating condition: RPPAdapter.Workflow.State eq 210
[DEBUG] Getting field value for path: RPPAdapter.Workflow.State (parts: [RPPAdapter Workflow State])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.State eq 210
[DEBUG] Rule did not match: rpp_no_response_resume
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_210
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_210
[DEBUG] No rules matched for transaction 428407c8e2484ab0856a94d0515dc7a8
[DEBUG] Identifying SOP case for transaction 9fd468b0b4c6474b9f5ff4a634527372 in environment my
[DEBUG] PaymentCore.Workflow count: 2
[DEBUG] PaymentEngine.Workflow.State: 220
[DEBUG] PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment
[DEBUG] PaymentEngine.Workflow.Attempt: 0
[DEBUG] RPPAdapter.Status: 
[DEBUG] RPPAdapter.Workflow.State: 
[DEBUG] RPPAdapter.Workflow.WorkflowID: 
[DEBUG] RPPAdapter.Workflow.Attempt: 0
[DEBUG] FastAdapter.Status: 
[DEBUG] Evaluating rule: pc_external_payment_flow_200_11
[DEBUG] Evaluating condition: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Getting field value for path: PaymentCore.Workflow.WorkflowID (parts: [PaymentCore Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentCore, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: slice
[DEBUG] Encountered slice with 2 elements
[DEBUG] Returning whole slice for path PaymentCore.Workflow.WorkflowID
[DEBUG] Field value for PaymentCore.Workflow.WorkflowID: [{internal_payment_flow 0 900 74684305c5fa497398786f64361b3d47} {external_payment_flow 11 200 80e807184e3c46c99e739fcad39c7e47}] (type: []domain.WorkflowInfo)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentCore.Workflow.WorkflowID eq external_payment_flow
[DEBUG] Rule did not match: pc_external_payment_flow_200_11
[DEBUG] Evaluating rule: pe_transfer_payment_210_0
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 210
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.State eq 210
[DEBUG] Rule did not match: pe_transfer_payment_210_0
[DEBUG] Evaluating rule: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating condition: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Getting field value for path: PaymentEngine.Workflow.WorkflowID (parts: [PaymentEngine Workflow WorkflowID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.WorkflowID: workflow_transfer_payment (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Workflow.WorkflowID eq workflow_transfer_collection
[DEBUG] Rule did not match: pe_220_0_fast_cashin_failed
[DEBUG] Evaluating rule: rpp_cashout_reject_101_19
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_cashout
[DEBUG] Rule did not match: rpp_cashout_reject_101_19
[DEBUG] Evaluating rule: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating condition: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Getting field value for path: RPPAdapter.Workflow.WorkflowID (parts: [RPPAdapter Workflow WorkflowID])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: WorkflowID, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.WorkflowID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.WorkflowID eq wf_ct_qr_payment
[DEBUG] Rule did not match: rpp_qr_payment_reject_210_0
[DEBUG] Evaluating rule: rpp_no_response_resume
[DEBUG] Evaluating condition: RPPAdapter.Workflow.State eq 210
[DEBUG] Getting field value for path: RPPAdapter.Workflow.State (parts: [RPPAdapter Workflow State])
[DEBUG] Processing part 0: RPPAdapter, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Field value for RPPAdapter.Workflow.State:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: RPPAdapter.Workflow.State eq 210
[DEBUG] Rule did not match: rpp_no_response_resume
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_900
[DEBUG] Evaluating rule: pc_external_payment_flow_201_0_RPP_210
[DEBUG] Evaluating condition: PaymentEngine.Workflow.State eq 220
[DEBUG] Getting field value for path: PaymentEngine.Workflow.State (parts: [PaymentEngine Workflow State])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: State, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Field value for PaymentEngine.Workflow.State: 220 (type: string)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.State eq 220
[DEBUG] Evaluating condition: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Getting field value for path: PaymentEngine.Workflow.Attempt (parts: [PaymentEngine Workflow Attempt])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Workflow, current value type: struct
[DEBUG] Processing part 2: Attempt, current value type: struct
[DEBUG] Final value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Field value for PaymentEngine.Workflow.Attempt: 0 (type: int)
[DEBUG] Condition evaluation result: true
[DEBUG] Condition passed: PaymentEngine.Workflow.Attempt eq 0
[DEBUG] Evaluating condition: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Getting field value for path: PaymentEngine.Transfers.ExternalID (parts: [PaymentEngine Transfers ExternalID])
[DEBUG] Processing part 0: PaymentEngine, current value type: ptr
[DEBUG] Processing part 1: Transfers, current value type: struct
[DEBUG] Processing part 2: ExternalID, current value type: struct
[DEBUG] Final value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Field value for PaymentEngine.Transfers.ExternalID:  (type: string)
[DEBUG] Condition evaluation result: false
[DEBUG] Condition failed: PaymentEngine.Transfers.ExternalID ne 
[DEBUG] Rule did not match: pc_external_payment_flow_201_0_RPP_210
[DEBUG] No rules matched for transaction 9fd468b0b4c6474b9f5ff4a634527372

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
frank.nguyen@DBSG-H4M0DVF2C7 buddy % 