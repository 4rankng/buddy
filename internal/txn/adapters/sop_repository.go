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
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.State",
					Operator:  "eq",
					Value:     "200",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.Attempt",
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
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "900",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
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
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.State",
					Operator:  "eq",
					Value:     "201",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.Attempt",
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
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.State",
					Operator:  "eq",
					Value:     "201",
				},
				{
					FieldPath: "PaymentCore.ExternalTransfer.Workflow.Attempt",
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
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "500", // stFailed
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess,
			Description: "PE capture processing, PC capture failed, but RPP succeeded",
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
					Value:     "230", // stCaptureProcessing
				},
				{
					FieldPath: "PaymentEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.State",
					Operator:  "eq",
					Value:     "500", // stFailed
				},
				{
					FieldPath: "PaymentCore.InternalCapture.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PaymentCore.InternalCapture.TxType",
					Operator:  "eq",
					Value:     "CAPTURE",
				},
				{
					FieldPath: "PaymentCore.InternalCapture.TxStatus",
					Operator:  "eq",
					Value:     "FAILED",
				},
				{
					FieldPath: "RPPAdapter.Workflow.WorkflowID",
					Operator:  "in",
					Value:     []string{"wf_ct_qr_payment", "wf_ct_cashout"},
				},
				{
					FieldPath: "RPPAdapter.Workflow.State",
					Operator:  "eq",
					Value:     "900", // stSuccess
				},
				{
					FieldPath: "RPPAdapter.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
		{
			CaseType:    domain.CaseEcotxnChargeFailedCaptureFailedTMError,
			Description: "Ecotxn Charge Failed Capture Failed with TMError",
			Conditions: []RuleCondition{
				{
					FieldPath: "PartnerpayEngine.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "workflow_charge",
				},
				{
					FieldPath: "PartnerpayEngine.Workflow.State",
					Operator:  "eq",
					Value:     "502",
				},
				{
					FieldPath: "PartnerpayEngine.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
				{
					FieldPath: "PartnerpayEngine.Charge.StatusReason",
					Operator:  "eq",
					Value:     "SYSTEM_ERROR",
				},
				{
					FieldPath: "PartnerpayEngine.Charge.StatusReasonDescription",
					Operator:  "eq",
					Value:     "error occurred in Thought Machine.",
				},
				{
					FieldPath: "PaymentCore.Workflow.WorkflowID",
					Operator:  "eq",
					Value:     "internal_payment_flow",
				},
				{
					FieldPath: "PaymentCore.Workflow.State",
					Operator:  "eq",
					Value:     "500",
				},
				{
					FieldPath: "PaymentCore.Workflow.Attempt",
					Operator:  "eq",
					Value:     0,
				},
			},
		},
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

// IdentifyCaseWithDebug identifies the SOP case with debug logging
func (r *SOPRepository) IdentifyCaseWithDebug(result *domain.TransactionResult, env string) domain.Case {
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

		if r.evaluateRuleWithDebug(rule, result) {
			result.CaseType = rule.CaseType
			return result.CaseType
		}
	}

	result.CaseType = domain.CaseNone
	return result.CaseType
}

// evaluateRule evaluates a single rule against a transaction result.
//
// This implementation is fully driven by FieldPath traversal. It does NOT special-case
// workflow paths; instead, getFieldValue is nil-safe and returns (value, ok).
func (r *SOPRepository) evaluateRule(rule CaseRule, result *domain.TransactionResult) bool {
	for _, condition := range rule.Conditions {
		if !r.evaluateCondition(condition, result) {
			return false
		}
	}
	return true
}

// evaluateRuleWithDebug evaluates a single rule against a transaction result with debug logging.
func (r *SOPRepository) evaluateRuleWithDebug(rule CaseRule, result *domain.TransactionResult) bool {
	fmt.Printf("Evaluating rule: %s (%s)\n", rule.CaseType, rule.Description)
	for _, condition := range rule.Conditions {
		fieldValue, ok := r.getFieldValue(condition.FieldPath, result)
		fmt.Printf("  Condition: %s %s %v", condition.FieldPath, condition.Operator, condition.Value)
		if !ok {
			fmt.Printf(" -> FAILED (field not found)\n")
			return false
		}
		fmt.Printf(" (actual: %v)", fieldValue)
		if !r.evaluateCondition(condition, result) {
			fmt.Printf(" -> FAILED\n")
			return false
		}
		fmt.Printf(" -> PASSED\n")
	}
	fmt.Printf("Rule PASSED: %s\n", rule.CaseType)
	return true
}

// evaluateCondition evaluates a single condition against the transaction result.
func (r *SOPRepository) evaluateCondition(condition RuleCondition, result *domain.TransactionResult) bool {
	fieldValue, ok := r.getFieldValue(condition.FieldPath, result)

	// Backward compatible behavior:
	// If the path isn't reachable due to nil pointers or missing fields,
	// treat `eq ""` as a match, otherwise fail.
	if !ok {
		return condition.Operator == "eq" && condition.Value == ""
	}

	switch condition.Operator {
	case "eq":
		// Special handling for nil comparison
		if condition.Value == nil {
			return fieldValue == nil
		}
		if fieldValue == nil {
			return false
		}

		// Handle string-to-int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if fieldStr, ok := fieldValue.(string); ok {
				if conditionInt, err1 := strconv.Atoi(conditionStr); err1 == nil {
					if fieldInt, err2 := strconv.Atoi(fieldStr); err2 == nil {
						return conditionInt == fieldInt
					}
				}
			}
		}
		return reflect.DeepEqual(fieldValue, condition.Value)

	case "ne":
		// Handle string-to-int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if fieldStr, ok := fieldValue.(string); ok {
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

// getFieldValue retrieves a field value from domain.TransactionResult using dot notation.
// It is nil-safe:
// - If a pointer along the path is nil, it returns (nil, false)
// - If a field name doesn't exist, it returns (nil, false)
// - If the path is valid but the final value is a nil pointer, it returns (nil, true)
func (r *SOPRepository) getFieldValue(fieldPath string, result *domain.TransactionResult) (interface{}, bool) {
	parts := strings.Split(fieldPath, ".")
	current := reflect.ValueOf(result)

	for _, part := range parts {
		// Dereference pointers safely
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil, false
			}
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			return nil, false
		}

		field := current.FieldByName(part)
		if !field.IsValid() {
			return nil, false
		}

		current = field
	}

	// If final value is a pointer, preserve nil vs non-nil
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return nil, true
		}
		current = current.Elem()
	}

	return current.Interface(), true
}

// compareValues compares two numeric or string values.
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

// isInSlice checks if value is in slice.
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

// matchRegex checks if value matches regex pattern.
func (r *SOPRepository) matchRegex(value, pattern interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	patternStr := fmt.Sprintf("%v", pattern)

	matched, err := regexp.MatchString(patternStr, valueStr)
	if err != nil {
		return false
	}
	return matched
}

// containsValue checks if value contains substring.
func (r *SOPRepository) containsValue(value, substr interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	substrStr := fmt.Sprintf("%v", substr)
	return strings.Contains(valueStr, substrStr)
}
