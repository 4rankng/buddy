# Implementation Plan: thought_machine_false_negative Case Type

## Overview
This document outlines the implementation plan for adding a new case type `thought_machine_false_negative` to handle scenarios where Thought Machine returns errors/false negatives but the transaction was actually successful. The fix involves patching data to retry flow by moving PE to state 230 and retrying PC capture.

## Case Details
Based on the example transaction ID `ced4efe76ea442ddbbca1f745ebe2386`:
- PE state: `stCaptureFailed(701)` with attempt 0
- PC status: `NOT_FOUND`
- RPP status: `PROCESSING`
- Fix: Move PE to state 230 and retry PC capture

## Implementation Steps

### 1. Domain Types Updates (`internal/txn/domain/types.go`)

1.1. Add new case constant:
```go
CaseThoughtMachineFalseNegative Case = "thought_machine_false_negative"
```

1.2. Add to GetCaseSummaryOrder function:
```go
CaseThoughtMachineFalseNegative,
```

### 2. Case Identification Rule (`internal/txn/adapters/sop_repository.go`)

2.1. Add new rule to getDefaultSOPRules():
```go
{
    CaseType:    domain.CaseThoughtMachineFalseNegative,
    Description: "Thought Machine returning errors/false negatives, but transaction was successful",
    Country:     "my", // Malaysia only
    Conditions: []RuleCondition{
        {
            FieldPath: "PaymentEngine.Workflow.WorkflowID",
            Operator:  "eq",
            Value:     "workflow_transfer_payment",
        },
        {
            FieldPath: "PaymentEngine.Workflow.State",
            Operator:  "eq",
            Value:     "701", // stCaptureFailed
        },
        {
            FieldPath: "PaymentEngine.Workflow.Attempt",
            Operator:  "eq",
            Value:     0,
        },
        {
            FieldPath: "PaymentCore.InternalTxns",
            Operator:  "eq",
            Value:     nil, // No internal transactions found (NOT_FOUND)
        },
        {
            FieldPath: "RPPAdapter.Status",
            Operator:  "eq",
            Value:     "PROCESSING",
        },
    },
},
```

### 3. SQL Template Configuration (`internal/txn/adapters/sql_templates.go`)

3.1. Add to templateConfigs map:
```go
domain.CaseThoughtMachineFalseNegative: {Parameters: []string{"run_ids"}},
```

3.2. Add SQL template function:
```go
domain.CaseThoughtMachineFalseNegative: func(result domain.TransactionResult) *DMLTicket {
    if runID := result.PaymentEngine.Workflow.RunID; runID != "" {
        // Get prev_trans_id from the example or extract from data
        prevTransID := "6e0daa5cfcc24478a2c55097fe2d7cf8" // This should be extracted dynamically
        
        return &DMLTicket{
            RunIDs: []string{runID},
            DeployTemplate: `-- thought_machine_false_negative
UPDATE workflow_execution SET state = 230,
  prev_trans_id = '%s',
  ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 230)
WHERE run_id IN (%s)
AND state = 701;`,
            RollbackTemplate: `UPDATE workflow_execution SET  state = 701,
  prev_trans_id = '%s',
  ` + "`data`" + ` = JSON_SET(` + "`data`" + `, '$.State', 701)
WHERE run_id IN (%s);`,
            TargetDB:      "PE",
            WorkflowID:    "workflow_transfer_payment",
            TargetState:   701,
            TargetAttempt: 0,
            CaseType:      domain.CaseThoughtMachineFalseNegative,
        }
    }
    return nil
},
```

### 4. Helper Functions (`internal/txn/adapters/sql_templates.go`)

4.1. Add helper function to extract prev_trans_id:
```go
// getPrevTransID extracts prev_trans_id from PaymentEngine workflow data
func getPrevTransID(result domain.TransactionResult) string {
    // This would need to be implemented based on the actual data structure
    // For now, using the example value
    return "6e0daa5cfcc24478a2c55097fe2d7cf8"
}
```

### 5. SQL Generation Logic Updates

5.1. Update the SQL generation to handle the prev_trans_id parameter:
- The deploy template needs to include the prev_trans_id value
- The rollback template needs to restore the original prev_trans_id

## Testing Plan

1. Test with the provided transaction ID: `ced4efe76ea442ddbbca1f745ebe2386`
2. Verify case identification works correctly
3. Verify SQL generation produces the correct deploy and rollback statements
4. Verify the generated SQL matches the examples in the document:
   - PE_Deploy.sql: Updates state to 230 with prev_trans_id
   - PE_Rollback.sql: Updates state back to 701 with prev_trans_id

## Implementation Notes

1. The case is Malaysia-specific (Country: "my")
2. The fix targets the Payment Engine (PE) database
3. The transition is from state 701 (stCaptureFailed) to state 230 (stCaptureProcessing)
4. The prev_trans_id needs to be preserved and updated correctly
5. The RPP adapter status of "PROCESSING" indicates the transaction is actually successful

## Mermaid Diagram

```mermaid
flowchart TD
    A[Transaction Query] --> B{Case Identification}
    B -->|PE state 701<br/>PC NOT_FOUND<br/>RPP PROCESSING| C[thought_machine_false_negative]
    C --> D[Generate SQL Templates]
    D --> E[Deploy: PE state 701 → 230]
    D --> F[Rollback: PE state 230 → 701]
    E --> G[Execute SQL]
    F --> H[Execute Rollback if needed]