# Requirements: RPP Adapter Auto-Fill Enhancement

## Context / Problem Statement

When users query RPP records, they currently receive only basic RPP information from the `credit_transfer` and `workflow_execution` tables. To get complete transaction context, users must manually query multiple other systems (PaymentEngine, PaymentCore, PartnerpayEngine) using correlation IDs like PartnerTxID and EndToEndID.

This manual process is:
- Time-consuming and error-prone
- Requires knowledge of cross-system correlation mappings
- Leads to incomplete transaction debugging
- Inefficient for support teams investigating payment issues

## Goals

**What success looks like:**
1. Querying an RPP record (via PartnerTxID or EndToEndID) automatically returns all related system information
2. Users get complete transaction context in a single query
3. Support teams can efficiently debug payment issues across all systems
4. No breaking changes to existing functionality

**Specific capabilities:**
- Auto-populate PaymentEngine transfer and workflow data
- Auto-populate PaymentCore internal/external transactions
- Auto-populate PartnerpayEngine charge information
- Maintain backward compatibility with existing `Query()` method

## Non-Goals (Explicitly Out of Scope)

- Modifying existing RPP `Query()` method behavior (must remain unchanged)
- Changing data structures or domain models (TransactionResult already supports all fields)
- Implementing real-time streaming or event-based updates (query-based only)
- Adding caching or performance optimizations beyond basic implementation
- Modifying FastAdapter (Singapore) - this is Malaysia-specific enhancement
- Creating new database tables or migrations

## Assumptions & Constraints

**Assumptions:**
1. Port interfaces (PaymentEnginePort, PaymentCorePort, PartnerpayEnginePort) already exist and work correctly
2. Existing query methods on ports handle time windows appropriately
3. Conversion logic from raw results to domain types exists in populators
4. RPP records contain PartnerTxID or EndToEndID that can be used for correlation
5. Workflow data contains run_id that can be used to query PartnerpayEngine

**Constraints:**
1. Must maintain backward compatibility - existing `Query()` method cannot change
2. Must use existing port interfaces - no new database clients
3. Must follow existing architectural patterns (separation of ports, adapters, strategies)
4. Must handle errors gracefully - failed system queries should not fail the overall RPP query
5. Must work within existing performance budgets (< 2 seconds for full query)

## Acceptance Criteria

**Testable requirements:**
- [ ] `QueryWithFullContext()` with PartnerTxID populates PaymentEngine, PaymentCore, PartnerpayEngine
- [ ] `QueryWithFullContext()` with EndToEndID populates PaymentEngine, PaymentCore, PartnerpayEngine
- [ ] QueryWithFullContext should be at strategy level not adapter level
- [ ] Existing `Query()` method remains unchanged (backward compatibility)
- [ ] Failed PaymentEngine query does not prevent PaymentCore/PartnerpayEngine population
- [ ] Missing PartnerTxID results in graceful degradation (no errors, partial data)
- [ ] Workflow run_id is used to query PartnerpayEngine charges
- [ ] Unit tests cover: success, partial fill, RPP-only, error handling, fallback logic
- [ ] All existing tests continue to pass
- [ ] Malaysia strategy can use new method (backward compatible)
- [ ] Structured logging added for debugging (populate success/failure for each system)

## Risks / Edge Cases / Rollback Considerations

**Risks:**
1. **Performance degradation**: Multiple sequential queries could slow down response time
   - *Mitigation*: Monitor performance, optimize later if needed (parallel queries, caching)
2. **Circular dependencies**: RPP adapter needs ports, but instantiation order might be complex
   - *Mitigation*: Use dependency injection, ensure proper initialization order
3. **Data inconsistency**: Related records might not exist or be out of sync
   - *Mitigation*: Graceful degradation, return partial data, log inconsistencies
4. **Time window edge cases**: Records might fall outside expected time windows
   - *Mitigation*: Use reasonable default windows (±1 hour PC, ±30 min PE), log when records found outside windows
5. **Multiple workflow runs**: RPP might have multiple workflows, unclear which run_id to use
   - *Mitigation*: Query all run_ids, return first successful charge, document behavior

**Edge Cases:**
1. RPP record has PartnerTxID but no EndToEndID
2. RPP record has EndToEndID but no PartnerTxID
3. RPP record has neither (shouldn't happen, but handle gracefully)
4. Related system queries time out or return errors
5. PaymentCore/PaymentEngine/PartnerpayEngine records don't exist for RPP transaction
6. Workflow data exists but has no run_id
7. Multiple workflows with same run_id

**Rollback Plan:**
1. Keep `Query()` method unchanged (always available as fallback)
2. New method can be disabled by not calling it (strategy level)
3. If critical issues found: revert RPPAdapter changes, restore old constructor, revert strategy
4. Consider feature flag for gradual rollout

## Open Questions

**None** - all clarifications resolved during planning phase.
