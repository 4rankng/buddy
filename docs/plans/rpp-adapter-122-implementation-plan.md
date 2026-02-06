# RPP Adapter 122 Implementation Plan

## Problem Analysis

### Issue Description
When querying an RPP adapter transaction with E2E ID `20251228TNGDMYNB010ORM77048250`, the system shows:
- RPP adapter data is correctly retrieved
- Workflow `wf_ct_cashin` with state `stFieldsValidationFailed(122)` and attempt `0` is found
- However, classification returns `NOT_FOUND` instead of the expected `rpp_cashin_validation_failed_122_0`
- No deploy/rollback SQL is generated

### Root Cause

The issue stems from how the RPP adapter returns workflow data and how the SOP rule evaluator processes it:

1. **RPP Adapter Returns a Slice of Workflows**: The `RPPAdapterInfo.Workflow` field is a `[]WorkflowInfo` slice containing multiple workflows (e.g., `wf_process_registry` and `wf_ct_cashin`).

2. **SOP Rule Evaluation Expects a Single Workflow**: The SOP rules use field paths like `RPPAdapter.Workflow.WorkflowID`, which the evaluator tries to access as a single struct, not a slice.

3. **getFieldValue Returns Nil for Slice Access**: When the evaluator tries to access `RPPAdapter.Workflow.WorkflowID` on a slice, it returns `(nil, false)` because the path doesn't match a struct field.

4. **getRPPWorkflowRunID Uses First Workflow**: The SQL template helper `getRPPWorkflowRunID()` simply returns `workflows[0].RunID`, which may not be the correct workflow (e.g., it might return `wf_process_registry` instead of `wf_ct_cashin`).

### Current Data Flow

```
RPPAdapter.QueryByE2EID()
  ↓
Returns RPPAdapterInfo with Workflow: []WorkflowInfo[
  {WorkflowID: "wf_process_registry", ...},
  {WorkflowID: "wf_ct_cashin", State: "122", Attempt: 0, ...}
]
  ↓
SOP Rule Evaluation: RPPAdapter.Workflow.WorkflowID
  ↓
getFieldValue("RPPAdapter.Workflow.WorkflowID") → (nil, false)
  ↓
Rule fails to match → Classification: NOT_FOUND
  ↓
No SQL generated
```

## Solution Design

### Approach 1: Modify SOP Rule Evaluation to Handle Workflow Slices (RECOMMENDED)

Update the SOP rule evaluator to:
1. Detect when accessing a field that is a slice
2. Check if ANY workflow in the slice matches the rule conditions
3. Return the matching workflow for SQL generation

**Pros:**
- Minimal changes to existing code
- Maintains backward compatibility with other adapters
- Flexible for future multi-workflow scenarios

**Cons:**
- Adds complexity to rule evaluation logic

### Approach 2: Add Helper to Extract Specific Workflow from RPP Adapter

Add a helper function to extract the workflow matching specific criteria (workflow_id, state, attempt) from the RPP adapter workflow slice.

**Pros:**
- Clear separation of concerns
- Easier to test and debug
- Can be reused in multiple places

**Cons:**
- Requires changes to multiple files
- May need to update SQL template helpers

## Implementation Plan

### Phase 1: Update SOP Rule Evaluator

**File: `internal/txn/adapters/sop_evaluator.go`**

1. Modify `getFieldValue()` to handle slice fields by returning the slice when accessed
2. Add `evaluateSliceCondition()` to check if ANY element in a workflow slice matches the rule
3. Ensure that when a workflow matches, we store which workflow matched for later SQL generation

### Phase 2: Update SQL Template Generation

**File: `internal/txn/adapters/sql_templates.go`**

1. Modify `getRPPWorkflowRunID()` to find the workflow that matches the case criteria
2. For `CaseRppCashinValidationFailed122_0`, find the workflow with:
   - `WorkflowID == "wf_ct_cashin"`
   - `State == "122"`
   - `Attempt == 0`

### Phase 3: Add Workflow Matching Helper

**File: `internal/txn/adapters/sql_templates.go`** (or new helper file)

Add a helper function:
```go
func findRPPWorkflow(workflows []domain.WorkflowInfo, workflowID, state string, attempt int) *domain.WorkflowInfo {
    for i := range workflows {
        if workflows[i].WorkflowID == workflowID &&
           workflows[i].State == state &&
           workflows[i].Attempt == attempt {
            return &workflows[i]
        }
    }
    return nil
}
```

### Phase 4: Update Domain Types (if needed)

**File: `internal/txn/domain/types.go`**

Consider adding a field to `RPPAdapterInfo` to store the matched workflow:
```go
type RPPAdapterInfo struct {
    ReqBizMsgID string
    PartnerTxID string
    EndToEndID  string
    Status      string
    CreatedAt   string
    Workflow    []WorkflowInfo
    MatchedWorkflow *WorkflowInfo // New field to store the workflow that matched the SOP rule
    Info        string
}
```

## Detailed Changes

### Change 1: Update `getFieldValue` in `sop_evaluator.go`

```go
func (r *SOPRepository) getFieldValue(fieldPath string, result *domain.TransactionResult) (interface{}, bool) {
    parts := strings.Split(fieldPath, ".")
    current := reflect.ValueOf(result)

    for i, part := range parts {
        // Dereference pointers safely
        if current.Kind() == reflect.Ptr {
            if current.IsNil() {
                return nil, false
            }
            current = current.Elem()
        }

        if current.Kind() != reflect.Struct {
            return nil, false
        }

        field := current.FieldByName(part)
        if !field.IsValid() {
            return nil, false
        }

        // If this is the last part and the field is a slice, return the slice
        if i == len(parts)-1 && field.Kind() == reflect.Slice {
            return field.Interface(), true
        }

        current = field
    }

    // If final value is a pointer, preserve nil vs non-nil
    if current.Kind() == reflect.Ptr {
        if current.IsNil() {
            return nil, true
        }
        current = current.Elem()
    }

    return current.Interface(), true
}
```

### Change 2: Update `evaluateCondition` in `sop_evaluator.go`

The existing `evaluateSliceCondition()` already handles slice evaluation correctly. We just need to ensure it's being called when the field is a slice.

### Change 3: Update SQL Template Helper in `sql_templates.go`

```go
func getRPPWorkflowRunID(workflows []domain.WorkflowInfo, workflowID, state string, attempt int) string {
    for _, wf := range workflows {
        if wf.WorkflowID == workflowID &&
           wf.State == state &&
           (attempt == -1 || wf.Attempt == attempt) {
            return wf.RunID
        }
    }
    // Fallback to first workflow if no match
    if len(workflows) > 0 {
        return workflows[0].RunID
    }
    return ""
}
```

### Change 4: Update SQL Template for CaseRppCashinValidationFailed122_0

```go
domain.CaseRppCashinValidationFailed122_0: func(result domain.TransactionResult) *domain.DMLTicket {
    if result.RPPAdapter == nil {
        return nil
    }

    // Find the specific workflow matching the case criteria
    runID := getRPPWorkflowRunID(
        result.RPPAdapter.Workflow,
        "wf_ct_cashin",
        "122",
        0,
    )

    if runID == "" {
        return nil // No matching workflow found
    }

    return &domain.DMLTicket{
        Deploy: []domain.TemplateInfo{
            {
                TargetDB: "RPP",
                SQLTemplate: `-- rpp_cashin_validation_failed_122_0, retry validation
UPDATE workflow_execution
SET state = 100,
    attempt = 1,
    data = JSON_SET(data, '$.State', 100)
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin'
AND state = 122;`,
                Params: []domain.ParamInfo{
                    {Name: "run_id", Value: runID, Type: "string"},
                },
            },
        },
        Rollback: []domain.TemplateInfo{
            {
                TargetDB: "RPP",
                SQLTemplate: `-- RPP Rollback: Move cashin workflow back to state 122
UPDATE workflow_execution
SET state = 122,
    attempt = 0,
    data = JSON_SET(data, '$.State', 122)
WHERE run_id = %s
AND workflow_id = 'wf_ct_cashin';`,
                Params: []domain.ParamInfo{
                    {Name: "run_id", Value: runID, Type: "string"},
                },
            },
        },
        CaseType: domain.CaseRppCashinValidationFailed122_0,
    }
},
```

## Testing Strategy

### Test Case 1: RPP Cashin Validation Failed 122/0

**Input:**
```
E2E ID: 20251228TNGDMYNB010ORM77048250
RPP Adapter:
  - req_biz_msg_id: 20251228RPPEMYKL010HRM30218865
  - partner_tx_id: 6ec50daa4a373a2f9fae4a6aec670679
  - Workflow[0]: wf_process_registry (state: 900, attempt: 0)
  - Workflow[1]: wf_ct_cashin (state: 122, attempt: 0, run_id: 6ec50daa4a373a2f9fae4a6aec670679)
```

**Expected Output:**
```
Classification: rpp_cashin_validation_failed_122_0
Deploy SQL:
  UPDATE workflow_execution
  SET state = 100, attempt = 1, data = JSON_SET(data, '$.State', 100)
  WHERE run_id = '6ec50daa4a373a2f9fae4a6aec670679'
  AND workflow_id = 'wf_ct_cashin' AND state = 122;

Rollback SQL:
  UPDATE workflow_execution
  SET state = 122, attempt = 0, data = JSON_SET(data, '$.State', 122)
  WHERE run_id = '6ec50daa4a373a2f9fae4a6aec670679'
  AND workflow_id = 'wf_ct_cashin';
```

### Test Case 2: RPP Cashout Reject 101/19

Verify that other RPP cases continue to work correctly with the workflow slice.

### Test Case 3: No Matching Workflow

Verify that when no workflow matches the criteria, the system returns NOT_FOUND without errors.

## Files to Modify

1. `internal/txn/adapters/sop_evaluator.go` - Update `getFieldValue()` to return slices
2. `internal/txn/adapters/sql_templates.go` - Update `getRPPWorkflowRunID()` and SQL templates

## Backward Compatibility

The changes are backward compatible because:
- Existing single-workflow scenarios will continue to work (a single workflow in a slice will match)
- The slice evaluation logic already exists in `evaluateSliceCondition()`
- SQL templates that don't specify workflow criteria will use the first workflow (existing behavior)

## Risk Assessment

**Low Risk:**
- The slice evaluation logic already exists and is tested
- We're only changing how workflows are selected for SQL generation
- The SOP rule matching logic remains the same

**Mitigation:**
- Add unit tests for the new workflow matching logic
- Test with real transaction data before deploying
- Monitor production logs for any unexpected behavior
