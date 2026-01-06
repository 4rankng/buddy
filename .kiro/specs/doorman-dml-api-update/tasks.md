# Implementation Plan: Doorman DML API Update

## Overview

This implementation plan updates the existing Doorman DML client to align with the complete API specification while maintaining backward compatibility. The approach focuses on enhancing the request/response structures, improving error handling, and adding comprehensive testing.

## Tasks

- [ ] 1. Update request and response data structures
  - Enhance CreateTicketRequest struct with all required fields
  - Create comprehensive CreateTicketResponse and TicketResult structs
  - Update JSON tags to match API specification exactly
  - _Requirements: 1.1, 1.2, 1.4, 1.5, 2.1, 2.2_

- [ ]* 1.1 Write property test for request payload structure
  - **Property 1: Complete Request Payload Structure**
  - **Validates: Requirements 1.1**

- [ ]* 1.2 Write property test for default field values
  - **Property 2: Database Field Default Value**
  - **Property 3: Tool Label Consistency**
  - **Property 4: Boolean Field Constants**
  - **Validates: Requirements 1.2, 1.4, 1.5**

- [ ] 2. Enhance CreateTicket method implementation
  - Update request payload construction to include all new fields
  - Set constant values: database="", toolLabel="direct", skipWhereClause=false, skipRollbackQuery=false
  - Handle skipRollbackQueryReason as optional field
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ]* 2.1 Write unit test for JSON serialization
  - Test null field handling for skipRollbackQueryReason
  - _Requirements: 1.3_

- [ ] 3. Implement enhanced response parsing
  - Update response parsing to handle complete TicketResult structure
  - Extract all ticket metadata fields from API response
  - Maintain backward compatibility by returning ticket ID as string
  - _Requirements: 2.1, 2.2, 5.4_

- [ ]* 3.1 Write property test for response parsing
  - **Property 5: Response Parsing Completeness**
  - **Property 6: Ticket Metadata Extraction**
  - **Validates: Requirements 2.1, 2.2**

- [ ] 4. Improve error handling and validation
  - Add comprehensive input validation before API calls
  - Enhance error message extraction from API responses
  - Implement service name validation per environment
  - _Requirements: 3.1, 3.2, 2.5, 4.3_

- [ ]* 4.1 Write property test for input validation
  - **Property 9: Input Validation Completeness**
  - **Property 10: Service Name Validation**
  - **Validates: Requirements 3.1, 3.2**

- [ ]* 4.2 Write unit tests for edge case validation
  - Test empty originalQuery validation
  - Test empty rollbackQuery validation
  - Test empty note validation
  - _Requirements: 3.3, 3.4, 3.5_

- [ ] 5. Enhance HTTP client and authentication handling
  - Improve authentication error messages
  - Add proper HTTP header setting for JSON requests
  - Enhance network error handling with status codes
  - Maintain session cookies for authenticated requests
  - _Requirements: 4.1, 4.2, 4.4, 4.5_

- [ ]* 5.1 Write property tests for HTTP handling
  - **Property 11: Authentication Error Handling**
  - **Property 12: Network Error Handling**
  - **Property 13: HTTP Header Consistency**
  - **Validates: Requirements 4.1, 4.2, 4.5**

- [ ]* 5.2 Write unit test for session management
  - Test cookie handling for authenticated requests
  - _Requirements: 4.4_

- [ ] 6. Update URL construction and display logic
  - Enhance ticket URL construction to use response data
  - Maintain backward compatibility with existing URL format
  - Improve ticket information display
  - _Requirements: 2.3, 5.5_

- [ ]* 6.1 Write property test for URL construction
  - **Property 7: URL Construction Consistency**
  - **Validates: Requirements 2.3, 5.5**

- [ ] 7. Ensure backward compatibility
  - Verify CreateTicket method signature remains unchanged
  - Test CLI parameter acceptance
  - Validate all existing service names are supported
  - Ensure return format consistency
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ]* 7.1 Write unit test for interface compatibility
  - Test method signature verification
  - _Requirements: 5.1_

- [ ]* 7.2 Write property tests for backward compatibility
  - **Property 8: Response Type Handling**
  - **Property 14: Service Name Support**
  - **Property 15: Return Format Consistency**
  - **Validates: Requirements 2.4, 5.3, 5.4**

- [ ]* 7.3 Write unit test for CLI parameter handling
  - Test command-line interface parameter acceptance
  - _Requirements: 5.2_

- [ ] 8. Integration testing and validation
  - Test complete request-response cycle with mock API
  - Validate error response parsing with various error scenarios
  - Test authentication flow with session management
  - _Requirements: 2.4, 2.5, 4.1, 4.2, 4.3_

- [ ]* 8.1 Write integration tests
  - Test end-to-end ticket creation flow
  - Test error handling scenarios
  - _Requirements: 2.4, 2.5, 4.1, 4.2, 4.3_

- [ ] 9. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties with minimum 100 iterations
- Unit tests validate specific examples and edge cases
- Integration tests ensure end-to-end functionality works correctly
- Backward compatibility is maintained throughout all changes
