package txn

import (
	"fmt"
	"testing"
)

func TestRppE2EIDPattern(t *testing.T) {
	// Test valid RPP E2E IDs
	validE2EIDs := []string{
		"20251209GXSPMYKL010ORB79174342",
		"20251209GXSPMYKL030OQR15900197",
		"20251209GXSPMYKL040OQR10829949",
		"20251209GXSPMYKL040OQR41308688",
		"20250101GXSPMYAB12345678901234",
		"20241231GXSPMYzz98765432109876",
	}

	for _, id := range validE2EIDs {
		if !RppE2EIDPattern.MatchString(id) {
			t.Errorf("Expected valid E2E ID to match pattern: %s", id)
		}
	}
}

func TestIsRppE2EID(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"20251209GXSPMYKL010ORB79174342", true},
		{"20251209GXSPMYKL030OQR15900197", true},
		{"20251209GXSPMYKL040OQR10829949", true},
		{"20251209GXSPMYKL040OQR41308688", true},
		{"20250101GXSPMYAB12345678901234", true},
		{"20251209GXSPMXKL010ORB79174342", true},

		{"20251209GXSPMY", false},                   // Too short
		{"20251209GXSPMYKL010ORB791743423", false},  // Too long
		{"ccc572052d6446a2b896fee381dcca3a", false}, // Transaction ID
		{"TS-4466.txt", false},                      // File path
		{"", false},                                 // Empty
	}

	for _, tc := range testCases {
		result := IsRppE2EID(tc.input)
		if result != tc.expected {
			t.Errorf("IsRppE2EID(%s) = %v; expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsFilePath(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"TS-4466.txt", true},
		{"transactions.csv", true},
		{"data.json", true},
		{"../path/to/file.txt", true},
		{"/absolute/path/to/file.txt", true},
		{"./relative/file.txt", true},
		{"file", false},                             // File without extension - ambiguous, classify as transaction ID
		{"20251209GXSPMYKL010ORB79174342", false},   // E2E ID
		{"ccc572052d6446a2b896fee381dcca3a", false}, // Transaction ID
		{"", false},                                 // Empty
		{"just-a-string", false},                    // No file-like pattern
		{".hiddenfile", true},                       // Hidden file
	}

	for _, tc := range testCases {
		result := IsFilePath(tc.input)
		if result != tc.expected {
			t.Errorf("IsFilePath(%s) = %v; expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestRppE2EIDFormatExamples(t *testing.T) {
	// Test the exact examples from the SOP
	examples := []string{
		"20251209GXSPMYKL010ORB79174342",
		"20251209GXSPMYKL030OQR15900197",
		"20251209GXSPMYKL040OQR10829949",
		"20251209GXSPMYKL040OQR41308688",
	}

	for _, example := range examples {
		// Check length
		if len(example) != 30 {
			t.Errorf("E2E ID %s has length %d; expected 30", example, len(example))
		}

		// Check pattern
		if !RppE2EIDPattern.MatchString(example) {
			t.Errorf("E2E ID %s does not match expected pattern", example)
		}

		// Check prefix
		if len(example) >= 14 {
			datePart := example[:8]
			prefixPart := example[8:14]
			if datePart < "20250101" || datePart > "20991231" {
				t.Errorf("E2E ID %s has invalid date part: %s", example, datePart)
			}
			if prefixPart != "GXSPMY" {
				t.Errorf("E2E ID %s has invalid prefix: %s", example, prefixPart)
			}
		}
	}
}

func TestWorkflowTransferCollectionStates(t *testing.T) {
	testCases := []struct {
		name     string
		state    int
		expected string
	}{
		{
			name:     "Initial state - stTransferPersisted",
			state:    100,
			expected: "stTransferPersisted",
		},
		{
			name:     "Processing state - stTransferProcessing",
			state:    210,
			expected: "stTransferProcessing",
		},
		{
			name:     "Auth success state - stAuthSuccess",
			state:    300,
			expected: "stAuthSuccess",
		},
		{
			name:     "Failure handling state - stPrepareFailureHandling",
			state:    501,
			expected: "stPrepareFailureHandling",
		},
		{
			name:     "Investigation required state - stCaptureFailed",
			state:    701,
			expected: "stCaptureFailed",
		},
		{
			name:     "Completion state - stTransferCompleted",
			state:    900,
			expected: "stTransferCompleted",
		},
		{
			name:     "Unknown state - should return formatted unknown",
			state:    999,
			expected: "stUnknown_999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result string
			if stateMap, exists := WorkflowStateMaps["workflow_transfer_collection"]; exists {
				if stateName, exists := stateMap[tc.state]; exists {
					result = stateName
				} else {
					result = fmt.Sprintf("stUnknown_%d", tc.state)
				}
			} else {
				result = fmt.Sprintf("stUnknown_%d", tc.state)
			}
			if result != tc.expected {
				t.Errorf("WorkflowStateMaps[\"workflow_transfer_collection\"][%d] = %v; expected %v", tc.state, result, tc.expected)
			}
		})
	}
}

func TestFormatWorkflowStateTransferCollection(t *testing.T) {
	testCases := []struct {
		name     string
		stateStr string
		expected string
	}{
		{
			name:     "Valid state number",
			stateStr: "220",
			expected: "stAuthProcessing(220)",
		},
		{
			name:     "Another valid state number",
			stateStr: "910",
			expected: "stCompletedNotified(910)",
		},
		{
			name:     "Invalid state number",
			stateStr: "999",
			expected: "stUnknown_999(999)",
		},
		{
			name:     "Non-numeric state string",
			stateStr: "invalid",
			expected: "invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatWorkflowState("workflow_transfer_collection", tc.stateStr)
			if result != tc.expected {
				t.Errorf("FormatWorkflowState(workflow_transfer_collection, %s) = %v; expected %v", tc.stateStr, result, tc.expected)
			}
		})
	}
}
