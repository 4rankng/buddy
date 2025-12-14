package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// CaseRule defines a rule for identifying SOP cases
type CaseRule struct {
	CaseType    domain.Case
	Description string
	Country     string // optional: "", "my", "sg" for country-specific rules
	Conditions  []RuleCondition
}

// RuleCondition defines a single condition in a rule
type RuleCondition struct {
	FieldPath string      // e.g., "PaymentEngine.Workflow.State"
	Operator  string      // eq, ne, lt, gt, in, not_in, regex, contains
	Value     interface{} // Expected value(s)
	Country   string      // optional: "", "my", "sg" for country-specific rules
}

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

// getDefaultSOPRules returns the default SOP case rules
func getDefaultSOPRules() []CaseRule {
	return []CaseRule{
		{
			CaseType:    domain.CasePcExternalPaymentFlow200_11,
			Description: "PC External Payment Flow stuck at state 200 with attempt 11",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentCore.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "external_payment_flow",
				},
				{
					FieldPath: "PaymentCore.Workflow.State",
					Operator:  "eq",
					Value:     "200",
				},
				{
					FieldPath: "PaymentCore.Workflow.Attempt",
					Operator:  "eq",
					Value:     11,
				},
			},
		},
		{
			CaseType:    domain.CasePeTransferPayment210_0,
			Description: "PE Transfer Payment stuck at state 210 with attempt 0",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "210",
				},
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePeStuck230RepublishPC,
			Description: "PE stuck at state 230 (capture) requires PC republish",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "230",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePe2200FastCashinFailed,
			Description: "PE Transfer Collection at state 220 with attempt 0 and Fast Adapter failed",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_collection",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "220",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "FastAdapter.Status",
					Operator:  "eq",
					Value:     "FAILED",
				},
			},
		},
		{
			CaseType:    domain.CaseRppCashoutReject101_19,
			Description: "RPP Cashout Reject at state 101 with attempt 19",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "wf_ct_cashout",
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "101",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     19,
				},
			},
		},
		{
			CaseType:    domain.CaseRppQrPaymentReject210_0,
			Description: "RPP QR Payment Reject at state 210 with attempt 0",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "wf_ct_qr_payment",
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "210",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CaseRppNoResponseResume,
			Description: "RPP No Response Resume (timeout scenario)",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "210",
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "in",
					Value:     []string{"wf_ct_cashout", "wf_ct_qr_payment"},
				},
			},
		},
		{
			CaseType:    domain.CasePcExternalPaymentFlow201_0RPP900,
			Description: "PC External Payment Flow 201/0 with RPP 900 (completed)",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "220",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentEngine.Transfers.ExternalID",
					Operator:  "ne",
					Value:     "",
				},
				{
					FieldPath: "PaymentCore.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "external_payment_flow",
				},
				{
					FieldPath: "PaymentCore.Workflow.State",
					Operator:  "eq",
					Value:     "201",
				},
				{
					FieldPath: "PaymentCore.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "RPPAdapter.Status",
					Operator:  "eq",
					Value:     "900",
				},
			},
		},
		{
			CaseType:    domain.CasePcExternalPaymentFlow201_0RPP210,
			Description: "PC External Payment Flow 201/0 with RPP not completed (stuck)",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "220",
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentEngine.Transfers.ExternalID",
					Operator:  "ne",
					Value:     "",
				},
				{
					FieldPath: "PaymentCore.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "external_payment_flow",
				},
				{
					FieldPath: "PaymentCore.Workflow.State",
					Operator:  "eq",
					Value:     "201",
				},
				{
					FieldPath: "PaymentCore.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "RPPAdapter.Status",
					Operator:  "ne",
					Value:     "900",
				},
			},
		},
		{
			CaseType:    domain.CaseThoughtMachineFalseNegative,
			Description: "Thought Machine returning errors/false negatives, but transaction was successful",
			Country:     "my",
			Conditions: []RuleCondition{
				{
					FieldPath: "PaymentEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_transfer_payment",
				},
				{
					FieldPath: "PaymentEngine.Workflow.State",
					Operator:  "eq",
					Value:     "701", // stCaptureFailed
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalTxns",
					Operator:  "eq",
					Value:     nil, // No internal transactions found (NOT_FOUND)
				},
				{
					FieldPath: "RPPAdapter.Status",
					Operator:  "eq",
					Value:     "PROCESSING",
				},
			},
		}}
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

// evaluateRule evaluates a single rule against a transaction result
func (r *SOPRepository) evaluateRule(rule CaseRule, result *domain.TransactionResult) bool {
	// Check if rule involves workflow fields
	hasWorkflowFields := false
	for _, condition := range rule.Conditions {
		if strings.Contains(condition.FieldPath, "PaymentCore.Workflow.") {
			hasWorkflowFields = true
			break
		}
	}

	if !hasWorkflowFields {
		// Simple evaluation for rules without workflow fields
		for _, condition := range rule.Conditions {
			if !r.evaluateConditionSimple(condition, result) {
				return false
			}
		}
		return true
	}

	// For workflow fields, check each workflow in the slice
	for _, workflow := range result.PaymentCore.Workflow {
		allConditionsMatch := true
		for _, condition := range rule.Conditions {
			if !strings.Contains(condition.FieldPath, "PaymentCore.Workflow.") {
				// For non-workflow conditions, use simple evaluation
				if !r.evaluateConditionSimple(condition, result) {
					allConditionsMatch = false
					break
				}
			} else {
				// For workflow conditions, check against this specific workflow
				if !r.evaluateWorkflowCondition(condition, &workflow) {
					allConditionsMatch = false
					break
				}
			}
		}

		if allConditionsMatch {
			return true
		}
	}

	return false
}

// evaluateConditionSimple evaluates a single condition for non-workflow fields
func (r *SOPRepository) evaluateConditionSimple(condition RuleCondition, result *domain.TransactionResult) bool {
	fieldValue := r.getFieldValue(condition.FieldPath, result)

	switch condition.Operator {
	case "eq":
		// Special handling for nil comparison
		if condition.Value == nil {
			return fieldValue == nil
		}
		if fieldValue == nil {
			return false
		}
		// Handle string to int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if fieldStr, ok := fieldValue.(string); ok {
				// Try to convert both to int for numeric comparison
				if conditionInt, err1 := strconv.Atoi(conditionStr); err1 == nil {
					if fieldInt, err2 := strconv.Atoi(fieldStr); err2 == nil {
						return conditionInt == fieldInt
					}
				}
			}
		}
		return reflect.DeepEqual(fieldValue, condition.Value)
	case "ne":
		// Handle string to int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if fieldStr, ok := fieldValue.(string); ok {
				// Try to convert both to int for numeric comparison
				if conditionInt, err1 := strconv.Atoi(conditionStr); err1 == nil {
					if fieldInt, err2 := strconv.Atoi(fieldStr); err2 == nil {
						return conditionInt != fieldInt
					}
				}
			}
		}
		return !reflect.DeepEqual(fieldValue, condition.Value)
	case "lt":
		return r.compareValues(fieldValue, condition.Value) < 0
	case "gt":
		return r.compareValues(fieldValue, condition.Value) > 0
	case "in":
		return r.isInSlice(fieldValue, condition.Value)
	case "not_in":
		return !r.isInSlice(fieldValue, condition.Value)
	case "regex":
		return r.matchRegex(fieldValue, condition.Value)
	case "contains":
		return r.containsValue(fieldValue, condition.Value)
	default:
		return false
	}
}

// evaluateWorkflowCondition evaluates a condition against a specific workflow
func (r *SOPRepository) evaluateWorkflowCondition(condition RuleCondition, workflow *domain.WorkflowInfo) bool {
	// Extract the field name from the path (last part after the dot)
	parts := strings.Split(condition.FieldPath, ".")
	fieldName := parts[len(parts)-1]

	var fieldValue interface{}
	switch fieldName {
	case "WorkflowID":
		fieldValue = workflow.WorkflowID
	case "State":
		fieldValue = workflow.State
	case "Attempt":
		fieldValue = workflow.Attempt
	default:
		return false
	}

	switch condition.Operator {
	case "eq":
		// Handle string to int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if fieldStr, ok := fieldValue.(string); ok {
				// Debug logging
				fmt.Printf("DEBUG EQ: condition='%s' (%T), field='%s' (%T)", conditionStr, condition.Value, fieldStr, fieldValue)
				// Try to convert both to int for numeric comparison
				if conditionInt, err1 := strconv.Atoi(conditionStr); err1 == nil {
					if fieldInt, err2 := strconv.Atoi(fieldStr); err2 == nil {
						fmt.Printf("DEBUG EQ: converted to int - condition=%d, field=%d, result=%v", conditionInt, fieldInt, conditionInt == fieldInt)
						return conditionInt == fieldInt
					}
				}
			}
		}
		return reflect.DeepEqual(fieldValue, condition.Value)
	case "ne":
		// Handle string to int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if fieldStr, ok := fieldValue.(string); ok {
				// Debug logging
				fmt.Printf("DEBUG NE: condition='%s' (%T), field='%s' (%T)", conditionStr, condition.Value, fieldStr, fieldValue)
				// Try to convert both to int for numeric comparison
				if conditionInt, err1 := strconv.Atoi(conditionStr); err1 == nil {
					if fieldInt, err2 := strconv.Atoi(fieldStr); err2 == nil {
						fmt.Printf("DEBUG NE: converted to int - condition=%d, field=%d, result=%v", conditionInt, fieldInt, conditionInt != fieldInt)
						return conditionInt != fieldInt
					}
				}
			}
		}
		return !reflect.DeepEqual(fieldValue, condition.Value)
	case "lt":
		return r.compareValues(fieldValue, condition.Value) < 0
	case "gt":
		return r.compareValues(fieldValue, condition.Value) > 0
	case "in":
		return r.isInSlice(fieldValue, condition.Value)
	case "not_in":
		return !r.isInSlice(fieldValue, condition.Value)
	default:
		return false
	}
}

// getFieldValue retrieves field value from domain.TransactionResult using dot notation
func (r *SOPRepository) getFieldValue(fieldPath string, result *domain.TransactionResult) interface{} {
	parts := strings.Split(fieldPath, ".")
	currentValue := reflect.ValueOf(result)

	for i, part := range parts {
		if currentValue.Kind() == reflect.Ptr {
			currentValue = currentValue.Elem()
		}

		if currentValue.Kind() == reflect.Struct {
			field := currentValue.FieldByName(part)
			if !field.IsValid() {
				return nil
			}
			currentValue = field
		} else if currentValue.Kind() == reflect.Slice {
			// For slices (like PaymentCore.InternalTxns), check if this is the last part
			if i == len(parts)-1 {
				// Return the whole slice when this is the final field
				sliceValue := currentValue.Interface()
				// Check if slice is empty or nil
				if sliceValue == nil || currentValue.Len() == 0 {
					return nil
				}
				return sliceValue
			}
			// If not the last part, we can't navigate further into slices
			return nil
		} else {
			return nil
		}
	}

	return currentValue.Interface()
}

// compareValues compares two numeric or string values
func (r *SOPRepository) compareValues(a, b interface{}) int {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Try to compare as numbers
	if aNum, aErr := strconv.ParseFloat(aStr, 64); aErr == nil {
		if bNum, bErr := strconv.ParseFloat(bStr, 64); bErr == nil {
			if aNum < bNum {
				return -1
			} else if aNum > bNum {
				return 1
			}
			return 0
		}
	}

	// Compare as strings
	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// isInSlice checks if value is in slice
func (r *SOPRepository) isInSlice(value, slice interface{}) bool {
	sliceValue := reflect.ValueOf(slice)
	if sliceValue.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < sliceValue.Len(); i++ {
		if reflect.DeepEqual(value, sliceValue.Index(i).Interface()) {
			return true
		}
	}
	return false
}

// matchRegex checks if value matches regex pattern
func (r *SOPRepository) matchRegex(value, pattern interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	patternStr := fmt.Sprintf("%v", pattern)

	matched, err := regexp.MatchString(patternStr, valueStr)
	if err != nil {
		return false
	}
	return matched
}

// containsValue checks if value contains substring
func (r *SOPRepository) containsValue(value, substr interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	substrStr := fmt.Sprintf("%v", substr)
	return strings.Contains(valueStr, substrStr)
}
