package adapters

import (
	"buddy/internal/txn/domain"
	"encoding/json"
	"strings"
)

// SOPRepository manages SOP case rules and identification
type SOPRepository struct {
	rules []CaseRule
}

// Global SOPRepo instance (singleton)
var SOPRepo = NewSOPRepository()

// NewSOPRepository creates a new SOP repository with predefined rules
func NewSOPRepository() *SOPRepository {
	return &SOPRepository{
		rules: getDefaultSOPRules(),
	}
}

// IdentifyCase identifies SOP case for a transaction result
func (r *SOPRepository) IdentifyCase(result *domain.TransactionResult, env string) domain.Case {
	// Check if we've already identified case
	if result.CaseType != domain.CaseNone && result.CaseType != "" {
		return result.CaseType
	}

	// Special handling for cash-in stuck at state 100 with timestamp analysis
	if caseType := r.identifyCashInStuck100Case(result, env); caseType != domain.CaseNone {
		result.CaseType = caseType
		return result.CaseType
	}

	// Check each rule in order
	for _, rule := range r.rules {
		// Skip country-specific rules if not matching
		if rule.Country != "" && rule.Country != env {
			continue
		}

		if r.evaluateRule(rule, result) {
			result.CaseType = rule.CaseType
			return result.CaseType
		}
	}

	result.CaseType = domain.CaseNone
	return result.CaseType
}

// identifyCashInStuck100Case performs timestamp analysis for cash-in workflows stuck at state 100
func (r *SOPRepository) identifyCashInStuck100Case(result *domain.TransactionResult, env string) domain.Case {
	// Only process Malaysia environment for now
	if env != "my" {
		return domain.CaseNone
	}

	// Check if this is a cash-in workflow stuck at state 100 with attempts > 0
	if result.RPPAdapter == nil {
		return domain.CaseNone
	}

	// Find workflow matching cash-in state 100 criteria
	var targetWorkflow *domain.WorkflowInfo
	for _, wf := range result.RPPAdapter.Workflow {
		if wf.WorkflowID == "wf_ct_cashin" && wf.State == "100" && wf.Attempt > 0 {
			targetWorkflow = &wf
			break
		}
	}

	if targetWorkflow == nil {
		return domain.CaseNone
	}

	// Extract timestamp from workflow data
	workflowUpdatedAt, err := r.extractWorkflowTimestamp(targetWorkflow.Data)
	if err != nil {
		// If we can't extract timestamp, default to retry case
		return domain.CaseCashInStuck100Retry
	}

	// For now, we don't have access to credit_transfer.updated_at in the TransactionResult
	// In a real implementation, this would need to be populated by the query strategy
	// For demonstration purposes, we'll use a placeholder logic

	// TODO: This needs to be enhanced to:
	// 1. Query credit_transfer table to get updated_at (UTC)
	// 2. Compare with workflow data UpdatedAt (GMT+8) using timezone conversion
	// 3. Return CaseCashInStuck100Retry if timestamps match after conversion
	// 4. Return CaseCashInStuck100UpdateMismatch if timestamps don't match

	// For now, default to retry case if we have workflow timestamp
	if workflowUpdatedAt != "" {
		return domain.CaseCashInStuck100Retry
	}

	return domain.CaseNone
}

// extractWorkflowTimestamp extracts the CreditTransfer.UpdatedAt timestamp from workflow data JSON
func (r *SOPRepository) extractWorkflowTimestamp(workflowData string) (string, error) {
	if workflowData == "" {
		return "", nil
	}

	// Parse the workflow data JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(workflowData), &data); err != nil {
		return "", err
	}

	// Navigate to CreditTransfer.UpdatedAt
	if creditTransfer, ok := data["CreditTransfer"].(map[string]interface{}); ok {
		if updatedAt, ok := creditTransfer["UpdatedAt"].(string); ok {
			return strings.TrimSpace(updatedAt), nil
		}
	}

	return "", nil
}
