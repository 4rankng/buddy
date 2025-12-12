# Workflow Transfer Collection Integration Plan

## Overview
This plan outlines the integration of the `workflow_transfer_collection` mapping with the provided state definitions into the existing codebase.

## State Definitions to Add
The following state mappings need to be added to the `WorkflowStateMaps` in `internal/shared/txn/types.go`:

```
"workflow_transfer_collection": {
    100: "stTransferPersisted",
    101: "stProcessingPublished",
    102: "stPreTransactionLimitCheck",
    103: "stPreRiskCheck",
    210: "stTransferProcessing",
    211: "stTransferStreamPersisted",
    220: "stAuthProcessing",
    221: "stAuthStreamPersisted",
    300: "stAuthSuccess",
    230: "stCapturePrepared",
    231: "stCaptureProcessing",
    232: "stCaptureStreamPersisted",
    240: "stCancelPrepared",
    241: "stCancelProcessing",
    242: "stCancelStreamPersisted",
    250: "stResumePrepared",
    501: "stPrepareFailureHandling",
    502: "stTransactionLimitCheckFailed",
    503: "stRiskCheckError",
    504: "stRiskCheckDeny",
    505: "stFailurePublished",
    510: "stFailureNotified",
    600: "stCanceled",
    610: "stCanceledPublished",
    701: "stCaptureFailed",
    702: "stCancelFailed",
    721: "stInvestigationRequiredPublished",
    722: "stInvestigationRequiredNotified",
    800: "stValidateSuccess",
    900: "stTransferCompleted",
    901: "stCaptureCompleted",
    902: "stTransferCompletedAutoPublish",
    905: "stCompletedPublished",
    910: "stCompletedNotified",
}
```

## Implementation Steps

### 1. Update WorkflowStateMaps
- Add the new `workflow_transfer_collection` mapping to the `WorkflowStateMaps` variable in `types.go`
- Ensure proper formatting and consistency with existing mappings

### 2. Update GetWorkflowStateName Function
- Add a new case for `workflow_transfer_collection` in the switch statement
- This will enable proper state name resolution for the new workflow type

### 3. Verify Integration
- Check that `FormatWorkflowState` function works correctly with the new workflow type
- Ensure all related functions can handle the new workflow type without issues

### 4. Create Tests
- Add test cases to `classification_test.go` to validate the new workflow state mapping
- Test key states from different ranges (100s, 200s, 500s, 700s, 900s) to ensure comprehensive coverage

### 5. Validation
- Test the integration with a few sample states to ensure proper functionality
- Verify that the workflow states are correctly formatted and displayed

## Key States to Test
- `stTransferPersisted` (100) - Initial state
- `stTransferProcessing` (210) - Processing state
- `stAuthSuccess` (300) - Success state
- `stPrepareFailureHandling` (501) - Failure handling
- `stCaptureFailed` (701) - Investigation required
- `stTransferCompleted` (900) - Final completion state

## Expected Outcome
After implementation, the system will be able to:
1. Recognize and map `workflow_transfer_collection` workflow states
2. Display human-readable state names for this workflow type
3. Properly format workflow states in output
4. Handle all the defined states from 100 to 910

## Dependencies
- No external dependencies required
- Integration with existing `WorkflowStateMaps` structure
- Uses existing `GetWorkflowStateName` and `FormatWorkflowState` functions