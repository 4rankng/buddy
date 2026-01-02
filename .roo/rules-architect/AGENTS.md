# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Architecture Rules (Non-Obvious Only)

- Two binaries (`mybuddy`/`sgbuddy`) share identical codebase - only differentiated by `BuildEnvironment` constant set at build time
- Singapore environment intentionally omits RPP adapter (`RPPAdapter: nil`) - any RPP-related features must be Malaysia-gated
- Dual adapter pattern: `internal/txn/service/adapters/` (data sources implementing ports) vs `internal/txn/adapters/` (application-facing SQL generation, output)
- TransactionQueryService uses singleton pattern with sync.Once - initialize once per process, not per request
- SQL generation uses template substitution (`%s` placeholders) rather than prepared statements - templates defined in `internal/txn/adapters/sql_templates*.go`
- SOP cases are cached in `TransactionResult.CaseType` after identification to avoid re-running `IdentifyCase()`
- Workflow state display requires `FormatWorkflowState(workflowID, state)` using `domain.WorkflowStateMaps` for human-readable output
