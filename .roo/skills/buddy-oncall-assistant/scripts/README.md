# Buddy On-Call Assistant Scripts

This directory contains helper scripts for the Buddy On-Call Assistant skill.

## download_jira_attachment.py

Downloads attachments from Jira tickets for both Malaysia and Singapore environments.

### Features

- **Auto-detects region** from ticket key (TS = Malaysia, TSE = Singapore)
- **Downloads all or specific attachments** by filename
- **Lists attachments** without downloading
- **Handles duplicates** automatically (adds _1, _2, etc.)
- **Works for both regions** (mybuddy and sgbuddy)

### Installation

```bash
# Install required dependency
pip install requests

# Or use system Python
python3 -m pip install requests
```

### Usage

```bash
# Download all attachments from a ticket
python download_jira_attachment.py TS-1234

# Download specific CSV file
python download_jira_attachment.py TS-1234 --filename transactions.csv

# List attachments without downloading
python download_jira_attachment.py TS-1234 --list-only

# Download to specific directory
python download_jira_attachment.py TS-1234 --output ./downloads

# Specify region manually
python download_jira_attachment.py TSE-567 --region sg
```

### Environment Variables

Required:
- `JIRA_USERNAME` - Your Jira username/email
- `JIRA_API_KEY` - Your Jira API token

### Examples

```bash
# Quick workflow for CSV investigation
python download_jira_attachment.py TS-1234 --filename failed_transactions.csv
awk -F',' 'NR>1 {print $1}' failed_transactions.csv > txn_ids.txt
mybuddy txn txn_ids.txt
```

### Help

```bash
python download_jira_attachment.py --help
```
