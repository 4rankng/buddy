# Buddy: Technical Design Document

## Overview

**Buddy** is a Go-based on-call companion CLI tool for G-Bank payment operations. It provides two region-specific binaries (`mybuddy` for Malaysia and `sgbuddy` for Singapore) that encode operational knowledge (SOPs) into executable code, enabling rapid incident response through automated transaction investigation, classification, and remediation SQL generation.

### Problem Statement

On-call engineers at G-Bank face several challenges during payment incidents:

1. **Multi-system complexity**: Transactions flow through Payment Engine, Payment Core, RPP Adapter, FAST, and PartnerPay engines - each with separate databases
2. **Knowledge fragmentation**: SOPs spread across wikis, past tickets, and tribal knowledge
3. **Manual error-prone workflows**: Querying multiple databases, classifying failure modes, and writing safe remediation SQL under time pressure
4. **slow response times**: Context switching between systems slows down mean-time-to-resolution (MTTR)

### Solution

Buddy consolidates operational knowledge into a single CLI that:
- **Orchestrates multi-database queries** across all payment systems
- **Auto-classifies failure patterns** using encoded SOP logic
- **Generates tested remediation SQL** with rollback statements
- **Manages JIRA workflows** for ticket tracking
- **Creates production DML tickets** through Doorman integration

### Target Users

- On-call engineers (primary)
- Payment operations teams
- SREs responding to payment incidents
- Backend engineers investigating transaction failures

---

## High-Level Architecture

### Dual Binary Design

Buddy uses build-time environment injection to produce two region-specific binaries from a single codebase:

```
┌─────────────────────────────────────────────────────────────┐
│                    Go Codebase                               │
│                  (shared implementation)                      │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
         ┌──────▼──────┐             ┌──────▼──────┐
         │  mybuddy    │             │  sgbuddy    │
         │  (Malaysia) │             │ (Singapore) │
         └─────────────┘             └─────────────┘
         Build-time ldflags:        Build-time ldflags:
         - JIRA_DOMAIN               - JIRA_DOMAIN
         - JIRA_USERNAME             - JIRA_USERNAME
         - DOORMAN_*                 (no Doorman)
         - env= "my"                 - env= "sg"
```

**Why this design?**
- **Regional isolation**: Prevents cross-region accidents
- **Credential security**: Credentials baked in at compile time, no runtime .env dependency
- **Subset functionality**: Singapore doesn't use RPP or Doorman, so sgbuddy excludes those commands

### Build-Time Injection

The Makefile injects environment-specific constants using Go ldflags:

```makefile
# From Makefile
build-my:
  go build -ldflags \
    "-X 'buddy/internal/buildinfo.JiraDomain=$(JIRA_DOMAIN_MY)' \
     -X 'buddy/internal/buildinfo.JiraUsername=$(JIRA_USERNAME_MY)' \
     -X 'buddy/internal/buildinfo.JiraApiKey=$(JIRA_API_KEY_MY)' \
     -X 'buddy/internal/buildinfo.DoormanUsername=$(DOORMAN_USERNAME_MY)' \
     -X 'buddy/internal/buildinfo.DoormanPassword=$(DOORMAN_PASSWORD_MY)' \
     -X 'buddy/internal/buildinfo.BuildEnvironment=my'" \
    -o bin/mybuddy cmd/mybuddy/main.go
```

This ensures:
- No runtime environment file dependencies
- Binary is self-contained for distribution
- Credentials are not exposed in process listings

### Component Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          CLI Layer (Cobra)                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │   txn    │ │   rpp    │ │ ecotxn   │ │   jira   │ │ doorman  │   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘   │
└───────┼────────────┼────────────┼────────────┼────────────┼──────────┘
        │            │            │            │            │
┌───────┴────────────┴────────────┴────────────┴────────────┴──────────┐
│                      Dependency Injection Container                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────────────────────────┐  │
│  │  Jira    │ │ Doorman  │ │   Transaction Query Service          │  │
│  │  Client  │ │  Client  │ │  ┌─────────────────────────────┐    │  │
│  └──────────┘ └──────────┘ │  │  Population Strategies      │    │  │
│                              │  └─────────────────────────────┘    │  │
│                              │  ┌─────────────────────────────┐    │  │
│                              │  │  DB Adapters (PE/PC/RPP/    │    │  │
│                              │  │              Fast/PPE)       │    │  │
│                              │  └─────────────────────────────┘    │  │
│                              └──────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
┌───────▼──────────┐     ┌─────────▼─────────┐     ┌──────────▼────────┐
│  Output Adapters │     │  SOP Repository   │     │   External Systems│
│  - SQL Generator │     │  - Case Detection │     │  - JIRA REST API  │
│  - File Writers  │◄────┤  - Classification │────►│  - Doorman API    │
│  - Terminal UI   │     │  - SQL Templates  │     │  - Databases      │
└──────────────────┘     └───────────────────┘     └───────────────────┘
```

### Design Patterns Used

| Pattern | Where Used | Purpose |
|---------|-----------|---------|
| **Strategy** | `txn/service/population_strategies.go` | Different data population for regular vs Eco transactions |
| **Adapter** | `txn/service/adapters/` | Abstract database-specific query logic |
| **Repository** | `txn/adapters/sop_repository.go` | Encapsulate SOP case classification rules |
| **Dependency Injection** | `di/container.go` | Centralized client lifecycle management |
| **Singleton** | Container clients | Shared Jira/Doorman/TransactionService instances |
| **Builder** | `txn/adapters/sql_generator.go` | Construct SQL from templates + parameters |

---

## Technical Components

### 1. Transaction Query Service

**Location**: `internal/txn/service/`

**Responsibility**: Orchestrate parallel queries across multiple payment system databases

#### Flow

```
User Input (TXN ID / E2E ID / Batch)
              │
              ▼
    ┌─────────────────┐
    │ Input Parser    │
    │ - Detect format │
    │ - Clean data    │
    └────────┬────────┘
             │
      ┌──────┴──────┐
      │             │
      ▼             ▼
Regular TXN    Eco TXN
(PE/PC/RPP/   (PPE
 FAST)         only)
      │             │
      └──────┬──────┘
             │
    ┌────────▼─────────┐
    │  DB Adapters     │
    │  - PE: transfers │
    │  - PC: internal  │
    │  - PC: external  │
    │  - RPP: response │
    │  - FAST: txn     │
    │  - PPE: workflow │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐
    │ Aggregate Result │
    │ - Transaction    │
    │   info           │
    │ - States         │
    │ - Errors         │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐
    │ SOP Repository   │
    │ - Classify case  │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐
    │ Output Adapters  │
    │ - Generate SQL   │
    │ - Write files    │
    │ - Terminal UI    │
    └─────────────────┘
```

#### Key Interface

```go
// internal/txn/ports/service.go
type TransactionQueryService interface {
    QueryTransaction(inputID string) (*domain.TransactionResult, error)
    QueryBatch(inputIDs []string) ([]*domain.TransactionResult, error)
}
```

#### Database Adapters

Each adapter encapsulates:
- Connection management
- Query construction
- Row parsing
- Error handling

**Adapters:**
- `PaymentEngineAdapter`: PE.transfers + PE.workflow_execution
- `PaymentCoreAdapter`: PC.internal_transactions + PC.external_transactions
- `RPPAdapter`: RPP.credit_transfer_response
- `FastAdapter`: FAST.cashin_transactions
- `PartnerpayEngineAdapter`: PPE.workflow_runs

---

### 2. SOP Case Classification System

**Location**: `internal/txn/adapters/sop_repository.go`

**Purpose**: Map transaction state combinations to known failure patterns with remediation protocols

#### Case Type Taxonomy

```go
// Example case types
type Case int

const (
    CaseNone Case = iota
    CasePcExternalPaymentFlow200_11        // PC stuck at 200, max attempts
    CasePcExternalPaymentFlow201_0_RPP_210 // PC 201, RPP 210 (no response)
    CasePeTransferPayment210_0             // PE payment stuck at 210
    CasePeStuckAtLimitCheck102_4           // PE stuck at limit check
    CaseRppNoResponseResumeAcsp            // RPP 210, resume when ACSP/ACTC
    CaseFastCashinStuck200                 // FAST cashin stuck at 200
    // ... 15+ more cases
)
```

#### Classification Logic

Each case has detection logic:

```go
// Example: RPP no response case
if pcInfo.State == 201 && pcInfo.Attempt == 0 &&
   rppInfo.State == 210 && rppInfo.Attempt == 0 {
    return CasePcExternalPaymentFlow201_0_RPP_210
}
```

**Detection factors:**
- Current state (200, 201, 210, 220, etc.)
- Attempt count (0, 1, 2, etc.)
- Error codes (DECLINED, TIMEOUT, etc.)
- State combinations across systems (PC 201 + RPP 210)

---

### 3. SQL Generation Engine

**Location**: `internal/txn/adapters/sql_generator.go`

**Purpose**: Generate safe, tested remediation SQL with rollback statements

#### Process

```
    ┌─────────────────┐
    │ Case Type       │
    │ (e.g., RPP 210) │
    └────────┬────────┘
             │
    ┌────────▼─────────┐
    │ SQL Template     │
    │ (parameterized)  │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐
    │ Extract Params   │
    │ - TransID        │
    │ - E2E ID         │
    │ - From State     │
    │ - To State       │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐
    │ Substitute       │
    │ - Quote strings  │
    │ - Format ints    │
    └────────┬─────────┘
             │
    ┌────────▼─────────┐
    │ SQL Statements   │
    │ ┌──────────────┐ │
    │ │ PE Deploy    │ │
    │ │ PE Rollback  │ │
    │ │ PC Deploy    │ │
    │ │ PC Rollback  │ │
    │ │ RPP Deploy   │ │
    │ │ RPP Rollback │ │
    │ └──────────────┘ │
    └─────────────────┘
```

#### Output Structure

```go
type SQLStatements struct {
    PCDeployStatements     []string
    PCRollbackStatements   []string
    PEDeployStatements     []string
    PERollbackStatements   []string
    RPPDeployStatements    []string
    RPPRollbackStatements  []string
    PPEDeployStatements    []string
    PPERollbackStatements  []string
}
```

#### Example Output

```sql
-- Payment Engine Deploy
UPDATE payment_engine.workflow_execution
SET state = 222
WHERE workflow_id = 'workflow_transfer_payment'
  AND transfer_id = 'TXN123456789'
  AND state = 210;

-- Payment Engine Rollback
UPDATE payment_engine.workflow_execution
SET state = 210
WHERE workflow_id = 'workflow_transfer_payment'
  AND transfer_id = 'TXN123456789'
  AND state = 222;
```

---

### 4. JIRA Integration

**Location**: `internal/clients/jira/`

**Features:**
- REST API client with basic auth
- Interactive ticket picker (terminal UI)
- Search by keywords
- View ticket details
- Download CSV attachments (mybuddy only)
- Browser integration for opening tickets

#### Workflow

```
mybuddy jira list
       │
       ▼
┌─────────────────┐
│ Search JQL:     │
│ assignee = currentUser() AND
│ status NOT IN (Closed, Completed)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Display Picker  │
│ 1. TS-1234 ...  │
│ 2. TS-1235 ...  │
│ 3. TS-1236 ...  │
└────────┬────────┘
         │
         ▼
    [User selects]
         │
         ▼
┌─────────────────┐
│ Show Details:   │
│ - Summary       │
│ - Status        │
│ - Description   │
│ - Attachments   │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
Download   Open in
CSV        Browser
```

#### API Client

```go
type JiraClient struct {
    domain   string
    username string
    apiKey   string
    httpClient *http.Client
}

type JiraInterface interface {
    SearchTickets(jql string) ([]Ticket, error)
    GetTicket(key string) (*Ticket, error)
    DownloadAttachment(ticketKey, filename string) ([]byte, error)
}
```

---

### 5. Doorman Integration

**Location**: `internal/clients/doorman/` (Malaysia only)

**Purpose**: Create production DML (Data Manipulation Language) tickets for database changes

#### API

```bash
mybuddy doorman create-dml \
  --service payment_engine \
  --original "UPDATE payment_engine.workflow_execution SET state = 222 WHERE ..." \
  --rollback "UPDATE payment_engine.workflow_execution SET state = 210 WHERE ..." \
  --note "Resume stuck RPP workflow for E2E ID: 20250101GXSPMY... Ref: TS-1234"
```

#### Response

```json
{
  "ticket_id": "DML-5678",
  "url": "https://doorman.gxbank.com/tickets/DML-5678",
  "status": "pending_approval"
}
```

**Integration with txn command:**
After `mybuddy txn` generates SQL, it prompts:
```
Create Doorman DML ticket? (y/n)
```
If yes, automatically creates ticket with pre-filled SQL.

---

### 6. Batch Processing

**Location**: `internal/apps/common/batch/`

**Purpose**: Process multiple transaction IDs from a file

#### Workflow

```
ids.txt
───────
TXN123456789
TXN123456790
TXN123456791

mybuddy txn ids.txt
       │
       ▼
┌─────────────────┐
│ Read & Parse    │
│ - Clean lines   │
│ - Deduplicate   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Process Loop    │
│ For each ID:    │
│   - Query DBs   │
│   - Classify    │
│   - Generate SQL│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Aggregate Output│
│ Write:          │
│ ids_results.txt │
└─────────────────┘
```

#### Output Format

```
=== Transaction 1: TXN123456789 ===
Status: Stuck at RPP 210
Case: CaseRppNoResponseResumeAcsp

Payment Engine Deploy:
UPDATE ...

RPP Adapter Deploy:
UPDATE ...

[... more transactions ...]
```

---

### 7. Dependency Injection Container

**Location**: `internal/di/container.go`

**Purpose**: Centralize client lifecycle management

#### Container Structure

```go
type Container struct {
    doormanClient doorman.DoormanInterface
    jiraClient    jira.JiraInterface
    txnService    *service.TransactionQueryService
    mu            sync.RWMutex
}

func (c *Container) InitializeForEnvironment(env string) error {
    // 1. Initialize config loader
    config.InitializeConfigLoader("config")

    // 2. Create clients for environment
    c.doormanClient = doorman.NewDoormanClient(env)
    c.jiraClient = jira.NewJiraClient(env)
    c.txnService = service.NewTransactionQueryService(env)

    return nil
}
```

#### Usage in Commands

```go
// cmd/mybuddy/main.go
container := di.NewContainer()
container.InitializeForEnvironment("my")
clientSet := container.GetClientSet()

rootCmd.AddCommand(mybuddyCmd.GetCommands(appCtx, clientSet)...)
```

---

## Domain Model

### Core Types

```go
// Transaction query result
type TransactionResult struct {
    Index            int
    InputID          string
    PaymentEngine    *PaymentEngineInfo
    PartnerpayEngine *PartnerpayEngineInfo
    PaymentCore      *PaymentCoreInfo
    FastAdapter      *FastAdapterInfo
    RPPAdapter       *RPPAdapterInfo
    CaseType         Case
    Error            string
}

// Payment engine workflow state
type WorkflowInfo struct {
    WorkflowID  string  // e.g., 'workflow_transfer_payment'
    Attempt     int     // Retry count (0, 1, 2, ...)
    State       int     // Current state (200, 210, 220, etc.)
    RunID       string  // Unique execution identifier
    PrevTransID string
    Data        string  // Full JSON data
}
```

### State Machine

Payment workflows progress through states:

```
Initial: 100
    │
    ▼
Processing: 200
    │
    ▼
Partner Processing: 210
    │
    ├─► Success: 220
    │       │
    │       ▼
    │   Completed: 400
    │
    └─► Failed: 500
```

**Common stuck states:**
- **200**: Processing timeout, needs retry
- **210**: Partner (RPP/FAST) not responding, needs resume
- **102**: Limit check failed, needs manual intervention

---

## Usage Guide

### Installation

#### From Source

```bash
git clone https://github.com/your-org/buddy.git
cd buddy

# Install dependencies
go mod download

# Build both binaries
make build

# Install to ~/bin
make deploy

# Verify
mybuddy --help
sgbuddy --help
```

#### Binary Distribution

```bash
# Copy pre-built binary
cp bin/mybuddy ~/bin/
chmod +x ~/bin/mybuddy

# Add to PATH (if ~/bin not already in PATH)
export PATH="$HOME/bin:$PATH"
```

### Environment Setup

Create `.env.my` (Malaysia):

```bash
JIRA_DOMAIN=gxbank.atlassian.net
JIRA_USERNAME=your.email@gxbank.com
JIRA_API_KEY=your_jira_api_token
DOORMAN_USERNAME=your_doorman_username
DOORMAN_PASSWORD=your_doorman_password
```

Create `.env.sg` (Singapore):

```bash
JIRA_DOMAIN=gxsbank.atlassian.net
JIRA_USERNAME=your.email@gxsbank.com
JIRA_API_KEY=your_jira_api_token
# No Doorman credentials needed
```

### Command Reference

#### mybuddy Commands

**Transaction Investigation**
```bash
# Single transaction
mybuddy txn TXN123456789

# By E2E ID
mybuddy txn 20250101GXSPMYXXXXXXXXXXXXXXXX

# Batch file
mybuddy txn ids.txt
```

**RPP Recovery** (Malaysia only)
```bash
# Resume stuck RPP workflow
mybuddy rpp resume 20250101GXSPMYXXXXXXXXXXXXXXXX

# Batch
mybuddy rpp resume e2e_ids.txt
```

**PartnerPay (Eco) Inspection**
```bash
mybuddy ecotxn <run_id>
```

**JIRA Operations**
```bash
# Interactive list
mybuddy jira list

# View ticket
mybuddy jira view TS-1234

# Search tickets
mybuddy jira search "payment" "failed"
```

**Doorman DML Creation** (Malaysia only)
```bash
mybuddy doorman create-dml \
  --service payment_engine \
  --original "UPDATE payment_engine.workflow_execution SET state = 222 WHERE transfer_id = 'TXN123'" \
  --rollback "UPDATE payment_engine.workflow_execution SET state = 210 WHERE transfer_id = 'TXN123'" \
  --note "Resume stuck workflow. Ref: TS-1234"
```

#### sgbuddy Commands

```bash
# Transaction (same as mybuddy)
sgbuddy txn TXN123456789

# Eco transactions
sgbuddy ecotxn <run_id>

# JIRA (no attachment download support)
sgbuddy jira list
sgbuddy jira view TSE-1234
sgbuddy jira search "payment"
```

### Common Workflows

#### 1. Incident Resolution Routine

**Trigger**: JIRA ticket assigned "Failed transaction TS-1234"

```bash
# Step 1: Fetch context
mybuddy jira view TS-1234

# Step 2: Extract transaction IDs from ticket description
# Example: "TXN123456789, TXN123456790"

# Step 3: Investigate
mybuddy txn TXN123456789

# Step 4: Review output
# ====================================
# Transaction: TXN123456789
# Case: CaseRppNoResponseResumeAcsp
#
# Payment Engine Deploy:
# UPDATE payment_engine.workflow_execution
# SET state = 222, attempt = 1
# WHERE transfer_id = 'TXN123456789';
#
# Payment Engine Rollback:
# UPDATE payment_engine.workflow_execution
# SET state = 210, attempt = 0
# WHERE transfer_id = 'TXN123456789';
# ====================================

# Step 5: Apply SQL manually OR create DML ticket
mybuddy doorman create-dml \
  --service payment_engine \
  --original "UPDATE payment_engine.workflow_execution SET state = 222, attempt = 1 WHERE transfer_id = 'TXN123456789'" \
  --rollback "UPDATE payment_engine.workflow_execution SET state = 210, attempt = 0 WHERE transfer_id = 'TXN123456789'" \
  --note "Resume stuck RPP workflow. Ref: TS-1234"

# Step 6: Update JIRA with resolution
```

#### 2. Batch CSV Processing

**Scenario**: Ticket has CSV attachment with 100 failed transactions

```bash
# Step 1: Download attachment
mybuddy jira list
# Select ticket → Download CSV → transactions.csv

# Step 2: Extract IDs (column 1)
awk -F',' 'NR>1 {print $1}' transactions.csv > txn_ids.txt

# Step 3: Clean and deduplicate
sort -u txn_ids.txt | grep -v '^$' | sed 's/\r$//' > clean_ids.txt

# Step 4: Batch process
mybuddy txn clean_ids.txt > results.log

# Step 5: Review results
cat results.log

# Step 6: Apply common SQL patterns
# (If all same case type, apply batch SQL)
```

#### 3. RPP Resume Workflow

**Scenario**: RPP transactions stuck at state 210, attempt 0

```bash
# Step 1: Identify stuck transactions
mybuddy rpp resume 20250101GXSPMYXXXXXXXXXXXXXXXX

# Step 2: Review classification
# ====================================
# E2E ID: 20250101GXSPMYXXXXXXXXXXXXXXXX
# Case: CaseRppNoResponseResumeAcsp
#
# RPP Adapter Deploy:
# UPDATE rpp_adapter.credit_transfer_response
# SET response_code = 'ACSP', state = 'completed'
# WHERE e2e_id = '20250101GXSPMYXXXXXXXXXXXXXXXX';
#
# RPP Adapter Rollback:
# UPDATE rpp_adapter.credit_transfer_response
# SET response_code = 'TIMEOUT', state = 'pending'
# WHERE e2e_id = '20250101GXSPMYXXXXXXXXXXXXXXXX';
# ====================================

# Step 3: Apply SQL (RPP operations are critical, always save rollback!)
# Copy SQL to safe location first
# Apply in production database
# Monitor for workflow completion
```

---

## SOP Encoding

### How SOPs Are Codified

Standard Operating Procedures (SOPs) are encoded directly into Go code:

1. **Case Detection Logic**: `internal/txn/adapters/sop_repository.go`
2. **SQL Templates**: Embedded in same file
3. **Reference Documentation**: `/docs/sops/MY_DML_SOP.md` and `SG_DML_SOP.md`

#### Example: RPP No Response Case

**Detection Logic:**
```go
// From SOP: "PC external_payment_flow stuck at 201, RPP stuck at 210"
if pcInfo != nil && rppInfo != nil {
    if pcInfo.State == 201 && pcInfo.Attempt == 0 &&
       rppInfo.State == 210 && rppInfo.Attempt == 0 {
        return CasePcExternalPaymentFlow201_0_RPP_210
    }
}
```

**SQL Template:**
```go
func (s *SOPRepository) GetSQL(caseType Case, result *TransactionResult) *SQLStatements {
    switch caseType {
    case CasePcExternalPaymentFlow201_0_RPP_210:
        return &SQLStatements{
            PCDeployStatements: []string{
                fmt.Sprintf(`UPDATE payment_core.external_transactions
                    SET state = 'completed'
                    WHERE transaction_id = '%s'`, result.PaymentCore.ExternalID),
            },
            PEDeployStatements: []string{
                fmt.Sprintf(`UPDATE payment_engine.workflow_execution
                    SET state = 222, attempt = 1
                    WHERE transfer_id = '%s'`, result.PaymentEngine.Transfers.ID),
            },
            RPPDeployStatements: []string{
                fmt.Sprintf(`UPDATE rpp_adapter.credit_transfer_response
                    SET response_code = 'ACSP'
                    WHERE e2e_id = '%s'`, result.RPPAdapter.E2EID),
            },
            // ... rollback statements
        }
    }
}
```

### Safety Mechanisms

1. **Parameterized Queries**: All user inputs are escaped/quoted
2. **Rollback Generation**: Every deploy SQL has corresponding rollback
3. **Case Validation**: Only known SOP cases generate SQL
4. **Environment Isolation**: Separate binaries prevent cross-region mistakes
5. **Review Prompts**: CLI prompts before creating DML tickets

### Adding New SOP Cases

To add a new failure pattern:

1. **Document the case** in `/docs/sops/MY_DML_SOP.md`
2. **Add case type enum** in `internal/txn/domain/types.go`
3. **Add detection logic** in `internal/txn/adapters/sop_repository.go`
4. **Add SQL templates** in same file
5. **Test with sample transaction IDs**

---

## Development Guide

### Project Structure

```
buddy/
├── cmd/                      # Binary entry points
│   ├── mybuddy/main.go      # Malaysia CLI
│   └── sgbuddy/main.go      # Singapore CLI
├── internal/
│   ├── apps/                # CLI command implementations
│   │   ├── common/          # Shared commands
│   │   ├── mybuddy/         # Malaysia commands
│   │   └── sgbuddy/         # Singapore commands
│   ├── clients/             # External service clients
│   ├── txn/                 # Transaction domain (core logic)
│   │   ├── service/         # Query orchestration
│   │   ├── adapters/        # Output adapters
│   │   ├── domain/          # Types & enums
│   │   └── ports/           # Interfaces
│   ├── di/                  # Dependency injection
│   ├── config/              # Configuration
│   └── utils/               # Utilities
├── docs/                    # Documentation
├── .roo/                    # Claude Code skill definitions
├── Makefile                 # Build automation
└── go.mod / go.sum
```

### Build Process

```bash
# Lint
make lint

# Build both
make build

# Build individual
make build-my   # Creates bin/mybuddy
make build-sg   # Creates bin/sgbuddy

# Run tests
make test

# Deploy to ~/bin
make deploy
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/txn/service/adapters

# Run with coverage
go test -cover ./...
```

### Adding New Commands

To add a new command to mybuddy:

1. Create file: `internal/apps/mybuddy/commands/newcmd.go`
2. Implement Cobra command interface:
```go
package commands

import (
    "github.com/spf13/cobra"
    "buddy/internal/apps/common"
)

func NewNewCmd(appCtx *common.Context, clientSet *di.ClientSet) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "newcmd",
        Short: "Description of new command",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Command logic
            return nil
        },
    }

    return cmd
}
```

3. Register in `internal/apps/mybuddy/commands/commands.go`:
```go
func GetCommands(appCtx *common.Context, clientSet *di.ClientSet) []*cobra.Command {
    return []*cobra.Command{
        NewNewCmd(appCtx, clientSet),
        // ... other commands
    }
}
```

---

## MCP Server Opportunity

### Why Convert to MCP?

**MCP (Model Context Protocol)** enables AI assistants to use external tools through a standardized protocol. Converting Buddy to an MCP server would provide:

#### Benefits

1. **Conversational On-Call Resolution**
   - Engineer: "Help me resolve TS-1234"
   - AI: Analyzes ticket, investigates transactions, proposes remediation, guides through execution

2. **Faster MTTR**
   - Natural language interface eliminates command memorization
   - AI can parallelize investigation steps
   - Context-aware suggestions based on historical resolutions

3. **Knowledge Transfer**
   - Junior engineers can learn by watching AI work through incidents
   - SOP explanations delivered interactively

4. **Automation Opportunities**
   - AI can batch-process similar incidents
   - Automated ticket triage and classification
   - Proactive monitoring and alert handling

### Proposed Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       Claude (AI Agent)                         │
│                  "Resolve TS-1234"                              │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      │ MCP Protocol (JSON-RPC)
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                    MCP Server (buddy-mcp)                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Tool Registry                                 │  │
│  │  - txn_investigate                                         │  │
│  │  - jira_list_tickets                                       │  │
│  │  - jira_get_ticket                                         │  │
│  │  - classify_sop_case                                       │  │
│  │  - generate_remediation_sql                                │  │
│  │  - create_dml_ticket                                       │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      │ Reuses existing Go code
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│              Buddy Core Libraries (existing)                    │
│  - Transaction Query Service                                    │
│  - SOP Repository                                               │
│  - SQL Generator                                                │
│  - JIRA/Doorman Clients                                         │
└─────────────────────────────────────────────────────────────────┘
```

### Tool Definitions

#### 1. `txn_investigate`

```json
{
  "name": "txn_investigate",
  "description": "Investigate a payment transaction by ID, E2E ID, or batch file. Returns transaction status across all systems, classifies SOP case, and generates remediation SQL.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "transaction_id": {
        "type": "string",
        "description": "Transaction ID (TXN...), E2E ID (2025...), or file path"
      },
      "region": {
        "type": "string",
        "enum": ["my", "sg"],
        "description": "Region (my=Malaysia, sg=Singapore)"
      }
    },
    "required": ["transaction_id", "region"]
  }
}
```

**Response:**
```json
{
  "transaction_id": "TXN123456789",
  "status": "stuck_at_rpp_210",
  "sop_case": "CaseRppNoResponseResumeAcsp",
  "remediation_sql": {
    "payment_engine": {
      "deploy": "UPDATE payment_engine.workflow_execution SET state = 222 WHERE transfer_id = 'TXN123456789'",
      "rollback": "UPDATE payment_engine.workflow_execution SET state = 210 WHERE transfer_id = 'TXN123456789'"
    },
    "rpp_adapter": {
      "deploy": "UPDATE rpp_adapter.credit_transfer_response SET response_code = 'ACSP' WHERE e2e_id = '2025...'",
      "rollback": "UPDATE rpp_adapter.credit_transfer_response SET response_code = 'TIMEOUT' WHERE e2e_id = '2025...'"
    }
  }
}
```

#### 2. `jira_list_tickets`

```json
{
  "name": "jira_list_tickets",
  "description": "List JIRA tickets assigned to current user, optionally filtered by status",
  "inputSchema": {
    "type": "object",
    "properties": {
      "region": {
        "type": "string",
        "enum": ["my", "sg"]
      },
      "status": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Filter by status (default: open tickets only)"
      }
    },
    "required": ["region"]
  }
}
```

#### 3. `jira_get_ticket`

```json
{
  "name": "jira_get_ticket",
  "description": "Get detailed information about a specific JIRA ticket including description, attachments, and comments",
  "inputSchema": {
    "type": "object",
    "properties": {
      "ticket_id": {
        "type": "string",
        "description": "JIRA ticket ID (e.g., TS-1234)"
      },
      "region": {
        "type": "string",
        "enum": ["my", "sg"]
      },
      "include_attachments": {
        "type": "boolean",
        "description": "Download and include attachment contents"
      }
    },
    "required": ["ticket_id", "region"]
  }
}
```

#### 4. `classify_sop_case`

```json
{
  "name": "classify_sop_case",
  "description": "Classify a transaction's SOP case type based on states across all payment systems",
  "inputSchema": {
    "type": "object",
    "properties": {
      "transaction_data": {
        "type": "object",
        "description": "Transaction data including payment engine, payment core, RPP, FAST states"
      },
      "region": {
        "type": "string",
        "enum": ["my", "sg"]
      }
    },
    "required": ["transaction_data", "region"]
  }
}
```

#### 5. `generate_remediation_sql`

```json
{
  "name": "generate_remediation_sql",
  "description": "Generate remediation SQL (deploy + rollback) for a given SOP case type and transaction",
  "inputSchema": {
    "type": "object",
    "properties": {
      "sop_case": {
        "type": "string",
        "description": "SOP case type (e.g., 'CaseRppNoResponseResumeAcsp')"
      },
      "transaction_data": {
        "type": "object",
        "description": "Transaction data with IDs, states, etc."
      },
      "region": {
        "type": "string",
        "enum": ["my", "sg"]
      }
    },
    "required": ["sop_case", "transaction_data", "region"]
  }
}
```

#### 6. `create_dml_ticket`

```json
{
  "name": "create_dml_ticket",
  "description": "Create a production DML ticket in Doorman (Malaysia only)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "service": {
        "type": "string",
        "enum": ["payment_engine", "payment_core", "rpp_adapter", "partnerpay_engine"]
      },
      "original_sql": {
        "type": "string",
        "description": "Deploy SQL statement"
      },
      "rollback_sql": {
        "type": "string",
        "description": "Rollback SQL statement"
      },
      "note": {
        "type": "string",
        "description": "Description and ticket reference"
      }
    },
    "required": ["service", "original_sql", "rollback_sql", "note"]
  }
}
```

### Implementation Considerations

#### Phase 1: Proof of Concept (2-3 weeks)

1. **MCP Server Skeleton**
   - Go library for MCP protocol (likely need to build)
   - Basic JSON-RPC handler
   - Tool registry

2. **Port 3 Core Tools**
   - `txn_investigate` (most critical)
   - `jira_list_tickets`
   - `jira_get_ticket`

3. **Integration Testing**
   - Connect to Claude via MCP client
   - Test basic workflows

#### Phase 2: Full Tool Coverage (2-3 weeks)

1. **Remaining Tools**
   - `classify_sop_case`
   - `generate_remediation_sql`
   - `create_dml_ticket`

2. **Batch Support**
   - Add `txn_investigate_batch` tool
   - Async job processing for large batches

3. **Error Handling**
   - Graceful failures
   - Helpful error messages for AI

#### Phase 3: Advanced Features (3-4 weeks)

1. **Conversational Context**
   - Session state management
   - Multi-turn workflows
   - Progress callbacks

2. **Safety Mechanisms**
   - SQL approval prompts
   - Dry-run mode
   - Rollback verification

3. **Monitoring & Observability**
   - Tool usage metrics
   - Resolution time tracking
   - Success/failure rates

### Example Conversations

#### Incident Resolution

**User**: "Help me resolve TS-1234"

**AI (via MCP)**:
1. Calls `jira_get_ticket(ticket_id="TS-1234", region="my")`
2. Extracts transaction IDs from description
3. Calls `txn_investigate(transaction_id="TXN123456789", region="my")`
4. Receives classification: `CaseRppNoResponseResumeAcsp`
5. Generates remediation SQL
6. **AI response**: "I've investigated TS-1234. Found 3 stuck transactions at RPP state 210. Here's the remediation SQL: [displays SQL]. Should I create a DML ticket for approval?"

**User**: "Yes, create DML ticket"

**AI**:
7. Calls `create_dml_ticket(...)` with generated SQL
8. Returns ticket URL: "Created DML-5678. Link: https://doorman.gxbank.com/DML-5678"

#### Batch Processing

**User**: "Process all transactions in attached CSV from TS-1234"

**AI**:
1. Calls `jira_get_ticket(..., include_attachments=true)`
2. Parses CSV attachment
3. Calls `txn_investigate` for each ID (parallel)
4. Groups results by SOP case type
5. **AI response**: "Found 100 transactions: 80 are RPP 210 timeouts, 15 are PC 200 max attempts, 5 are other cases. For the 80 RPP cases, I can generate batch SQL. Apply to all 80?"

**User**: "Apply to RPP cases only"

**AI**:
6. Generates batch SQL for 80 transactions
7. Creates DML ticket
8. Provides summary and next steps

### Technical Challenges

1. **State Management**: MCP is stateless, but some workflows need context (ticket → transactions → SQL → DML)
2. **Long-Running Operations**: Batch processing may take minutes; need async job pattern
3. **Error Recovery**: AI needs clear error messages to self-correct
4. **Security**: MCP server needs proper auth (credentials already baked in, but need MCP auth layer)
5. **Rate Limiting**: Prevent AI from flooding databases with queries

### Recommended Path Forward

1. **Start small**: Build MCP wrapper around existing `txn` command
2. **Reuse all code**: Don't rewrite business logic, just expose via MCP
3. **Parallel CLI + MCP**: Keep CLI as primary interface, MCP as enhancement
4. **Measure impact**: Track MTTR before/after, incident resolution time
5. **Iterate**: Add tools based on real on-call scenarios

---

## Conclusion

Buddy represents a sophisticated approach to operational tooling:

- **Knowledge encoding**: SOPs become executable code
- **Safety first**: Rollback SQL, environment isolation, approval gates
- **Developer experience**: Single CLI, batch processing, interactive UI
- **Extensibility**: Clear patterns for adding new SOP cases and commands

The MCP server opportunity represents the next evolution: from CLI tool to AI collaborator, enabling conversational incident resolution while maintaining all safety mechanisms and operational knowledge embedded in the system.

---

**Document Version**: 1.0
**Last Updated**: 2025-01-05
**Author**: G-Bank Payment Operations Team
