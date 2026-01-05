# Requirements: Batch DML Generation Fix

## Context / Problem Statement
The mybuddy batch processing tool successfully queries transaction statuses and writes analysis to a text file, but fails to generate the corresponding SQL DML scripts (*.sql files) required to fix identified issues.

**Current Behavior:**
- Input: TS-4558.txt (transaction IDs)
- Output: TS-4558.txt_results.txt (created successfully)
- Missing: PE_Deploy.sql, PE_Rollback.sql, RPP_Deploy.sql, etc.

**Specific Issue - pe_stuck_at_limit_check_102_4:**
This error pattern requires a two-step fix:
1. Update workflow_execution state to 221 (Manual Reject)
2. Inject AuthorisationID from Payment Core into the transfer table

The root cause is in `internal/apps/common/batch/processor.go` - it calculates results but never invokes adapters to generate SQL statements and write them to disk.

## Goals
1. Enable batch processor to generate SQL DML files after analysis
2. Ensure pe_stuck_at_limit_check_102_4 generates proper UPDATE statements for both workflow_execution and transfer tables
3. Generate both Deploy and Rollback SQL file pairs
4. Preserve transfer.updated_at timestamp when updating

## Non-goals
- Modifying existing adapter logic for SQL statement generation
- Changing the transaction query/analysis logic
- Modifying the text file output format

## Assumptions & Constraints
- Adapters already have GenerateSQLStatements() functionality
- Adapters already have WriteSQLFiles() functionality
- SQL files should be written to current directory (where command is run)
- Existing pe_stuck_at_limit_check_102_4 adapter logic needs enhancement for workflow_execution updates

## Acceptance Criteria
- [ ] Running `mybuddy txn TS-4558.txt` generates PE_Deploy.sql and PE_Rollback.sql
- [ ] PE_Deploy.sql contains UPDATE workflow_execution statement (state=221)
- [ ] PE_Deploy.sql contains UPDATE transfer statement with AuthorisationID injection
- [ ] UPDATE transfer statement explicitly sets updated_at to preserve original timestamp
- [ ] Console output shows "SQL DML files generated: [...]"
- [ ] If no SQL fixes needed, console shows "No SQL fixes required for these transactions."

## Risks / Edge Cases / Rollback Considerations
- **Risk:** Incorrect timestamp preservation could cause transfer.updated_at to change unexpectedly
- **Risk:** Missing workflow_execution updates could leave transactions in stuck state
- **Edge Case:** Transactions may not have PaymentCore.InternalAuth.TxID available
- **Rollback:** If implementation breaks, revert processor.go changes; existing adapter logic remains untouched

## Open Questions
1. Should we clear previous SQL files before generating new ones? (Plan recommends yes via ClearSQLFiles()) --> YES
2. What should happen if SQL file generation fails partially? (Plan recommends logging error and continuing) --> Show user error and stop
