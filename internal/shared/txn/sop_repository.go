package txn

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// SOPCaseRule defines a rule for identifying SOP cases
type SOPCaseRule struct {
	CaseType    SOPCase
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
			CaseType:    SOPCasePcExternalPaymentFlow200_11,
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
			CaseType:    SOPCasePeTransferPayment210_0,
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
			CaseType:    SOPCasePe2200FastCashinFailed,
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
			CaseType:    SOPCaseRppCashoutReject101_19,
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
			CaseType:    SOPCaseRppQrPaymentReject210_0,
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
			CaseType:    SOPCaseRppNoResponseResume,
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
			CaseType:    SOPCasePcExternalPaymentFlow201_0RPP900,
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
			CaseType:    SOPCasePcExternalPaymentFlow201_0RPP210,
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
func (r *SOPRepository) IdentifySOPCase(result *TransactionResult, env string) SOPCase {
	// Check if we've already identified the case
	if result.CaseType != SOPCaseNone {
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

	result.CaseType = SOPCaseNone
	return result.CaseType
}

// evaluateRule evaluates a single rule against a transaction result
func (r *SOPRepository) evaluateRule(rule SOPCaseRule, result *TransactionResult) bool {
	for _, condition := range rule.Conditions {
		if !r.evaluateCondition(condition, result) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (r *SOPRepository) evaluateCondition(condition RuleCondition, result *TransactionResult) bool {
	fieldValue := r.getFieldValue(condition.FieldPath, result)
	if fieldValue == nil {
		return false
	}

	switch condition.Operator {
	case "eq":
		return reflect.DeepEqual(fieldValue, condition.Value)
	case "ne":
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

// getFieldValue retrieves field value from TransactionResult using dot notation
func (r *SOPRepository) getFieldValue(fieldPath string, result *TransactionResult) interface{} {
	parts := strings.Split(fieldPath, ".")
	currentValue := reflect.ValueOf(result)

	for _, part := range parts {
		if currentValue.Kind() == reflect.Ptr {
			currentValue = currentValue.Elem()
		}

		if currentValue.Kind() == reflect.Struct {
			field := currentValue.FieldByName(part)
			if !field.IsValid() {
				return nil
			}
			currentValue = field
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
