# Implementation Tasks: RPP Record Auto-Fill Enhancement

## Task 1: Read Existing Strategy Population Logic

**Goal**: Understand how strategies currently populate data across multiple adapters.

**Steps**:
1. Read `internal/txn/service/strategies/base_strategy.go`
2. Identify existing populatePaymentCore() method
3. Identify existing populateAdaptersFromPaymentEngine() method
4. Read `internal/txn/service/strategies/malaysia_strategy.go`
5. Understand current PopulationStrategy interface
6. Note how data flows between adapters

**Done when**:
- Strategy population patterns understood
- Reusable methods identified

---

## Task 2: Read Existing Populator Implementations

**Goal**: Understand how populators query and convert data.

**Steps**:
1. Read `internal/txn/service/population/payment_engine_populator.go`
2. Identify QueryByTransactionID and QueryByExternalID methods
3. Read `internal/txn/service/population/payment_core_populator.go`
4. Identify QueryInternal and QueryExternal methods
5. Read `internal/txn/service/population/partnerpay_populator.go`
6. Understand conversion patterns from raw results to domain types

**Done when**:
- Populator query patterns understood
- Conversion logic identified

---

## Task 3: Enhance Malaysia Strategy for RPP Auto-Fill

**Goal**: Modify MalaysiaPopulationStrategy to auto-fill all systems when RPP data is found.

**File**: `internal/txn/service/strategies/malaysia_strategy.go`

**Steps**:
1. Locate Populate method (line 32-71)
2. Find section where RPP adapter is queried (line 37-43)
3. After RPPAdapter is populated, add new logic:
   - If result.RPPAdapter != nil and result.RPPAdapter.PartnerTxID != "":
     - Call pePopulator.QueryByTransactionID(result.RPPAdapter.PartnerTxID)
     - Assign to result.PaymentEngine
     - Call pcPopulator methods using result.RPPAdapter.PartnerTxID and CreatedAt
     - Assign to result.PaymentCore
     - Iterate through result.RPPAdapter.Workflow
     - For each workflow with RunID, call partnerpayPopulator
     - Assign to result.PartnerpayEngine
4. Add logging for each system population attempt
5. Handle errors gracefully (log warning, continue)

**Done when**:
- After RPP data populated, automatically populate PE, PC, PPE
- Uses RPPAdapter.PartnerTxID for correlation
- Uses RPPAdapter.Workflow.RunID for Partnerpay
- Graceful degradation - partial data OK
- Structured logging added

---

## Task 4: Add RPP-to-PaymentEngine Population Helper

**Goal**: Create reusable method to populate PaymentEngine from RPP data.

**File**: `internal/txn/service/strategies/base_strategy.go`

**Steps**:
1. Create `populatePaymentEngineFromRPP(result *domain.TransactionResult) error` method
2. Check if result.RPPAdapter is nil, return if so
3. Extract PartnerTxID from result.RPPAdapter
4. If PartnerTxID is empty, return nil
5. Call pePopulator.QueryByTransactionID(PartnerTxID)
6. If successful, assign to result.PaymentEngine
7. If failed and EndToEndID exists, try pePopulator.QueryByExternalID()
8. Add info log for success/failure
9. Return nil (graceful degradation)

**Done when**:
- Method queries PaymentEngine using RPP PartnerTxID
- Fallback to EndToEndID if PartnerTxID lookup fails
- Assigns to result.PaymentEngine
- Logs success/failure
- Never returns error (graceful degradation)

---

## Task 5: Add RPP-to-PaymentCore Population Helper

**Goal**: Create reusable method to populate PaymentCore from RPP data.

**File**: `internal/txn/service/strategies/base_strategy.go`

**Steps**:
1. Create `populatePaymentCoreFromRPP(result *domain.TransactionResult) error` method
2. Check if result.RPPAdapter is nil or PartnerTxID is empty, return if so
3. Check if CreatedAt is empty (needed for time window), return if so
4. Extract PartnerTxID and CreatedAt from result.RPPAdapter
5. Call pcPopulator.QueryInternal(PartnerTxID, CreatedAt)
6. Convert and assign to result.PaymentCore.Internal
7. Call pcPopulator.QueryExternal(PartnerTxID, CreatedAt)
8. Convert and assign to result.PaymentCore.External
9. Add info log for success/failure with transaction counts
10. Return nil (graceful degradation)

**Done when**:
- Method queries PaymentCore using RPP PartnerTxID
- Time-windowed queries using CreatedAt
- Assigns to result.PaymentCore
- Logs success/failure with counts
- Never returns error (graceful degradation)

---

## Task 6: Add RPP-to-PartnerpayEngine Population Helper

**Goal**: Create reusable method to populate PartnerpayEngine from RPP workflow data.

**File**: `internal/txn/service/strategies/base_strategy.go`

**Steps**:
1. Create `populatePartnerpayFromRPP(result *domain.TransactionResult) error` method
2. Check if result.RPPAdapter is nil or Workflow is empty, return if so
3. Iterate through result.RPPAdapter.Workflow
4. For each workflow with non-empty RunID:
   - Call partnerpayPopulator.QueryByInputID(wf.RunID)
   - If successful, assign to result.PartnerpayEngine
   - Break after first successful result
5. Add info log for success/failure
6. Return nil (graceful degradation)

**Done when**:
- Method queries PartnerpayEngine using workflow RunIDs
- Assigns to result.PartnerpayEngine
- Logs success/failure
- Never returns error (graceful degradation)

---

## Task 7: Update Malaysia Strategy Populate Method

**Goal**: Integrate the new helper methods into Malaysia strategy.

**File**: `internal/txn/service/strategies/malaysia_strategy.go`

**Steps**:
1. Locate Populate method (line 32-71)
2. After RPP adapter query section (after line 43)
3. Add new section: "Auto-fill related systems from RPP data"
4. Call populatePaymentEngineFromRPP(result)
5. Call populatePaymentCoreFromRPP(result)
6. Call populatePartnerpayFromRPP(result)
7. Each call ignores returned error (graceful degradation)
8. Add summary log showing which systems were populated

**Done when**:
- Three helper methods called after RPP data populated
- Errors ignored (graceful degradation)
- Summary log added
- No breaking changes to existing logic

---

## Task 8: Add Structured Logging

**Goal**: Add comprehensive logging for RPP auto-fill process.

**File**: `internal/txn/service/strategies/base_strategy.go` and `malaysia_strategy.go`

**Steps**:
1. Import `log/slog` package
2. In Malaysia strategy Populate method: Add log when starting RPP auto-fill
3. In each helper method: Add outcome logs
4. Use Info level for successful population
5. Use Warn level for failed population (include error)
6. Include relevant context: IDs, counts, timestamps

**Done when**:
- Entry log shows RPP PartnerTxID and EndToEndID
- Each system population logged with outcome
- Warnings include error details
- Summary log shows final state

---

## Task 9: Write Unit Tests for Helper Methods

**Goal**: Create unit tests for the new population helper methods.

**File**: `internal/txn/service/strategies/base_strategy_test.go` (or create new)

**Steps**:
1. Create test setup with mock populators
2. Test populatePaymentEngineFromRPP - success with PartnerTxID
3. Test populatePaymentEngineFromRPP - fallback to EndToEndID
4. Test populatePaymentEngineFromRPP - both fail (graceful degradation)
5. Test populatePaymentCoreFromRPP - success with internal and external
6. Test populatePaymentCoreFromRPP - only internal found
7. Test populatePaymentCoreFromRPP - query fails (graceful degradation)
8. Test populatePartnerpayFromRPP - success with workflow run_id
9. Test populatePartnerpayFromRPP - no workflows (graceful)
10. Run tests and ensure 100% pass rate

**Done when**:
- 9 test cases implemented
- All tests pass
- Code coverage >80% for new methods

---

## Task 10: Write Integration Test for Malaysia Strategy

**Goal**: Test the complete RPP auto-fill flow in Malaysia strategy.

**File**: `internal/txn/service/strategies/malaysia_strategy_test.go`

**Steps**:
1. Create test with RPP E2E ID as input
2. Mock adapterPopulator to return RPP data with PartnerTxID
3. Mock pePopulator.QueryByTransactionID to return PE data
4. Mock pcPopulator to return PC data
5. Mock partnerpayPopulator to return PPE data
6. Call Populate(input)
7. Verify all four systems populated in result
8. Test with partial data (some systems not found)
9. Test with RPP data only (all systems fail)
10. Run tests and ensure 100% pass rate

**Done when**:
- Integration test covers full RPP auto-fill flow
- Tests success, partial fill, and RPP-only scenarios
- All tests pass

---

## Task 11: Run Existing Test Suite

**Goal**: Ensure no regressions in existing functionality.

**Steps**:
1. Run full test suite: `go test ./internal/txn/...`
2. Fix any failures caused by changes
3. Verify all existing tests still pass
4. Check for any compilation errors

**Done when**:
- All existing tests pass
- No new compilation errors
- No regressions detected

---

## Task 12: Manual Integration Testing

**Goal**: Validate end-to-end functionality with real data.

**Steps**:
1. Deploy changes to dev environment
2. Find real RPP record with PartnerTxID
3. Call Malaysia strategy Populate with RPP E2E ID
4. Verify RPPAdapter populated correctly
5. Verify PaymentEngine populated via PartnerTxID
6. Verify PaymentCore populated via PartnerTxID
7. Verify PartnerpayEngine populated via workflow RunID
8. Check logs for appropriate messages
9. Measure query performance
10. Test with RPP record that has no related data

**Done when**:
- All systems populate correctly for real RPP data
- Logs show expected info/warning messages
- Performance < 2 seconds
- Edge cases handled gracefully

---

## Task 13: Update KNOWLEDGE.md

**Goal**: Document the new RPP auto-fill workflow.

**File**: `KNOWLEDGE.md`

**Steps**:
1. Add new section: "Auto-Filling Related Systems from RPP Records"
2. Document that Malaysia strategy automatically populates all systems
3. Explain correlation logic: PartnerTxID → PE/PC, RunID → PPE
4. Provide code example showing usage
5. Add to summary section

**Done when**:
- KNOWLEDGE.md updated with RPP auto-fill workflow
- Clear documentation of correlation logic
- Usage examples provided

---

## Task 14: Code Review and Final Polish

**Goal**: Final review and cleanup.

**Steps**:
1. Review all code changes
2. Add/update code comments as needed
3. Ensure all files have proper package headers
4. Run gofmt on all modified files
5. Run golangci-lint and fix issues
6. Verify all logging is appropriate

**Done when**:
- Code reviewed and clean
- Comments added where appropriate
- Code formatted and linted
- Ready for production deployment
