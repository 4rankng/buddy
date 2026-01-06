# Requirements Document

## Introduction

This feature adds automatic handling for RPP adapter workflows of type `wf_ct_cashin` that are stuck in state 100 (`stTransferPersisted`). When such workflows are detected, the system will automatically generate a deploy script to update the timestamp and set `attempt=1`, resolving optimistic lock failures and allowing the workflow to retry and progress from the transfer persistence state.

## Glossary

- **RPP_Adapter**: Real-time Payment Processing adapter component that handles payment workflows
- **Workflow_Execution**: Database table storing workflow execution state and metadata
- **wf_ct_cashin**: Specific workflow type for processing cash-in operations in RPP
- **stTransferPersisted**: State 100 representing successful transfer persistence but stuck before next transition
- **Deploy_Script**: SQL script that modifies database state to fix stuck workflows
- **Rollback_Script**: SQL script that reverses changes made by deploy script
- **DML_Ticket**: Data Manipulation Language ticket containing deploy and rollback scripts
- **Optimistic_Lock**: Concurrency control mechanism using timestamps to prevent conflicting updates

## Requirements

### Requirement 1: Detect RPP Cashin Transfer Persisted State

**User Story:** As a system operator, I want the system to automatically detect `wf_ct_cashin` workflows stuck in state 100, so that I can apply appropriate fixes without manual identification.

#### Acceptance Criteria

1. WHEN analyzing transaction results, THE System SHALL identify RPP adapter workflows with workflow_id `wf_ct_cashin`
2. WHEN a `wf_ct_cashin` workflow has state 100 and attempt 0, THE System SHALL flag it as requiring timestamp update fix
3. WHEN multiple `wf_ct_cashin` workflows exist in state 100 with attempt 0, THE System SHALL handle each workflow individually
4. THE System SHALL extract the run_id from qualifying workflows for use in SQL generation

### Requirement 2: Generate Deploy Script for Timestamp and Attempt Update

**User Story:** As a system operator, I want the system to generate a deploy script that updates the timestamp and sets attempt=1 for stuck `wf_ct_cashin` workflows, so that optimistic lock failures are resolved and the workflow can retry.

#### Acceptance Criteria

1. WHEN a `wf_ct_cashin` workflow is in state 100 with attempt 0, THE System SHALL generate a deploy script
2. THE Deploy_Script SHALL update the workflow_execution table
3. THE Deploy_Script SHALL set attempt=1 for the specific run_id
4. THE Deploy_Script SHALL set updated_at=NOW() to resolve timestamp conflicts
5. THE Deploy_Script SHALL update the JSON data field to set State=100
6. THE Deploy_Script SHALL target workflows where workflow_id='wf_ct_cashin' AND state=100
7. THE Deploy_Script SHALL include the specific run_id in the WHERE clause for safety
8. THE Deploy_Script SHALL include descriptive comments explaining the optimistic lock fix purpose

### Requirement 3: Generate Rollback Script for Recovery

**User Story:** As a system operator, I want a rollback script that can reverse the timestamp and attempt updates, so that I can undo changes if the fix causes issues.

#### Acceptance Criteria

1. WHEN generating a deploy script, THE System SHALL also generate a corresponding rollback script
2. THE Rollback_Script SHALL reset attempt back to 0 for the same run_id
3. THE Rollback_Script SHALL preserve the original updated_at timestamp if possible
4. THE Rollback_Script SHALL target the same workflow_id and run_id as the deploy script
5. THE Rollback_Script SHALL include descriptive comments explaining the rollback purpose

### Requirement 4: Integrate with Existing Template System

**User Story:** As a developer, I want the new functionality to integrate seamlessly with the existing SQL template system, so that it follows established patterns and conventions.

#### Acceptance Criteria

1. THE System SHALL create a new Case type for this cashin scenario
2. THE System SHALL register the new template function in the RPP templates map
3. THE System SHALL follow the existing TemplateFunc signature pattern
4. THE System SHALL return a DMLTicket with properly formatted TemplateInfo structures
5. THE System SHALL use the existing helper functions for run_id extraction
6. THE System SHALL target the "RPP" database for all generated SQL statements

### Requirement 5: Validate Input Data and State Conditions

**User Story:** As a system operator, I want the system to validate input data and state conditions before generating scripts, so that invalid or inappropriate data doesn't result in incorrect SQL generation.

#### Acceptance Criteria

1. WHEN RPPAdapter data is nil, THE System SHALL return nil (no ticket generated)
2. WHEN no workflows match the criteria, THE System SHALL return nil (no ticket generated)
3. WHEN run_id is empty or invalid, THE System SHALL return nil (no ticket generated)
4. THE System SHALL validate that workflow_id exactly matches 'wf_ct_cashin'
5. THE System SHALL validate that state exactly matches '100'
6. THE System SHALL validate that attempt equals 0 (stuck at initial attempt)

### Requirement 6: Handle JSON Data Field Updates

**User Story:** As a system operator, I want the system to properly update the JSON data field in workflow_execution, so that the workflow state remains consistent across all data representations.

#### Acceptance Criteria

1. WHEN updating workflow state, THE System SHALL use JSON_SET to update the data field
2. THE Deploy_Script SHALL set the State property in the JSON data to 100
3. THE System SHALL preserve other JSON data properties during the update
4. THE Rollback_Script SHALL restore the original JSON data state if modified
