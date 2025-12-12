package txn

import (
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

func TestRppE2EIDPatternInvalid(t *testing.T) {
	// Test invalid RPP E2E IDs
	invalidE2EIDs := []string{
		// Wrong date format
		"2025-12-09GXSPMYKL010ORB79174342",
		"2025129GXSPMYKL010ORB79174342",   // 7 digits
		"202512099GXSPMYKL010ORB79174342", // 9 digits
		// Wrong prefix
		"20251209GXSPMXKL010ORB79174342",
		"20251209GSPMYKL010ORB79174342",
		// Wrong length
		"20251209GXSPMYKL010ORB7917434",   // 31 chars
		"20251209GXSPMYKL010ORB791743423", // 33 chars
		// Special characters
		"20251209GXSPMYKL010ORB79174342@",
		"20251209GXSPMYKL010ORB79174342!",
		// Empty
		"",
		// Transaction IDs
		"ccc572052d6446a2b896fee381dcca3a",
		"f4e858c9f47f4a469f09126f94f42ace",
		// File paths
		"TS-4466.txt",
		"transactions.csv",
		"../path/to/file.txt",
		// Numbers only
		"12345678901234567890123456789012",
		// Lowercase prefix
		"20251209gxspmyKL010ORB79174342",
	}

	for _, id := range invalidE2EIDs {
		if RppE2EIDPattern.MatchString(id) {
			t.Errorf("Expected invalid E2E ID to NOT match pattern: %s", id)
		}
	}
}

func TestClassifyInput(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected InputType
	}{
		{
			name:     "Valid RPP E2E ID 1",
			input:    "20251209GXSPMYKL010ORB79174342",
			expected: InputTypeE2EID,
		},
		{
			name:     "Valid RPP E2E ID 2",
			input:    "20251209GXSPMYKL030OQR15900197",
			expected: InputTypeE2EID,
		},
		{
			name:     "Valid RPP E2E ID 3",
			input:    "20251209GXSPMYKL040OQR10829949",
			expected: InputTypeE2EID,
		},
		{
			name:     "Valid RPP E2E ID 4",
			input:    "20251209GXSPMYKL040OQR41308688",
			expected: InputTypeE2EID,
		},
		{
			name:     "Transaction ID",
			input:    "ccc572052d6446a2b896fee381dcca3a",
			expected: InputTypeTransactionID,
		},
		{
			name:     "Another Transaction ID",
			input:    "f4e858c9f47f4a469f09126f94f42ace",
			expected: InputTypeTransactionID,
		},
		{
			name:     "Short transaction ID",
			input:    "abc123",
			expected: InputTypeTransactionID,
		},
		{
			name:     "File path with .txt extension",
			input:    "TS-4466.txt",
			expected: InputTypeFilePath,
		},
		{
			name:     "File path with .csv extension",
			input:    "transactions.csv",
			expected: InputTypeFilePath,
		},
		{
			name:     "File path with directories",
			input:    "/path/to/file.txt",
			expected: InputTypeFilePath,
		},
		{
			name:     "Relative file path",
			input:    "../data/transactions.json",
			expected: InputTypeFilePath,
		},
		{
			name:     "Invalid E2E ID - wrong prefix",
			input:    "20251209GXSPMXKL010ORB79174342",
			expected: InputTypeTransactionID,
		},
		{
			name:     "Invalid E2E ID - wrong date format",
			input:    "2025-12-09GXSPMYKL010ORB79174342",
			expected: InputTypeTransactionID,
		},
		{
			name:     "Invalid E2E ID - wrong length",
			input:    "20251209GXSPMYKL010ORB791743",
			expected: InputTypeTransactionID,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: InputTypeTransactionID,
		},
		{
			name:     "Numeric string",
			input:    "1234567890",
			expected: InputTypeTransactionID,
		},
		{
			name:     "Lowercase E2E prefix",
			input:    "20251209gxspmymy010orb79174342",
			expected: InputTypeTransactionID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ClassifyInput(tc.input)
			if result != tc.expected {
				t.Errorf("ClassifyInput(%s) = %v; expected %v", tc.input, result, tc.expected)
			}
		})
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
		{"20251209GXSPMXKL010ORB79174342", false},   // Wrong prefix
		{"20251209gxspmymy010orb79174342", false},   // Lowercase
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
