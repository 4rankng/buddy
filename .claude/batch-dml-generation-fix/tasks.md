# Tasks: Batch DML Generation Fix

## Task 1: Update WriteSQLFiles to return created files list
**Goal**: Modify `WriteSQLFiles()` to track and return list of files created

**Steps**:
1. Open `/Users/dev/Documents/buddy/internal/txn/adapters/sql_writer.go`
2. Change function signature from `func WriteSQLFiles(...) error` to `func WriteSQLFiles(...) ([]string, error)`
3. Add `var filesCreated []string` at start of function
4. After each successful file write, append filename to `filesCreated`:
   - After writing PC_Deploy.sql: `filesCreated = append(filesCreated, "PC_Deploy.sql")`
   - After writing PC_Rollback.sql: `filesCreated = append(filesCreated, "PC_Rollback.sql")`
   - Repeat for PE, PPE, RPP Deploy and Rollback files
5. Return `filesCreated, nil` at end instead of just `nil`

**Done when**: Function returns list of created filenames

---

## Task 2: Update batch processor to clear old SQL files
**Goal**: Clear stale SQL files before generating new ones

**Steps**:
1. Open `/Users/dev/Documents/buddy/internal/apps/common/batch/processor.go`
2. After line 54 (after writing batch results text file), add new line:
   ```go
   adapters.ClearSQLFiles()
   ```

**Done when**: Old SQL files are removed before each batch run

---

## Task 3: Update batch processor to handle SQL file creation result
**Goal**: Display which SQL files were created and stop on error

**Steps**:
1. Open `/Users/dev/Documents/buddy/internal/apps/common/batch/processor.go`
2. Replace line 97 (`if err := adapters.WriteSQLFiles(statements, filePath)`) with:
   ```go
   filesCreated, err := adapters.WriteSQLFiles(statements, filePath)
   ```
3. Replace error handling at lines 97-99 with:
   ```go
   if err != nil {
       fmt.Printf("%sError writing SQL files: %v\n", appCtx.GetPrefix(), err)
       return
   }
   ```
4. After error handling (new line 100), add output to show created files:
   ```go
   if len(filesCreated) > 0 {
       fmt.Printf("%sSQL DML files generated: %v\n", appCtx.GetPrefix(), filesCreated)
   } else {
       fmt.Printf("%sNo SQL fixes required for these transactions.\n", appCtx.GetPrefix())
   }
   ```

**Done when**: Console shows which SQL files were created, and processing stops on error

---

## Task 4: Test the implementation
**Goal**: Verify SQL files are generated with proper output

**Steps**:
1. Build: `make deploy`
2. Run batch: `mybuddy txn TS-4558.txt` (or your test file)
3. Verify console shows: `SQL DML files generated: [PE_Deploy.sql, PE_Rollback.sql]`
4. Verify files exist in current directory
5. Check PE_Deploy.sql content:
   - Should contain workflow_execution UPDATE (state=221)
   - Should contain transfer table UPDATE with AuthorisationID
   - Should preserve updated_at timestamp
6. Verify no old/stale SQL statements in files

**Done when**: All verification steps pass

---

## Task 5: Verify rollback SQL generation
**Goal**: Ensure PE_Rollback.sql is also generated

**Steps**:
1. Check PE_Rollback.sql exists
2. Verify content contains workflow_execution rollback (state back to 102)
3. Verify AuthorisationID is set to NULL in rollback

**Done when**: Rollback SQL properly reverses deploy changes
