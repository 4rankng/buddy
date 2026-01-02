# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Coding Rules (Non-Obvious Only)

- Always use `config.Get(key, defaultValue)` instead of `os.Getenv` - returns build-time constants, not runtime environment variables
- SQL templates use `%s` positional placeholders with `buildSQLFromTemplate` - NOT prepared statements
- `TransactionResult.CaseType` stores identified SOP case to avoid re-identification (don't re-run `IdentifyCase`)
- Workflow state mappings in `domain.WorkflowStateMaps` - use `FormatWorkflowState` for display
- When adding new adapters, place data source implementations in `internal/txn/service/adapters/`, application-facing adapters in `internal/txn/adapters/`
- Singapore environment sets `RPPAdapter: nil` - always check for nil before using
- Error wrapping: use `fmt.Errorf("message: %w", err)` pattern
- Maximum 500 lines per source file - split files if exceeded
- Use `utils.GetStringValue(row, "key")` for safe map access from database results
