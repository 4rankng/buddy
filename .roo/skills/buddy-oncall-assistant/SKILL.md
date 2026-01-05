---
name: buddy-oncall-assistant
description: CLI tools (mybuddy/sgbuddy) for database queries, transaction investigation, Jira integration, and payment operations.
---

# MyBuddy & SGBuddy CLI Tools

## Tool Selection

| Region | Tool | Env File | Jira Project |
|:---|:---|:---|:---|
| Malaysia | `mybuddy` | `.env.my` | `TS` |
| Singapore | `sgbuddy` | `.env.sg` | `TSE` |

## Environment Variables

```bash
JIRA_USERNAME / JIRA_API_KEY
DOORMAN_USERNAME / DOORMAN_PASSWORD
JIRA_DOMAIN  # Optional: gxbank.atlassian.net (MY) or gxsbank.atlassian.net (SG)
```

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
# Download attachment
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
| Investigate transaction | `txn <id>` |
| Batch process | `txn <file>` |
| RPP recovery | `rpp resume` (MY) |
| Query database | `doorman query` |
| Create DML ticket | `doorman create-dml` (MY) |

## Notes

- **Regional:** MyBuddy has RPP, SGBuddy has Fast Adapter
- **SGBuddy:** Cannot download attachments (use browser/curl)
- **Privacy:** Never output PII or credentials
