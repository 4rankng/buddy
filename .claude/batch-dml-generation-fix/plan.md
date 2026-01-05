# Plan: Batch DML Generation Fix

## High-Level Approach
The batch processor already has SQL generation infrastructure in place. The issue is that **the SQL writer doesn't produce console output** showing which files were created. We need to:

1. **Enhance console output** in `WriteSQLFiles()` to show which files were generated
2. **Verify transfer table updates** are being generated (logic exists in sql.go:64-72)
3. **Clear old SQL files** before generating new ones (prevent stale data)
4. **Stop processing on SQL write errors** (per user requirement)

## Architecture / Modules Touched

### Primary Files:
1. **internal/txn/adapters/sql_writer.go**
   - Modify `WriteSQLFiles()` to return list of created files
   - Add console output showing generated files
   - Return error if write fails

2. **internal/apps/common/batch/processor.go**
   - Add call to `ClearSQLFiles()` before generation
   - Update error handling to stop on SQL write failures
   - Display generated file list to user

### Secondary Files (verification only):
3. **internal/txn/adapters/sql.go** (lines 64-128)
   - Verify transfer update generation logic is working
   - Logic already exists - no changes needed

4. **internal/txn/adapters/sql_templates_pe_basic.go** (lines 58-102)
   - Verify pe_stuck_at_limit_check_102_4 template
   - Template already generates workflow_execution UPDATE
   - Transfer update handled separately in sql.go

## Data/Model Changes
None - data structures remain the same

## API/UX Behavior Changes
- **Before**: Silent SQL file generation, user doesn't know what was created
- **After**: Clear console output:
  ```
  SQL DML files generated: [PE_Deploy.sql, PE_Rollback.sql]
  ```
- **Error handling**: Stop processing on SQL write failure (per user requirement)

## Testing Strategy
1. **Manual test**: Run `mybuddy txn TS-4558.txt`
2. **Verify**: PE_Deploy.sql and PE_Rollback.sql created
3. **Verify**: Console shows generated file list
4. **Verify**: workflow_execution UPDATE statement present
5. **Verify**: transfer table UPDATE with AuthorisationID present
6. **Verify**: updated_at timestamp preserved

## Migration/Rollout Steps
1. Make code changes
2. Rebuild: `make deploy`
3. Test with TS-4558.txt
4. Verify console output and SQL file content
5. **Rollback**: Revert processor.go and sql_writer.go changes if issues arise

## Observability
- Console output shows which SQL files were created
- Error messages displayed if SQL write fails
- Summary shows count of statements per database

## Alternatives Considered

### Alternative 1: Append mode (rejected)
- Keep old SQL files and append new ones
- **Rejected**: Confusing, hard to distinguish runs
- **Chosen**: Clear files first, start fresh

### Alternative 2: Continue on error (rejected)
- Log error but continue processing
- **Rejected**: Per user requirement - show error and stop
- **Chosen**: Return error immediately, stop batch processing

### Alternative 3: Generate to subdirectory (rejected)
- Create TS-4558/PE_Deploy.sql
- **Rejected**: Over-engineering for this use case
- **Chosen**: Write to current directory with clear filenames

## Critical Implementation Details

### Change 1: sql_writer.go - Return created files list
```go
func WriteSQLFiles(statements domain.SQLStatements, basePath string) ([]string, error) {
    var filesCreated []string

    // Track each file written
    if len(statements.PEDeployStatements) > 0 {
        filesCreated = append(filesCreated, "PE_Deploy.sql")
        // ... write file
    }
    // ... repeat for all DB types

    return filesCreated, nil
}
```

### Change 2: processor.go - Clear files first
```go
// After writing text results (line 54)
adapters.ClearSQLFiles()

// Generate SQL statements (line 57)
statements := adapters.GenerateSQLStatements(results)

// Write SQL files (line 97)
filesCreated, err := adapters.WriteSQLFiles(statements, filePath)
if err != nil {
    fmt.Printf("%sError writing SQL files: %v\n", appCtx.GetPrefix(), err)
    return  // Stop processing per user requirement
}
```

### Change 3: processor.go - Display generated files
```go
if len(filesCreated) > 0 {
    fmt.Printf("%sSQL DML files generated: %v\n", appCtx.GetPrefix(), filesCreated)
} else {
    fmt.Printf("%sNo SQL fixes required for these transactions.\n", appCtx.GetPrefix())
}
```

## Files to Modify
1. `/Users/dev/Documents/buddy/internal/txn/adapters/sql_writer.go` - Update WriteSQLFiles signature
2. `/Users/dev/Documents/buddy/internal/apps/common/batch/processor.go` - Add ClearSQLFiles call and error handling
