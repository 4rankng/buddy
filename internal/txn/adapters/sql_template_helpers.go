package adapters

import (
	"buddy/internal/txn/domain"
)

// getRPPWorkflowRunIDByCriteria finds and returns the run_id of a workflow matching specific criteria.
// Parameters:
//   - workflows: slice of workflows to search
//   - workflowID: workflow_id to match (empty string means any workflow_id)
//   - state: state to match (empty string means any state)
//   - attempt: attempt number to match (-1 means any attempt)
//
// Returns empty string if no matching workflow is found.
func getRPPWorkflowRunIDByCriteria(workflows []domain.WorkflowInfo, workflowID, state string, attempt int) string {
	for _, wf := range workflows {
		// Check workflow_id if specified
		if workflowID != "" && wf.WorkflowID != workflowID {
			continue
		}
		// Check state if specified
		if state != "" && wf.State != state {
			continue
		}
		// Check attempt if specified (and not -1 which means any attempt)
		if attempt != -1 && wf.Attempt != attempt {
			continue
		}
		// All criteria matched
		return wf.RunID
	}
	// No matching workflow found
	return ""
}
