Add Case_PeCaptureProcessing_PcCaptureFailed_RppSuccess

check

payment engine workflow_transfer_payment:

   state=stCaptureProcessing(230) attempt=0



payment core    type=CAPTURE status=FAILED

   workflow:

      workflow_id=internal_payment_flow

      state=stFailed(500) attempt=0



Rpp-adapter

wf_ct_qr_payment or wf_ct_cashout

   state=stSuccess(900) attempt=0

Based on the "DML SOP: Payment Transaction Fix Protocols" provided, here are the classifications for the identified cases.

[2] e2e_id: 20251212GXSPMYKL040OQR32194316
Classification: thought_machine_false_negative Reasoning:

Payment Core (PC) State: The internal_payment_flow (CAPTURE) is at state 500 (stFailed). This matches the specific state criteria listed under the thought_machine_false_negative case in the SOP (payment core internal_payment_flow state 500).

RPP Adapter State: The RPP state is 900 (Success), indicating the transaction was successful externally, but the internal core failed/returned a false negative.

Fix Protocol: The required fix is to patch the data to retry the flow (reset PC from 500 to 0) as detailed in PC_Deploy.sql.   