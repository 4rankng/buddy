package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// SOPCaseRule defines a rule for identifying SOP cases
type SOPCaseRule struct {
	CaseType    domain.SOPCase
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
	rules []SOPCaseRule
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
func getDefaultSOPRules() []SOPCaseRule {
	return []SOPCaseRule{
		{
			CaseType:    domain.SOPCasePcExternalPaymentFlow200_11,
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
			CaseType:    domain.SOPCasePeTransferPayment210_0,
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
			CaseType:    domain.SOPCasePe2200FastCashinFailed,
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
			CaseType:    domain.SOPCaseRppCashoutReject101_19,
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
			CaseType:    domain.SOPCaseRppQrPaymentReject210_0,
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
			CaseType:    domain.SOPCaseRppNoResponseResume,
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
			CaseType:    domain.SOPCasePcExternalPaymentFlow201_0RPP900,
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
			CaseType:    domain.SOPCasePcExternalPaymentFlow201_0RPP210,
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
		}}
}

// IdentifySOPCase identifies the SOP case for a transaction result
func (r *SOPRepository) IdentifySOPCase(result *domain.TransactionResult, env string) domain.SOPCase {
	// Check if we've already identified the case
	if result.CaseType != domain.SOPCaseNone {
		return result.CaseType
	}

	fmt.Printf("[DEBUG] Identifying SOP case for transaction %s in environment %s\n", result.TransactionID, env)
	fmt.Printf("[DEBUG] PaymentCore.Workflow count: %d\n", len(result.PaymentCore.Workflow))
	fmt.Printf("[DEBUG] PaymentEngine.Workflow.State: %s\n", result.PaymentEngine.Workflow.State)
	fmt.Printf("[DEBUG] PaymentEngine.Workflow.WorkflowID: %s\n", result.PaymentEngine.Workflow.WorkflowID)
	fmt.Printf("[DEBUG] PaymentEngine.Workflow.Attempt: %d\n", result.PaymentEngine.Workflow.Attempt)
	fmt.Printf("[DEBUG] RPPAdapter.Status: %s\n", result.RPPAdapter.Status)
	fmt.Printf("[DEBUG] RPPAdapter.Workflow.State: %s\n", result.RPPAdapter.Workflow.State)
	fmt.Printf("[DEBUG] RPPAdapter.Workflow.WorkflowID: %s\n", result.RPPAdapter.Workflow.WorkflowID)
	fmt.Printf("[DEBUG] RPPAdapter.Workflow.Attempt: %d\n", result.RPPAdapter.Workflow.Attempt)
	fmt.Printf("[DEBUG] FastAdapter.Status: %s\n", result.FastAdapter.Status)

	// Check each rule in order
	for _, rule := range r.rules {
		// Skip country-specific rules if not matching
		if rule.Country != "" && rule.Country != env {
			fmt.Printf("[DEBUG] Skipping rule %s due to country mismatch (rule: %s, env: %s)\n", rule.CaseType, rule.Country, env)
			continue
		}

		fmt.Printf("[DEBUG] Evaluating rule: %s\n", rule.CaseType)
		if r.evaluateRule(rule, result) {
			fmt.Printf("[DEBUG] Rule matched: %s\n", rule.CaseType)
			result.CaseType = rule.CaseType
			return result.CaseType
		} else {
			fmt.Printf("[DEBUG] Rule did not match: %s\n", rule.CaseType)
		}
	}

	fmt.Printf("[DEBUG] No rules matched for transaction %s\n", result.TransactionID)
	result.CaseType = domain.SOPCaseNone
	return result.CaseType
}

// evaluateRule evaluates a single rule against a transaction result
func (r *SOPRepository) evaluateRule(rule SOPCaseRule, result *domain.TransactionResult) bool {
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
			fmt.Printf("[DEBUG] Evaluating condition: %s %s %v\n", condition.FieldPath, condition.Operator, condition.Value)
			if !r.evaluateConditionSimple(condition, result) {
				fmt.Printf("[DEBUG] Condition failed: %s %s %v\n", condition.FieldPath, condition.Operator, condition.Value)
				return false
			}
			fmt.Printf("[DEBUG] Condition passed: %s %s %v\n", condition.FieldPath, condition.Operator, condition.Value)
		}
		return true
	}

	// For workflow fields, check each workflow in the slice
	fmt.Printf("[DEBUG] Rule has workflow fields, checking each workflow in PaymentCore.Workflow slice\n")
	for i, workflow := range result.PaymentCore.Workflow {
		fmt.Printf("[DEBUG] Checking workflow %d: %+v\n", i, workflow)

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
			fmt.Printf("[DEBUG] All conditions matched for workflow %d\n", i)
			return true
		}
	}

	fmt.Printf("[DEBUG] No workflows matched all conditions\n")
	return false
}

// evaluateConditionSimple evaluates a single condition for non-workflow fields
func (r *SOPRepository) evaluateConditionSimple(condition RuleCondition, result *domain.TransactionResult) bool {
	fieldValue := r.getFieldValue(condition.FieldPath, result)
	if fieldValue == nil {
		fmt.Printf("[DEBUG] Field value is nil for path: %s\n", condition.FieldPath)
		return false
	}

	fmt.Printf("[DEBUG] Field value for %s: %v (type: %T)\n", condition.FieldPath, fieldValue, fieldValue)

	var evalResult bool
	switch condition.Operator {
	case "eq":
		evalResult = reflect.DeepEqual(fieldValue, condition.Value)
	case "ne":
		evalResult = !reflect.DeepEqual(fieldValue, condition.Value)
	case "lt":
		evalResult = r.compareValues(fieldValue, condition.Value) < 0
	case "gt":
		evalResult = r.compareValues(fieldValue, condition.Value) > 0
	case "in":
		evalResult = r.isInSlice(fieldValue, condition.Value)
	case "not_in":
		evalResult = !r.isInSlice(fieldValue, condition.Value)
	case "regex":
		evalResult = r.matchRegex(fieldValue, condition.Value)
	case "contains":
		evalResult = r.containsValue(fieldValue, condition.Value)
	default:
		evalResult = false
	}

	fmt.Printf("[DEBUG] Condition evaluation result: %t\n", evalResult)
	return evalResult
}

// evaluateWorkflowCondition evaluates a condition against a specific workflow
func (r *SOPRepository) evaluateWorkflowCondition(condition RuleCondition, workflow *domain.WorkflowInfo) bool {
	fmt.Printf("[DEBUG] Evaluating workflow condition: %s %s %v\n", condition.FieldPath, condition.Operator, condition.Value)

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
		fmt.Printf("[DEBUG] Unknown workflow field: %s\n", fieldName)
		return false
	}

	fmt.Printf("[DEBUG] Workflow field %s value: %v (type: %T)\n", fieldName, fieldValue, fieldValue)

	var evalResult bool
	switch condition.Operator {
	case "eq":
		evalResult = reflect.DeepEqual(fieldValue, condition.Value)
	case "ne":
		evalResult = !reflect.DeepEqual(fieldValue, condition.Value)
	case "lt":
		evalResult = r.compareValues(fieldValue, condition.Value) < 0
	case "gt":
		evalResult = r.compareValues(fieldValue, condition.Value) > 0
	case "in":
		evalResult = r.isInSlice(fieldValue, condition.Value)
	case "not_in":
		evalResult = !r.isInSlice(fieldValue, condition.Value)
	default:
		evalResult = false
	}

	fmt.Printf("[DEBUG] Workflow condition evaluation result: %t\n", evalResult)
	return evalResult
}

// getFieldValue retrieves field value from domain.TransactionResult using dot notation
func (r *SOPRepository) getFieldValue(fieldPath string, result *domain.TransactionResult) interface{} {
	parts := strings.Split(fieldPath, ".")
	fmt.Printf("[DEBUG] Getting field value for path: %s (parts: %v)\n", fieldPath, parts)
	currentValue := reflect.ValueOf(result)

	for i, part := range parts {
		fmt.Printf("[DEBUG] Processing part %d: %s, current value type: %s\n", i, part, currentValue.Kind())
		if currentValue.Kind() == reflect.Ptr {
			currentValue = currentValue.Elem()
		}

		if currentValue.Kind() == reflect.Struct {
			field := currentValue.FieldByName(part)
			if !field.IsValid() {
				fmt.Printf("[DEBUG] Field %s not found in struct\n", part)
				return nil
			}
			currentValue = field
		} else if currentValue.Kind() == reflect.Slice {
			fmt.Printf("[DEBUG] Encountered slice with %d elements\n", currentValue.Len())
			// If we're at a slice and there are more parts to process,
			// we need to find an element in the slice that has the requested field
			if i < len(parts)-1 {
				// For PaymentCore.Workflow.WorkflowID, State, or Attempt, we need to search for elements matching criteria
				// Return the whole slice and let the evaluation handle the matching
				fmt.Printf("[DEBUG] Returning whole slice for field %s to be evaluated by condition\n", part)
				return currentValue.Interface()
			} else {
				// This is the last part and it's a slice, return the whole slice
				fmt.Printf("[DEBUG] Returning whole slice for path %s\n", fieldPath)
				return currentValue.Interface()
			}
		} else {
			fmt.Printf("[DEBUG] Unexpected type %s at part %s\n", currentValue.Kind(), part)
			return nil
		}
	}

	finalValue := currentValue.Interface()
	fmt.Printf("[DEBUG] Final value for %s: %v (type: %T)\n", fieldPath, finalValue, finalValue)
	return finalValue
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
