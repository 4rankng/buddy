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

### 4.1 Download CSV Attachments and Extract Transaction IDs

CSV attachments are commonly attached to Jira tickets and contain transaction IDs that need batch investigation.

#### Download CSV from Jira (mybuddy only)

```bash
# Method 1: Via interactive picker
mybuddy jira list
# → Select ticket → Choose "Download attachment"
# → Select the CSV file attachment
# → File downloads to current directory

# Method 2: Direct download via browser
# If interactive picker doesn't work:
# 1. Open ticket in browser (via picker or URL)
# 2. Download CSV attachment manually
# 3. Save to working directory
```

#### Extract Transaction IDs from CSV

Once you have the CSV file downloaded, you need to extract transaction IDs from it.

**Understand CSV Structure:**
- CSV files may have various column names for transaction IDs
- Common column names: `transaction_id`, `transactionId`, `txn_id`, `id`, `e2e_id`
- May contain headers or be data-only
- May need to skip first row (headers)

**Methods to Extract Transaction IDs:**

##### Method 1: Using awk (Quick extraction)

```bash
# Extract from first column (assuming transaction ID is in column 1)
awk 'NR>1 {print $1}' transactions.csv > txn_ids.txt

# Extract from specific column by header name
# Example: Get transaction_id column
awk -F',' 'NR==1 {for(i=1;i<=NF;i++)if($i=="transaction_id")col=i} NR>1{print $col}' transactions.csv > txn_ids.txt

# Extract all non-empty transaction IDs from any column
awk -F',' 'NR>1 {for(i=1;i<=NF;i++)if($i~/^[A-Z0-9]+$/)print $i}' transactions.csv > txn_ids.txt
```

##### Method 2: Extract specific patterns (Transaction ID formats)

```bash
# Extract regular transaction IDs (e.g., TXN123, 20251228TNGDMYNB...)
grep -oE '[A-Z]{3}[0-9]{9,}' transactions.csv > txn_ids.txt

# Extract RPP E2E IDs (format: YYYYMMDDGXSPMY...)
grep -oE '20[0-9]{6}GXSPMY[0-9A-Z]{16,}' transactions.csv > e2e_ids.txt

# Extract transaction IDs with date prefix (common format)
grep -oE '20[0-9]{6}TNGD[A-Z0-9]{20,}' transactions.csv > txn_ids.txt

# Combine all transaction ID formats
grep -oE '(20[0-9]{6}(TNGD|GXSPMY)[A-Z0-9]{15,}|[A-Z]{3}[0-9]{9,})' transactions.csv > txn_ids.txt
```

##### Method 3: Clean and deduplicate IDs

```bash
# Remove duplicates, empty lines, and spaces
sort -u txn_ids.txt | grep -v '^$' > txn_ids_unique.txt

# Remove carriage returns (Windows line endings)
sed -i '' 's/\r$//' txn_ids_unique.txt

# Verify the extracted IDs
cat txn_ids_unique.txt
```

#### Batch Process Extracted Transaction IDs

Once you have the transaction IDs extracted:

```bash
# Batch process all extracted transaction IDs
mybuddy txn txn_ids_unique.txt

# Or process line by line for better control
while IFS= read -r txn_id; do
    echo "Processing: $txn_id"
    mybuddy txn "$txn_id"
    echo "---"
done < txn_ids_unique.txt > investigation_results.txt

# Review results
cat investigation_results.txt
```

#### Example Workflow: CSV Batch Investigation

```bash
# Complete workflow for CSV attachment processing

# Step 1: Download CSV from Jira
mybuddy jira list
# → Select ticket → Download CSV attachment (e.g., failed_transactions.csv)

# Step 2: Inspect CSV structure
head -5 failed_transactions.csv
# → Identify which column contains transaction IDs

# Step 3: Extract transaction IDs
# If transaction IDs are in column 1:
awk -F',' 'NR>1 {print $1}' failed_transactions.csv > txn_ids.txt

# Step 4: Clean and deduplicate
sort -u txn_ids.txt | grep -v '^$' | sed 's/\r$//' > txn_ids_clean.txt

# Step 5: Verify extracted IDs (check first few)
head -10 txn_ids_clean.txt

# Step 6: Batch investigate
mybuddy txn txn_ids_clean.txt

# Step 7: Update Jira ticket with findings
# → Summarize investigation results
# → Note how many transactions were processed
# → Highlight any critical findings or patterns
```

#### Tips for CSV Processing

1. **Always inspect the CSV first** to understand its structure:
   ```bash
   head -10 file.csv        # View first 10 lines
   wc -l file.csv           # Count total lines
   ```

2. **Check for different delimiters** (comma, semicolon, tab):
   ```bash
   # Try different delimiters if comma doesn't work
   awk -F';' 'NR>1 {print $1}' file.csv    # Semicolon
   awk -F'\t' 'NR>1 {print $1}' file.csv   # Tab
   ```

3. **Handle quoted fields** (commas within quotes):
   ```bash
   # Use Python for complex CSV parsing
   python3 -c "import csv,sys; [print(row[0]) for row in csv.reader(sys.stdin)]" < file.csv > txn_ids.txt
   ```

4. **Validate transaction IDs** before processing:
   ```bash
   # Check if IDs match expected format
   grep -vE '^[A-Z0-9]{15,}$' txn_ids.txt  # Show invalid IDs
   ```

5. **Save investigation results** for documentation:
   ```bash
   # Save both output and any generated SQL
   mybuddy txn txn_ids.txt | tee investigation_$(date +%Y%m%d_%H%M%S).log
   ```

#### Special Case: sgbuddy (Singapore) CSV Attachments

Since sgbuddy doesn't support attachment downloads via CLI:

```bash
# Alternative for sgbuddy:

# Option 1: Manual browser download
# 1. sgbuddy jira list → Select ticket → Open in browser
# 2. Download CSV manually from browser
# 3. Use same extraction commands as above

# Option 2: Use curl with attachment URL (if URL is visible)
# Extract URL from ticket details, then:
curl -L -o transactions.csv "<attachment-url>"

# Then proceed with extraction steps:
awk -F',' 'NR>1 {print $1}' transactions.csv > txn_ids.txt
sgbuddy txn txn_ids.txt
```

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

### Workflow 5: CSV Attachment Batch Investigation

```bash
# Scenario: Jira ticket has CSV attachment with hundreds of failed transactions

# Step 1: Download CSV from Jira
mybuddy jira list
# → Select ticket → Download attachment → Choose CSV file

# Step 2: Inspect CSV to understand structure
head -10 transactions.csv
# Output shows: transaction_id,status,timestamp,error_code
#            TXN001,failed,2025-01-04,TIMESTAMP

# Step 3: Extract transaction IDs (column 1)
awk -F',' 'NR>1 {print $1}' transactions.csv > txn_ids.txt

# Step 4: Clean and verify
sort -u txn_ids.txt | grep -v '^$' | sed 's/\r$//' > txn_ids_clean.txt
wc -l txn_ids_clean.txt  # Count transactions to investigate
head -5 txn_ids_clean.txt  # Verify format

# Step 5: Batch investigate with timestamped log
mybuddy txn txn_ids_clean.txt | tee investigation_$(date +%Y%m%d_%H%M%S).log

# Step 6: Review results and patterns
# → Look for common failure modes
# → Identify transactions needing remediation
# → Group by SOP case type

# Step 7: Update Jira ticket with comprehensive summary
# "Investigated 250 transactions from CSV attachment.
#  - 150 pending timeout → Generated SQL for remediation
#  - 50 already successful → No action needed
#  - 30 declined → Require manual review
#  - 20 require database investigation
# Remediation SQL attached. Applied to production at [time]."
```

### Workflow 6: Multi-CSV Processing (Complex Tickets)

```bash
# Scenario: Multiple CSV attachments in one ticket

# Step 1: Download all CSV attachments
mybuddy jira list
# → Select ticket → Download each CSV
# Files: failed_txns.csv, pending_txns.csv, retry_txns.csv

# Step 2: Extract IDs from all CSVs into single file
for file in *.csv; do
    awk -F',' 'NR>1 {print $1}' "$file"
done > all_txn_ids.txt

# Step 3: Deduplicate across all files
sort -u all_txn_ids.txt | grep -v '^$' | sed 's/\r$//' > unique_txn_ids.txt

# Step 4: Categorize by source for tracking
# Track which CSV each ID came from
for file in failed_txns.csv pending_txns.csv retry_txns.csv; do
    awk -F',' 'NR>1 {print $1}' "$file" | while read id; do
        echo "$id,$file"
    done
done > txn_sources.txt

# Step 5: Batch investigate
mybuddy txn unique_txn_ids.txt > investigation_results.log

# Step 6: Cross-reference results with sources
# Create summary by original category
paste txn_ids.txt txn_sources.txt | grep -f investigation_results.log

# Step 7: Update Jira with categorized findings
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
