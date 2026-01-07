frank.nguyen@DBSG-H4M0DVF2C7 buddy % 
mybuddy txn TS-4583.txt
[MY] Processing batch file: TS-4583.txt
[MY] Found 6 transaction IDs to process
[MY] Processing 1/6: 20260102GXSPMYKL010ORB83560088
[MY] Processing 2/6: 20260102GXSPMYKL010ORB90952799
[MY] Processing 3/6: 20260102GXSPMYKL030OQR30035724
[MY] Processing 4/6: 20260102GXSPMYKL040OQR15579388
[MY] Processing 5/6: 20260102GXSPMYKL040OQR21626342
[MY] Processing 6/6: 20260102GXSPMYKL040OQR73934108
[MY] 
Writing batch results to: TS-4583.txt_results.txt
[MY] Batch processing completed. Results written to TS-4583.txt_results.txt

================================================================================
### [1] e2e_id: 20260102GXSPMYKL010ORB83560088
[payment-engine]
type: PAYMENT
subtype: RPP_NETWORK
domain: DEPOSITS
status: PROCESSING
created_at: 2026-01-02T08:13:42.9042Z
reference_id: 7E651B98-61E7-45FE-A3C6-EE9FEEE68629
workflow_transfer_payment:
   state=workflow_transfer_payment:220 (stTransferProcessing) attempt=0
   run_id=7E651B98-61E7-45FE-A3C6-EE9FEEE68629

[payment-core]
internal_auth: SUCCESS
   tx_id=60f09005d0dc4b2d9e82153be492dc70
   group_id=ba2c3de7f1f1417a92105223bc04bc7e
   type=AUTH
   error_code='' error_msg=''
   workflow:
      workflow_id=internal_payment_flow
      state=internal_payment_flow:900 (stSuccess) attempt=0
      run_id=60f09005d0dc4b2d9e82153be492dc70
external_transaction:
   ref_id=c100947c022545368d68da492043b701
   group_id=ba2c3de7f1f1417a92105223bc04bc7e
   type=TRANSFER status=PROCESSING
   workflow:
      workflow_id=external_payment_flow
      state=external_payment_flow:201 (stProcessing) attempt=0
      run_id=c100947c022545368d68da492043b701

[rpp-adapter]
e2e_id: 20260102GXSPMYKL010ORB83560088
credit_transfer.status: PROCESSING
partner_tx_id: ba2c3de7f1f1417a92105223bc04bc7e
wf_process_registry:
   state=wf_process_registry:0 (stInit) attempt=0
   run_id=d9e1fbc16edf3f4fac8ff33f4c0b6f88
wf_ct_cashout:
   state=wf_ct_cashout:210 (stTransferProcessing) attempt=0
   run_id=ba2c3de7f1f1417a92105223bc04bc7e

[Classification]
cashout_rpp210_pe220_pc201


Choose an option:
1. Resume to Success (Manual Success) - This once
2. Reject/Fail (Manual Rejection) - This once
3. Resume to Success (Manual Success) - Apply to all similar
4. Reject/Fail (Manual Rejection) - Apply to all similar

Enter your choice (1, 2, 3, or 4): 

add flag --auto to the command
mybuddy txn TS-4583.txt --auto

then we will choose to resume all if ticket title contains 
"Debit Account confirmation" or " Crebit Account confirmation"

