---
name: buddy-oncall-assistant
description: Operational CLI tools (mybuddy/sgbuddy) for Jira management, transaction investigation, RPP recovery, and DML creation in G-Bank payment operations.
---

# Buddy On-Call Assistant Skill

## Context & Tool Selection
Select the tool based on the Jira project or region.

| Scope | Region | Tool Command | Env File | Jira Project |
| :--- | :--- | :--- | :--- | :--- |
| **Malaysia** | MY | `mybuddy` | `.env.my` | `TS` |
| **Singapore** | SG | `sgbuddy` | `.env.sg` | `TSE` |

## Environment Variables
Ensure these are set before execution:
* `JIRA_USERNAME` / `JIRA_API_KEY`
* `DOORMAN_USERNAME` / `DOORMAN_PASSWORD`
* `JIRA_DOMAIN` (Optional defaults: `gxbank.atlassian.net` for MY, `gxsbank.atlassian.net` for SG)

## Core Commands

### 1. Jira Operations
Interact with tickets. Use interactive pickers when possible.

```bash
# List assigned tickets (interactive)
mybuddy jira list

# Search (Summary/Description)
mybuddy jira search "keyword" "keyword2"

# View Details (Direct)
mybuddy jira view TS-1234
```

### 2. Transaction Investigation
Diagnose status and generate remediation SQL. Input: Transaction ID (TXN...), E2E ID (2025...), or file path.

```bash
# Single Transaction
mybuddy txn <txn_id_or_e2e_id>

# Batch Processing (File containing IDs)
mybuddy txn ids.txt
```

### 3. RPP Recovery (Malaysia Only)
Resume stuck RPP adapter workflows.

```bash
# Analyze and generate Deploy/Rollback SQL
mybuddy rpp resume
```

### 4. PartnerPay (Eco) Inspection
Investigate PartnerPay workflow runs.

```bash
mybuddy ecotxn <run_id>
```

### 5. Doorman DML Creation (Malaysia Only)
Generate production database change tickets.

Syntax:

```bash
mybuddy doorman create-dml \
  --service <service_name> \
  --original "<update_query>" \
  --rollback "<rollback_query>" \
  --note "<description_and_reference>"
```
Valid Services: payment_engine, payment_core, rpp_adapter, partnerpay_engine.

## Advanced Workflows

### Batch CSV Processing (Jira Attachments)
Use for tickets with attached CSVs of failed transactions.

#### 1. Download Attachment
Prefer the Python script for reliability and auto-region detection.

```bash
# Download specific file
python .kilocode/skills/buddy-oncall-assistant/scripts/download_jira_attachment.py TS-1234 --filename transactions.csv

# List available attachments
python .kilocode/skills/buddy-oncall-assistant/scripts/download_jira_attachment.py TS-1234 --list-only
```

#### 2. Extract IDs
Parse CSV (handle headers and delimiters) to create a clean list.

```bash
# Extract column 1 (standard)
awk -F',' 'NR>1 {print $1}' transactions.csv > txn_ids.txt

# Extract column by header name "transaction_id"
awk -F',' 'NR==1 {for(i=1;i<=NF;i++)if($i=="transaction_id")col=i} NR>1{print $col}' transactions.csv > txn_ids.txt

# Clean/Deduplicate
sort -u txn_ids.txt | grep -v '^$' | sed 's/\r$//' > clean_ids.txt
```

#### 3. Execute Batch

```bash
mybuddy txn clean_ids.txt > results.log
```

### Resolution Routine (End-to-End)
Trigger: "Resolve TS-XXXX"

1. **Context**: `mybuddy jira view TS-XXXX` (Check description/region).
2. **Fetch Data**:
    * If IDs in text -> Copy to file.
    * If CSV -> Run Batch CSV Processing.
3. **Investigate**: Run `mybuddy txn` (or `rpp resume` / `ecotxn` based on context).
4. **Analyze**: Review generated SQL against SOP case type (Timeout, Declined, etc.).
5. **Action**:
    * If valid -> Apply SQL.
    * If complex -> Create DML ticket (`doorman create-dml`).
6. **Close**: Update Jira with findings, SQL used, and result.

## Troubleshooting & Rules
* **SgBuddy Attachments**: `sgbuddy` CLI cannot download attachments. Use Browser or curl.
* **Hex Strings**: If CSV contains 32-char hex strings, check for a "Batch ID" column 
* **Safety**: Always generate and save Rollback SQL for RPP operations.
* **Privacy**: Never output raw customer PII or credentials in chat.
