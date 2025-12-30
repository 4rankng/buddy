package adapters

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"buddy/internal/txn/domain"
)

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

// evaluateCondition evaluates a single condition against the transaction result.
func (r *SOPRepository) evaluateCondition(condition RuleCondition, result *domain.TransactionResult) bool {
	fieldValue, ok := r.getFieldValue(condition.FieldPath, result)

	// Backward compatible behavior:
	// If the path isn't reachable due to nil pointers or missing fields,
	// treat `eq ""` as a match, otherwise fail.
	if !ok {
		return condition.Operator == "eq" && condition.Value == ""
	}

	// Handle slice values: check if ANY element in slice matches condition
	if r.isSliceValue(fieldValue) {
		return r.evaluateSliceCondition(condition, fieldValue)
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

// evaluateSliceCondition evaluates a condition against a slice value.
// Returns true if ANY element in slice matches condition.
func (r *SOPRepository) evaluateSliceCondition(condition RuleCondition, sliceValue interface{}) bool {
	sliceVal := reflect.ValueOf(sliceValue)
	if sliceVal.Kind() != reflect.Slice {
		return false
	}

	// Empty slice handling
	if sliceVal.Len() == 0 {
		if condition.Operator == "eq" && condition.Value == "" {
			return true
		}
		if condition.Operator == "ne" && condition.Value != "" {
			return true
		}
		return false
	}

	// Check if ANY element matches condition
	for i := 0; i < sliceVal.Len(); i++ {
		element := sliceVal.Index(i).Interface()
		if r.evaluateConditionWithElement(condition, element) {
			return true
		}
	}
	return false
}

// evaluateConditionWithElement evaluates a condition against a single element value.
func (r *SOPRepository) evaluateConditionWithElement(condition RuleCondition, element interface{}) bool {
	switch condition.Operator {
	case "eq":
		if condition.Value == nil {
			return element == nil
		}
		if element == nil {
			return false
		}
		// Handle string-to-int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if elementStr, ok := element.(string); ok {
				if conditionInt, err1 := strconv.Atoi(conditionStr); err1 == nil {
					if elementInt, err2 := strconv.Atoi(elementStr); err2 == nil {
						return conditionInt == elementInt
					}
				}
			}
		}
		return reflect.DeepEqual(element, condition.Value)

	case "ne":
		// Handle string-to-int conversion for numeric comparisons
		if conditionStr, ok := condition.Value.(string); ok {
			if elementStr, ok := element.(string); ok {
				if conditionInt, err1 := strconv.Atoi(conditionStr); err1 == nil {
					if elementInt, err2 := strconv.Atoi(elementStr); err2 == nil {
						return conditionInt != elementInt
					}
				}
			}
		}
		return !reflect.DeepEqual(element, condition.Value)

	case "in":
		return r.isInSlice(element, condition.Value)
	case "not_in":
		return !r.isInSlice(element, condition.Value)
	case "regex":
		return r.matchRegex(element, condition.Value)
	case "contains":
		return r.containsValue(element, condition.Value)
	default:
		return false
	}
}

// isSliceValue checks if a value is a slice type.
func (r *SOPRepository) isSliceValue(value interface{}) bool {
	if value == nil {
		return false
	}
	return reflect.TypeOf(value).Kind() == reflect.Slice
}
