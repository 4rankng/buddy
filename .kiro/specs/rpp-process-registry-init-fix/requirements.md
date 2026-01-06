# Requirements Document

## Introduction

This feature adds automatic handling for RPP adapter workflows of type `wf_process_registry` that are stuck in state 0 (`stInit`). When such workflows are detected, the system will automatically generate a deploy script to set `attempt=1`, allowing the workflow to retry and potentially recover from the initialization state.

## Glossary

- **RPP_Adapter**: Real-time Payment Processing adapter component that handles payment workflows
- **Workflow_Execution**: Database table storing workflow execution state and metadata
- **wf_process_registry**: Specific workflow type for processing registry operations in RPP
- **stInit**: Initial state (state 0) of a workflow execution
- **Deploy_Script**: SQL script that modifies database state to fix stuck workflows
- **Rollback_Script**: SQL script that reverses changes made by deploy script
- **DML_Ticket**: Data Manipulation Language ticket containing deploy and rollback scripts

## Requirements

### Requirement 1: Detect RPP Process Registry Init State

**User Story:** As a system operator, I want the system to automatically detect `wf_process_registry` workflows stuck in state 0, so that I can apply appropriate fixes without manual identification.

#### Acceptance Criteria

1. WHEN analyzing transaction results, THE System SHALL identify RPP adapter workflows with workflow_id `wf_process_registry`
2. WHEN a `wf_process_registry` workflow has state 0, THE System SHALL flag it as requiring initialization fix
3. WHEN multiple `wf_process_registry` workflows exist in state 0, THE System SHALL handle each workflow individually
4. THE System SHALL extract the run_id from qualifying workflows for use in SQL generation

### Requirement 2: Generate Deploy Script for Attempt Reset

**User Story:** As a system operator, I want the system to generate a deploy script that sets attempt=1 for stuck `wf_process_registry` workflows, so that the workflow can retry initialization.

#### Acceptance Criteria

1. WHEN a `wf_process_registry` workflow is in state 0, THE System SHALL generate a deploy script
2. THE Deploy_Script SHALL update the workflow_execution table
3. THE Deploy_Script SHALL set attempt=1 for the specific run_id
4. THE Deploy_Script SHALL target workflows where workflow_id='wf_process_registry' AND state=0
5. THE Deploy_Script SHALL include the specific run_id in the WHERE clause for safety
6. THE Deploy_Script SHALL include descriptive comments explaining the fix purpose

### Requirement 3: Generate Rollback Script for Recovery

**User Story:** As a system operator, I want a rollback script that can reverse the attempt reset, so that I can undo changes if the fix causes issues.

#### Acceptance Criteria

1. WHEN generating a deploy script, THE System SHALL also generate a corresponding rollback script
2. THE Rollback_Script SHALL reset attempt back to 0 for the same run_id
3. THE Rollback_Script SHALL target the same workflow_id and run_id as the deploy script
4. THE Rollback_Script SHALL include descriptive comments explaining the rollback purpose

### Requirement 4: Integrate with Existing Template System

**User Story:** As a developer, I want the new functionality to integrate seamlessly with the existing SQL template system, so that it follows established patterns and conventions.

#### Acceptance Criteria

1. THE System SHALL create a new Case type for this scenario
2. THE System SHALL register the new template function in the RPP templates map
3. THE System SHALL follow the existing TemplateFunc signature pattern
4. THE System SHALL return a DMLTicket with properly formatted TemplateInfo structures
5. THE System SHALL use the existing helper functions for run_id extraction
6. THE System SHALL target the "RPP" database for all generated SQL statements

### Requirement 5: Validate Input Data

**User Story:** As a system operator, I want the system to validate input data before generating scripts, so that invalid or missing data doesn't result in incorrect SQL generation.

#### Acceptance Criteria

1. WHEN RPPAdapter data is nil, THE System SHALL return nil (no ticket generated)
2. WHEN no workflows match the criteria, THE System SHALL return nil (no ticket generated)
3. WHEN run_id is empty or invalid, THE System SHALL return nil (no ticket generated)
4. THE System SHALL validate that workflow_id exactly matches 'wf_process_registry'
5. THE System SHALL validate that state exactly matches '0'
