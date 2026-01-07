# Design Document: JIRA Download Attachment Command

## Overview

This design implements a native `download-attachment` subcommand for the existing JIRA command structure in both mybuddy and sgbuddy applications. The command will leverage the existing JIRA client infrastructure to download CSV attachments from JIRA tickets, providing a streamlined alternative to the current Python script.

The implementation follows the established patterns in the codebase, integrating seamlessly with the existing Cobra CLI framework and JIRA client interface.

## Architecture

### Command Structure
```
mybuddy jira download-attachment [TICKET_ID] [flags]
sgbuddy jira download-attachment [TICKET_ID] [flags]
```

The command will be added as a subcommand to the existing JIRA command structure in both applications, following the same pattern as `list`, `view`, and `search` commands.

### Integration Points

1. **CLI Layer**: New Cobra command integrated into existing JIRA command structure
2. **Service Layer**: New attachment download service using existing JIRA client
3. **Client Layer**: Existing JIRA client interface with attachment methods
4. **Configuration**: Existing JIRA configuration and authentication

## Components and Interfaces

### 1. Download Attachment Command

**Location**:
- `internal/apps/mybuddy/commands/jira_download.go`
- `internal/apps/sgbuddy/commands/jira_download.go`

**Responsibilities**:
- Parse command line arguments and flags
- Validate ticket ID format
- Coordinate the download process
- Display progress and results to user

**Interface**:
```go
func NewJiraDownloadAttachmentCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command
```

### 2. Attachment Download Service

**Location**: `internal/apps/common/jira/download_service.go`

**Responsibilities**:
- Filter attachments for CSV files
- Manage download process for multiple files
- Handle file naming conflicts
- Provide progress reporting

**Interface**:
```go
type AttachmentDownloadService struct {
    jiraClient jira.JiraInterface
    logger     *logging.Logger
}

type DownloadOptions struct {
    TicketID    string
    OutputDir   string
    CSVOnly     bool
}

type DownloadResult struct {
    TicketID        string
    TotalFound      int
    CSVFound        int
    Downloaded      int
    Failed          int
    DownloadedFiles []string
    Errors          []error
}

func (s *AttachmentDownloadService) DownloadAttachments(ctx context.Context, opts DownloadOptions) (*DownloadResult, error)
```

### 3. Enhanced JIRA Client Interface

The existing JIRA client interface already includes the necessary methods:
- `GetIssueDetails(ctx context.Context, issueKey string) (*JiraTicket, error)`
- `DownloadAttachment(ctx context.Context, attachment Attachment, savePath string) error`

No changes needed to the interface, but we'll add a helper method for CSV filtering.

## Data Models

### Enhanced Attachment Model

The existing `Attachment` struct will be used as-is:

```go
type Attachment struct {
    ID       string `json:"id"`
    Filename string `json:"filename"`
    MimeType string `json:"mimeType"`
    URL      string `json:"url"`
    Content  string `json:"content"`
}
```

### Download Progress Model

```go
type DownloadProgress struct {
    TicketID       string
    CurrentFile    string
    FilesProcessed int
    TotalFiles     int
    Status         string
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Let me analyze the acceptance criteria to determine testable properties.

Based on the prework analysis, I can identify several properties that can be consolidated to eliminate redundancy:

**Property Reflection:**
- Properties 1.1 and 1.2 (mybuddy/sgbuddy behavior) can be combined since both apps should behave identically
- Properties 1.5 and 6.2 (filename conflict handling) are duplicates and can be combined
- Properties 6.1 and 6.3 (filename preservation and path display) can be combined into a comprehensive file handling property
- Properties 3.2, 3.3, 3.4, 3.5 (progress reporting) can be combined into a comprehensive progress reporting property

### Property 1: CSV Download Completeness
*For any* valid JIRA ticket with CSV attachments, when the download command is executed, all CSV files should be successfully downloaded to the output directory with their original filenames preserved.
**Validates: Requirements 1.1, 1.2, 1.4, 6.1**

### Property 2: CSV Filtering Accuracy
*For any* JIRA ticket with mixed attachment types, when attachments are processed, only files with CSV mime type or .csv extension should be included in the download process.
**Validates: Requirements 2.1, 2.3**

### Property 3: Filename Conflict Resolution
*For any* download operation where target filenames already exist, the system should append numeric suffixes to prevent file overwrites while preserving the original extension.
**Validates: Requirements 1.5, 6.2**

### Property 4: Default Directory Behavior
*For any* command execution without explicit output directory specification, files should be saved to the current working directory.
**Validates: Requirements 1.3**

### Property 5: Progress Reporting Completeness
*For any* download operation, the system should display ticket information at start, progress during downloads, success/failure messages per file, and a final summary upon completion.
**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 6.3**

### Property 6: Error Handling Robustness
*For any* invalid input or system error condition, the system should display appropriate error messages and handle failures gracefully without crashing.
**Validates: Requirements 4.1, 4.4, 4.5**

## Error Handling

### Input Validation
- Ticket ID format validation (basic format checking)
- Output directory permissions checking
- JIRA client initialization verification

### Network Error Handling
- Connection timeout handling with retry suggestions
- Authentication failure detection with configuration guidance
- API rate limiting awareness

### File System Error Handling
- Directory creation for output paths
- File permission checking before write operations
- Disk space validation for large downloads

### Graceful Degradation
- Continue processing remaining files if individual downloads fail
- Provide partial success reporting
- Clean up incomplete downloads on interruption

## Testing Strategy

### Unit Testing Approach
The testing strategy will use a dual approach combining unit tests for specific scenarios and property-based tests for comprehensive coverage.

**Unit Tests Focus:**
- Specific error conditions (authentication failures, no attachments, permissions errors)
- Edge cases (empty ticket responses, malformed attachment data)
- Integration points between components
- CLI argument parsing and validation

**Property-Based Testing Focus:**
- Universal properties that hold across all valid inputs
- Comprehensive input coverage through randomization
- File handling behaviors across different scenarios
- Error handling consistency across various failure modes

### Property-Based Testing Configuration
- Minimum 100 iterations per property test
- Each property test references its design document property
- Tag format: **Feature: jira-download-attachment, Property {number}: {property_text}**
- Use Go's testing framework with a property-based testing library (e.g., gopter or rapid)

### Test Data Generation
- Generate random ticket IDs with various formats
- Create mock JIRA responses with different attachment combinations
- Simulate various file system states and permissions
- Generate network error conditions for robustness testing

The testing approach ensures both specific examples work correctly and universal properties hold across all possible inputs, providing comprehensive validation of the download functionality.
