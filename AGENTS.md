# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Build Commands

- `make deploy` - Build and install binaries to ~/bin

## Critical Conventions

- **Build-time credentials**: JIRA/Doorman credentials are injected via ldflags at build time, not from environment at runtime
- **File size limit**: Maximum 500 lines per source file
- **Post-implementation**: Always run `make build` and `make lint` after making changes
- **Singleton services**: TransactionQueryService uses sync.Once pattern - initialize once per process

## Architecture Notes

- Two binaries (mybuddy/sgbuddy) share same codebase, differentiated by BuildEnvironment constant
- Singapore environment does NOT use RPP adapter (nil in adapter set)
- Dual adapter locations: `internal/txn/adapters/` (application layer) and `internal/txn/service/adapters/` (data sources)
