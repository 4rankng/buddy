package sgtxn

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ProcessSGBatchFile processes multiple transactions from a file
func ProcessSGBatchFile(filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	index := 1

	// Process each line as a transaction ID
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		// Query the transaction
		result := QuerySGTransaction(line)

		// Write the result
		WriteSGResult(os.Stdout, *result, index)

		index++
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

// IsFile checks if the given path is a file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
