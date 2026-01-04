---
name: jira-buddy
description: Use this skill when working with Jira tickets, investigating payment issues, or resolving operational tasks using mybuddy or sgbuddy tools.
---

# Jira-Buddy Skill

You are an expert at using **mybuddy** and **sgbuddy** - operational CLI tools for investigating payment issues and resolving Jira tickets.

## Tool Overview

### Environment Selection
- **mybuddy**: Malaysia (MY) environment operations
- **sgbuddy**: Singapore (SG) environment operations

**IMPORTANT:** Always ask the user which environment to work with if not specified. Do not assume.

### What These Tools Do
Both tools provide:
1. **Jira Integration** - List, search, and view assigned tickets (including direct ticket viewing)
2. **Transaction Investigation** - Query payment engine database for transaction status
3. **Remediation** - Generate SQL fixes and process workflows
4. **Eco/PartnerPay** - Publish and manage Eco transactions

### Key Differences
| Feature | mybuddy (MY) | sgbuddy (SG) |
|---------|--------------|--------------|
| Jira attachments | ✅ Yes (shown & downloadable) | ❌ No |
| RPP commands | ✅ Yes (resume, rtp-cashin) | ❌ No |
| Eco transactions | ✅ Yes | ✅ Yes |
| Transaction lookup | ✅ Yes | ✅ Yes |

---

## Standard Workflow for Resolving Jira Tickets

Follow this step-by-step process when working with Jira tickets:

### Step 1: List Assigned Tickets
```bash
# Malaysia
mybuddy jira list

# Singapore
sgbuddy jira list
```

This launches an interactive picker showing:
- Ticket key (clickable link)
- Summary
- Status
- Priority
- Assignee
- Created date
- Due date (if set)
- Description
- Attachments (mybuddy only)

### Step 2: View Specific Ticket Details
If you have a specific ticket key and want to see its details directly:
```bash
mybuddy jira view TS-4565
sgbuddy jira view TSE-123
```

This displays:
- Ticket key (clickable link)
- Summary
- Status
- Priority
- Assignee
- Created and due dates
- Full description
- Attachments (mybuddy only)

### Step 3: Search for Specific Tickets (Optional)
If you need to find tickets matching specific criteria:
```bash
mybuddy jira search "payment issue"
mybuddy jira search "API" "error"
sgbuddy jira search "refund"
```

Search looks in:
- Ticket summary
- Ticket description
- Only your unresolved tickets (not Closed/Completed)

### Step 4: Download Attachments
**For mybuddy:**
- Use the interactive picker's "Download attachment" action
- Attachments download to current directory

**For sgbuddy:**
- sgbuddy doesn't show attachments in the picker
- Manually extract attachment URLs from ticket details
- Use curl or browser to download

### Step 5: Extract Transaction IDs
Transaction IDs can be found in:
1. **Ticket description** - Look for patterns like:
   - `20251228TNGDMYNB010ORM77048250` (regular txn ID)
   - `YYYYMMDDGXSPMYXXXXXXXXXXXXXXXX` (RPP E2E ID format)
   - UUIDs: `fd230a01dcd04282851b7b9dd6260c93`

2. **CSV attachments** - Common format:
   - Batch files with multiple transaction IDs
   - One ID per line or in columns

3. **Filenames** - Attachments may be named like `TSE-833.txt`

### Step 6: Investigate Transactions

#### Single Transaction
```bash
mybuddy txn 20251228TNGDMYNB010ORM77048250
sgbuddy txn 9392fb12b6c64db18e779ae60bdf4307
```

#### Batch Processing
```bash
# Process multiple IDs from a file
mybuddy txn /path/to/transaction-ids.txt
sgbuddy txn /path/to/tickets.txt
```

#### RPP Transactions (mybuddy only)
```bash
# RPP E2E ID format
mybuddy txn 20251228GXSPMY202512281234567890ABCDEF
```

The tool will:
1. Query the payment engine database
2. Display transaction status
3. Generate remediation SQL if needed
4. Show applicable rules and workflows

### Step 7: Process Eco Transactions (If Applicable)
```bash
mybuddy ecotxn fd230a01dcd04282851b7b9dd6260c93
mybuddy ecotxn TSE-833.txt
sgbuddy ecotxn publish <txnid>
```

### Step 8: Handle RPP Workflows (mybuddy only)
```bash
# Inspect RPP adapter workflows
mybuddy rpp resume <workflow-id>

# Handle RTP cashin operations
mybuddy rpp rtp-cashin <transaction-id>
```

### Step 9: Document and Resolve
Based on investigation findings:
- **Generate SQL** - The tool provides remediation SQL
- **Verify status** - Check if transaction needs manual intervention
- **Update ticket** - Document findings and actions taken
- **Close ticket** - Use Jira interface or CLI if resolved

---

## Complete Command Reference

### Jira Operations

#### List Tickets
```bash
mybuddy jira list    # Malaysia - with attachments
sgbuddy jira list    # Singapore - no attachments
```
**Interactive features:**
- Search/filter tickets
- View ticket details
- Open in browser
- Download attachments (mybuddy only)
- Quit to exit

#### View Ticket Details
```bash
mybuddy jira view <ticket-key>
sgbuddy jira view <ticket-key>

# Examples
mybuddy jira view TS-4565
sgbuddy jira view TSE-123
```

Displays full ticket details including:
- Ticket key (clickable link)
- Summary, status, priority, assignee
- Created and due dates
- Full description
- Attachments (mybuddy only)

#### Search Tickets
```bash
mybuddy jira search [terms...]
sgbuddy jira search [terms...]

# Examples
mybuddy jira search "payment failed"
mybuddy jira search "DuitNow" "error"
sgbuddy jira search PayNow
```

### Transaction Investigation

#### Query Single Transaction
```bash
mybuddy txn <transaction-id-or-e2e-id>
sgbuddy txn <transaction-id-or-e2e-id>

# Examples
mybuddy txn 20251228TNGDMYNB010ORM77048250
sgbuddy txn 9392fb12b6c64db18e779ae60bdf4307
```

**Supported ID formats:**
- Regular transaction IDs (alphanumeric)
- RPP E2E IDs: `YYYYMMDDGXSPMYXXXXXXXXXXXXXXXX`
- UUIDs

#### Batch Process Transactions
```bash
mybuddy txn /path/to/file.txt
sgbuddy txn /path/to/file.txt

# File format: one transaction ID per line
# Example file contents:
# 20251228TNGDMYNB010ORM77048250
# 20251228TNGDMYNB010ORM77048251
# fd230a01dcd04282851b7b9dd6260c93
```

### Eco/PartnerPay Operations

#### Publish Eco Transaction
```bash
mybuddy ecotxn <txnid-or-file>
sgbuddy ecotxn publish <txnid-or-file>

# Examples
mybuddy ecotxn fd230a01dcd04282851b7b9dd6260c93
mybuddy ecotxn TSE-833.txt
sgbuddy ecotxn publish 9392fb12b6c64db18e779ae60bdf4307
```

### RPP Operations (mybuddy only)

#### Resume RPP Workflow
```bash
mybuddy rpp resume <workflow-id>
```
Inspects RPP adapter workflow status and details.

#### Handle RTP Cashin
```bash
mybuddy rpp rtp-cashin <transaction-id>
```
Processes RTP cashin operations for RPP transactions.

---

## Best Practices

### 1. Environment Detection
- **Always clarify environment first**: "Are you working with Malaysia (mybuddy) or Singapore (sgbuddy) operations?"
- Check for context clues in ticket:
  - Project prefixes (TSE-MY vs TSE-SG)
  - Currency mentions (MYR, SGD)
  - Bank names (Maybank, CIMB vs DBS, OCBC)

### 2. Batch Processing
- When tickets have 5+ transactions, use batch files
- Create file with one ID per line
- Saves time and generates consolidated output

### 3. Transaction ID Extraction
- Check ticket description first
- Download and inspect CSV attachments
- Look for common patterns:
  - Date prefixes: `20251228...`
  - Environment codes: `MY`, `SG`
  - Service codes: `TNG`, `DNG`, `GPS`

### 4. Output Handling
- mybuddy/sgbuddy generate **remediation SQL**
- Review SQL before executing
- Test in non-prod if possible
- Keep SQL for ticket documentation

### 5. Workflow Efficiency
**For single tickets:**
```bash
mybuddy jira list
# → select ticket
# → download attachments
# → extract IDs
# → investigate with `mybuddy txn`
```

**For batch operations:**
```bash
# Create IDs file first
cat > txn_ids.txt
20251228TNGDMYNB010ORM77048250
20251228TNGDMYNB010ORM77048251
^D

# Process all at once
mybuddy txn txn_ids.txt
```

### 6. Error Handling
If you encounter errors:
- **"JIRA client not initialized"** → Check `.env.my` or `.env.sg` configuration
- **"JIRA_USERNAME not configured"** → Set username in environment file
- **"transaction not found"** → Verify ID format and environment
- **Network errors** → Check VPN/connection to internal services

---

## Common Workflows

### Workflow 1: Payment Investigation Ticket
```bash
# 1. List tickets
mybuddy jira list

# 2. Select ticket and download attachments

# 3. Extract transaction IDs from description/CSV
# Ticket: TSE-123
# Txn ID: 20251228TNGDMYNB010ORM77048250

# 4. Investigate
mybuddy txn 20251228TNGDMYNB010ORM77048250

# 5. Review output and remediation SQL

# 6. Document findings in ticket
# 7. Close if resolved
```

### Workflow 2: Batch Transaction Processing
```bash
# 1. Search for batch-related tickets
mybuddy jira search "batch" "failed"

# 2. Download CSV attachment

# 3. Extract all transaction IDs to file
cat > batch_txns.txt
20251228TNGDMYNB010ORM77048250
20251228TNGDMYNB010ORM77048251
20251228TNGDMYNB010ORM77048252

# 4. Process all transactions
mybuddy txn batch_txns.txt

# 5. Review consolidated output
# 6. Generate summary for ticket
```

### Workflow 3: RPP Transaction Investigation (Malaysia only)
```bash
# 1. List tickets
mybuddy jira list

# 2. Find RPP-related ticket
mybuddy jira search "RPP" "GXSPMY"

# 3. Get E2E ID from ticket
# Format: 20251228GXSPMY202512281234567890ABCDEF

# 4. Investigate RPP workflow
mybuddy rpp resume 20251228GXSPMY202512281234567890ABCDEF

# 5. If RTP cashin issue
mybuddy rpp rtp-cashin 20251228GXSPMY202512281234567890ABCDEF

# 6. Document findings
```

### Workflow 4: Eco Transaction Publishing
```bash
# 1. Find Eco-related ticket
mybuddy jira search "Eco" "PartnerPay"

# 2. Get Eco transaction ID
# Format: UUID like fd230a01dcd04282851b7b9dd6260c93

# 3. Publish Eco transaction
mybuddy ecotxn fd230a01dcd04282851b7b9dd6260c93

# 4. Verify result
```

---

## Quick Reference Card

### Choose Your Tool
| Environment | Command | Attachments | RPP |
|-------------|---------|-------------|-----|
| Malaysia | `mybuddy` | ✅ | ✅ |
| Singapore | `sgbuddy` | ❌ | ❌ |

### Core Commands
```bash
# List tickets
mybuddy jira list | sgbuddy jira list

# Search tickets
mybuddy jira search "terms" | sgbuddy jira search "terms"

# Investigate transaction
mybuddy txn <id> | sgbuddy txn <id>

# Batch process
mybuddy txn file.txt | sgbuddy txn file.txt

# Eco publish
mybuddy ecotxn <id> | sgbuddy ecotxn publish <id>

# RPP (mybuddy only)
mybuddy rpp resume <id>
mybuddy rpp rtp-cashin <id>
```

### Transaction ID Formats
- **Regular**: `20251228TNGDMYNB010ORM77048250`
- **RPP E2E**: `20251228GXSPMYXXXXXXXXXXXXXXXX`
- **UUID**: `fd230a01dcd04282851b7b9dd6260c93`

---

## Troubleshooting

### Issue: "JIRA client not initialized"
**Solution:**
- Ensure `.env.my` or `.env.sg` exists
- Check `JIRA_DOMAIN`, `JIRA_USERNAME`, `JIRA_API_KEY` are set
- Verify file is in project root

### Issue: "No tickets found"
**Solution:**
- Verify you're using correct environment
- Check JIRA username matches logged-in user
- Ensure tickets are not Closed/Completed status

### Issue: "transaction not found"
**Solution:**
- Verify transaction ID format
- Check you're using correct environment (MY vs SG)
- Ensure transaction exists in database

### Issue: Command not found
**Solution:**
- Build tools: `make deploy`
- Verify binary in PATH or use `./mybuddy`, `./sgbuddy`

---

## Tips for Efficiency

1. **Batch everything** - Use files for 3+ transactions
2. **Search smart** - Use specific search terms in ticket search
3. **Document as you go** - Copy SQL output to ticket immediately
4. **Use picker wisely** - mybuddy's interactive picker saves time
5. **Check attachments** - Often contain the transaction IDs you need
6. **Know your environment** - Wrong environment = wrong data

---

## Summary

You are now equipped to:
- ✅ List, search, and view Jira tickets
- ✅ Download and inspect attachments
- ✅ Extract transaction IDs from various sources
- ✅ Investigate single and batch transactions
- ✅ Process RPP workflows (mybuddy)
- ✅ Publish Eco transactions
- ✅ Generate remediation SQL
- ✅ Resolve operational tickets efficiently

**Remember:** Always clarify environment first, follow the workflow, and document your findings.
