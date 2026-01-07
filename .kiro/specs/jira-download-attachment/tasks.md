# Implementation Plan: JIRA Download Attachment Command

## Overview

This implementation plan breaks down the development of the JIRA download attachment command into discrete, manageable tasks. The approach follows the existing codebase patterns and integrates seamlessly with the current JIRA command structure in both mybuddy and sgbuddy applications.

## Tasks

- [ ] 1. Create shared attachment download service
  - Create `internal/apps/common/jira/download_service.go` with attachment filtering and download logic
  - Implement CSV file filtering by mime type and extension
  - Add filename conflict resolution with numeric suffixes
  - _Requirements: 2.1, 2.3, 1.5, 6.2_

- [ ]* 1.1 Write property test for CSV filtering
  - **Property 2: CSV Filtering Accuracy**
  - **Validates: Requirements 2.1, 2.3**

- [ ]* 1.2 Write property test for filename conflict resolution
  - **Property 3: Filename Conflict Resolution**
  - **Validates: Requirements 1.5, 6.2**

- [ ] 2. Implement download command for mybuddy
  - Create `internal/apps/mybuddy/commands/jira_download.go`
  - Add `NewJiraDownloadAttachmentCmd` function following existing patterns
  - Integrate command into existing JIRA command structure
  - _Requirements: 1.1, 5.4_

- [ ]* 2.1 Write property test for download completeness
  - **Property 1: CSV Download Completeness**
  - **Validates: Requirements 1.1, 1.4, 6.1**

- [ ]* 2.2 Write unit tests for mybuddy command
  - Test CLI argument parsing and validation
  - Test integration with download service
  - _Requirements: 1.1_

- [ ] 3. Implement download command for sgbuddy
  - Create `internal/apps/sgbuddy/commands/jira_download.go`
  - Add `NewJiraDownloadAttachmentCmd` function following existing patterns
  - Integrate command into existing JIRA command structure
  - _Requirements: 1.2, 5.4_

- [ ]* 3.1 Write property test for sgbuddy download completeness
  - **Property 1: CSV Download Completeness** (sgbuddy variant)
  - **Validates: Requirements 1.2, 1.4, 6.1**

- [ ]* 3.2 Write unit tests for sgbuddy command
  - Test CLI argument parsing and validation
  - Test integration with download service
  - _Requirements: 1.2_

- [ ] 4. Add progress reporting and user feedback
  - Implement progress display during downloads
  - Add success/failure reporting per file
  - Add final summary with download statistics
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 6.3_

- [ ]* 4.1 Write property test for progress reporting
  - **Property 5: Progress Reporting Completeness**
  - **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 6.3**

- [ ] 5. Implement comprehensive error handling
  - Add input validation for ticket IDs
  - Add authentication error handling with guidance
  - Add network error handling with retry suggestions
  - Add file system error handling
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 6.4_

- [ ]* 5.1 Write property test for error handling
  - **Property 6: Error Handling Robustness**
  - **Validates: Requirements 4.1, 4.4, 4.5**

- [ ]* 5.2 Write unit tests for specific error cases
  - Test authentication failure scenarios
  - Test no attachments scenarios
  - Test permissions error scenarios
  - _Requirements: 4.2, 4.3, 6.4_

- [ ] 6. Add default directory behavior
  - Implement current working directory as default output location
  - Add directory creation if needed
  - Add permissions checking for output directory
  - _Requirements: 1.3_

- [ ]* 6.1 Write property test for default directory behavior
  - **Property 4: Default Directory Behavior**
  - **Validates: Requirements 1.3**

- [ ] 7. Integration and command registration
  - Update mybuddy commands.go to include new download command
  - Update sgbuddy commands.go to include new download command
  - Ensure proper dependency injection of JIRA client
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ]* 7.1 Write integration tests
  - Test end-to-end command execution
  - Test integration with existing JIRA infrastructure
  - _Requirements: 5.1, 5.2, 5.3_

- [ ] 8. Update documentation and skill registry
  - Update `.roo/skills/oncall-buddy/SKILL.md` to document new command
  - Add command usage examples and descriptions
  - Document the new native Go alternative to Python script
  - _Requirements: All requirements (documentation)_

- [ ] 9. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- The implementation follows existing codebase patterns for consistency
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests ensure proper wiring with existing infrastructure
