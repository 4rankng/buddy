package adapters

import (
	"buddy/internal/txn/domain"
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

// IdentifyCase identifies the SOP case for a transaction result
func (r *SOPRepository) IdentifyCase(result *domain.TransactionResult, env string) domain.Case {
	// Check if we've already identified the case
	if result.CaseType != domain.CaseNone {
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
