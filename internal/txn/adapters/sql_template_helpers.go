package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"strings"
	"time"
)

// getRPPWorkflowRunIDByCriteria finds and returns the run_id of a workflow matching specific criteria.
// Parameters:
//   - workflows: slice of workflows to search
//   - workflowID: workflow_id to match (empty string means any workflow_id)
//   - state: state to match (empty string means any state)
//   - attempt: attempt number to match (-1 means any attempt)
//
// Returns empty string if no matching workflow is found.
func getRPPWorkflowRunIDByCriteria(workflows []domain.WorkflowInfo, workflowID, state string, attempt int) string {
	for _, wf := range workflows {
		// Check workflow_id if specified
		if workflowID != "" && wf.WorkflowID != workflowID {
			continue
		}
		// Check state if specified
		if state != "" && wf.State != state {
			continue
		}
		// Check attempt if specified (and not -1 which means any attempt)
		if attempt != -1 && wf.Attempt != attempt {
			continue
		}
		// All criteria matched
		return wf.RunID
	}
	// No matching workflow found
	return ""
}

// convertUTCToGMT8 converts a UTC timestamp to GMT+8 format by adding 8 hours.
// Handles standard MySQL datetime format: "2006-01-02 15:04:05"
// Parameters:
//   - utcTimestamp: UTC timestamp string in MySQL datetime format
//
// Returns the GMT+8 timestamp string and any parsing error.
func convertUTCToGMT8(utcTimestamp string) (string, error) {
	if utcTimestamp == "" {
		return "", fmt.Errorf("UTC timestamp cannot be empty")
	}

	// Parse the UTC timestamp using MySQL datetime format
	utcTime, err := time.Parse("2006-01-02 15:04:05", utcTimestamp)
	if err != nil {
		return "", fmt.Errorf("invalid timestamp format '%s': expected 'YYYY-MM-DD HH:MM:SS', got error: %w", utcTimestamp, err)
	}

	// Add 8 hours to convert UTC to GMT+8
	gmt8Time := utcTime.Add(8 * time.Hour)

	// Format back to MySQL datetime format
	return gmt8Time.Format("2006-01-02 15:04:05"), nil
}

// compareTimestampsWithTimezone compares a UTC timestamp with a GMT+8 timestamp
// by converting the UTC timestamp to GMT+8 and checking if they match.
// Parameters:
//   - utcTimestamp: UTC timestamp string in MySQL datetime format
//   - gmt8Timestamp: GMT+8 timestamp string in MySQL datetime format
//
// Returns true if timestamps match after timezone conversion, false otherwise, and any error.
func compareTimestampsWithTimezone(utcTimestamp, gmt8Timestamp string) (bool, error) {
	if utcTimestamp == "" || gmt8Timestamp == "" {
		return false, fmt.Errorf("timestamps cannot be empty: UTC='%s', GMT+8='%s'", utcTimestamp, gmt8Timestamp)
	}

	// Convert UTC to GMT+8
	convertedTimestamp, err := convertUTCToGMT8(utcTimestamp)
	if err != nil {
		return false, fmt.Errorf("failed to convert UTC timestamp: %w", err)
	}

	// Validate GMT+8 timestamp format
	_, err = time.Parse("2006-01-02 15:04:05", gmt8Timestamp)
	if err != nil {
		return false, fmt.Errorf("invalid GMT+8 timestamp format '%s': expected 'YYYY-MM-DD HH:MM:SS', got error: %w", gmt8Timestamp, err)
	}

	// Compare the converted UTC timestamp with the GMT+8 timestamp
	// Use strings.TrimSpace to handle any potential whitespace differences
	return strings.TrimSpace(convertedTimestamp) == strings.TrimSpace(gmt8Timestamp), nil
}
