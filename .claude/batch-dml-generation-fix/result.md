7 buddy % git pull && make deploy && mybuddy txn TS-4558.txt
remote: Enumerating objects: 11, done.
remote: Counting objects: 100% (11/11), done.
remote: Compressing objects: 100% (2/2), done.
remote: Total 6 (delta 4), reused 6 (delta 4), pack-reused 0 (from 0)
Unpacking objects: 100% (6/6), 1.22 KiB | 83.00 KiB/s, done.
From github.com:4rankng/buddy
   98960d5..680bed8  main       -> origin/main
Updating 98960d5..680bed8
Fast-forward
 internal/txn/adapters/sql_generator.go | 17 ++++++++++++++---
 1 file changed, 14 insertions(+), 3 deletions(-)
Building mybuddy with Malaysia environment...
mybuddy built successfully
Building sgbuddy with Singapore environment...
sgbuddy built successfully
Building and deploying binaries...
Building mybuddy with Malaysia environment...
mybuddy built successfully
Building sgbuddy with Singapore environment...
sgbuddy built successfully
Deployed to /Users/frank.nguyen/bin
You can now use 'mybuddy' and 'sgbuddy' commands from anywhere.
[MY] Processing batch file: TS-4558.txt
[MY] Found 4 transaction IDs to process
[MY] Processing 1/4: 253c9e27c69f465bbeed564eb16a4f0e
[MY] Processing 2/4: 8d69bd2672a041c78d2c18784f83d8eb
[MY] Processing 3/4: 198fe80766cb48b4aca3cf8a38f5baa5
[MY] Processing 4/4: 90a8976b531446be8e00d42f02ff2d0d
[MY] 
Writing batch results to: TS-4558.txt_results.txt
[MY] Batch processing completed. Results written to TS-4558.txt_results.txt
[DEBUG] generateSQLFromTicket: case=pe_stuck_at_limit_check_102_4, deploy=2, rollback=2
[DEBUG] Processing rollback templates for case pe_stuck_at_limit_check_102_4: 1 groups
[DEBUG] Processing rollback group: targetDB=PE, runIDs=2
[DEBUG] Generated rollback SQL length: 249
[MY] 
SQL Generation Summary:
[MY]   Generated 4 SQL statements:
[MY]     PE Deploy: 4 statements

SQL statements written to PE_Deploy.sql
[MY] SQL DML files generated: [PE_Deploy.sql]
frank.nguyen@DBSG-H4M0DVF2C7 buddy % 


The PE_Rollback.sql is not generated

and content of PE_Deploy.sql is not correct

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', '57cd93f7d6534e6ca2b19183c8d03538'),
    updated_at = '2026-01-05T10:43:02.56458Z'
WHERE transaction_id = '253c9e27c69f465bbeed564eb16a4f0e';

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', 'e190d5d404274615a97b228dc0cdbdf5'),
    updated_at = '2025-12-15T03:48:44.683021Z'
WHERE transaction_id = '8d69bd2672a041c78d2c18784f83d8eb';

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', '757f9263d2394b3eb03bb66687092957'),
    updated_at = '2025-12-31T02:19:46.13746Z'
WHERE transaction_id = '198fe80766cb48b4aca3cf8a38f5baa5';

-- Update transfer table with AuthorisationID from payment-core internal_auth
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', 'ef8a3114ccab4c309cd7855270b5f221'),
    updated_at = '2025-12-26T13:32:25.422421Z'
WHERE transaction_id = '90a8976b531446be8e00d42f02ff2d0d';

