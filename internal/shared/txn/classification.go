package txn

import (
	"os"
	"path/filepath"
	"strings"
)

// InputType represents the type of input provided
type InputType string

const (
	InputTypeE2EID         InputType = "E2E_ID"
	InputTypeTransactionID InputType = "TRANSACTION_ID"
	InputTypeFilePath      InputType = "FILE_PATH"
)

// ClassifyInput determines the type of input provided
func ClassifyInput(input string) InputType {
	if IsRppE2EID(input) {
		return InputTypeE2EID
	}

	if IsFilePath(input) {
		return InputTypeFilePath
	}

	return InputTypeTransactionID
}

// IsRppE2EID checks if the input matches the RPP E2E ID pattern
func IsRppE2EID(input string) bool {
	return RppE2EIDPattern.MatchString(input)
}

// IsFilePath checks if the input looks like a file path
// This is a simple check - it looks for common file path patterns
func IsFilePath(input string) bool {
	// Empty string is not a file path
	if input == "" {
		return false
	}

	// Check for path separators first (most reliable indicator)
	if strings.Contains(input, "/") || strings.Contains(input, "\\") {
		return true
	}

	// Check for relative path indicators
	if strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../") {
		return true
	}

	// Check for file extensions (but be more strict)
	if strings.Contains(input, ".") {
		// Split on last dot
		parts := strings.Split(input, ".")
		if len(parts) > 1 && parts[len(parts)-1] != "" {
			// Check if the extension looks like a real file extension (1-5 chars)
			ext := parts[len(parts)-1]
			if len(ext) >= 1 && len(ext) <= 5 && !strings.Contains(ext, "/") && !strings.Contains(ext, "\\") {
				return true
			}
		}
		// Hidden files start with dot
		if strings.HasPrefix(input, ".") && len(input) > 1 {
			return true
		}
	}

	// Check if it's an existing file (quick check)
	if _, err := os.Stat(input); err == nil {
		return true
	}

	// Don't classify random strings as files unless they have clear file-like patterns
	return false
}

// IsAbsolutePath checks if a path is absolute
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// NormalizeFilePath normalizes a file path
func NormalizeFilePath(path string) string {
	// Use filepath.Clean to normalize the path
	cleaned := filepath.Clean(path)

	// Convert forward slashes to OS-specific separators
	return filepath.FromSlash(cleaned)
}
