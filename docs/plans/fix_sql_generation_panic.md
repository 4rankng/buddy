# Fix SQL Generation Panic Bug

## Problem Description
The application panics with a nil pointer dereference when generating SQL files. The stack trace shows the panic occurs at line 114 in `processSingleTransaction` function in `internal/apps/mybuddy/commands/txn.go`.

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x2 addr=0x10 pc=0x102f81974]

goroutine 1 [running]:
buddy/internal/apps/mybuddy/commands.processSingleTransaction(0x14000198700, {0x16d19ec44, 0x20})
        /Users/frank.nguyen/Documents/buddy/internal/apps/mybuddy/commands/txn.go:114 +0x134
```

## Root Cause Analysis
The issue is in the `processSingleTransaction` function at line 114:

```go
if result.PaymentEngine.Transfers.Status == domain.NotFoundStatus || result.PartnerpayEngine.Transfers.Status == domain.NotFoundStatus || result.Error != "" {
    return
}
```

The code attempts to access `result.PaymentEngine.Transfers.Status` and `result.PartnerpayEngine.Transfers.Status` without first checking if `result.PaymentEngine` or `result.PartnerpayEngine` are nil. According to the `TransactionResult` struct definition in `domain/types.go`, these fields are pointers:

```go
type TransactionResult struct {
    InputID          string
    PaymentEngine    *PaymentEngineInfo    // Pointer type
    PartnerpayEngine *PartnerpayEngineInfo // Pointer type
    PaymentCore      *PaymentCoreInfo
    FastAdapter      *FastAdapterInfo
    RPPAdapter       *RPPAdapterInfo
    CaseType         Case
    Error            string
}
```

When either `PaymentEngine` or `PartnerpayEngine` is nil, attempting to access their `Transfers` field causes a nil pointer dereference.

## Solution Plan

### 1. Fix the Nil Pointer Dereference
Add proper nil checks before accessing the `Transfers` field:

```go
// Check if PaymentEngine or PartnerpayEngine have NotFoundStatus
paymentEngineNotFound := result.PaymentEngine != nil && result.PaymentEngine.Transfers.Status == domain.NotFoundStatus
partnerpayEngineNotFound := result.PartnerpayEngine != nil && result.PartnerpayEngine.Transfers.Status == domain.NotFoundStatus

if paymentEngineNotFound || partnerpayEngineNotFound || result.Error != "" {
    return
}
```

### 2. Additional Defensive Programming
While fixing the immediate issue, we should also add defensive checks in other parts of the code that might have similar issues:

1. In `sql_generator.go`, the `getInternalPaymentFlowRunID` function already has proper nil checks
2. In `sql_templates.go`, most templates check for nil values before accessing fields
3. We should verify that all template functions handle nil cases properly

### 3. Testing
After implementing the fix:
1. Test with various transaction IDs that might result in nil PaymentEngine or PartnerpayEngine
2. Verify that SQL generation works correctly for valid transactions
3. Ensure the error handling displays appropriate messages when transactions are not found

## Implementation Steps
1. Modify the `processSingleTransaction` function in `internal/apps/mybuddy/commands/txn.go`
2. Add nil checks before accessing PaymentEngine.Transfers and PartnerpayEngine.Transfers
3. Test the fix with sample transactions
4. Verify SQL generation works correctly after the fix

## Files to Modify
- `internal/apps/mybuddy/commands/txn.go` - Fix the nil pointer dereference at line 114

## Testing Strategy
1. Test with invalid transaction IDs that result in nil PaymentEngine
2. Test with invalid transaction IDs that result in nil PartnerpayEngine
3. Test with valid transaction IDs to ensure normal operation
4. Verify that SQL generation still works for all supported SOP cases