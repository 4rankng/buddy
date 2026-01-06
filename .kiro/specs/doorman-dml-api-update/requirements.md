# Requirements Document

## Introduction

Update the existing doorman DML generation code to align with the new API specification and response format as demonstrated in the plan document. The current implementation needs to be enhanced to properly handle the complete request payload structure and parse the detailed response format returned by the Doorman API.

## Glossary

- **Doorman**: The database management service that handles DML (Data Manipulation Language) ticket creation and approval workflows
- **DML_Ticket**: A database change request ticket created through Doorman for production database modifications
- **API_Client**: The Go client that interfaces with the Doorman REST API
- **Request_Payload**: The JSON structure sent to the Doorman API for ticket creation
- **Response_Payload**: The JSON structure returned by the Doorman API after ticket creation

## Requirements

### Requirement 1

**User Story:** As a developer, I want to create DML tickets with the complete API payload structure, so that all ticket metadata is properly captured and processed by Doorman.

#### Acceptance Criteria

1. WHEN creating a DML ticket, THE API_Client SHALL include all required fields in the request payload: accountID, clusterName, database, schema, originalQuery, rollbackQuery, toolLabel, skipWhereClause, skipRollbackQuery, skipRollbackQueryReason, and note
2. WHEN the database field is empty, THE API_Client SHALL send an empty string value
3. WHEN skipRollbackQueryReason is null, THE API_Client SHALL omit the field or send null value
4. THE API_Client SHALL set toolLabel to "direct" to match the expected API format
5. THE API_Client SHALL always set skipWhereClause and skipRollbackQuery to boolean false

### Requirement 2

**User Story:** As a developer, I want to receive detailed ticket information from the API response, so that I can access all ticket metadata including ID, status, owners, and URLs.

#### Acceptance Criteria

1. WHEN a DML ticket is created successfully, THE API_Client SHALL parse the complete response structure including code, errors, message, result array, and requestID
2. WHEN parsing the result array, THE API_Client SHALL extract ticket metadata: id, submitter, status, owners, oncallUsers, env, accountID, accountName, clusterName, instanceName, schema, note, and pagePath
3. WHEN displaying ticket information, THE API_Client SHALL show the ticket ID and construct the proper ticket URL
4. THE API_Client SHALL handle both success responses (code 200) and error responses appropriately
5. WHEN the API returns an error, THE API_Client SHALL display meaningful error messages from the response

### Requirement 3

**User Story:** As a developer, I want the client to validate input parameters before making API calls, so that invalid requests are caught early and clear error messages are provided.

#### Acceptance Criteria

1. WHEN required parameters are missing, THE API_Client SHALL return validation errors before making API calls
2. WHEN serviceName is provided, THE API_Client SHALL validate it against supported services for the current environment
3. WHEN originalQuery is empty, THE API_Client SHALL return a validation error
4. WHEN rollbackQuery is empty, THE API_Client SHALL return a validation error
5. WHEN note is empty, THE API_Client SHALL return a validation error

### Requirement 4

**User Story:** As a developer, I want the API client to handle authentication and HTTP communication robustly, so that network issues and authentication failures are handled gracefully.

#### Acceptance Criteria

1. WHEN authentication fails, THE API_Client SHALL return clear error messages indicating authentication issues
2. WHEN network requests fail, THE API_Client SHALL return descriptive error messages including HTTP status codes
3. WHEN the API returns non-200 status codes, THE API_Client SHALL parse and display error details from the response body
4. THE API_Client SHALL maintain session cookies for authenticated requests
5. THE API_Client SHALL set appropriate HTTP headers including Content-Type for JSON requests

### Requirement 5

**User Story:** As a developer, I want the updated client to maintain backward compatibility with existing command-line interfaces, so that current workflows are not disrupted.

#### Acceptance Criteria

1. THE API_Client SHALL maintain the existing CreateTicket method signature for backward compatibility
2. WHEN called through the command-line interface, THE API_Client SHALL accept the same parameters: serviceName, originalQuery, rollbackQuery, and note
3. THE API_Client SHALL continue to support all existing service names: payment_engine, payment_core, fast_adapter, rpp_adapter, partnerpay_engine
4. THE API_Client SHALL return ticket IDs in the same string format as before
5. THE API_Client SHALL construct ticket URLs using the same format as the current implementation
