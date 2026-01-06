# Design Document

## Overview

This design updates the existing Doorman DML client to align with the complete API specification. The current implementation will be enhanced to support the full request payload structure and parse the detailed response format, while maintaining backward compatibility with existing interfaces.

## Architecture

The design follows the existing client architecture pattern with minimal changes to the public interface:

```
CLI Commands → DoormanInterface → DoormanClient → Doorman API
```

The main changes will be:
1. Enhanced request payload structure in `CreateTicket` method
2. Improved response parsing to extract detailed ticket metadata
3. Better error handling and validation
4. Maintained backward compatibility for existing callers

## Components and Interfaces

### Enhanced Request Structure

The `CreateTicketRequest` struct will be updated to match the complete API specification:

```go
type CreateTicketRequest struct {
    AccountID               string  `json:"accountID"`
    ClusterName             string  `json:"clusterName"`
    Database                string  `json:"database"`                // New field, empty string
    Schema                  string  `json:"schema"`
    OriginalQuery           string  `json:"originalQuery"`
    RollbackQuery           string  `json:"rollbackQuery"`
    ToolLabel               string  `json:"toolLabel"`               // Changed to "direct"
    SkipWhereClause         bool    `json:"skipWhereClause"`
    SkipRollbackQuery       bool    `json:"skipRollbackQuery"`
    SkipRollbackQueryReason *string `json:"skipRollbackQueryReason,omitempty"`
    Note                    string  `json:"note"`
}
```

### Enhanced Response Structure

The `CreateTicketResponse` struct will be expanded to capture all response metadata:

```go
type CreateTicketResponse struct {
    Code      int                    `json:"code"`
    Errors    interface{}           `json:"errors"`
    Message   interface{}           `json:"message"`
    Result    []TicketResult        `json:"result"`
    RequestID string                `json:"requestID"`
}

type TicketResult struct {
    ID                      int      `json:"id"`
    Submitter              string   `json:"submitter"`
    Status                 string   `json:"status"`
    Owners                 []string `json:"owners"`
    OncallUsers            []string `json:"oncallUsers"`
    Env                    string   `json:"env"`
    AccountID              string   `json:"accountID"`
    AccountName            string   `json:"accountName"`
    DbsManaged             bool     `json:"dbsManaged"`
    ClusterName            string   `json:"clusterName"`
    ClusterType            string   `json:"clusterType"`
    ClusterID              int      `json:"clusterID"`
    InstanceName           string   `json:"instanceName"`
    InstanceID             int      `json:"instanceID"`
    OncallGroup            string   `json:"oncallGroup"`
    TechFamily             string   `json:"techFamily"`
    PagePath               string   `json:"pagePath"`
    Note                   string   `json:"note"`
    Batch                  int      `json:"batch"`
    Schema                 string   `json:"schema"`
    EvaluateRows           int      `json:"evaluateRows"`
    AffectRows             int      `json:"affectRows"`
    Percentage             float64  `json:"percentage"`
    OriginalQuery          string   `json:"originalQuery"`
    RollbackQuery          string   `json:"rollbackQuery"`
    SubQuery               string   `json:"subQuery"`
    SubMinID               int      `json:"subMinID"`
    SubMaxID               int      `json:"subMaxID"`
    Encrypted              bool     `json:"encrypted"`
    Pattern                string   `json:"pattern"`
    EoApprover             string   `json:"eoApprover"`
    DbaApprover            string   `json:"dbaApprover"`
    ToolLabel              string   `json:"toolLabel"`
    Database               string   `json:"database"`
    FileDir                string   `json:"fileDir"`
    FileType               string   `json:"fileType"`
    FileSize               int      `json:"fileSize"`
    PauseLabel             int      `json:"pauseLabel"`
    WarningMsg             string   `json:"warningMsg"`
    Remark                 string   `json:"remark"`
    PeakTime               *string  `json:"peakTime"`
    Archived               bool     `json:"archived"`
    CreatedAt              string   `json:"createdAt"`
    SkipWhereClause        bool     `json:"skipWhereClause"`
    SkipRollbackQuery      bool     `json:"skipRollbackQuery"`
    SkipRollbackQueryReason string  `json:"skipRollbackQueryReason"`
}
```

### Interface Compatibility

The `DoormanInterface` will remain unchanged to maintain backward compatibility:

```go
type DoormanInterface interface {
    // Existing methods remain the same
    CreateTicket(serviceName, originalQuery, rollbackQuery, note string) (string, error)
    // ... other methods
}
```

## Data Models

### Input Validation

The client will implement comprehensive input validation:

1. **Service Name Validation**: Check against environment-specific supported services
2. **Query Validation**: Ensure originalQuery and rollbackQuery are non-empty
3. **Note Validation**: Ensure note field is provided
4. **Environment Validation**: Validate service availability per environment

### Error Handling Model

Enhanced error handling will provide detailed context:

```go
type DoormanError struct {
    Type    string // "validation", "authentication", "api", "network"
    Message string
    Details interface{} // Additional error context
}
```

## Error Handling

### Validation Errors
- Pre-request validation for required fields
- Service availability checks per environment
- Clear error messages for missing or invalid parameters

### Authentication Errors
- Detailed authentication failure messages
- Session management for cookie-based auth
- Retry logic for authentication failures

### API Errors
- Parse error responses from API
- Extract meaningful error messages from response body
- Handle different HTTP status codes appropriately

### Network Errors
- Timeout handling with configurable timeouts
- Connection error handling
- Retry logic for transient failures

## Testing Strategy

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property-Based Testing

The testing approach will use property-based testing to validate universal behaviors across different inputs, combined with unit tests for specific examples and edge cases.

## Correctness Properties

Based on the prework analysis, the following properties will be validated through property-based testing:

### Property 1: Complete Request Payload Structure
*For any* valid DML ticket creation request, the generated request payload should contain all required fields: accountID, clusterName, database, schema, originalQuery, rollbackQuery, toolLabel, skipWhereClause, skipRollbackQuery, skipRollbackQueryReason, and note
**Validates: Requirements 1.1**

### Property 2: Database Field Default Value
*For any* DML ticket creation request, the database field should always be set to an empty string
**Validates: Requirements 1.2**

### Property 3: Tool Label Consistency
*For any* DML ticket creation request, the toolLabel field should always be set to "direct"
**Validates: Requirements 1.4**

### Property 4: Boolean Field Constants
*For any* DML ticket creation request, skipWhereClause and skipRollbackQuery should always be set to false
**Validates: Requirements 1.5**

### Property 5: Response Parsing Completeness
*For any* successful API response, the client should parse all response fields: code, errors, message, result array, and requestID
**Validates: Requirements 2.1**

### Property 6: Ticket Metadata Extraction
*For any* response with a result array, the client should extract all ticket metadata fields including id, submitter, status, owners, and other specified fields
**Validates: Requirements 2.2**

### Property 7: URL Construction Consistency
*For any* ticket ID, the constructed ticket URL should follow the expected format with the correct base URL and ticket ID
**Validates: Requirements 2.3, 5.5**

### Property 8: Response Type Handling
*For any* API response, the client should handle both success (code 200) and error responses appropriately based on the response code
**Validates: Requirements 2.4**

### Property 9: Input Validation Completeness
*For any* request with missing required parameters, the client should return validation errors before making API calls
**Validates: Requirements 3.1**

### Property 10: Service Name Validation
*For any* service name, the client should validate it against the list of supported services for the current environment
**Validates: Requirements 3.2**

### Property 11: Authentication Error Handling
*For any* authentication failure, the client should return clear error messages indicating authentication issues
**Validates: Requirements 4.1**

### Property 12: Network Error Handling
*For any* network request failure, the client should return descriptive error messages including HTTP status codes
**Validates: Requirements 4.2**

### Property 13: HTTP Header Consistency
*For any* JSON request, the client should set appropriate HTTP headers including Content-Type
**Validates: Requirements 4.5**

### Property 14: Service Name Support
*For any* of the existing service names (payment_engine, payment_core, fast_adapter, rpp_adapter, partnerpay_engine), the client should continue to support them
**Validates: Requirements 5.3**

### Property 15: Return Format Consistency
*For any* successful ticket creation, the client should return the ticket ID as a string in the same format as before
**Validates: Requirements 5.4**

### Unit Test Coverage

Unit tests will complement property-based tests by covering:

- **Edge Cases**: Empty query validation (Requirements 3.3, 3.4, 3.5)
- **Interface Compatibility**: Method signature verification (Requirements 5.1)
- **CLI Parameter Handling**: Command-line interface parameter acceptance (Requirements 5.2)
- **Error Response Parsing**: Specific error message extraction (Requirements 2.5, 4.3)
- **Session Management**: Cookie handling for authenticated requests (Requirements 4.4)
- **JSON Serialization**: Null field handling for skipRollbackQueryReason (Requirements 1.3)

### Testing Framework

The implementation will use:
- **Go's testing package** for unit tests
- **Property-based testing library** (such as `gopter` or `quick`) for property tests
- **HTTP mocking** (such as `httptest`) for API interaction testing
- **Minimum 100 iterations** per property test to ensure comprehensive coverage

Each property test will be tagged with a comment referencing its design document property:
```go
// Feature: doorman-dml-api-update, Property 1: Complete Request Payload Structure
```
