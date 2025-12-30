# RPP Workflow ID Validation Implementation Plan

## Summary
Modify the `GetDMLTicketForRppResume` function in [`internal/txn/adapters/sql.go`](internal/txn/adapters/sql.go) to validate workflow IDs before using them for SQL generation. Instead of always picking `Workflow[0]`, iterate through the workflow slice to find a matching workflow with `WorkflowID` equal to `wf_ct_cashout` or `wf_ct_qr_payment`.

## Current Issue
The function currently assumes `result.RPPAdapter.Workflow[0]` is the correct workflow, but the SQL query filters by `workflow_id IN ('wf_ct_cashout', 'wf_ct_qr_payment')`. If the first workflow has a different `WorkflowID`, the generated SQL would use an incorrect `run_id`.

## Implementation Steps

### 1. Create Helper Function
Add a new helper function in [`internal/txn/adapters/sql.go`](internal/txn/adapters/sql.go) to find a matching workflow:

```go
// findMatchingWorkflow searches for a workflow with the specified WorkflowID
// Returns the workflow and its index, or nil/-1 if not found
func findMatchingWorkflow(workflows []domain.WorkflowInfo, targetIDs []string) *domain.WorkflowInfo {
    for i := range workflows {
        for _, targetID := range targetIDs {
            if workflows[i].WorkflowID == targetID {
                return &workflows[i]
            }
        }
    }
    return nil
}
```

### 2. Update GetDMLTicketForRppResume
Modify the function to:
- Call the helper function to find the matching workflow
- Return `nil` if no matching workflow is found
- Use the found workflow's `RunID` for both Deploy and Rollback SQL parameters

### 3. Add Unit Tests
Create test cases in a new or existing test file:
- Test case: First workflow matches (`wf_ct_cashout`)
- Test case: Second workflow matches (`wf_ct_qr_payment`)
- Test case: No matching workflow found (returns `nil`)
- Test case: Empty workflow slice (returns `nil`)

### 4. Verification
Run tests to ensure:
- The function correctly finds matching workflows
- Returns `nil` when no match is found
- SQL generation uses the correct `run_id`

## Files to Modify
1. [`internal/txn/adapters/sql.go`](internal/txn/adapters/sql.go) - Add helper function and update `GetDMLTicketForRppResume`
2. [`internal/txn/adapters/sql_generator_test.go`](internal/txn/adapters/sql_generator_test.go) or new test file - Add unit tests

## Expected Behavior
- When a workflow with `WorkflowID == 'wf_ct_cashout'` or `'wf_ct_qr_payment'` exists in the slice, use its `RunID`
- When no matching workflow exists, return `nil` (no DML ticket generated)
- The SQL query's `workflow_id` filter matches the workflow whose `run_id` is used
