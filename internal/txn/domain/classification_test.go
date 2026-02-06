package domain

import (
	"fmt"
	"slices"
	"testing"

	"buddy/internal/config"
)

func init() {
	// Initialize config loader for tests
	// Config files are embedded at build time, no path needed
	_ = config.InitializeConfigLoader()
}

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
			if stateMap, exists := GetWorkflowStateMap("workflow_transfer_collection"); exists {
				if stateName, exists := stateMap[tc.state]; exists {
					result = stateName
				} else {
					result = fmt.Sprintf("stUnknown_%d", tc.state)
				}
			} else {
				result = fmt.Sprintf("stUnknown_%d", tc.state)
			}
			if result != tc.expected {
				t.Errorf("GetWorkflowStateMap(\"workflow_transfer_collection\")[%d] = %v; expected %v", tc.state, result, tc.expected)
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
			expected: "220 (stAuthProcessing)",
		},
		{
			name:     "Another valid state number",
			stateStr: "910",
			expected: "910 (stCompletedNotified)",
		},
		{
			name:     "Invalid state number",
			stateStr: "999",
			expected: "999",
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

func TestThoughtMachineFalseNegativeCase(t *testing.T) {
	// Test E2E ID from documentation
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid E2E ID from documentation",
			input:    "20251212GXSPMYKL040OQR32194316",
			expected: true,
		},
		{
			name:     "Another valid E2E ID",
			input:    "20251202GXSPMYKL010ORB62198922",
			expected: true,
		},
		{
			name:     "Invalid E2E ID",
			input:    "invalid-id",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRppE2EID(tc.input)
			if result != tc.expected {
				t.Errorf("IsRppE2EID(%s) = %v; expected %v", tc.input, result, tc.expected)
			}
		})
	}

	// Test workflow state formatting
	state := FormatWorkflowState("workflow_transfer_payment", "701")
	expected := "701 (stCaptureFailed)"
	if state != expected {
		t.Errorf("Expected workflow state %s, got %s", expected, state)
	}

	// Test case constant
	if CaseThoughtMachineFalseNegative != "thought_machine_false_negative" {
		t.Errorf("Case constant mismatch: got %s, expected thought_machine_false_negative", CaseThoughtMachineFalseNegative)
	}

	// Test that case is in summary order
	summaryOrder := GetCaseSummaryOrder()
	found := slices.Contains(summaryOrder, CaseThoughtMachineFalseNegative)
	if !found {
		t.Error("CaseThoughtMachineFalseNegative not found in summary order")
	}
}
