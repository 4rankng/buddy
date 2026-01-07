# Requirements Document

## Introduction

This feature adds a native JIRA attachment download command to both mybuddy and sgbuddy CLI applications. Users will be able to download CSV attachments from JIRA tickets using a simple command interface, replacing the need for the external Python script.

## Glossary

- **Mybuddy**: CLI application for Malaysia operations
- **Sgbuddy**: CLI application for Singapore operations
- **JIRA_Client**: The existing JIRA client interface for API operations
- **CSV_Attachment**: A CSV file attached to a JIRA ticket
- **Ticket_ID**: A JIRA ticket identifier (e.g., TS-1234, TSE-567)
- **Download_Command**: The new CLI command for downloading attachments

## Requirements

### Requirement 1: Command Interface

**User Story:** As an operations engineer, I want to download CSV attachments from JIRA tickets using a simple CLI command, so that I can quickly access transaction data for analysis.

#### Acceptance Criteria

1. WHEN a user runs `mybuddy jira download-attachment [TICKET_ID]`, THE Download_Command SHALL download all CSV attachments from the specified ticket
2. WHEN a user runs `sgbuddy jira download-attachment [TICKET_ID]`, THE Download_Command SHALL download all CSV attachments from the specified ticket
3. WHEN the command is executed, THE Download_Command SHALL save files to the current working directory by default
4. WHEN multiple CSV files exist, THE Download_Command SHALL download all of them with their original filenames
5. WHEN a filename conflict occurs, THE Download_Command SHALL append a numeric suffix to avoid overwriting existing files

### Requirement 2: CSV File Filtering

**User Story:** As an operations engineer, I want the command to automatically filter for CSV files, so that I only download relevant transaction data files.

#### Acceptance Criteria

1. WHEN attachments are retrieved from a ticket, THE Download_Command SHALL filter for files with CSV mime type or .csv extension
2. WHEN no CSV attachments are found, THE Download_Command SHALL display an informative message and exit gracefully
3. WHEN non-CSV attachments exist, THE Download_Command SHALL ignore them and only process CSV files

### Requirement 3: Progress and Status Reporting

**User Story:** As an operations engineer, I want to see clear feedback about the download process, so that I know what files are being downloaded and if the operation succeeded.

#### Acceptance Criteria

1. WHEN the command starts, THE Download_Command SHALL display the ticket ID and number of CSV attachments found
2. WHEN downloading each file, THE Download_Command SHALL show the filename and download progress
3. WHEN a download completes successfully, THE Download_Command SHALL display a success message with the saved file path
4. WHEN a download fails, THE Download_Command SHALL display an error message and continue with remaining files
5. WHEN all downloads complete, THE Download_Command SHALL display a summary of successful and failed downloads

### Requirement 4: Error Handling

**User Story:** As an operations engineer, I want clear error messages when something goes wrong, so that I can understand and resolve issues quickly.

#### Acceptance Criteria

1. WHEN an invalid ticket ID is provided, THE Download_Command SHALL display a descriptive error message
2. WHEN JIRA authentication fails, THE Download_Command SHALL display an authentication error with configuration guidance
3. WHEN a ticket has no attachments, THE Download_Command SHALL display an informative message
4. WHEN network errors occur, THE Download_Command SHALL display a network error message and suggest retry
5. WHEN file system errors occur, THE Download_Command SHALL display a file system error with the specific issue

### Requirement 5: Integration with Existing JIRA Infrastructure

**User Story:** As a developer, I want the new command to use existing JIRA client infrastructure, so that authentication and configuration are consistent across all JIRA operations.

#### Acceptance Criteria

1. WHEN the command executes, THE Download_Command SHALL use the existing JIRA_Client interface for API operations
2. WHEN authentication is required, THE Download_Command SHALL use the same authentication mechanism as other JIRA commands
3. WHEN configuration is needed, THE Download_Command SHALL use the same environment variables and config files as existing JIRA commands
4. WHEN the command is added, THE Download_Command SHALL be integrated into the existing JIRA command structure for both applications

### Requirement 6: File Management

**User Story:** As an operations engineer, I want downloaded files to be saved with clear naming and organization, so that I can easily identify and work with the downloaded data.

#### Acceptance Criteria

1. WHEN files are downloaded, THE Download_Command SHALL preserve the original attachment filenames
2. WHEN filename conflicts occur, THE Download_Command SHALL append a numeric suffix (e.g., file_1.csv, file_2.csv)
3. WHEN downloads complete, THE Download_Command SHALL display the full path of each saved file
4. WHEN the current directory is not writable, THE Download_Command SHALL display a permissions error message
