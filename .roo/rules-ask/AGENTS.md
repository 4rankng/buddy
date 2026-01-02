# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Documentation Rules (Non-Obvious Only)

- Two binaries (`mybuddy` and `sgbuddy`) share same codebase - only differ by build-time `BuildEnvironment` constant
- Singapore environment intentionally has no RPP adapter (nil) - RPP-related features are Malaysia-only
- `internal/txn/adapters/` contains application-facing adapters (SQL generation, output), while `internal/txn/service/adapters/` contains data source implementations
- Workflow states are stored as numeric strings in database but displayed via `FormatWorkflowState()` using `domain.WorkflowStateMaps`
- SOP documentation in `docs/sops/` drives SQL generation - not just documentation but executable rules
- Build-time credentials are injected via ldflags - see `internal/buildinfo/constants.go` for available variables
