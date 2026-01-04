---
name: buddy-oncall-assistant
description: Use this skill when working with Jira tickets, mybuddy/sgbuddy tools, G-Bank payment operations investigations, or transaction remediation workflows
---

# Buddy On-Call Assistant Skill

This skill guides you through using mybuddy and sgbuddy tools to investigate and resolve Jira tickets for G-Bank payment operations.

## Tool Overview

### mybuddy (Malaysia Operations)
- **Region**: Malaysia (MY)
- **Config**: Uses `.env.my` environment file
- **Commands**:
  - `txn` - Query transaction status and generate remediation SQL
  - `rpp` - RPP (Regional Payment Partner) adapter operations
  - `rpp resume` - Resume RPP workflows with SQL generation
  - `ecotxn` - PartnerPay (Eco) workflow inspection
  - `jira` - JIRA ticket operations (list, search, attachments)

### sgbuddy (Singapore Operations)
- **Region**: Singapore (SG)
- **Config**: Uses `.env.sg` environment file
- **Commands**:
  - `txn` - Query transaction status and generate remediation SQL
  - `jira` - JIRA ticket operations (list, search)
  - `ecotxn` - PartnerPay (Eco) workflow inspection

### Prerequisites

Before using these tools, ensure the following environment variables are set:

**For Jira access:**
- `JIRA_DOMAIN` - Jira instance domain (optional, has defaults)
- `JIRA_USERNAME` - Jira username/email
- `JIRA_API_KEY` - Jira API key for authentication

**For database access:**
- `DOORMAN_USERNAME` - Doorman database username
- `DOORMAN_PASSWORD` - Doorman database password

**Default Jira domains:**
- Malaysia: `https://gxbank.atlassian.net` (Project: TS)
- Singapore: `https://gxsbank.atlassian.net` (Project: TSE)

## Jira Ticket Workflow

### 1. List Assigned Tickets

List active Jira tickets assigned to the current user:

```bash
# Malaysia
mybuddy jira list

# Singapore
sgbuddy jira list
```

**What it does:**
- Fetches tickets assigned to you using JQL
- Excludes completed/closed tickets
- Shows up to 50 tickets by default
- Launches interactive picker for ticket selection

### 2. Search Tickets

Search for tickets matching specific criteria:

```bash
# Search by keywords in summary/description
mybuddy jira search "payment failed"
sgbuddy jira search "API error"

# Multiple search terms
mybuddy jira search "refund" "pending"
```

**What it does:**
- Searches in both summary and description fields
- Returns unresolved tickets only
- Provides clickable links to tickets
- Supports full-text search queries

### 3. Read Ticket Details

When you select a ticket from the list/search results (via interactive picker):

**Available actions:**
- View full ticket details (description, status, assignee)
- Open ticket in browser
- Download attachments (mybuddy only)
- View transaction IDs and other key information

**For mybuddy (Malaysia):**
- Full attachment support in picker
- Download attachments directly to current directory
- Handles duplicate filenames automatically

**For sgbuddy (Singapore):**
- No attachment download in picker
- View attachment URLs manually
- Use browser to download attachments

### 4. Download Attachments

```bash
# Via interactive picker (mybuddy only)
mybuddy jira list
# → Select ticket → Choose "Download attachment"

# Downloads to current directory
# Duplicates are named: file_1.pdf, file_2.pdf, etc.
```

**Note**: sgbuddy does not support attachment downloads via CLI. Use browser or curl with attachment URLs.

### 5. Transition Ticket Status

Close or transition tickets programmatically:

```bash
# Available through the Jira client API
# Supports fetching available transitions
# Executes state changes (e.g., close, resolve)
```

This is typically done through the interactive picker or via Jira API calls.

## Transaction Investigation Workflow

### Step 1: Extract Transaction ID from Jira Ticket

When reading a Jira ticket, look for:
- Regular transaction IDs (e.g., `TXN123456789`)
- RPP E2E IDs (format: `YYYYMMDDGXSPMYXXXXXXXXXXXXXXXX`)
- Transaction IDs in ticket description or comments

### Step 2: Investigate the Transaction

```bash
# Malaysia
mybuddy txn <transaction-id-or-e2e-id-or-file>

# Singapore
sgbuddy txn <transaction-id-or-e2e-id-or-file>

# Examples
mybuddy txn 20251228TNGDMYNB010ORM77048250
sgbuddy txn TXN123456789

# Batch processing from file
mybuddy txn transactions.txt
```

**What it does:**
- Queries Payment Engine database
- Retrieves transaction status and details
- Identifies the SOP case type
- Generates remediation SQL based on Standard Operating Procedures
- Provides interactive prompts for case classification

### Step 3: Interpret the Output

The tool will show:
- **Transaction Status**: Current state of the transaction
- **Case Type**: SOP classification (e.g., "timeout", "declined", "pending")
- **Generated SQL**: Remediation SQL to apply (if applicable)
- **Recommendations**: Next steps based on transaction state

### Step 4: Apply Remediation

The generated SQL is ready to apply:
1. Review the generated SQL for correctness
2. Apply to the appropriate database (Payment Engine, Payment Core, etc.)
3. Verify the transaction state after applying
4. Update the Jira ticket with actions taken

**Important**: Always document remediation steps in Jira ticket comments.

## RPP Workflow Recovery (Malaysia Only)

### When to Use RPP Resume

Use `mybuddy rpp resume` when:
- RPP adapter workflows are stuck
- Transactions are pending at RPP adapter
- Jira tickets indicate RPP-related issues
- Need to generate deploy/rollback SQL for RPP workflows

### How to Use

```bash
# Inspect RPP adapter records and generate SQL
mybuddy rpp resume

# This will:
# 1. Query RPP adapter tables
# 2. Identify stuck workflows
# 3. Generate deploy SQL to resume workflows
# 4. Generate rollback SQL for safety
```

### Interpreting RPP Output

- **Adapter Status**: Current state of RPP workflows
- **Stuck Workflows**: Workflows that need intervention
- **Generated SQL**: Deploy SQL to resume, rollback SQL for safety
- **Recommendations**: Whether to proceed with remediation

### Applying RPP Remediation

1. Review the generated SQL thoroughly
2. Test in non-production if possible
3. Apply deploy SQL to resume workflows
4. Monitor for successful completion
5. Keep rollback SQL for emergency rollback
6. Document actions in Jira ticket

## PartnerPay Workflow Inspection

### When to Use Eco Transaction

Use `ecotxn` when:
- Jira ticket mentions PartnerPay or Eco transactions
- Ticket contains run_id for Eco workflows
- Need to investigate PartnerPay transaction status

### How to Use

```bash
# Malaysia
mybuddy ecotxn <run_id>

# Singapore
sgbuddy ecotxn <run_id>

# Example
mybuddy ecotxn RUN12345
sgbuddy ecotxn 20251228-ECO-001
```

**What it does:**
- Queries PartnerPay database
- Retrieves Eco workflow details
- Shows workflow status and steps
- Identifies failures or pending steps
- Provides diagnostic information

## Best Practices

### 1. Identify Correct Region

Always verify which region the Jira ticket belongs to:
- **Malaysia**: Use mybuddy, `.env.my`, gxbank.atlassian.net
- **Singapore**: Use sgbuddy, `.env.sg`, gxsbank.atlassian.net

**How to identify:**
- Check Jira project (TS = Malaysia, TSE = Singapore)
- Check ticket components or labels
- Check ticket description for region mentions
- Ask the user if unclear

### 2. Verify Environment Setup

Before running commands:
```bash
# Check that environment variables are set
echo $JIRA_USERNAME
echo $JIRA_API_KEY
echo $DOORMAN_USERNAME

# Verify tool is available
mybuddy --help   # or sgbuddy --help
```

### 3. Use Interactive Features

- Use `jira list` with the interactive picker for complex cases
- Let the tool guide you through SOP classification
- Review all generated SQL before applying
- Take advantage of browser-opening features for detailed review

### 4. Document in Jira

Always update the Jira ticket with:
- Investigation findings
- Transaction IDs reviewed
- Remediation applied (include SQL)
- Verification results
- Next steps or resolution

### 5. Safety First

- Review generated SQL before applying
- Keep rollback SQL for RPP operations
- Test in non-production when possible
- Verify transactions after remediation
- Document everything for audit trail

## Common Workflows

### Workflow 1: Payment Timeout Investigation

```bash
# 1. List or search for timeout-related tickets
mybuddy jira search "timeout"

# 2. Select ticket and read details
# → Extract transaction ID from description

# 3. Investigate transaction
mybuddy txn <transaction-id>

# 4. Review generated remediation SQL
# → Apply if appropriate

# 5. Update Jira ticket with findings
```

### Workflow 2: RPP Adapter Stuck Workflows

```bash
# 1. Search for RPP-related tickets
mybuddy jira search "RPP" "stuck"

# 2. Use RPP resume to investigate and generate SQL
mybuddy rpp resume

# 3. Review generated deploy and rollback SQL

# 4. Apply deploy SQL (if safe)
# → Keep rollback SQL for safety

# 5. Update Jira with actions taken
```

### Workflow 3: PartnerPay Transaction Investigation

```bash
# 1. Identify Eco/PartnerPay ticket
sgbuddy jira search "PartnerPay"

# 2. Extract run_id from ticket

# 3. Investigate Eco workflow
sgbuddy ecotxn <run_id>

# 4. Identify failure point

# 5. Update Jira with findings and next steps
```

### Workflow 4: Batch Transaction Processing

```bash
# 1. Create file with transaction IDs (one per line)
echo "TXN001" > transactions.txt
echo "TXN002" >> transactions.txt
echo "TXN003" >> transactions.txt

# 2. Process batch
mybuddy txn transactions.txt

# 3. Review each transaction's status and SQL
# → Apply remediation as needed

# 4. Update all affected Jira tickets
```

## Troubleshooting

### Issue: "Jira authentication failed"

**Solution:**
- Verify `JIRA_USERNAME` and `JIRA_API_KEY` are set
- Check API key is valid and not expired
- Ensure account has access to the Jira instance
- Try accessing Jira in browser to verify credentials

### Issue: "Command not found: mybuddy"

**Solution:**
- Build the tools: `make build` or `make build-my` / `make build-sg`
- Ensure the built binary is in your PATH
- Check if you need to use `./bin/mybuddy` or similar

### Issue: "Transaction not found"

**Solution:**
- Verify the transaction ID is correct
- Check if you're using the correct region (mybuddy vs sgbuddy)
- Some transaction IDs may be E2E IDs - try with the tool anyway
- Check if transaction exists in the database (access issues)

### Issue: "Database connection failed"

**Solution:**
- Verify `DOORMAN_USERNAME` and `DOORMAN_PASSWORD` are set
- Check VPN/network connectivity to database
- Ensure database credentials are valid and not expired
- Verify you have access to the specific database instance

### Issue: "No attachments available" (sgbuddy)

**Expected behavior**: sgbuddy does not support attachment downloads via CLI.

**Workaround:**
- Use the browser to open the ticket
- Download attachments manually from Jira web UI
- Or use curl with the attachment URL if available

### Issue: "Generated SQL looks wrong"

**Solution:**
- Verify the SOP case type selected is correct
- Check if transaction state matches assumptions
- Review transaction details manually
- Don't apply SQL if you're unsure - ask for review
- Consider testing in non-production first

## Related Code Reference

For deeper understanding or debugging, relevant code locations:

- **mybuddy Jira commands**: `internal/apps/mybuddy/commands/jira.go`
- **sgbuddy Jira commands**: `internal/apps/sgbuddy/commands/jira.go`
- **Jira search/list**: `internal/clients/jira/search.go`
- **Attachment handling**: `internal/clients/jira/attachments.go`
- **Interactive picker UI**: `internal/ui/jira_picker.go`
- **Transaction query**: `internal/apps/common/txn/`
- **RPP operations**: `internal/apps/mybuddy/commands/rpp.go`
- **Eco operations**: `internal/apps/common/ecotxn/`

## Important Notes

### Environment-Specific Features

| Feature | mybuddy (MY) | sgbuddy (SG) |
|---------|--------------|--------------|
| Transaction queries | ✅ | ✅ |
| Jira list/search | ✅ | ✅ |
| Jira attachments | ✅ | ❌ |
| RPP operations | ✅ | ❌ |
| Eco transactions | ✅ | ✅ |
| CSV attachments | ✅ | ❌ |

### Data Privacy and Security

- Never output full credentials or API keys in chat
- Transaction data may contain sensitive information
- Follow data handling policies for customer data
- Use secure channels for sharing sensitive information
- Audit trail is critical - document all actions

### When in Doubt

1. **Ask for clarification**: Which region? Which environment?
2. **Verify assumptions**: Don't guess - check with user
3. **Review before applying**: Especially for SQL remediation
4. **Document everything**: Jira tickets are the source of truth
5. **Escalate if needed**: Some issues may require senior engineer review

## Quick Reference Commands

```bash
# Jira Operations
mybuddy jira list              # List assigned tickets
mybuddy jira search "query"    # Search tickets

# Transaction Investigation
mybuddy txn <id>               # Investigate transaction
mybuddy txn file.txt           # Batch process

# RPP Operations (Malaysia only)
mybuddy rpp resume             # Resume stuck RPP workflows

# PartnerPay Investigation
mybuddy ecotxn <run_id>        # Investigate Eco workflow

# Same commands for Singapore (using sgbuddy)
sgbuddy jira list
sgbuddy txn <id>
sgbuddy ecotxn <run_id>
```
