package utils

import (
	"bufio"
	"os"
	"strings"
)

// ReadTransactionIDsFromFile reads transaction IDs from a file, one per line
func ReadTransactionIDsFromFile(filePath string) ([]string, error) {
	var ids []string
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint:errcheck // Safe to ignore as we're only reading from the file

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id := strings.TrimSpace(scanner.Text())
		if id != "" {
			ids = append(ids, id)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

// IsSimpleFilePath checks if the input looks like a file path (simple version for service layer)
// For more sophisticated checking, use domain.IsFilePath
func IsSimpleFilePath(input string) bool {
	// Empty string is not a file path
	if input == "" {
		return false
	}

	// Check for path separators
	if strings.Contains(input, "/") || strings.Contains(input, "\\") {
		return true
	}

	// Check for file extensions
	if strings.Contains(input, ".txt") || strings.Contains(input, ".csv") || strings.Contains(input, ".log") {
		return true
	}

	// Check if it exists as a file
	if _, err := os.Stat(input); err == nil {
		return !strings.Contains(input, " ")
	}

	return false
}
