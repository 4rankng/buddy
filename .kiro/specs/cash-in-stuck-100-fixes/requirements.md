# Requirements Document

## Introduction

This feature implements two specialized fixes for cash-in workflows that get stuck at state 100 due to different types of timestamp-related issues. The system needs to handle both simple retry cases (where timestamps match after timezone conversion) and update mismatch cases (where actual timestamp synchronization is required).

## Glossary

- **Workflow_Execution**: The database table storing workflow execution state and data
- **Credit_Transfer**: The database table storing credit transfer records with UTC timestamps
- **State_100**: The specific workflow state where cash-in operations can get stuck
- **Run_ID**: Unique identifier for a workflow execution instance
- **UpdatedAt_Mismatch**: Condition where credit_transfer.updated_at (UTC) doesn't match workflow_execution.data UpdatedAt (GMT+8) even after timezone conversion
- **Timezone_Conversion**: Process of converting UTC timestamps to GMT+8 by adding 8 hours
- **RPP_Database**: The target database where these workflow fixes are applied

## Requirements

### Requirement 1: Simple Retry Fix

**User Story:** As an operations engineer, I want to retry stuck cash-in workflows when timestamps match after timezone conversion, so that I can resolve simple retry cases without data modification.

#### Acceptance Criteria

1. WHEN a cash-in workflow is stuck at state 100 with attempts, THE System SHALL identify if this is a simple retry case
2. WHEN timestamps match after timezone conversion from UTC to GMT+8, THE System SHALL classify this as a retry case
3. WHEN applying the retry fix, THE System SHALL set attempt=1 without modifying timestamp data
4. THE System SHALL target workflows with workflow_id='wf_ct_cashin' and state=100
5. WHEN the retry fix is applied, THE System SHALL generate the appropriate SQL UPDATE statement

### Requirement 2: Update Mismatch Fix

**User Story:** As an operations engineer, I want to fix cash-in workflows with actual timestamp mismatches, so that I can synchronize workflow data and resolve stuck operations.

#### Acceptance Criteria

1. WHEN a cash-in workflow is stuck at state 100 with attempts, THE System SHALL identify if this is an update mismatch case
2. WHEN timestamps don't match even after timezone conversion, THE System SHALL classify this as an update mismatch case
3. WHEN applying the mismatch fix, THE System SHALL set attempt=1 AND update the workflow data timestamp
4. THE System SHALL convert credit_transfer.updated_at from UTC to GMT+8 for workflow data synchronization
5. THE System SHALL use JSON_SET to update the $.CreditTransfer.UpdatedAt field in workflow data
6. WHEN the mismatch fix is applied, THE System SHALL generate the appropriate SQL UPDATE statement with converted timestamp

### Requirement 3: Timestamp Conversion and Validation

**User Story:** As an operations engineer, I want accurate timezone conversion between UTC and GMT+8, so that timestamp comparisons and updates are performed correctly.

#### Acceptance Criteria

1. WHEN converting timestamps, THE System SHALL add 8 hours to UTC timestamps to get GMT+8
2. WHEN comparing timestamps, THE System SHALL account for the timezone difference between credit_transfer (UTC) and workflow data (GMT+8)
3. THE System SHALL validate timestamp format before performing conversions
4. WHEN timezone conversion fails, THE System SHALL return an appropriate error message

### Requirement 4: SQL Generation and Safety

**User Story:** As an operations engineer, I want safe and accurate SQL generation for workflow fixes, so that database operations are performed correctly without data corruption.

#### Acceptance Criteria

1. THE System SHALL generate SQL UPDATE statements targeting the workflow_execution table
2. WHEN generating SQL, THE System SHALL include proper WHERE clauses with run_id, workflow_id, and state conditions
3. THE System SHALL validate run_id format before including in SQL statements
4. WHEN generating mismatch fix SQL, THE System SHALL properly escape and format the converted timestamp
5. THE System SHALL include comments in generated SQL indicating the fix type being applied

### Requirement 5: Fix Type Detection and Routing

**User Story:** As an operations engineer, I want automatic detection of fix types, so that the appropriate resolution is applied based on the specific timestamp condition.

#### Acceptance Criteria

1. WHEN analyzing a stuck workflow, THE System SHALL compare credit_transfer.updated_at with workflow data UpdatedAt
2. WHEN timestamps match after timezone conversion, THE System SHALL recommend the retry fix
3. WHEN timestamps don't match after timezone conversion, THE System SHALL recommend the mismatch fix
4. THE System SHALL provide clear indication of which fix type is being applied
5. WHEN fix type cannot be determined, THE System SHALL request additional information or manual review
