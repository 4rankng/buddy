if we cannot find info for 20251231HBMBMYKL040OQR32713402 in rpp-adapter, most likely because it is stuck
at wf_process_registry
we can extract the [:8] from 20251231HBMBMYKL040OQR32713402
to get a start date
then query workflow_execution table to find the record by



select * from workflow_execution where
created_at >= '2025-12-31T00:00:00.00000Z' // extract from the end2end id
and created_at <= '2025-12-31T01:00:00.00000Z' // created_at + 1hour
and data like '%20251231HBMBMYKL040OQR32713402%'
and workflow_id='wf_process_registry'