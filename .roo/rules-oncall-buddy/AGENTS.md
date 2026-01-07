---
name: oncall-buddy
description: CLI tools (mybuddy/sgbuddy) for database queries, transaction investigation, Jira integration, and payment operations.
---

# Quick Decision Tree

1. **Start with ticket ID?** → TS → mybuddy | TSE → sgbuddy
2. **Need transaction info?** → txn [id_or_file]
3. **Grab/PartnerPay transaction?** → ecotxn [run_id]
4. **Need database query?** → doorman query -s [service] -q "[sql]"
5. **RPP stuck (MY only)?** → rpp resume
6. **PayNow unlink (SG only)?** → paynow unlink [safeid]
7. **Search logs (SG only)?** → dd search "[query]" --last [hours]

---

# Critical Workflow

When given a ticket ID (TS-XXXX or TSE-XXXX):
1. Extract Batch IDs or TAR02 BMIDs from ticket
2. Save to file: `{ticket_id}.txt` (one ID per line)
3. Run investigation:
   - TS ticket → `mybuddy txn {ticket_id}.txt`
   - TSE ticket → `sgbuddy txn {ticket_id}.txt`
4. If Grab transaction → use `mybuddy ecotxn` or `sgbuddy ecotxn`

# Tool Selection

| Region | Tool | Env File | Jira Project |
|:---|:---|:---|:---|
| Malaysia | `mybuddy` | `.env.my` | `TS` |
| Singapore | `sgbuddy` | `.env.sg` | `TSE` |

# Commands

## Jira Integration
```bash
# List tickets (interactive)
mybuddy jira list
sgbuddy jira list

# Search tickets
mybuddy jira search "keyword"

# View ticket
mybuddy jira view TS-1234
sgbuddy jira view TSE-5678

# Download CSV attachments
mybuddy jira download-attachment TS-1234
sgbuddy jira download-attachment TSE-5678

# Download to specific directory
mybuddy jira download-attachment TS-1234 --output ./downloads
sgbuddy jira download-attachment TSE-5678 -o ./data
```

## Transaction Investigation
```bash
# Single transaction (TXN ID, E2E ID, or Batch ID)
mybuddy txn TXN123
mybuddy txn 20250101120000

# Batch processing (file with one ID per line)
mybuddy txn ids.txt
sgbuddy txn ids.txt
```

## RPP Operations (Malaysia Only)
```bash
mybuddy rpp resume
```
Output: Deploy SQL + Rollback SQL

## PartnerPay Inspection
```bash
# View transaction
mybuddy ecotxn <run_id>
sgbuddy ecotxn <run_id>

# MyBuddy - Auto-create DML tickets
mybuddy ecotxn <run_id> --create-dml "TS-4558"

# SGBuddy - Interactive mode
sgbuddy ecotxn <txn-id> --publish

# SGBuddy - Auto-create DML ticket
sgbuddy ecotxn <txn-id> --publish --create-dml "TSE-1234"
```

## Database Queries

### Query (Read-Only)
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

### Create DML Ticket (Malaysia Only)
```bash
mybuddy doorman create-dml \
  --service payment_core \
  --original "UPDATE transactions SET status = 'done' WHERE id = 'TXN123'" \
  --rollback "UPDATE transactions SET status = 'pending' WHERE id = 'TXN123'" \
  --note "Fix TXN123 - Ref TS-456"
```

## CSV Processing
```bash
# Download attachment
mybuddy jira download-attachment TS-1234
sgbuddy jira download-attachment TSE-5678

# Extract IDs (first column)
awk -F',' 'NR>1 {print $1}' data.csv | sort -u > ids.txt

# Batch process
mybuddy txn ids.txt
```

# Quick Decision Table

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
| PayNow unlink | `paynow unlink <safeid>` (SG) |
| Search logs | `dd search "<query>" --last <hours>` (SG) |

---

# Singapore-Exclusive Commands

## PayNow Operations
```bash
sgbuddy paynow unlink <safeid>
```
Unlinks PayNow for a given SafeID. Triggers deregistration in Pairing Service.

## Datadog Integration
```bash
# Search logs
sgbuddy dd search "service:payment error" --last 3

# Aggregate metrics
sgbuddy dd aggregate "error"

# Submit log events
sgbuddy dd submit
```
Search logs, aggregate metrics, submit log events.

---

# Malaysia-Specific Notes

## E2E ID Usage
Malaysia supports E2E IDs (timestamp format: `20250101120000`) for batch identification:
```bash
mybuddy txn 20250101120000
```
This is more efficient than extracting individual transaction IDs when investigating batches.

---

# Regional Differences

- **Malaysia (MyBuddy):** Has RPP adapter, no Fast adapter
- **Singapore (SGBuddy):** Has Fast adapter, no RPP adapter
- **Common:** Both support native CSV attachment downloads

# Database Schema Reference

Consult `DATABASE_SCHEMA.md` for:
- Table structures, column names, data types
- Service-specific schema differences
- Relationship mappings between tables

Use before executing `doorman query` commands to verify table/column names.
