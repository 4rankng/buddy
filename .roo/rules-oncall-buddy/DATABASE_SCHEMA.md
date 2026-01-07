Database Schema Reference
Malaysia Services (mybuddy)
Payment Core (payment_core)
external_transaction (id number, tx_id string, ref_id string, tx_type string, core_acct_id string, holder_name string, ext_acct string, amount number, currency string, rails_tx_id string, status string, remarks string, group_id string, payment_rail string, error_code string, error_message string, metadata string, properties string, client_name string, group_tag string, created_at string, updated_at string)

internal_transaction (id number, tx_id string, ref_id string, src_acct_id string, dest_acct_id string, amount number, currency string, tx_type string, auth_tx_id string, group_id string, remarks string, metadata string, status string, error_code string, error_msg string, txn_domain string, txn_type string, txn_subtype string, client_name string, group_tag string, created_at string, updated_at string)

refund_transaction (id number, tx_id string, ref_id string, source_account_id string, destination_account_id string, amount number, cumulative_amount number, currency string, group_id string, original_tx_id string, version number, status string, error_code string, error_msg string, remarks string, metadata string, client_name string, group_tag string, txn_domain string, txn_type string, txn_subtype string, created_at string, updated_at string)

workflow_execution (id number, workflow_id string, run_id string, transition_id string, prev_trans_id string, attempt number, state number, data string, created_at string, updated_at string)

Payment Engine (payment_engine)
refund (id number, dakota_service_id string, transaction_id string, reference_id string, status string, version number, amount number, transfer_tx_id string, remarks string, status_reason string, status_reason_description string, txn_domain string, txn_type string, txn_subtype string, created_at string, valued_at string, updated_at string)

transfer (id number, type string, customer_id string, dakota_service_id string, transaction_id string, reference_id string, status string, amount number, currency string, source_account_id string, source_account string, destination_account_id string, destination_account string, properties string, status_reason string, status_reason_description string, capture_method string, metadata string, remarks string, mandate_id string, txn_domain string, txn_type string, txn_subtype string, external_id string, created_at string, valued_at string, updated_at string)

transfer_reversal (id number, type string, dakota_service_id string, transaction_id string, reference_id string, transfer_tx_id string, status string, amount number, currency string, source_account_id string, source_account string, destination_account_id string, destination_account string, properties string, metadata string, remarks string, status_reason string, status_reason_description string, capture_method string, txn_domain string, txn_type string, txn_subtype string, external_id string, created_at string, valued_at string, updated_at string)

workflow_execution (id number, workflow_id string, run_id string, transition_id string, prev_trans_id string, attempt number, state number, data string, created_at string, updated_at string)

RPP Adapter (rpp_adapter)
credit_transfer (id number, service_id string, partner_msg_id string, partner_tx_id string, partner_tx_sts string, partner_tx_sts_rsn string, tx_code string, tx_created_at string, req_biz_msg_id string, req_msg_id string, end_to_end_id string, end_to_end_id_signature string, tx_id string, amount number, currency string, settlement_cycle string, bank_settlement_date string, dbtr_id string, dbtr_name string, dbtr_acct_id string, dbtr_acct_type string, dbtr_agt_id string, dbtr_splmtry_info string, instr_for_dbtr_acct string, cdtr_agt_id string, cdtr_name string, cdtr_acct_id string, cdtr_acct_type string, instr_for_cdtr_agt string, cdtr_splmtry_info string, receipt_ref string, payment_description string, proxy_lookup_ref string, resp_biz_msg_id string, resp_msg_id string, resp_created_at string, tx_sts string, tx_sts_rsn string, tx_sts_rsn_description string, qr_tx_info string, created_at string, updated_at string)

workflow_execution (id number, workflow_id string, run_id string, transition_id string, prev_trans_id string, attempt number, state number, data string, created_at string, updated_at string)

Partnerpay Engine (partnerpay_engine)
charge (id number, customer_id string, partner_id string, transaction_id string, reference_id string, status string, amount number, currency string, source_account string, destination_account string, properties string, status_reason string, status_reason_description string, capture_method string, captured_amount number, metadata string, remarks string, billing_token string, txn_domain string, txn_type string, txn_subtype string, external_id string, transaction_payload string, created_at string, updated_at string, valued_at string)

intent (id number, intent_id string, idempotency_key string, partner_transaction_id string, payment_transaction_id string, customer_id string, partner_id string, source_account string, destination_account string, payment_method string, amount number, currency string, type string, status string, status_reason string, status_reason_description string, metadata string, created_at string, updated_at string, properties string)

workflow_execution (id number, workflow_id string, run_id string, transition_id string, prev_trans_id string, attempt number, state number, data string, created_at string, updated_at string)

Singapore Services (sgbuddy)
Payment Core (payment_core)
Same schemas as Malaysia: external_transaction, internal_transaction, refund_transaction, workflow_execution

Payment Engine (payment_engine)
Same schemas as Malaysia: refund, transfer, transfer_reversal, workflow_execution

Fast Adapter (fast_adapter)
transactions (id number, type string, instruction_id string, transaction_id string, end_to_end_id string, case_id string, originator_msg_id string, originator_cre_dt string, interbank_settlement_amt number, instructing_agent_bic string, instructed_agent_bic string, debtor_account_id string, debtor_account_nm string, creditor_account_id string, creditor_account_nm string, user_role number, status number, cancel_reason_code string, reject_reason_code string, purpose_txn_cd string, bcs_origin string, created_at string, updated_at string, cashout_txn_id string, postscript string, mandate_id string, reversal_id string, additional_info string)

Partnerpay Engine (partnerpay_engine)
Same schemas as Malaysia: charge, intent, workflow_execution

Common Patterns & Query Examples
Query Examples
Get table list: mybuddy doorman query -s <service> -q "SHOW TABLES" Get schema: mybuddy doorman query -s <service> -q "DESCRIBE <table>" Get indexes: mybuddy doorman query -s <service> -q "SHOW INDEX FROM <table>"

Services Summary
Malaysia (mybuddy): payment_core, payment_engine, rpp_adapter, partnerpay_engine

Singapore (sgbuddy): payment_core, payment_engine, fast_adapter, partnerpay_engine