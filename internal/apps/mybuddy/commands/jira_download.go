package mybuddy

import (
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/apps/common/jira"
	"buddy/internal/di"

	"github.com/spf13/cobra"
)

// NewJiraDownloadAttachmentCmd creates a command to download CSV attachments from JIRA tickets
func NewJiraDownloadAttachmentCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "download-attachment [ticket-id]",
		Short: "Download CSV attachments from JIRA ticket",
		Long: `Download all CSV attachments from a JIRA ticket to the current directory.

This command fetches all CSV attachments from the specified JIRA ticket and downloads
them to the current working directory (or specified output directory).

Only CSV files (based on file extension .csv or CSV mime type) will be downloaded.
If filename conflicts occur, numeric suffixes will be added automatically.

Examples:
  mybuddy jira download-attachment TS-1234
  mybuddy jira download-attachment TS-1234 --output ./downloads`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ticketID := args[0]

			// Check if JIRA client is initialized
			if clients.Jira == nil {
				fmt.Printf("Error: JIRA client not initialized. Please ensure JIRA is properly configured.\n")
				os.Exit(1)
			}

			// Get JIRA configuration
			jiraConfig := clients.GetJiraConfig(appCtx.Environment)

			// Validate JIRA username is configured
			if jiraConfig.Auth.Username == "" {
				fmt.Printf("Error: JIRA_USERNAME not configured. Please set it in your .env.%s file\n", appCtx.Environment)
				os.Exit(1)
			}

			// Create download service
			downloadService := jira.NewAttachmentDownloadService(clients.Jira)

			// Prepare download options
			opts := jira.DownloadOptions{
				TicketID:  ticketID,
				OutputDir: outputDir,
				CSVOnly:   true,
			}

			// Create context with timeout
			ctx := cmd.Context()

			fmt.Printf("Fetching attachments for %s...\n", ticketID)

			// Download attachments
			result, err := downloadService.DownloadAttachments(ctx, opts)
			if err != nil {
				fmt.Printf("Error downloading attachments: %v\n", err)
				os.Exit(1)
			}

			// Display results
			displayDownloadResults(result)
		},
	}

	// Add output directory flag
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for downloaded files")

	return cmd
}

// displayDownloadResults shows the results of the download operation
func displayDownloadResults(result *jira.DownloadResult) {
	fmt.Printf("\nDownload Summary for %s:\n", result.TicketID)
	fmt.Printf("Total attachments found: %d\n", result.TotalFound)
	fmt.Printf("CSV files found: %d\n", result.CSVFound)

	if result.CSVFound == 0 {
		fmt.Printf("No CSV attachments found in ticket %s.\n", result.TicketID)
		return
	}

	fmt.Printf("Successfully downloaded: %d\n", result.Downloaded)
	if result.Failed > 0 {
		fmt.Printf("Failed downloads: %d\n", result.Failed)
	}

	// Show downloaded files
	if len(result.DownloadedFiles) > 0 {
		fmt.Printf("\nDownloaded files:\n")
		for _, filePath := range result.DownloadedFiles {
			fmt.Printf("  ✓ %s\n", filePath)
		}
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors encountered:\n")
		for _, err := range result.Errors {
			fmt.Printf("  ✗ %v\n", err)
		}
	}

	fmt.Printf("\n")
}
