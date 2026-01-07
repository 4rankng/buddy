---
name: buddy-oncall-assistant
description: CLI tools (mybuddy/sgbuddy) for database queries, transaction investigation, Jira integration, and payment operations.
---

NOTE: When you are given a ticket ID (eg TS-4558 or TSE-1113) your job is to find a list of Batch ID or TAR02 BMID then save to a file
eg TS-4558.txt

if TS ticket you call mybuddy txn TS-4558.txt, it will generate solution for the stuck txn
if TSE ticket you call sgbuddy txn TSE-1113.txt, it will generate solution for the stuck txn

if the transaction is Grab txn you should call mybuddy ecotxn or sgbuddy ecotxn



# MyBuddy & SGBuddy CLI Tools

## Tool Selection

| Region | Tool | Env File | Jira Project |
|:---|:---|:---|:---|
| Malaysia | `mybuddy` | `.env.my` | `TS` |
| Singapore | `sgbuddy` | `.env.sg` | `TSE` |


## Commands

### Jira Integration

```bash
# List tickets (interactive)
mybuddy jira list
sgbuddy jira list

# Search tickets
mybuddy jira search "keyword"

# View ticket
mybuddy jira view TS-1234
sgbuddy jira view TSE-5678

# Download CSV attachments (NEW - Native Go Implementation)
mybuddy jira download-attachment TS-1234
sgbuddy jira download-attachment TSE-5678

# Download to specific directory
mybuddy jira download-attachment TS-1234 --output ./downloads
sgbuddy jira download-attachment TSE-5678 -o ./data
```

### Transaction Investigation

```bash
# Single transaction (accepts TXN ID, E2E ID, or Batch ID)
mybuddy txn TXN123
mybuddy txn 20250101120000

# Batch processing (file with one ID per line)
mybuddy txn ids.txt
sgbuddy txn ids.txt
```

### RPP Operations (Malaysia Only)

```bash
mybuddy rpp resume
```

Output: Deploy SQL + Rollback SQL

### PartnerPay Inspection

```bash
mybuddy ecotxn <run_id>
sgbuddy ecotxn <run_id>
```

### Database Queries

#### Query (Read-Only)

```bash
mybuddy doorman query --service <service> --query "<sql>"
sgbuddy doorman query --service <service> --query "<sql>"
```

**Services:**
- MyBuddy: `payment_engine`, `payment_core`, `rpp_adapter`, `partnerpay_engine`
- SGBuddy: `payment_engine`, `payment_core`, `fast_adapter`, `partnerpay_engine`

**Flags:**
- `--service, -s`: Service name (required)
- `--query, -q`: SQL query (required)
- `--format, -f`: `table` (default) or `json`

**Examples:**
```bash
mybuddy doorman query -s payment_core -q "SELECT * FROM transactions WHERE id = 'TXN123'"
sgbuddy doorman query -s fast_adapter -q "SELECT * FROM orders WHERE status = 'pending'" --format json
```

#### Create DML Ticket (Malaysia Only)

```bash
mybuddy doorman create-dml \
  --service payment_core \
  --original "UPDATE transactions SET status = 'done' WHERE id = 'TXN123'" \
  --rollback "UPDATE transactions SET status = 'pending' WHERE id = 'TXN123'" \
  --note "Fix TXN123 - Ref TS-456"
```

## CSV Processing

```bash
# Download attachment (Native Go - Recommended)
mybuddy jira download-attachment TS-1234
sgbuddy jira download-attachment TSE-5678

# Download attachment (Legacy Python Script)
python .roo/skills/buddy-oncall-assistant/scripts/download_jira_attachment.py TS-1234 --filename data.csv

# Extract IDs (first column)
awk -F',' 'NR>1 {print $1}' data.csv | sort -u > ids.txt

# Batch process
mybuddy txn ids.txt
```

## Tool Selection Guide

| Task | Command |
|:---|:---|
| Find tickets | `jira search` |
| View ticket | `jira view <id>` |
| Download CSV attachments | `jira download-attachment <id>` |
| Investigate transaction | `txn <id>` |
| Batch process | `txn <file>` |
| RPP recovery | `rpp resume` (MY) |
| Query database | `doorman query` |
| Create DML ticket | `doorman create-dml` (MY) |

## Notes

- **Regional:** MyBuddy has RPP, SGBuddy has Fast Adapter
- **Attachments:** Both MyBuddy and SGBuddy now support native CSV attachment downloads
- **Privacy:** Never output PII or credentials

## Database Schema Reference

The `DATABASE_SCHEMA.md` file contains detailed table structures for all services across both Malaysia and Singapore regions. Use it to understand table schemas, column names, and data types before constructing queries.

### When to Reference DATABASE_SCHEMA.md

**Before executing `doorman query` commands:**
- Identify correct table names for the target service
- Verify column names and data types
- Understand relationships between tables
- Example: Check `payment_core.external_transaction` schema before querying by `tx_id`

**During transaction investigation (`txn` commands):**
- Understand the data structure of transaction tables
- Identify relevant columns for filtering (e.g., `status`, `error_code`)
- Map transaction IDs to appropriate tables across services

**For RPP operations (`rpp resume`):**
- Review `rpp_adapter.credit_transfer` schema for RPP-specific fields
- Understand workflow execution state transitions

**When debugging workflows:**
- Reference `workflow_execution` table structure across all services
- Understand state numbers and transition IDs
- Interpret the `data` JSON field structure

**To understand regional differences:**
- Malaysia: `rpp_adapter` (unique to MY)
- Singapore: `fast_adapter` (unique to SG)
- Common services: `payment_core`, `payment_engine`, `partnerpay_engine`

### Recommended Workflow

1. **Identify the service** for your task (e.g., `payment_core` for transaction queries)
2. **Consult DATABASE_SCHEMA.md** to review the relevant table structure
3. **Construct your query** using verified column names and data types
4. **Execute** using the appropriate `doorman query` command
5. **Iterate** by refining queries based on schema understanding

### Example Usage

```bash
# Step 1: Check DATABASE_SCHEMA.md for payment_core.external_transaction
# Step 2: Construct query using verified columns
mybuddy doorman query -s payment_core -q "SELECT tx_id, status, amount FROM external_transaction WHERE tx_id = 'TXN123'"

# For Singapore Fast Adapter
sgbuddy doorman query -s fast_adapter -q "SELECT transaction_id, status FROM transactions WHERE status = 1"
```

**Key Services Reference:**
- Malaysia: `payment_core`, `payment_engine`, `rpp_adapter`, `partnerpay_engine`
- Singapore: `payment_core`, `payment_engine`, `fast_adapter`, `partnerpay_engine`
