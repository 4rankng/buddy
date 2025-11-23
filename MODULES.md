# Backend Modules Documentation

This document provides detailed information about the four backend modules implemented for the oncall application.

## Architecture Overview

The application follows a **Ports & Adapters** architecture pattern:
- **Ports**: Interface definitions in `pkg/ports/` that define module contracts
- **Adapters**: Concrete implementations in `pkg/modules/` that fulfill those contracts
- **Container**: Dependency injection in `pkg/core/container.go` that manages module lifecycle

## Module Implementations

### 1. Doorman Module
**Location**: `pkg/modules/doorman/`

#### Purpose
SQL query execution and database operations for payment systems.

#### Capabilities
- **Multi-cluster SQL execution**: Query different database clusters with connection management
- **Specialized database access**: Pre-configured access to payment-related databases:
  - Payment Engine (`sg-prd-m-payment-engine`)
  - Payment Core (`sg-prd-m-payment-core`)
  - Partner Pay Engine (`sg-prd-m-partnerpay-engine`)
  - Pairing Service (`sg-prd-m-pairing-service`)
  - Transaction Limit (`sg-prd-m-transaction-limit`)
- **Stuck transaction management**: Identify, analyze, and fix stuck transactions
- **Security**: Query validation to prevent SQL injection

#### Key Methods
```go
type DoormanPort interface {
    ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error)
    QueryPaymentEngine(query string) ([]map[string]interface{}, error)
    QueryPaymentCore(query string) ([]map[string]interface{}, error)
    GetStuckTransactions(state string, hours int) ([]map[string]interface{}, error)
    FixStuckTransaction(transactionID, fixType string) ([]map[string]interface{}, error)
    HealthCheck() error
}
```

#### Usage Example
```go
container := core.MustGetContainer()
doorman := container.Doorman()

// Query stuck transactions
stuckTxns, err := doorman.GetStuckTransactions("PENDING", 24)
if err != nil {
    log.Fatal(err)
}

// Fix a specific transaction
result, err := doorman.FixStuckTransaction("txn_12345", "retry")
```

#### Configuration
```bash
DOORMAN_BASE_URL=https://doorman.sgbank.pr
DOORMAN_USERNAME=your-username
DOORMAN_PASSWORD=your-password
DOORMAN_TIMEOUT=30s
DOORMAN_RETRY_ATTEMPTS=3
```

---

### 2. Jira Module
**Location**: `pkg/modules/jira/`

#### Purpose
Jira ticket management and SHIPRM creation for incident tracking.

#### Capabilities
- **Ticket Management**: Create, retrieve, update, and search Jira tickets
- **Advanced Search**: JQL-based ticket searching with customizable filters
- **Rich Comments**: Support for Atlassian Document Format (ADF) for rich text
- **SHIPRM Creation**: Automated SHIPRM (System High Impact Production Risk Management) ticket creation
- **Team Integration**: Get tickets assigned to specific teams or users

#### Key Methods
```go
type JiraPort interface {
    GetTicket(ticketKey string) (*JiraTicket, error)
    GetAssignedTickets(team string) ([]JiraTicket, error)
    SearchTickets(query string) ([]JiraTicket, error)
    CreateTicket(ticket *JiraTicket) (string, string, error)
    CreateSHIPRM(request *SHIPRMRequest) (string, string, error)
    AddComment(ticketKey string, comment string) error
    HealthCheck() error
}
```

#### SHIPRM Creation
The module provides specialized SHIPRM ticket creation with:
- Service impact analysis
- Risk assessment
- Implementation steps with curl commands
- Validation requirements
- Automated metadata population

#### Usage Example
```go
jira := container.Jira()

// Get assigned tickets
tickets, err := jira.GetAssignedTickets("payment-team")

// Create a SHIPRM request
shiprmReq := &SHIPRMRequest{
    Title:       "Deregister PayNow for merchant accounts",
    Description: "Automated deregistration for requested accounts",
    ChangeType:  "PAYNOW_DEREGISTRATION",
    CURL:        "curl -X POST https://api.payment.com/deregister",
    // ... other fields
}
ticketID, url, err := jira.CreateSHIPRM(shiprmReq)
```

#### Configuration
```bash
JIRA_BASE_URL=https://your-domain.atlassian.net
JIRA_EMAIL=your-email@company.com
JIRA_TOKEN=your-jira-api-token
JIRA_TIMEOUT=30s
```

---

### 3. Datadog Module
**Location**: `pkg/modules/datadog/`

#### Purpose
Log search and monitoring operations with Datadog integration.

#### Capabilities
- **Log Search**: Advanced log searching with filters and pagination
- **Log Submission**: Submit logs to Datadog for centralized monitoring
- **Log Aggregation**: Perform aggregations on log data for analytics
- **Health Monitoring**: Validate Datadog connectivity and credentials

#### Key Methods
```go
type DatadogPort interface {
    SearchLogs(params *LogSearchParams) (*LogSearchResponse, error)
    SubmitLogs(logs []LogEvent) (*LogSubmissionResponse, error)
    AggregateLogs(request *LogAggregationRequest) (*LogAggregationResponse, error)
    GetMetricQuery(query string, from, to time.Time) (*MetricResponse, error)
    HealthCheck() error
}
```

#### Search Parameters
```go
type LogSearchParams struct {
    Query    string            `json:"query"`
    From     string            `json:"from"`
    To       string            `json:"to"`
    Limit    int               `json:"limit"`
    Tags     map[string]string `json:"tags,omitempty"`
}
```

#### Usage Example
```go
datadog := container.Datadog()

// Search logs
params := &LogSearchParams{
    Query: "service:payment-engine AND status:error",
    From:  "now-1h",
    To:    "now",
    Limit: 50,
}
results, err := datadog.SearchLogs(params)

// Submit custom logs
logs := []LogEvent{{
    Message:   "Transaction processing started",
    Service:   "payment-engine",
    Severity:  "info",
    Tags:      []string{"env:production", "team:payment"},
}}
response, err := datadog.SubmitLogs(logs)
```

#### Configuration
```bash
DATADOG_BASE_URL=https://api.datadoghq.com
DATADOG_API_KEY=your-datadog-api-key
DATADOG_APP_KEY=your-datadog-app-key
DATADOG_TIMEOUT=30s
```

---

### 4. Storage Module
**Location**: `pkg/modules/storage/`

#### Purpose
Lightweight, in-memory data persistence using SQLite with JSON1 extension.

#### Architecture
- **In-memory SQLite**: High-performance storage with ACID properties
- **JSON1 Extension**: Native JSON querying capabilities
- **TTL Support**: Automatic data expiration
- **Advanced Querying**: Complex queries with metadata filtering

#### Capabilities
- **CRUD Operations**: Store, retrieve, update, delete data with arbitrary structure
- **Metadata Support**: Rich metadata for each stored item
- **TTL Management**: Time-to-live support with automatic cleanup
- **Batch Operations**: Efficient bulk operations
- **Complex Queries**: Advanced querying with JSON functions and filtering

#### Key Methods
```go
type StoragePort interface {
    Store(key string, data interface{}) error
    StoreWithTTL(key string, data interface{}, ttl time.Duration) error
    Retrieve(key string) (interface{}, error)
    Query(criteria *QueryCriteria) ([]StorageItem, error)
    StoreBatch(items map[string]interface{}) error
    RetrieveBatch(keys []string) (map[string]interface{}, error)
    Cleanup() error
    HealthCheck() error
}
```

#### Query Capabilities
```go
type QueryCriteria struct {
    Prefix         string            `json:"prefix,omitempty"`
    Tags           map[string]string `json:"tags,omitempty"`
    CreatedAfter   *time.Time        `json:"created_after,omitempty"`
    ExpiresBefore  *time.Time        `json:"expires_before,omitempty"`
    SortBy         string            `json:"sort_by,omitempty"`
    SortOrder      string            `json:"sort_order,omitempty"`
    Limit          int               `json:"limit,omitempty"`
    Offset         int               `json:"offset,omitempty"`
}
```

#### Usage Example
```go
storage := container.Storage()

// Store data with TTL
err := storage.StoreWithTTL("user-session", sessionData, 2*time.Hour)

// Store with metadata
metadata := map[string]string{
    "user_id":    "12345",
    "session_type": "premium",
}
err = storage.StoreWithMetadata("preferences", prefs, metadata)

// Complex querying
criteria := &QueryCriteria{
    Prefix: "user-",
    Tags: map[string]string{
        "status": "active",
        "tier": "premium",
    },
    CreatedAfter: time.Now().Add(-24 * time.Hour),
    SortBy:      "created_at",
    SortOrder:   "desc",
    Limit:       100,
}
results, err := storage.Query(criteria)
```

#### Configuration
```bash
STORAGE_BASE_PATH=./data
STORAGE_MAX_SIZE=104857600
STORAGE_DEFAULT_TTL=24h
STORAGE_CLEANUP_INTERVAL=1h
```

---

## Module Integration

### Dependency Injection

All modules are managed through a centralized container:

```go
// Initialize all modules
container, err := core.NewContainer()
if err != nil {
    log.Fatal(err)
}

// Access modules
doorman := container.Doorman()
jira := container.Jira()
datadog := container.Datadog()
storage := container.Storage()

// Health check all modules
health := container.HealthCheck()
```

### Configuration Management

Configuration is environment-based with validation:

```go
// Load configuration
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal(err)
}

// Validate required fields
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

### Error Handling

All modules follow consistent error handling:
- Structured error messages with context
- Retry logic for network operations
- Graceful degradation for optional features
- Health check capabilities

### Performance Considerations

- **Connection Reuse**: HTTP clients are reused across requests
- **Batch Operations**: Support for efficient bulk operations
- **TTL Cleanup**: Automatic cleanup of expired data
- **Memory Management**: SQLite in-memory with WAL mode for performance
- **Timeout Configuration**: Configurable timeouts for all network operations

---

## Security

### Authentication
- Environment variable based credential storage
- No hardcoded credentials in source code
- Secure transmission (HTTPS for all API calls)

### Data Protection
- Input validation for SQL queries
- Query parameter sanitization
- TTL-based data expiration
- Optional encryption support (storage module)

### Access Control
- Service account-based authentication
- Principle of least privilege for API tokens
- Audit logging for all operations