# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Debug Rules (Non-Obvious Only)

- Build-time credentials (JIRA/Doorman) are injected via ldflags - check `internal/buildinfo/constants.go` for available variables
- Singapore build (`make build-sg`) sets `RPPAdapter: nil` - any RPP-related code will fail if not nil-checked
- TransactionQueryService is a singleton using sync.Once - initialized once per process, not per request
- SQL generation uses template substitution with `%s` placeholders - not prepared statements, so debug generated SQL strings directly
- `config.Get()` returns build-time constants, NOT runtime environment values - don't expect `os.Getenv` behavior
- Database queries return `[]map[string]interface{}` - use `utils.GetStringValue(row, "key")` for safe access
- Workflow states are numeric strings in database - use `FormatWorkflowState(workflowID, state)` for human-readable output
