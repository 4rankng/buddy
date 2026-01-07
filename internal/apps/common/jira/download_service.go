package jira

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"buddy/internal/clients/jira"
	"buddy/internal/logging"
)

// AttachmentDownloadService handles downloading attachments from JIRA tickets
type AttachmentDownloadService struct {
	jiraClient jira.JiraInterface
	logger     *logging.Logger
}

// DownloadOptions contains options for downloading attachments
type DownloadOptions struct {
	TicketID  string
	OutputDir string
	CSVOnly   bool
}

// DownloadResult contains the results of a download operation
type DownloadResult struct {
	TicketID        string
	TotalFound      int
	CSVFound        int
	Downloaded      int
	Failed          int
	DownloadedFiles []string
	Errors          []error
}

// NewAttachmentDownloadService creates a new download service
func NewAttachmentDownloadService(jiraClient jira.JiraInterface) *AttachmentDownloadService {
	return &AttachmentDownloadService{
		jiraClient: jiraClient,
		logger:     logging.NewDefaultLogger("jira-download"),
	}
}

// DownloadAttachments downloads attachments from a JIRA ticket
func (s *AttachmentDownloadService) DownloadAttachments(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	result := &DownloadResult{
		TicketID:        opts.TicketID,
		DownloadedFiles: make([]string, 0),
		Errors:          make([]error, 0),
	}

	// Validate ticket ID format
	if err := s.validateTicketID(opts.TicketID); err != nil {
		return nil, fmt.Errorf("invalid ticket ID: %w", err)
	}

	// Get ticket details
	fmt.Printf("Retrieving ticket details for %s...\n", opts.TicketID)
	ticket, err := s.jiraClient.GetIssueDetails(ctx, opts.TicketID)
	if err != nil {
		return nil, s.handleJiraError(err, opts.TicketID)
	}

	if ticket == nil {
		fmt.Printf("Ticket %s not found. Please verify the ticket ID exists and you have access to it.\n", opts.TicketID)
		return nil, fmt.Errorf("ticket %s not found", opts.TicketID)
	}

	result.TotalFound = len(ticket.Attachments)
	fmt.Printf("Found %d total attachment(s) in ticket %s\n", result.TotalFound, opts.TicketID)

	if result.TotalFound == 0 {
		fmt.Printf("No attachments found in ticket %s\n", opts.TicketID)
		return result, nil
	}

	// Filter for CSV files if requested
	attachments := ticket.Attachments
	if opts.CSVOnly {
		attachments = s.filterCSVAttachments(ticket.Attachments)
	}
	result.CSVFound = len(attachments)

	if result.CSVFound == 0 {
		fmt.Printf("No CSV attachments found in ticket %s\n", opts.TicketID)
		return result, nil
	}

	fmt.Printf("Found %d CSV attachment(s) to download\n", result.CSVFound)

	// Ensure output directory exists and is writable
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}

	if err := s.ensureOutputDirectory(opts.OutputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory '%s': %w\nPlease check directory permissions", opts.OutputDir, err)
	}

	// Check if output directory is writable
	if err := s.checkDirectoryWritable(opts.OutputDir); err != nil {
		return nil, fmt.Errorf("output directory '%s' is not writable: %w\nPlease check directory permissions", opts.OutputDir, err)
	}

	// Download each attachment
	for i, attachment := range attachments {
		fmt.Printf("Downloading %d/%d: %s", i+1, len(attachments), attachment.Filename)

		filePath := filepath.Join(opts.OutputDir, attachment.Filename)

		// Handle filename conflicts
		finalPath := s.resolveFilenameConflict(filePath)
		if finalPath != filePath {
			fmt.Printf(" (saved as %s)", filepath.Base(finalPath))
		}
		fmt.Printf("...")

		err := s.jiraClient.DownloadAttachment(ctx, attachment, finalPath)
		if err != nil {
			result.Failed++
			downloadErr := s.handleDownloadError(err, attachment.Filename)
			result.Errors = append(result.Errors, downloadErr)
			s.logger.Error("Failed to download attachment %s: %v", attachment.Filename, err)
			fmt.Printf(" ✗ FAILED\n")
		} else {
			result.Downloaded++
			result.DownloadedFiles = append(result.DownloadedFiles, finalPath)
			s.logger.Info("Successfully downloaded: %s", finalPath)
			fmt.Printf(" ✓ SUCCESS\n")
		}
	}

	return result, nil
}

// filterCSVAttachments filters attachments to only include CSV files
func (s *AttachmentDownloadService) filterCSVAttachments(attachments []jira.Attachment) []jira.Attachment {
	var csvAttachments []jira.Attachment

	for _, attachment := range attachments {
		if s.isCSVFile(attachment) {
			csvAttachments = append(csvAttachments, attachment)
		}
	}

	return csvAttachments
}

// isCSVFile checks if an attachment is a CSV file based on filename extension or mime type
func (s *AttachmentDownloadService) isCSVFile(attachment jira.Attachment) bool {
	// Check file extension
	if strings.HasSuffix(strings.ToLower(attachment.Filename), ".csv") {
		return true
	}

	// Check mime type
	mimeType := strings.ToLower(attachment.MimeType)
	return mimeType == "text/csv" ||
		mimeType == "application/csv" ||
		strings.Contains(mimeType, "csv")
}

// ensureOutputDirectory creates the output directory if it doesn't exist
func (s *AttachmentDownloadService) ensureOutputDirectory(dir string) error {
	if dir == "." || dir == "" {
		return nil // Current directory always exists
	}

	return os.MkdirAll(dir, 0755)
}

// resolveFilenameConflict handles filename conflicts by appending numeric suffixes
func (s *AttachmentDownloadService) resolveFilenameConflict(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath // No conflict
	}

	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s_%d%s", base, counter, ext)
		newPath := filepath.Join(dir, newFilename)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

// validateTicketID performs basic validation on the ticket ID format
func (s *AttachmentDownloadService) validateTicketID(ticketID string) error {
	if ticketID == "" {
		return fmt.Errorf("ticket ID cannot be empty")
	}

	// Basic format check - should contain at least one letter and one number with a dash
	if !strings.Contains(ticketID, "-") {
		return fmt.Errorf("ticket ID '%s' should be in format like 'TS-1234' or 'TSE-567'", ticketID)
	}

	return nil
}

// handleJiraError provides user-friendly error messages for JIRA API errors
func (s *AttachmentDownloadService) handleJiraError(err error, ticketID string) error {
	errMsg := err.Error()

	// Check for common error patterns
	if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized") {
		return fmt.Errorf("authentication failed: %w\n\nPlease check your JIRA credentials:\n- Verify JIRA_USERNAME is set correctly\n- Verify JIRA_API_KEY is valid\n- Check your .env file configuration", err)
	}

	if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden") {
		return fmt.Errorf("access denied to ticket %s: %w\n\nPossible causes:\n- You don't have permission to view this ticket\n- The ticket is in a restricted project\n- Your JIRA account lacks necessary permissions", ticketID, err)
	}

	if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found") {
		return fmt.Errorf("ticket %s not found: %w\n\nPlease verify:\n- The ticket ID is correct\n- The ticket exists in your JIRA instance\n- You have access to the project", ticketID, err)
	}

	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "connection") {
		return fmt.Errorf("network error connecting to JIRA: %w\n\nSuggestions:\n- Check your internet connection\n- Verify JIRA server is accessible\n- Try again in a few moments", err)
	}

	// Generic error with helpful context
	return fmt.Errorf("failed to get ticket details for %s: %w\n\nIf this persists, please check your JIRA configuration and network connectivity", ticketID, err)
}

// handleDownloadError provides user-friendly error messages for download failures
func (s *AttachmentDownloadService) handleDownloadError(err error, filename string) error {
	errMsg := err.Error()

	if strings.Contains(errMsg, "permission denied") || strings.Contains(errMsg, "access denied") {
		return fmt.Errorf("failed to save %s: permission denied\n\nPlease check that you have write permissions to the output directory", filename)
	}

	if strings.Contains(errMsg, "no space left") || strings.Contains(errMsg, "disk full") {
		return fmt.Errorf("failed to save %s: insufficient disk space\n\nPlease free up disk space and try again", filename)
	}

	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "connection") {
		return fmt.Errorf("failed to download %s: network error\n\nSuggestion: Try again in a few moments", filename)
	}

	return fmt.Errorf("failed to download %s: %w", filename, err)
}

// checkDirectoryWritable tests if a directory is writable by creating a temporary file
func (s *AttachmentDownloadService) checkDirectoryWritable(dir string) error {
	testFile := filepath.Join(dir, ".jira_download_test")

	// Try to create a test file
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}

	// Clean up the test file
	if err := file.Close(); err != nil {
		s.logger.Warn("Failed to close test file: %v", err)
	}
	if err := os.Remove(testFile); err != nil {
		s.logger.Warn("Failed to remove test file: %v", err)
	}

	return nil
}
