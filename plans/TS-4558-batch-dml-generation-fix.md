
why I dont see PE_Rollback.sql created???? We need to have a pair of deploy and rollback. and also for pe_stuck_at_limit_check_102_4 we need to update workflow_execution
/Users/frank.nguyen/Documents/buddy/internal/txn/adapters/sql_templates_pe_basic.go

Issue: Batch Processing Missing DML Generation (TS-4558)
1. Problem Statement
The mybuddy tool successfully queries transaction statuses in batch mode and writes the analysis to a text file. However, it fails to generate the corresponding SQL DML scripts (*.sql files) required to fix the identified issues.

Current Observation:

Input: TS-4558.txt (contains transaction IDs).

Output: TS-4558.txt_results.txt (created successfully).

Missing: PE_Deploy.sql, RPP_Deploy.sql, etc. (Files are not generated).

2. Specific SOP Logic: pe_stuck_at_limit_check_102_4
This specific error pattern requires a two-step fix. The Payment Engine (PE) is stuck at 102 (Limit Checked), while the Payment Core has successfully authorized the transaction.

To resolve this, we must manually inject the AuthorisationID from Payment Core into the Transfer table and reject the stuck workflow.

Logic Requirements:

Identify Auth ID: Extract internal_auth.tx_id from the Payment Core status (e.g., ef8a3114ccab4c309cd7855270b5f221).

Update Workflow: Set state to 221 (Manual Reject).

Update Transfer: Inject the AuthorisationID into the JSON properties and preserve the original updated_at timestamp.

Target SQL Output:

SQL

-- 1. Reset/Reject the Workflow Execution
UPDATE workflow_execution
SET state = 221, 
    attempt = 1, 
    `data` = JSON_SET(`data`, 
        '$.StreamMessage', JSON_OBJECT('Status', 'FAILED', 'ErrorCode', "ADAPTER_ERROR", 'ErrorMessage', 'Manual Rejected'),
        '$.State', 221)
WHERE run_id = 'D060C5AD-C53F-4CEC-AC60-E3B04AB9DE46' 
  AND state = 102 
  AND workflow_id = 'workflow_transfer_payment';

-- 2. Link the Authorisation ID to the Transfer Table
UPDATE transfer
SET properties = JSON_SET(properties, '$.AuthorisationID', 'ef8a3114ccab4c309cd7855270b5f221'),
    updated_at = '2025-12-26T13:32:25.308547Z' -- PRESERVE ORIGINAL TIMESTAMP
WHERE id = 228567995;
3. Root Cause Analysis
The issue lies in internal/apps/common/batch/processor.go. The batch processing flow calculates the results but never invokes the adapters responsible for converting those results into SQL statements and writing them to disk.

Missing Calls
adapters.GenerateSQLStatements(results): Converts the analysis into SQL strings.

adapters.WriteSQLFiles(statements, basePath): Writes those strings to the PE_Deploy.sql, PC_Deploy.sql, etc.

4. Implementation Plan
Step A: Update Batch Processor
Modify ProcessTransactionFile in internal/apps/common/batch/processor.go to include the SQL generation logic immediately after writing the text results.

Logical Flow:

Read File.

Query Transactions (Loop).

Write _results.txt.

[NEW] Generate SQL Statements from Results.

[NEW] Write SQL Files to current directory.

Step B: Code Logic (Go)
You need to inject the following logic into the processor:

Go

// ... existing code ...
// After WriteBatchResults(results, outputFilename)

// 1. Generate the DML objects
sqlStatements := adapters.GenerateSQLStatements(results)

// 2. Clear previous SQL files to avoid appending to old runs
adapters.ClearSQLFiles()

// 3. Write the new SQL files
// Note: WriteSQLFiles internally handles separating PE, PC, RPP, etc.
filesCreated, err := adapters.WriteSQLFiles(sqlStatements, ".")
if err != nil {
    fmt.Printf("Error writing SQL files: %v\n", err)
} else {
    if len(filesCreated) > 0 {
        fmt.Printf("SQL DML files generated: %v\n", filesCreated)
    } else {
        fmt.Println("No SQL fixes required for these transactions.")
    }
}
Step C: Update SOP Adapter
Ensure the adapter handling pe_stuck_at_limit_check_102_4 correctly implements the logic to fetch the AuthorisationID and format the Transfer table update.

Logic:

Check if PaymentCore.InternalAuth.Status == "SUCCESS".

Capture PaymentCore.InternalAuth.TxID.

Generate UPDATE transfer statement using JSON_SET with the captured ID.

Ensure the WHERE clause uses the Transfer ID.

Ensure updated_at is explicitly set to the current DB value (to prevent auto-update behavior from changing it).

5. Verification Checklist
To confirm the fix works, run the following:

Rebuild: make deploy

Run Batch: mybuddy txn TS-4558.txt

Check Console: Look for "SQL DML files generated: [...]"

Verify Files: cat PE_Deploy.sql

Verify Content: Ensure the UPDATE transfer statement contains the correct ef8a... ID and updated_at.