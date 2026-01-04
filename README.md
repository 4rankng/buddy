# Buddy

Buddy is a Go-based on-call companion for G-Bank payment operations. The
`mybuddy` and `sgbuddy` binaries wrap common transaction-health checks, SOP
verification, remediation SQL generation, and JIRA workflows into a single CLI so
engineers can work quickly during incidents in either the Malaysia or Singapore
region.

## Binaries

| Name | Description |
| --- | --- |
| `mybuddy` | Malaysia-focused CLI entry point. Initializes the Malaysia environment, JIRA/Doorman clients, and exposes commands tuned for MY payment flows. |
| `sgbuddy` | Singapore-focused CLI entry point. Shares most commands with `mybuddy` but runs with the Singapore build-time environment and credentials. |

Both binaries share the same codebase and only differ by the build-time
constants injected via `ldflags`.

## Setup & Build

1. Copy `env_example` to `.env.my` and `.env.sg`, then populate the required
   secrets (JIRA + Doorman credentials per region).
2. Ensure Go 1.21+ is installed.
4. Run `make lint` before sending a change to catch formatting or vet issues.
5. Optionally run `make deploy` to copy the binaries into `~/bin` for global
   use; `make help` lists every available target.

## Project Layout

```
buddy/
├── cmd/                 # Region-specific entrypoints
├── internal/            # Application code grouped by domain
│   ├── apps/            # CLI commands and shared app context
│   ├── buildinfo/       # Build-time constants
│   ├── clients/         # Outbound integrations (JIRA, Doorman, DB clients)
│   ├── config/          # Runtime configuration wrapper
│   ├── txn/             # Transaction remediation domain logic
│   └── ui/              # Terminal utilities & interactive pickers
├── docs/                # SOPs and architectural notes
├── plans/               # Design documents for ongoing work
├── bin/                 # Build artifacts (ignored in CI/CD)
├── Makefile             # Build, lint, and deploy automation
├── go.mod / go.sum      # Module definition and dependencies
└── TS-4475*.txt         # Sample output from regression experiments
```

## Directory and Package Guide

### `cmd/`
- `cmd/mybuddy/main.go` wires the Malaysia environment: loads config, builds the
  root Cobra command, registers MY-specific subcommands, and executes the CLI.
- `cmd/sgbuddy/main.go` mirrors the above but injects the Singapore environment
  before executing.

### `internal/apps`
- `common/app.go` defines the shared `Context` (environment prefix helpers,
  binary name) and is used by every command tree.
- `common/cobra/root.go` creates the base Cobra root command, root `version`
  subcommand, and user-facing description text.
- `common/jira/common.go` contributes `buddy jira search` which both binaries
  reuse to search unresolved tickets with consistent validation.
- `mybuddy/commands/commands.go` aggregates all Malaysia CLI commands.
  - `txn.go` powers `mybuddy txn`, supporting transaction lookup by ID/E2E ID or
    file, SOP case verification, and remediation SQL generation.
  - `rpp.go` exposes the `mybuddy rpp` command group and nests the resume
    workflow.
  - `rpp_resume.go` implements `mybuddy rpp resume` to inspect RPP adapter
    records, classify SOP cases, and emit deploy/rollback SQL for workflow
    resumption.
  - `ecotxn.go` adds `mybuddy ecotxn` for PartnerPay (Eco) workflow inspection by
    `run_id`.
  - `jira.go` adds `mybuddy jira list` plus wires in the common search command
    and interactive picker for Malaysia engineers.
- `sgbuddy/commands` mirrors the Malaysia structure but only registers the
  command set needed for Singapore on-call (transaction, Eco transaction, JIRA).

### `internal/clients`
- `jira/`
  - `jira_config.go` and `jira_client.go` describe the REST client surface used
    by commands.
  - `jira_init.go` exposes a singleton initializer so both binaries share one
    configured client per process.
  - `jira_interface.go` documents the interface consumed by the UI layer for
    listing, searching, and downloading attachments.
- `doorman/` contains the Doorman authentication client definitions used to reach
  protected databases.

### `internal/config` & `internal/buildinfo`
- `internal/buildinfo/constants.go` holds the build-time variables injected via
  `ldflags` (JIRA/Doorman credentials and the target environment) and validates
  that they were provided.
- `internal/config/config.go` loads those constants once per process and provides
  getters (including fallbacks for optional values) so the rest of the codebase
  never touches `os.Getenv` directly.

### `internal/txn`
- `service/`
  - `transaction_service.go`, `query.go`, `batch.go`, and `client.go` coordinate
    transaction lookups, batch file processing, and shared client plumbing.
  - `service/adapters/*.go` houses low-level data sources: `payment_engine.go`
    fetches transfer/workflow state from the Payment Engine, `partnerpay_engine.go`
    handles PartnerPay lookups, `payment_core.go` talks to Payment Core tables,
    `fast_adapter.go` queries FAST transactions, and `rpp_adapter.go` provides
    the Malaysia-only adapter and E2E-ID inspection logic.
- `adapters/` (note the different path) is the application-facing output layer:
  SQL generation (`sql_generator.go`, `sql_templates.go`), SOP ticket helpers
  (`sop_repository.go`), file writers (`output.go`, `sql_writer.go`), and ECO txn
  publishers.
- `domain/` captures the typed structs (`types.go`) and classification helpers
  (`classification.go`) that represent a transaction plus SOP metadata.
- `ports/` defines the interfaces implemented by the various DB clients so that
  services can be unit tested.
- `utils/` offers shared helpers such as `file.go` (reading transaction IDs from
  batch files) and `query.go` (SQL row parsing helpers).

### `internal/ui`
- `formatter.go` truncates/word-wraps text and formats ANSI hyperlinks for
  terminal output.
- `jira_picker.go` implements the interactive ticket chooser, details view, and
  attachment downloader used by `buddy jira list`.
- `terminal.go` contains terminal detection, hyperlink helpers, and cross-platform
  browser launching logic.

### Documentation & Plans
- `docs/plans/*.md` capture research notes (JIRA client, Txn result format,
  specific investigations). Use these when onboarding or spec’ing new work.
- `docs/sops/MY_DML_SOP.md` and `docs/sops/SG_DML_SOP.md` are the playbooks that
  drive the auto-generated SQL.
- `plans/TxnResultOutputFormat.md` in the repository root is the working draft
  for the CLI output redesign.

### Supporting Files
- `env_example` shows the required secrets for each `.env.<region>` file.
- `bin/` stores compiled binaries; it is safe to delete/rebuild as needed.
- `TS-4475.txt` and `TS-4475.txt-output.txt` contain recorded sample runs that
  help verify regression reproductions.
- `CLAUDE.md` summarises local repo conventions (e.g., preferred tooling).

Use this guide as a reference when navigating the codebase or onboarding new
contributors; update it whenever packages move or new commands are introduced.
