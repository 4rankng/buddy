# Implementation Plan: Cash-in Stuck 100 Fixes

## Overview

This implementation plan converts the design into discrete coding tasks for implementing two specialized cash-in workflow fixes. The tasks build incrementally on the existing RPP template architecture, adding timezone conversion utilities and new SQL template functions.

## Tasks

- [ ] 1. Add new case constants to domain types
  - Add `CaseCashInStuck100Retry` and `CaseCashInStuck100UpdateMismatch` constants to `internal/txn/domain/types.go`
  - Update `GetCaseSummaryOrder()` function to include new cases
  - _Requirements: 1.4, 2.5_

- [ ] 2. Implement timezone conversion utilities
  - [ ] 2.1 Create timezone conversion functions in `internal/txn/adapters/sql_template_helpers.go`
    - Implement `convertUTCToGMT8(utcTimestamp string) (string, error)`
    - Implement `compareTimestampsWithTimezone(utcTimestamp, gmt8Timestamp string) (bool, error)`
    - Handle MySQL datetime format parsing and validation
    - _Requirements: 2.4, 3.1, 3.2, 3.3, 3.4_

  - [ ]* 2.2 Write property test for timezone conversion
    - **Property 5: Timezone conversion accuracy**
    - **Validates: Requirements 2.4, 3.1**

  - [ ]* 2.3 Write unit tests for timezone utilities
    - Test valid timestamp conversions
    - Test invalid timestamp format handling
    - Test edge cases (leap years, daylight saving boundaries)
    - _Requirements: 3.3, 3.4_

- [ ] 3. Implement retry fix template function
  - [ ] 3.1 Add `cashInStuck100Retry` function to `internal/txn/adapters/sql_templates_rpp_basic.go`
    - Use `getRPPWorkflowRunIDByCriteria` to find workflows with state=100 and attempts>0
    - Generate SQL template that sets `attempt=1` without timestamp modifications
    - Include proper deploy and rollback SQL templates
    - _Requirements: 1.1, 1.3, 1.4, 1.5_

  - [ ]* 3.2 Write property test for retry SQL generation
    - **Property 3: Retry SQL generation**
    - **Validates: Requirements 1.3, 1.5**

- [ ] 4. Implement update mismatch fix template function
  - [ ] 4.1 Add `cashInStuck100UpdateMismatch` function to `internal/txn/adapters/sql_templates_rpp_basic.go`
    - Use timezone conversion utilities to determine converted timestamp
    - Generate SQL template with both `attempt=1` and JSON_SET for timestamp update
    - Include proper parameter handling for converted timestamp
    - _Requirements: 2.1, 2.3, 2.5, 2.6_

  - [ ]* 4.2 Write property test for mismatch SQL generation
    - **Property 4: Mismatch SQL generation**
    - **Validates: Requirements 2.3, 2.5, 2.6**

- [ ] 5. Register new template functions
  - [ ] 5.1 Update template registration in `registerRPPBasicTemplates` function
    - Register both new case types with their respective template functions
    - Ensure proper integration with existing template system
    - _Requirements: 1.5, 2.6_

- [ ] 6. Enhance case detection logic (if needed)
  - [ ] 6.1 Review and update case detection to support new timestamp analysis
    - Extend existing detection logic to identify stuck state 100 workflows
    - Add timestamp extraction and comparison logic
    - Route to appropriate template based on timestamp analysis
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

  - [ ]* 6.2 Write property tests for case classification
    - **Property 1: Retry case classification**
    - **Property 2: Mismatch case classification**
    - **Validates: Requirements 1.1, 1.2, 2.1, 2.2**

- [ ] 7. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ]* 8. Write comprehensive integration tests
  - [ ]* 8.1 Write property test for SQL safety and targeting
    - **Property 6: SQL safety and targeting**
    - **Validates: Requirements 1.4, 4.1, 4.2**

  - [ ]* 8.2 Write property test for timestamp validation
    - **Property 7: Timestamp validation**
    - **Validates: Requirements 3.3, 3.4**

  - [ ]* 8.3 Write property test for SQL comment inclusion
    - **Property 8: SQL comment inclusion**
    - **Validates: Requirements 4.5**

  - [ ]* 8.4 Write integration tests for end-to-end workflow
    - Test complete flow from transaction result to generated SQL
    - Test both retry and mismatch scenarios
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 9. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- The implementation leverages existing RPP template patterns for consistency
