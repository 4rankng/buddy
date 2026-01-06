# Implementation Plan: RPP Process Registry Init Fix

## Overview

This implementation plan adds automatic detection and remediation for RPP adapter workflows of type `wf_process_registry` stuck in state 0 (`stInit`). The solution integrates with the existing SQL template system by adding a new case type and template function.

## Tasks

- [x] 1. Add new case type and update case ordering
  - Add `CaseRppProcessRegistryStuckInit` constant to domain types
  - Update `GetCaseSummaryOrder()` to include the new case
  - _Requirements: 4.1_

- [ ] 2. Implement core template function
  - [x] 2.1 Create `rppProcessRegistryStuckInit` template function
    - Implement function signature matching `TemplateFunc` type
    - Add input validation for nil RPPAdapter
    - Use `getRPPWorkflowRunIDByCriteria` helper to find qualifying workflows
    - _Requirements: 1.1, 1.2, 5.1, 5.2_

  - [ ]* 2.2 Write property test for workflow qualification
    - **Property 1: Workflow Qualification**
    - **Validates: Requirements 1.1, 1.2, 5.4, 5.5**

  - [x] 2.3 Generate deploy script template
    - Create TemplateInfo for deploy operation
    - Set TargetDB to "RPP"
    - Include SQL template with UPDATE statement for workflow_execution
    - Add parameterized run_id to WHERE clause
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [ ]* 2.4 Write property test for deploy script correctness
    - **Property 2: Deploy Script Correctness**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 4.6**

  - [x] 2.5 Generate rollback script template
    - Create TemplateInfo for rollback operation
    - Ensure consistency with deploy script targeting
    - Set attempt back to 0 in rollback SQL
    - _Requirements: 3.1, 3.2, 3.3_

  - [ ]* 2.6 Write property test for rollback script consistency
    - **Property 3: Rollback Script Consistency**
    - **Validates: Requirements 3.1, 3.2, 3.3**

- [ ] 3. Register template function
  - [x] 3.1 Add function registration in RPP basic templates
    - Update `registerRPPBasicTemplates` function
    - Map new case to template function
    - _Requirements: 4.2_

  - [ ]* 3.2 Write unit tests for template registration
    - Test that function is properly registered in templates map
    - Test that case mapping works correctly
    - _Requirements: 4.2_

- [ ] 4. Add comprehensive input validation
  - [x] 4.1 Implement validation logic
    - Check for nil RPPAdapter
    - Validate run_id is not empty or whitespace-only
    - Ensure exact matching for workflow_id and state
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

  - [ ]* 4.2 Write property test for input validation
    - **Property 7: Input Validation**
    - **Validates: Requirements 5.1, 5.2, 5.3**

- [ ] 5. Implement SQL content validation
  - [x] 5.1 Add SQL comment generation
    - Include descriptive comments in both deploy and rollback templates
    - Explain the purpose of each operation
    - _Requirements: 2.6, 3.4_

  - [ ]* 5.2 Write property test for SQL documentation
    - **Property 4: SQL Documentation**
    - **Validates: Requirements 2.6, 3.4**

  - [x] 5.3 Ensure proper DMLTicket structure
    - Return properly formatted DMLTicket with Deploy and Rollback arrays
    - Set correct CaseType field
    - _Requirements: 4.4_

  - [ ]* 5.4 Write property test for ticket structure
    - **Property 6: Ticket Structure Validation**
    - **Validates: Requirements 4.4, 4.6**

- [ ] 6. Handle multiple workflow processing
  - [x] 6.1 Implement individual workflow processing
    - Ensure each qualifying workflow is processed separately
    - Extract run_id for each workflow individually
    - _Requirements: 1.3, 1.4_

  - [ ]* 6.2 Write property test for individual processing
    - **Property 5: Individual Workflow Processing**
    - **Validates: Requirements 1.3, 1.4**

- [-] 7. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 8. Integration testing and validation
  - [ ] 8.1 Create integration test with sample data
    - Test end-to-end functionality with realistic transaction data
    - Verify SQL generation produces expected output
    - _Requirements: All requirements_

  - [ ]* 8.2 Write integration tests for SQL execution
    - Test that generated SQL can be executed safely
    - Verify rollback operations properly reverse deploy operations
    - _Requirements: All requirements_

- [ ] 9. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties across randomized inputs
- Unit tests validate specific examples and edge cases
- The implementation follows existing codebase patterns and conventions
