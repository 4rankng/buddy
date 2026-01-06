# Implementation Plan: RPP Cashin Workflow Fix

## Overview

This implementation plan adds automatic detection and remediation for RPP adapter workflows of type `wf_ct_cashin` stuck in state 100 (`stTransferPersisted`) with attempt 0. The solution integrates with the existing SQL template system by adding a new case type and template function that resolves optimistic lock failures through timestamp updates.

## Tasks

- [x] 1. Add new case type and update case ordering
  - Add `CaseRppCashinStuck100_0` constant to domain types
  - Update `GetCaseSummaryOrder()` to include the new case
  - _Requirements: 4.1_

- [ ] 2. Implement core template function
  - [x] 2.1 Create `rppCashinStuck100_0` template function
    - Implement function signature matching `TemplateFunc` type
    - Add input validation for nil RPPAdapter
    - Use `getRPPWorkflowRunIDByCriteria` helper to find qualifying workflows
    - Validate workflow_id='wf_ct_cashin', state='100', attempt=0
    - _Requirements: 1.1, 1.2, 5.1, 5.4, 5.5, 5.6_

  - [ ]* 2.2 Write property test for workflow qualification
    - **Property 1: Workflow Qualification**
    - **Validates: Requirements 1.1, 1.2, 5.4, 5.5, 5.6**

  - [x] 2.3 Generate deploy script template
    - Create TemplateInfo for deploy operation
    - Set TargetDB to "RPP"
    - Include SQL template with UPDATE statement for workflow_execution
    - Set attempt=1, updated_at=NOW(), and JSON data State=100
    - Add parameterized run_id to WHERE clause with workflow_id and state conditions
    - Include descriptive comments explaining optimistic lock fix
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8_

  - [ ]* 2.4 Write property test for deploy script generation
    - **Property 2: Deploy Script Generation**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 4.6**

  - [x] 2.5 Generate rollback script template
    - Create TemplateInfo for rollback operation
    - Ensure consistency with deploy script targeting
    - Set attempt back to 0 in rollback SQL
    - Include descriptive comments explaining rollback purpose
    - _Requirements: 3.1, 3.2, 3.4, 3.5_

  - [ ]* 2.6 Write property test for rollback script consistency
    - **Property 3: Rollback Script Consistency**
    - **Validates: Requirements 3.1, 3.2, 3.4**

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
    - Ensure exact matching for workflow_id='wf_ct_cashin'
    - Ensure exact matching for state='100'
    - Ensure exact matching for attempt=0
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

  - [ ]* 4.2 Write property test for input validation
    - **Property 7: Input Validation**
    - **Validates: Requirements 5.1, 5.2, 5.3**

- [ ] 5. Implement SQL content validation
  - [x] 5.1 Add SQL comment generation
    - Include descriptive comments in both deploy and rollback templates
    - Explain optimistic lock fix purpose in deploy comments
    - Explain rollback purpose in rollback comments
    - _Requirements: 2.8, 3.5_

  - [ ]* 5.2 Write property test for SQL documentation
    - **Property 4: SQL Documentation**
    - **Validates: Requirements 2.8, 3.5**

  - [x] 5.3 Ensure proper DMLTicket structure
    - Return properly formatted DMLTicket with Deploy and Rollback arrays
    - Set correct CaseType field
    - Ensure TargetDB="RPP" for all templates
    - _Requirements: 4.4, 4.6_

  - [ ]* 5.4 Write property test for ticket structure
    - **Property 6: Ticket Structure Validation**
    - **Validates: Requirements 4.4, 4.6**

- [ ] 6. Handle JSON data field updates
  - [x] 6.1 Implement JSON_SET usage for data field
    - Use JSON_SET function to update State property in data field
    - Ensure deploy script sets State=100 in JSON data
    - Ensure rollback script handles JSON data appropriately
    - _Requirements: 6.1, 6.2, 6.4_

  - [ ]* 6.2 Write property test for JSON data handling
    - **Property 8: JSON Data Handling**
    - **Validates: Requirements 6.1, 6.2, 6.4**

- [ ] 7. Handle multiple workflow processing
  - [x] 7.1 Implement individual workflow processing
    - Ensure each qualifying workflow is processed separately
    - Extract run_id for each workflow individually using helper function
    - Return ticket for first qualifying workflow found
    - _Requirements: 1.3, 1.4_

  - [ ]* 7.2 Write property test for individual processing
    - **Property 5: Individual Workflow Processing**
    - **Validates: Requirements 1.3, 1.4**

- [ ] 8. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 9. Integration testing and validation
  - [ ] 9.1 Create integration test with sample data
    - Test end-to-end functionality with realistic cashin transaction data
    - Verify SQL generation produces expected output for optimistic lock scenario
    - Test with transaction data matching the original plan example
    - _Requirements: All requirements_

  - [ ]* 9.2 Write integration tests for SQL execution
    - Test that generated SQL can be executed safely
    - Verify rollback operations properly reverse deploy operations
    - Test timestamp update behavior
    - _Requirements: All requirements_

- [ ] 10. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties across randomized inputs
- Unit tests validate specific examples and edge cases
- The implementation follows existing codebase patterns and conventions
- Focus on optimistic lock resolution through timestamp updates distinguishes this from other workflow fixes
