

please create a cli tool named 'oncall' (you can reference the code of /Users/dev/Documents/clients/oncall-app/assets/archived-app for the development)
that

1. The app must be highly extensible, plug-and-play principle, port and adapter design pattern
I am building this for my payment team but I want to make it extensible for other teams to use it for their own use cases

2. The app folder strategy is as follows:
- One folder to contain all reusable modules for all teams to build their own logic, just like lego blocks
  Currently I need to have
   Doorman module: allow user to execute SQL queries, to create DML to fix stuck transactions
   Jira module: to retreive Jira tickets details, to create new Jira tickets, create SHIPRM for different services
   Datadog module: to search datadog logs and metrics
   Storage module: to store data in very light weight file based storage

- One folder to contain custom logic for each team eg internal/payment/

3. The app should have terminal gui, user can navigate with arrow keys or tab key, and select options with enter key
4. The first phase I want to see the GUI desgin first
5. The second phase you implement functionality (refer to /Users/dev/Documents/clients/oncall-app/assets/archived-app for the development)

NOTE

Folder Structure:

pkg/modules/: Contains reusable "Lego block" modules (Doorman, Jira, Datadog).
internal/teams/: Contains team-specific logic (e.g., internal/teams/payment).
pkg/ports/: Defines interfaces for modules to ensure pluggability.
Proposed Changes
Project Structure
oncall/
├── cmd/
│   └── oncall/
│       └── main.go          # Entry point
├── pkg/
│   ├── core/                # Core application logic & plugin manager
│   ├── ports/               # Interfaces (Ports)
│   ├── tui/                 # Terminal UI implementation (Bubbletea)
│   └── modules/             # Reusable Modules (Adapters/Plugins)
│       ├── doorman/
│       ├── jira/
│       └── datadog/
└── internal/
    └── teams/               # Team-specific implementations
        └── payment/         # Payment team's custom logic

if I type oncall payment, you should display GUI for the payment team,
list of info I need
- List of the pending Jira assign to payment team (use mock up for now)
- View a transaction flow given transaction id
- Create deregister paynow SHIPRM


+---------------------------------------------------------------+
|  ONCALL CLI  >  TEAM: PAYMENT                                 |
+---------------------------------------------------------------+
|  [1] DASHBOARD OVERVIEW                                       |
|                                                               |
|  > 1. Pending Jira Tickets (3)                                |
|    * [PAY-1024] Fix double charge race condition [CRITICAL]   |
|    * [PAY-1025] Update mTLS certs for Provider X [HIGH]       |
|    * [PAY-1030] Investigate settlement delay [MED]            |
|                                                               |
|  > 2. Transaction Tools                                       |
|    [ Search Transaction ID ] ______________________________   |
|    (Press Enter to fetch Flow from Datadog/Doorman)           |
|                                                               |
|  > 3. Admin Actions                                           |
|    [ Create Deregister PayNow SHIPRM ]                        |
|                                                               |
+---------------------------------------------------------------+
|  ↑/↓: Move  |  Enter: Select  |  Tab: Next Section  |  q: Quit|
+---------------------------------------------------------------+


ALTERNATIVE

┌──────────────────────────────────────────────────────────────────────────┐
│                          Oncall Tool: Payment Team                       │
└──────────────────────────────────────────────────────────────────────────┘
┌───────────────────────┬──────────────────────────────────────────────────┐
│  Pending Jira Tickets │                  Quick Actions                   │
│ (Use Arrow Keys)      │               (Press Tab to switch)              │
├───────────────────────┼──────────────────────────────────────────────────┤
│ > [PAY-1234] Refunds.. │   1. View Transaction Flow                       │
│   [PAY-1235] Settle... │   2. Create Deregister PayNow SHIPRM            │
│   [PAY-1236] Docs...   │                                                  │
│                       │                                                  │
│                       │                                                  │
└───────────────────────┴──────────────────────────────────────────────────┘
└──────────────────────────────────────────────────────────────────────────┘
│ [Enter: Select] [Tab: Switch Pane] [q: Quit]                            │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│                         Jira Ticket: PAY-1235                           │
├──────────────────────────────────────────────────────────────────────────┤
│ Title: Medium severity - Discrepancy in daily settlement report         │
│ Assignee: oncall-payment                                                 │
│ Status: In Progress                                                      │
│                                                                          │
│ Description:                                                             │
│ The daily settlement report for 2023-10-27 shows a discrepancy of       │
│ $1,250.45 compared to the ledger totals. Initial investigation points   │
│ to a potential issue with batch processing for refunds.                  │
│                                                                          │
│ Link: https://your-org.atlassian.net/browse/PAY-1235                     │
│                                                                          │
│                           [Press Enter to close]                         │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│                      View Transaction Flow                              │
├──────────────────────────────────────────────────────────────────────────┤
│ Enter Transaction ID: txn_abc_123_456_                                  │
│                                                                          │
│ [Enter: Search] [Esc: Cancel]                                            │
└──────────────────────────────────────────────────────────────────────────┘



┌────────────────────────────────────────────────────────────────────────────┐
│ ONCALL ▸ TEAM: PAYMENT                      ENV: prod     USER: frankng   │
├────────────────────────────────────────────────────────────────────────────┤
│ Jira Tickets (↑/↓ to move, Enter to select, Tab to switch pane)          │
│                                                                            │
│ > [PAY-1024] Double charge race condition             [CRITICAL]          │
│   [PAY-1025] Update mTLS certs for Provider X         [HIGH]              │
│   [PAY-1030] Investigate settlement delay             [MED]               │
│   [PAY-1042] PayNow dereg stuck in PENDING            [HIGH]              │
│   [PAY-1050] Settlement report mismatch               [MED]               │
│                                                                            │
├─────────────────────────────┬──────────────────────────────────────────────┤
│ Ticket Summary              │  Actions / Details (active pane)            │
│                             │                                              │
│ ID: PAY-1024                │  Transaction Flow                           │
│ Title: Double charge race   │  ────────────────────────────────────────   │
│        condition            │  Transaction ID: txn____________________    │
│ Assignee: oncall-payment    │  [Enter: Fetch Flow] [Ctrl+L: Use last ID]  │
│ Status: In Progress         │                                              │
│ Priority: CRITICAL          │  ────────────────────────────────────────   │
│                             │  Create Deregister PayNow SHIPRM           │
│ Jira Link:                  │  ────────────────────────────────────────   │
│ https://jira/PAY-1024       │  From Ticket: PAY-1024                      │
│                             │  Reason: ________________________________   │
│                             │  [Enter: Generate Draft]                    │
└─────────────────────────────┴──────────────────────────────────────────────┘
│ [↑/↓]: Move ticket  [Tab]: Switch pane  [Enter]: Activate  [q]: Quit      │
└────────────────────────────────────────────────────────────────────────────┘


┌────────────────────────────────────────────────────────────────────────────┐
│ Transaction Flow ▸ txn_abc_123_456                                        │
├────────────────────────────────────────────────────────────────────────────┤
│ 1. API_GATEWAY     → 200 OK   ts=2025-11-23T16:23:11Z                      │
│ 2. PAYMENT_CORE    → AUTH_OK  ts=2025-11-23T16:23:11Z                      │
│ 3. PAYNOW_ADAPTER  → ACCEPTED ts=2025-11-23T16:23:12Z                      │
│ 4. LEDGER_SERVICE  → POSTED   ts=2025-11-23T16:23:12Z                      │
│ 5. SETTLEMENT_SVC  → QUEUED   ts=2025-11-23T16:23:15Z                      │
│                                                                            │
│ Datadog Logs:                                                              │
│   payment-core:trace_id=abc-123  status=success                            │
│   paynow-adapter:txn_state=PENDING_REMOTE                                  │
│                                                                            │
│ Doorman Check (DB):                                                        │
│   txns.id = txn_abc_123_456  state='PENDING_SETTLEMENT'                    │
├────────────────────────────────────────────────────────────────────────────┤
│ [Esc]: Close  [d]: Open in Datadog  [s]: Generate Fix SQL (Doorman)       │
└────────────────────────────────────────────────────────────────────────────┘


┌────────────────────────────────────────────────────────────────────────────┐
│ Create Deregister PayNow SHIPRM                                           │
├────────────────────────────────────────────────────────────────────────────┤
│ From Jira Ticket: PAY-1042                                                │
│ Title: PayNow dereg stuck in PENDING                                      │
│                                                                            │
│ Customer Account:  ________________________________                        │
│ PayNow Proxy ID:  ________________________________                        │
│ Incident Summary:                                                         │
│   [PayNow deregistration request stuck in PENDING for > 30 minutes]       │
│                                                                            │
│ Business Impact:                                                          │
│   [Customer unable to deregister PayNow handle; risk of double mapping]   │
│                                                                            │
│ Requested Action:                                                         │
│   [Force deregister on provider + reconcile internal state]               │
│                                                                            │
├────────────────────────────────────────────────────────────────────────────┤
│ [Enter]: Generate SHIPRM Draft   [Esc]: Cancel (discard)                  │
└────────────────────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────────┐
│  ONCALL CLI  >  TEAM: PAYMENT                                 [ONLINE]   │
└──────────────────────────────────────────────────────────────────────────┘
┌─ [1] PENDING JIRA TICKETS (Active) ─┐┌─ [2] QUICK ACTIONS ───────────────┐
│                                     ││                                   │
│ > [PAY-1024] Fix double charge...   ││   1. View Transaction Flow        │
│   [PAY-1025] Update mTLS certs...   ││   2. Create Deregister PayNow...  │
│   [PAY-1030] Settlement delay...    ││   3. Database Maintenance         │
│                                     ││                                   │
│                                     ││                                   │
│                                     ││                                   │
│                                     ││                                   │
│                                     ││                                   │
└──────────────────────────────────────────────────────────────────────────┘
  [↑/↓: Navigate]  [Tab: Switch Pane]  [Enter: Details]  [q: Quit]


┌──────────────────────────────────────────────────────────────────────────┐
│  ONCALL CLI  >  TEAM: PAYMENT                                 [ONLINE]   │
└──────────────────────────────────────────────────────────────────────────┘
┌─ [1] PENDING JIRA TICKETS ──────────┐┌─ [2] QUICK ACTIONS (Active) ──────┐
│   [PAY-1024] Fix double charge...   ││   1. View Transaction Flow        │
│   [PAY-1025] Update mTLS certs...   ││ > 2. Create Deregister PayNow...  │
│   [PAY-1030] Settlement delay...    ││   3. Database Maintenance         │
│                                     ││                                   │
├─────────────────────────────────────┴────────────────────────────────────┤
│  INPUT REQUIRED                                                          │
│                                                                          │
│  Enter Transaction ID to trace flow:                                     │
│  > txn_sg_9988123_                                                       │
│                                                                          │
│  [Enter: Search Datadog]  [Esc: Cancel]                                  │
└──────────────────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────────────────────┐
│                         ONCALL CLI > TEAM: PAYMENT                                   │
└──────────────────────────────────────────────────────────────────────────────────────┘
┌───────────────────────────────┬──────────────────────────────────────────────────────┐
│  Pending Jira Tickets (3)      │                  Quick Actions                        │
│  (Use ↑/↓ to navigate)         │              (Press Tab to switch pane)                 │
├───────────────────────────────┼──────────────────────────────────────────────────────┤
│ > [PAY-1024] Fix double charge  │  [ View Transaction Flow ]                            │
│   race condition [CRITICAL]    │                                                        │
│   [PAY-1025] Update mTLS certs  │  [ Create Deregister PayNow SHIPRM ]                  │
│   for Provider X [HIGH]        │                                                        │
│   [PAY-1030] Investigate       │                                                        │
│   settlement delay [MED]        │  [ Refresh All Data ]                                 │
│                               │                                                        │
│                               │                                                        │
└───────────────────────────────┴──────────────────────────────────────────────────────┘
└──────────────────────────────────────────────────────────────────────────────────────┘
│ ↑/↓: Move | Tab: Switch Pane | Enter: Select | q: Quit                                 │
└──────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────┐
│                      Transaction Flow: txn_abc_123_456_                               │
├──────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                        │
│  1. [✓ SUCCESS] Payment Initiated                                                       │
│     Timestamp: 2023-10-27T10:00:01Z                                                    │
│     Source: [DATADOG LOG] - payment-gateway-service                                    │
│     Detail: "Received payment request for $50.00"                                      │
│                                                                                        │
│  2. [✓ SUCCESS] Provider Authorization                                                  │
│     Timestamp: 2023-10-27T10:00:03Z                                                    │
│     Source: [DATADOG LOG] - provider-x-adapter                                         │
│     Detail: "Auth code: 12345"                                                          │
│                                                                                        │
│  3. [✗ FAILURE] Ledger Update                                                          │
     Timestamp: 2023-10-27T10:00:05Z                                                    │
│     Source: [DOORMAN DB] - `transactions` table                                         │
│     Detail: "ERROR: Duplicate entry for idempotency_key 'xyz789'"                      │
│                                                                                        │
│                                                                                        │
│                                                                    [Esc: Back to Dashboard]│
└──────────────────────────────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────────────┐
│  ONCALL CLI  ▸  TEAM: PAYMENT                             ENV: PRODUCTION    │
├──────────────────────────────────────────────────────────────────────────────┤
│  1. INCIDENT QUEUE (Jira Module)    │  2. DIAGNOSTICS (Datadog + Doorman)    │
│  [Esc]: Unfocus Pane                │  [Tab]: Switch Pane                    │
├─────────────────────────────────────┼────────────────────────────────────────┤
│                                     │                                        │
│  > [PAY-1024] Fix double charge...  │  ┌─ Trace Transaction Flow ─────────┐  │
│    Status:  In Progress             │  │ Enter Txn ID:                    │  │
│    Prio:    CRITICAL                │  │ > txn_sg_998123________________  │  │
│    Assign:  oncall-payment          │  │                                  │  │
│                                     │  │ [Enter] Fetch Flow (DD + SQL)    │  │
│    [PAY-1025] Update mTLS certs...  │  └──────────────────────────────────┘  │
│    Status:  Open                    │                                        │
│    Prio:    HIGH                    │                                        │
│                                     │  3. ADMIN ACTIONS (Doorman/Jira)       │
│    [PAY-1030] Settlement delay...   │  ────────────────────────────────────  │
│    Status:  ToDo                    │  [ ] Create Deregister PayNow SHIPRM   │
│    Prio:    MED                     │  [ ] Force Unlock Transaction          │
│                                     │  [ ] Restart Settlement Consumer       │
│                                     │                                        │
│                                     │                                        │
│                                     │                                        │
│                                     │                                        │
│                                     │                                        │
│  [Enter]: Open Details              │  [↑/↓]: Select Action                  │
├─────────────────────────────────────┴────────────────────────────────────────┤
│  LOGS & FEEDBACK                                                             │
│  [16:04:22] INFO: Jira module loaded 3 tickets.                              │
│  [16:04:23] INFO: Datadog client connected. Waiting for input...             │
└──────────────────────────────────────────────────────────────────────────────┘
   ^N: New Ticket   ^F: Find Txn   ^Q: Quit   ?: Help
