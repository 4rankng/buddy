Migration Plan: Move txn Package to internal/shared

 Overview

 Move the entire txn/ package from the project root to internal/shared/txn/ to properly encapsulate it as internal
 infrastructure while maintaining its shared usage between mybuddy and sgbuddy.

 Current State Analysis

 Based on the exploration, the txn package contains:
 - 13 Go files handling transaction management, SOP case detection, and SQL generation
 - Used by 4 files across internal/mybuddy (3 files) and internal/sgbuddy (1 file)
 - Supports both "my" (Malaysia) and "sg" (Singapore) environments via explicit parameters
 - Well-designed WithEnv function pattern for environment handling

 Migration Steps

 1. Create Directory Structure

 mkdir -p internal/shared/txn

 2. Move All Files

 Move all 13 files from txn/ to internal/shared/txn/:
 - batch.go
 - cases.go
 - classification.go
 - classification_test.go
 - output.go
 - query.go
 - sql.go
 - sql_generator.go
 - sql_templates.go
 - sql_writer.go
 - stats.go
 - types.go

 3. Update Import Paths

 In internal/mybuddy (3 files to update):
 - internal/mybuddy/txn.go
 - internal/mybuddy/rpp_resume.go
 - internal/mybuddy/ecotxn.go

 Change import from:
 "buddy/txn"
 To:
 "buddy/internal/shared/txn"

 In internal/sgbuddy (1 file to update):
 - internal/sgbuddy/txn.go

 Change import from:
 "buddy/txn"
 To:
 "buddy/internal/shared/txn"

 4. No Code Changes Required

 - The existing WithEnv function pattern already handles environment differentiation
 - All function calls remain identical
 - No behavioral changes needed

 Verification Steps

 1. Run make lint to ensure no formatting issues
 2. Run make build to verify compilation
 3. Test both mybuddy and sgbuddy transaction commands
 4. Verify batch processing works for both environments

 Benefits

 1. Proper Encapsulation: txn becomes properly internal to the project
 2. Clear Intent: internal/shared signals this is shared infrastructure
 3. Maintains Reusability: Both tools continue sharing the same logic
 4. Zero Breaking Changes: All functionality remains identical
 5. Follows Go Best Practices: Internal packages belong under internal/

 Rollback Plan

 If issues arise:
 1. Move files back from internal/shared/txn/ to txn/
 2. Revert import paths in the 4 affected files
 3. Run make lint and make build to verify

 Critical File Paths

 Files to move:
 - /Users/dev/Documents/buddy/txn/* → /Users/dev/Documents/buddy/internal/shared/txn/*

 Files to update imports:
 - /Users/dev/Documents/buddy/internal/mybuddy/txn.go
 - /Users/dev/Documents/buddy/internal/mybuddy/rpp_resume.go
 - /Users/dev/Documents/buddy/internal/mybuddy/ecotxn.go
 - /Users/dev/Documents/buddy/internal/sgbuddy/txn.go
