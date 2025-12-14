when mybuddy txn {text}

and text is in format 20251211MBBEMYKL070ORB42241026
first 8 chars is YYYYMMDD (please ensure it is date format)
last 8 chars is digits
string is 30 chars length

then text is e2e id

you need to search rpp-adapter for the record
select * from credit_transfer where end_to_end_id='20251211MBBEMYKL070ORB42241026'

and get partner_tx_id to search 
select * from workflow_execution where run_id='{partner_tx_id}'

having data from these two query you can fill up the 

RPPAdapterInfo


then with e2e id and created_at date from above to fill in PaymentEngineInfo

SELECT * FROM transfer where external_id='{e2e id}'
and created_at >= '{created_at - 30min}'
and created_at <= '{created_at + 30min}'

then you have transaction_id

with transaction_id you can fill up PaymentEngineInfo, PaymentCoreInfo

