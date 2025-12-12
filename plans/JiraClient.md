# JIRA Integration Documentation

## Overview

This Rails application integrates with JIRA to manage and track payment-related issues. The integration allows for fetching JIRA tickets, parsing attachment data (especially CSV files), creating tickets, and managing the lifecycle of issues through various payment systems. The JIRA integration is primarily used by the oncall team to monitor and resolve payment processing issues.

## 1. Detailed API Specifications

### 1.1 JIRA REST API Endpoints

The system uses the following JIRA REST API v3 endpoints:

#### Authentication
- **Method**: Basic Authentication
- **Headers**:
  - `Authorization: Basic <base64(username:api_key)>`
  - `Accept: application/json`

#### Get Assigned Issues
- **Endpoint**: `GET /rest/api/3/search/jql`
- **Parameters**:
  - `jql`: JQL query string
  - `fields`: Comma-separated list of fields to return
  - `maxResults`: Maximum number of results (default: 50)
- **Example Request**:
  ```http
  GET /rest/api/3/search/jql?jql=assignee%20IN%20(%22user@example.com%22)%20AND%20project%20=%20TS%20AND%20status%20NOT%20IN%20(Done,%20Resolved,%20Closed,%20Completed)%20ORDER%20BY%20created%20ASC&fields=assignee,summary,issuetype,key,priority,status,created,duedate,customfield_10060&maxResults=50
  ```
- **Response Format**:
  ```json
  {
    "expand": "schema,names",
    "startAt": 0,
    "maxResults": 50,
    "total": 10,
    "issues": [
      {
        "id": "12345",
        "key": "TS-1234",
        "fields": {
          "assignee": {
            "displayName": "John Doe"
          },
          "summary": "Issue summary",
          "issuetype": {
            "name": "Task"
          },
          "priority": {
            "name": "Medium"
          },
          "status": {
            "name": "In Progress"
          },
          "created": "2023-01-01T10:00:00.000+0000",
          "duedate": "2023-01-15",
          "customfield_10060": "2023-01-10"
        }
      }
    ]
  }
  ```

#### Get Issue Details
- **Endpoint**: `GET /rest/api/3/issue/{issueKey}`
- **Parameters**: None
- **Response Format**: Similar to above but with additional fields including attachments

#### Get Attachment Content
- **Endpoint**: `GET {attachment.content}`
- **Headers**: Same authentication as above
- **Response**: Raw file content (CSV, text, etc.)

#### Transition Issue (Close/Update Status)
- **Endpoint**: `POST /rest/api/3/issue/{issueKey}/transitions`
- **Request Body**:
  ```json
  {
    "transition": {
      "id": "3"
    }
  }
  ```

### 1.2 Request/Response Field Mappings

#### JIRA Ticket Fields
| JIRA Field | Internal Field | Type | Description |
|------------|----------------|------|-------------|
| id | external_id | String | Internal JIRA ID |
| key | key | String | Ticket key (e.g., "TS-1234") |
| fields.summary | summary | String | Ticket title |
| fields.description | description | String | Ticket description (ADF format) |
| fields.assignee.displayName | assignee_name | String | Assigned user name |
| fields.status.name | status | String | Current status |
| fields.priority.name | priority | String | Priority level |
| fields.created | created_at | DateTime | Creation timestamp |
| fields.duedate | due_at | DateTime | Due date |
| fields.customfield_10060 | due_at | DateTime | Custom due date field |
| fields.attachment | attachments | Array | Attachment metadata |

### 1.3 HTTP Methods and Headers

#### Standard Headers
```http
Authorization: Basic <base64(username:api_key)>
Accept: application/json
Content-Type: application/json
```

#### Method Usage
- `GET`: For retrieving issues, issue details, and attachments
- `POST`: For transitioning issues (closing/updating status)

### 1.4 Error Handling Patterns

#### HTTP Status Codes
- `200 OK`: Successful request
- `401 Unauthorized`: Invalid credentials
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server-side error

#### Error Response Format
```json
{
  "errorMessages": ["Error message"],
  "errors": {
    "field": "Field-specific error"
  }
}
```

## 2. Implementation Specifics

### 2.1 Step-by-Step Algorithms

#### Login/Authentication Algorithm
1. Validate environment variables (JIRA_USERNAME, JIRA_API_KEY)
2. Create base64 encoded credentials: `base64(username:api_key)`
3. Set up HTTP client with authentication headers
4. Test connection with a simple API call
5. Return authenticated client or error

#### Get Tickets Algorithm
1. Build JQL query based on parameters:
   - Filter by assignee emails
   - Filter by project key
   - Exclude completed/closed statuses
   - Order by creation date
2. Make API request to `/rest/api/3/search/jql`
3. Parse response JSON
4. Convert each issue to `JiraApi::Ticket` objects
5. Return array of tickets or empty array on error

#### Parse CSV Algorithm
1. Download attachment content from URL
2. Handle redirects (max 5 hops)
3. Parse CSV content without headers first
4. Identify header row by matching expected field names
5. Map column indices to field names
6. Process data rows:
   - Skip empty rows
   - Skip summary rows (containing "total")
   - Extract values based on column mapping
7. Return array of structured data

#### Create Ticket Algorithm (Not Implemented)
1. Build request payload with required fields
2. Make POST request to `/rest/api/3/issue`
3. Parse response for ticket ID and key
4. Return ticket details or error

#### Close Ticket Algorithm
1. Validate reason_type parameter
2. Get current issue details to verify state
3. Build transition request with target status ID
4. Make POST request to `/rest/api/3/issue/{issueKey}/transitions`
5. Return success or error response

### 2.2 Data Transformation Logic

#### Atlassian Document Format (ADF) to Plain Text
```ruby
def extract_text_from_adf(node)
  return "" unless node.is_a?(Hash)
  
  text_content = []
  
  if node["type"] == "text"
    text_content << node["text"]
  elsif node["content"].is_a?(Array)
    node["content"].each do |child|
      text_content << extract_text_from_adf(child)
    end
  end
  
  # Add line breaks for paragraphs
  if node["type"] == "paragraph"
    text_content.join("") + "\n\n"
  else
    text_content.join("")
  end
end
```

### 2.3 CSV Parsing Logic with Field Mappings

#### Field Mapping Configuration
```ruby
fields = {
  transaction_date: {
    index: -1,
    fields: [ "date" ]
  },
  batch_id: {
    index: -1,
    fields: [ "batch id", "partner_tx_id" ]
  },
  end_to_end_id: {
    index: -1,
    fields: [ "tar02 bmid", "original_bizmsgid" ]
  },
  transaction_id: {
    index: -1,
    fields: [ "transaction id" ]
  },
  req_biz_msg_id: {
    index: -1,
    fields: [ "req_biz_msg_id" ]
  },
  internal_status: {
    index: -1,
    fields: [ "dbmy status" ]
  },
  paynet_status: {
    index: -1,
    fields: [ "column_status", "tar02 sts", "rpp_status" ]
  }
}
```

#### CSV Processing Steps
1. Parse CSV content into rows
2. Find header row by matching field names
3. Map column indices to field names
4. Process each data row:
   - Skip empty rows
   - Skip summary rows (containing "total")
   - Extract values based on column mapping
5. Return structured data array

### 2.4 JQL Query Construction Details

#### Base JQL Template
```
assignee IN ({emails}) AND project = {projectKey} AND status NOT IN (Done, Resolved, Closed, Completed) ORDER BY created ASC
```

#### Email Formatting
```ruby
emails.map { |email| "\"#{email}\"" }.join(", ")
```

#### Field Selection
```
assignee,summary,issuetype,key,priority,status,created,duedate,customfield_10060
```

## 3. Configuration Details

### 3.1 Environment Variables

#### Required Variables
```bash
# JIRA Configuration
JIRA_DOMAIN=https://your-domain.atlassian.net
JIRA_USERNAME=your_username@example.com
JIRA_API_KEY=your_api_token

# Doorman Integration
DOORMAN_USERNAME=your_doorman_username
DOORMAN_PASSWORD=your_doorman_password
```

#### Optional Variables
```bash
# Logging
LOG_LEVEL=info

# Performance
MAX_RESULTS_PER_REQUEST=50
CSV_PROCESSING_TIMEOUT=30
```

### 3.2 Settings Structure

#### config/settings.yml
```yaml
jira:
  domain: https://gxbank.atlassian.net

doorman:
  host: https://doorman.infra.prd.g-bank.app
```

### 3.3 Custom Field IDs and Purposes

| Field ID | Name | Purpose |
|----------|------|---------|
| customfield_10060 | Ticket Due Date | Custom due date field used in addition to standard due date |

## 4. Data Models

### 4.1 Complete Data Structures

#### Issue Model
```go
type Issue struct {
    ID               int       `json:"id" db:"id"`
    ExternalID       string    `json:"external_id" db:"external_id"`
    Key              string    `json:"key" db:"key"`
    IssueType        int       `json:"issue_type" db:"issue_type"`
    Summary          string    `json:"summary" db:"summary"`
    Description      string    `json:"description" db:"description"`
    AssigneeName     string    `json:"assignee_name" db:"assignee_name"`
    Status           string    `json:"status" db:"status"`
    Priority         string    `json:"priority" db:"priority"`
    AttachmentFilename *string `json:"attachment_filename" db:"attachment_filename"`
    AttachmentURL    *string   `json:"attachment_url" db:"attachment_url"`
    CreatedAt        time.Time `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
    DueAt            time.Time `json:"due_at" db:"due_at"`
    Data             string    `json:"data" db:"data"` // JSON string
}
```

#### IssueItem Model
```go
type IssueItem struct {
    ID              int       `json:"id" db:"id"`
    IssueID         int       `json:"issue_id" db:"issue_id"`
    BatchID         *string   `json:"batch_id" db:"batch_id"`
    EndToEndID      *string   `json:"end_to_end_id" db:"end_to_end_id"`
    TransactionID   *string   `json:"transaction_id" db:"transaction_id"`
    CausedBy        *string   `json:"caused_by" db:"caused_by"`
    TransactionDate *time.Time `json:"transaction_date" db:"transaction_date"`
    InternalStatus  *string   `json:"internal_status" db:"internal_status"`
    PaynetStatus    *string   `json:"paynet_status" db:"paynet_status"`
    DoormanData     string    `json:"doorman_data" db:"doorman_data"` // JSON string
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
```

#### DoormanDMLRecord Model
```go
type DoormanDMLRecord struct {
    ID           int       `json:"id" db:"id"`
    ClusterName  string    `json:"cluster_name" db:"cluster_name"`
    ContentHash  string    `json:"content_hash" db:"content_hash"`
    Service      string    `json:"service" db:"service"`
    Status       string    `json:"status" db:"status"`
    Submitter    string    `json:"submitter" db:"submitter"`
    TicketID     string    `json:"ticket_id" db:"ticket_id"`
    Original     string    `json:"original" db:"original"`
    Rollback     *string   `json:"rollback" db:"rollback"`
    Note         *string   `json:"note" db:"note"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
```

### 4.2 Field Types and Validations

#### IssueType Enum Values
```go
const (
    IssueTypeUnknown int = iota
    IssueTypeDebitAccount
    IssueTypeCreditAccount
    IssueTypeMarkFailed
    IssueTypeUnderProcessing
    IssueTypeRetryFundInFailure
)
```

#### Validation Rules
- `external_id`: Required, unique
- `key`: Required, unique
- `issue_type`: Required, must be valid enum value
- `summary`: Required, max 255 characters
- `status`: Required, max 100 characters
- `priority`: Required, max 50 characters
- `created_at`: Required
- `due_at`: Required
- `data`: Required, valid JSON

### 4.3 Database Schema Details

#### Issues Table
```sql
CREATE TABLE issues (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(255) NOT NULL UNIQUE,
    key VARCHAR(255) NOT NULL UNIQUE,
    issue_type INTEGER NOT NULL DEFAULT 0,
    summary VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    assignee_name VARCHAR(255) NOT NULL,
    status VARCHAR(100) NOT NULL,
    priority VARCHAR(50) NOT NULL,
    attachment_filename VARCHAR(255),
    attachment_url TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    due_at TIMESTAMP NOT NULL,
    data JSONB NOT NULL
);

CREATE INDEX idx_issues_external_id ON issues(external_id);
CREATE INDEX idx_issues_key ON issues(key);
CREATE INDEX idx_issues_data ON issues USING GIN(data);
```

#### IssueItems Table
```sql
CREATE TABLE issue_items (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL REFERENCES issues(id),
    batch_id VARCHAR(255),
    end_to_end_id VARCHAR(255),
    transaction_id VARCHAR(255),
    caused_by VARCHAR(255),
    transaction_date TIMESTAMP,
    internal_status VARCHAR(255),
    paynet_status VARCHAR(255),
    doorman_data JSONB DEFAULT '[]',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(issue_id, batch_id, end_to_end_id, transaction_id)
);

CREATE INDEX idx_issue_items_issue_id ON issue_items(issue_id);
```

## 5. Integration Flows

### 5.1 Detailed Sequence of Operations

#### Issue Retrieval Flow
1. User accesses oncall issues page
2. System loads engineer email configurations
3. Client makes API request to JIRA with JQL query
4. JIRA returns list of assigned issues
5. System displays issues in UI

#### Issue Processing Flow
1. User clicks on an issue to view details
2. System fetches full issue details from JIRA
3. Issue is stored/updated in local database
4. If issue has CSV attachment:
   - Download attachment content
   - Parse CSV data
   - Create IssueItem records
5. System enriches IssueItems with payment system data
6. Generate DML queries based on issue analysis
7. Store queries in IssueItem.doorman_data

#### Doorman Integration Flow
1. User clicks "Create Doorman Tickets"
2. System groups IssueItems by caused_by
3. For each group:
   - Check if DML already exists (by content hash)
   - If exists, reuse existing Doorman ticket
   - If new, create Doorman ticket via API
   - Store response in DoormanDMLRecord
4. Update IssueItems with Doorman ticket IDs
5. Redirect back to issue page with updated status

### 5.2 State Management and Transitions

#### Issue States
- `Open`: Initial state when issue is created
- `In Progress`: When issue is being worked on
- `Resolved`: When issue is resolved
- `Closed`: When issue is closed

#### Workflow States
- `thought_machine_issue`: RPP state 900 + PE state 701
- `republish_credit_transfer`: RPP state 900 + PE state 220 + PC state 201
- `resume_rpp_no_response`: RPP state 210 + PE state 220 + PC state 900
- `reject_rpp_no_response`: Same as above but for different issue types
- `resume_pe_stuck_timeout`: RPP state 900 + PE state 221 + PC state 900
- `resume_pe_stuck_231`: RPP state 900 + PE state 231 + PC state 900
- `republish_pc_internal_wf`: RPP state 200 + PE state 210 + PC state 900
- `reattempt_pc_902`: RPP state 200 + PE state 210 + PC state 902
- `republish_pc_internal_wf_pe230`: RPP state 900 + PE state 230 + PC state 900
- `pc_cashout_stuck_200_11`: PE state 210 + PC state 200 + attempt 11
- `reject_pe_210`: PE state 210 + PC state 900
- `resume_pe_stuck_220`: RPP state 900 + PE state 220 + PC state 900
- `reject_rpp_qr_stuck_0`: RPP state 0 + PE state 220 + PC state 201
- `resume_pe_220_pc_202`: RPP state 900 + PE state 220 + PC state 202
- `resume_rpp_stuck_301`: RPP state 301 + PE state 220 + PC state 201

### 5.3 Error Recovery Mechanisms

#### API Error Handling
1. Check HTTP status codes
2. Parse error responses
3. Log detailed error information
4. Return appropriate error messages to user
5. Implement retry logic for transient errors

#### CSV Processing Errors
1. Validate CSV format before processing
2. Handle malformed CSV gracefully
3. Log parsing errors with context
4. Skip problematic rows but continue processing
5. Report summary of processed vs. skipped rows

#### Doorman Integration Errors
1. Validate SQL queries before submission
2. Handle authentication failures
3. Implement content hash checking to avoid duplicates
4. Store failed attempts for retry
5. Provide clear error messages for debugging

## 6. Go-Specific Considerations

### 6.1 Suggested Go Packages/Libraries

#### HTTP Client
```go
import (
    "net/http"
    "net/url"
    "encoding/base64"
    "encoding/json"
    "time"
)

// Use standard net/http package with custom client
client := &http.Client{
    Timeout: 30 * time.Second,
}
```

#### CSV Processing
```go
import (
    "encoding/csv"
    "strings"
    "bytes"
)

// Use standard encoding/csv package
reader := csv.NewReader(strings.NewReader(csvContent))
records, err := reader.ReadAll()
```

#### Database Operations
```go
import (
    "database/sql"
    "github.com/lib/pq" // PostgreSQL driver
    "github.com/jmoiron/sqlx"
)

// Use sqlx for enhanced SQL operations
db, err := sqlx.Connect("postgres", connectionString)
```

#### JSON Handling
```go
import (
    "encoding/json"
    "github.com/google/jsonapi" // For JSON API format if needed
)
```

### 6.2 Go Struct Definitions for Key Data Models

#### JIRA Client
```go
type JiraClient struct {
    domain     string
    username   string
    apiKey     string
    httpClient *http.Client
}

func NewJiraClient(domain, username, apiKey string) *JiraClient {
    return &JiraClient{
        domain:   domain,
        username: username,
        apiKey:   apiKey,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

#### JIRA Ticket
```go
type JiraTicket struct {
    ID          string      `json:"id"`
    Key         string      `json:"key"`
    Summary     string      `json:"summary"`
    Description interface{} `json:"description"` // Can be string or ADF JSON
    Assignee    string      `json:"assignee"`
    Status      string      `json:"status"`
    Priority    string      `json:"priority"`
    CreatedAt   time.Time   `json:"created_at"`
    DueAt       *time.Time  `json:"due_at"`
    Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
    ID       string `json:"id"`
    Filename string `json:"filename"`
    MimeType string `json:"mimeType"`
    URL      string `json:"url"`
    Content  string `json:"content"`
}
```

#### Doorman Service
```go
type DoormanService struct {
    client    *http.Client
    baseURL   string
    authCookie string
}

type DoormanRequest struct {
    AccountID         string `json:"accountID"`
    ClusterName       string `json:"clusterName"`
    Schema            string `json:"schema"`
    OriginalQuery     string `json:"originalQuery"`
    RollbackQuery     string `json:"rollbackQuery"`
    ToolLabel         string `json:"toolLabel"`
    SkipWhereClause   bool   `json:"skipWhereClause"`
    SkipRollbackQuery bool   `json:"skipRollbackQuery"`
    Note              string `json:"note"`
}

type DoormanResponse struct {
    Result []struct {
        ID          string    `json:"id"`
        Status      string    `json:"status"`
        Submitter   string    `json:"submitter"`
        ClusterName string    `json:"clusterName"`
        CreatedAt   time.Time `json:"createdAt"`
        Note        string    `json:"note"`
    } `json:"result"`
}
```

### 6.3 Go-Specific Patterns for Authentication and API Calls

#### Authentication Pattern
```go
func (c *JiraClient) setAuth(req *http.Request) {
    auth := base64.StdEncoding.EncodeToString(
        []byte(fmt.Sprintf("%s:%s", c.username, c.apiKey)),
    )
    req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")
}
```

#### API Call Pattern with Context
```go
func (c *JiraClient) GetAssignedIssues(ctx context.Context, projectKey string, emails []string) ([]JiraTicket, error) {
    // Build JQL query
    emailList := make([]string, len(emails))
    for i, email := range emails {
        emailList[i] = fmt.Sprintf("\"%s\"", email)
    }
    jql := fmt.Sprintf(
        "assignee IN (%s) AND project = %s AND status NOT IN (Done, Resolved, Closed, Completed) ORDER BY created ASC",
        strings.Join(emailList, ", "),
        projectKey,
    )
    
    // Build request URL
    u, err := url.Parse(c.domain + "/rest/api/3/search/jql")
    if err != nil {
        return nil, fmt.Errorf("failed to parse URL: %w", err)
    }
    
    q := u.Query()
    q.Set("jql", jql)
    q.Set("fields", "assignee,summary,issuetype,key,priority,status,created,duedate,customfield_10060")
    q.Set("maxResults", "50")
    u.RawQuery = q.Encode()
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    c.setAuth(req)
    
    // Make request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
    }
    
    // Parse response
    var response struct {
        Issues []json.RawMessage `json:"issues"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    // Convert to JiraTicket objects
    tickets := make([]JiraTicket, len(response.Issues))
    for i, issueData := range response.Issues {
        var ticket JiraTicket
        if err := json.Unmarshal(issueData, &ticket); err != nil {
            return nil, fmt.Errorf("failed to unmarshal ticket: %w", err)
        }
        tickets[i] = ticket
    }
    
    return tickets, nil
}
```

#### Error Handling Pattern
```go
type JiraError struct {
    StatusCode int
    Message    string
    Details    map[string]interface{}
}

func (e *JiraError) Error() string {
    return fmt.Sprintf("JIRA API error %d: %s", e.StatusCode, e.Message)
}

func handleJiraError(resp *http.Response) error {
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return nil
    }
    
    var errResp struct {
        ErrorMessages []string          `json:"errorMessages"`
        Errors       map[string]string `json:"errors"`
    }
    
    json.NewDecoder(resp.Body).Decode(&errResp)
    
    message := "Unknown error"
    if len(errResp.ErrorMessages) > 0 {
        message = errResp.ErrorMessages[0]
    }
    
    details := make(map[string]interface{})
    for k, v := range errResp.Errors {
        details[k] = v
    }
    
    return &JiraError{
        StatusCode: resp.StatusCode,
        Message:    message,
        Details:    details,
    }
}
```

#### CSV Processing Pattern
```go
type CSVFieldMapping struct {
    Index  int
    Fields []string
}

type CSVProcessor struct {
    fields map[string]CSVFieldMapping
}

func NewCSVProcessor() *CSVProcessor {
    return &CSVProcessor{
        fields: map[string]CSVFieldMapping{
            "transaction_date": {
                Index:  -1,
                Fields: []string{"date"},
            },
            "batch_id": {
                Index:  -1,
                Fields: []string{"batch id", "partner_tx_id"},
            },
            "end_to_end_id": {
                Index:  -1,
                Fields: []string{"tar02 bmid", "original_bizmsgid"},
            },
            // ... other field mappings
        },
    }
}

func (p *CSVProcessor) ProcessCSV(content string) ([]map[string]string, error) {
    reader := csv.NewReader(strings.NewReader(content))
    records, err := reader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("failed to parse CSV: %w", err)
    }
    
    if len(records) == 0 {
        return nil, fmt.Errorf("empty CSV")
    }
    
    // Find header row
    headerRowIndex := -1
    for i, row := range records {
        cleanedRow := make([]string, len(row))
        for j, cell := range row {
            cleanedRow[j] = strings.ToLower(strings.TrimSpace(cell))
        }
        
        for _, mapping := range p.fields {
            for _, field := range mapping.Fields {
                for _, cell := range cleanedRow {
                    if cell == field {
                        headerRowIndex = i
                        break
                    }
                }
                if headerRowIndex >= 0 {
                    break
                }
            }
            if headerRowIndex >= 0 {
                break
            }
        }
        if headerRowIndex >= 0 {
            break
        }
    }
    
    if headerRowIndex < 0 {
        return nil, fmt.Errorf("header row not found")
    }
    
    // Map column indices
    headerRow := records[headerRowIndex]
    for fieldName, mapping := range p.fields {
        for j, header := range headerRow {
            for _, field := range mapping.Fields {
                if strings.EqualFold(strings.TrimSpace(header), field) {
                    mapping.Index = j
                    break
                }
            }
            if mapping.Index >= 0 {
                break
            }
        }
    }
    
    // Process data rows
    var results []map[string]string
    for _, row := range records[headerRowIndex+1:] {
        if isEmptyRow(row) {
            continue
        }
        
        if isSummaryRow(row) {
            continue
        }
        
        rowData := make(map[string]string)
        hasData := false
        
        for fieldName, mapping := range p.fields {
            if mapping.Index >= 0 && mapping.Index < len(row) {
                value := strings.TrimSpace(row[mapping.Index])
                if value != "" && value != "-" {
                    rowData[fieldName] = value
                    hasData = true
                }
            }
        }
        
        if hasData {
            results = append(results, rowData)
        }
    }
    
    return results, nil
}

func isEmptyRow(row []string) bool {
    for _, cell := range row {
        if strings.TrimSpace(cell) != "" {
            return false
        }
    }
    return true
}

func isSummaryRow(row []string) bool {
    if len(row) == 0 {
        return false
    }
    return strings.Contains(strings.ToLower(strings.TrimSpace(row[0])), "total")
}
```

## Configuration

Required environment variables:

```bash
JIRA_DOMAIN=https://your-domain.atlassian.net
JIRA_USERNAME=your_username@example.com
JIRA_API_KEY=your_api_token
```

Additional configuration in `config/settings.yml`:

```yaml
jira:
  domain: https://gxbank.atlassian.net
```

Doorman integration requires:

```bash
DOORMAN_USERNAME=your_doorman_username
DOORMAN_PASSWORD=your_doorman_password
```

## Data Flow

1. **Issue Retrieval**: The `Oncall::IssuesController` fetches assigned issues from JIRA
2. **Issue Storage**: Issues are stored in the local `issues` table with full JSON data
3. **Attachment Processing**: CSV attachments are downloaded and parsed into `IssueItem` records
4. **Data Enrichment**: `IssueItem` records are enriched with data from various payment systems
5. **Doorman Integration**: SQL queries are generated and sent to Doorman for database operations
6. **Status Updates**: Issue status is tracked and updated throughout the process

## Integration Points

### Doorman Integration

The system integrates with Doorman for database operations:

1. SQL queries are generated from issue data
2. Doorman tickets are created with original and rollback queries
3. Doorman responses are stored and linked back to issues

### Payment Systems Integration

The system integrates with multiple payment systems:

- **Payment Core**: Internal transaction processing
- **Payment Engine**: Transfer management
- **RPP Adapter**: Payment network adapter
- **PartnerPay Engine**: Partner payment processing

Each system provides data that enriches the issue items and helps identify root causes.

## Code Examples

### Fetching and Processing Issues

```ruby
# Initialize the client
client = JiraApi::Client.new

# Get assigned issues
emails = ["engineer@example.com"]
issues = client.get_assigned_issues("TS", emails)

# Process each issue
issues.each do |issue|
  # Store in local database
  local_issue = Issue.find_or_initialize_by(external_id: issue.id, key: issue.key)
  local_issue.update!(
    issue_type: determine_issue_type(issue.summary),
    status: issue.status,
    assignee_name: issue.assignee,
    summary: issue.summary,
    description: client.format_description(issue.description),
    priority: issue.priority,
    created_at: issue.created_at,
    due_at: issue.due_at,
    data: issue.to_json
  )

  # Process attachments
  if issue.attachment
    rows = client.get_csv_attachment(issue.attachment.content)
    rows.each do |row|
      issue_item = IssueItem.find_or_create_from_issue(local_issue, row)
      issue_item.upsert_content
    end
  end
end
```

### Creating Doorman Tickets

```ruby
# Get issue items with SQL queries
issue_items = IssueItem.where(caused_by: "payment_core")

# Group by cause and create Doorman tickets
issue_items.group_by(&:caused_by).each do |caused_by, items|
  items.each do |item|
    if item.doorman_data.present?
      item.doorman_data.each do |data|
        response = Doorman::Ticket.create(
          data["service"],
          data["original"],
          data["rollback"],
          "https://gxbank.atlassian.net/browse/#{item.issue.key}"
        )
        
        # Store the response
        if response
          result = response.dig("result", 0)
          Doorman::DmlRecord.create!(
            content_hash: Digest::SHA256.hexdigest(data.to_json),
            status: result["status"],
            submitter: result["submitter"],
            cluster_name: result["clusterName"],
            service: data["service"],
            original: data["original"],
            rollback: data["rollback"],
            ticket_id: result["id"],
            note: result["note"],
            created_at: result["createdAt"].to_time
          )
        end
      end
    end
  end
end
```

## Troubleshooting

### Common Issues and Solutions

1. **Authentication Errors**
   - Verify `JIRA_USERNAME` and `JIRA_API_KEY` are correctly set
   - Ensure the API token has the necessary permissions
   - Check if the token has expired

2. **CSV Parsing Errors**
   - Verify the CSV format matches expected headers
   - Check for special characters that might break parsing
   - Ensure the CSV is properly encoded

3. **Attachment Download Failures**
   - Verify attachment URLs are accessible
   - Check redirect handling in the attachment download method
   - Ensure proper authentication for attachment access

4. **Doorman Integration Issues**
   - Verify Doorman credentials are correctly configured
   - Check if the Doorman service is accessible
   - Ensure SQL queries are properly formatted

5. **Issue Type Detection**
   - Verify issue summaries contain expected keywords
   - Check the `get_issue_type` method for proper pattern matching
   - Consider adding new issue types if needed

### Debugging Tips

1. Enable detailed logging to track API calls and responses
2. Use the Rails console to test JIRA API interactions
3. Check the JSON data stored in the `issues` table for debugging
4. Verify the mapping between CSV columns and database fields
5. Test Doorman integration with simple SQL queries first

### Performance Considerations

1. Cache frequently accessed JIRA data
2. Limit the number of issues fetched per request
3. Process large CSV attachments in background jobs
4. Optimize database queries for issue items
5. Consider pagination for large result sets