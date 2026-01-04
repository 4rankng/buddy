package jira

import (
	"fmt"
	"os"
	"strings"

	"buddy/internal/apps/common"
	"buddy/internal/clients/jira"
	"buddy/internal/ui"

	"github.com/spf13/cobra"
)

// NewJiraViewCmd creates a command to view a specific JIRA ticket details
func NewJiraViewCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [ticket-key]",
		Short: "View JIRA ticket details",
		Long: `View detailed information about a specific JIRA ticket including:
- Summary, status, priority, assignee
- Created and due dates
- Full description
- Attachments

Example:
  buddy jira view TS-4565`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ticketKey := args[0]

			// Check if JIRA client is initialized
			if jira.Jira == nil {
				fmt.Printf("Error: JIRA client not initialized. Please ensure JIRA is properly configured.\n")
				os.Exit(1)
			}

			// Get JIRA configuration
			jiraConfig := jira.GetJiraConfig(appCtx.Environment)

			// Validate JIRA username is configured
			if jiraConfig.Auth.Username == "" {
				fmt.Printf("Error: JIRA_USERNAME not configured. Please set it in your .env.%s file\n", appCtx.Environment)
				os.Exit(1)
			}

			// Create context with timeout
			ctx := cmd.Context()

			// Get ticket details
			ticket, err := jira.Jira.GetIssueDetails(ctx, ticketKey)
			if err != nil {
				fmt.Printf("Error fetching JIRA ticket %s: %v\n", ticketKey, err)
				os.Exit(1)
			}

			// Map jira.JiraTicket to ui.JiraIssue
			uiIssue := ui.JiraIssue{
				Key:         ticket.Key,
				Summary:     ticket.Summary,
				Status:      ticket.Status,
				Priority:    ticket.Priority,
				Assignee:    ticket.Assignee,
				IssueType:   ticket.IssueType,
				CreatedAt:   ticket.CreatedAt,
				DueAt:       ticket.DueAt,
				Description: ticket.Description,
				Attachments: ticket.Attachments,
			}

			// Build BaseBrowseURL from JIRA_BASE_URL or jiraConfig.Domain
			baseURL := os.Getenv("JIRA_BASE_URL")
			if baseURL == "" && jiraConfig.Domain != "" {
				baseURL = jiraConfig.Domain
			}

			// Create picker config with view settings
			config := ui.JiraPickerConfig{
				ProjectKey:        jiraConfig.Project,
				BaseBrowseURL:     baseURL + "/browse",
				ShowAttachments:   true,
				MaxDescriptionLen: 0, // No limit - show full description
				HyperlinksMode:    ui.HyperlinksAuto,
				JiraClient:        jira.Jira,
			}

			// Print ticket details
			ui.PrintDetails(uiIssue, config, baseURL+"/browse")
		},
	}
	return cmd
}

// NewJiraSearchCmd creates a generic JIRA search command that can be used by both mybuddy and sgbuddy
func NewJiraSearchCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [search terms...]",
		Short: "Search your unresolved JIRA tickets",
		Long: `Search through your unresolved JIRA tickets (not Closed or Completed).
Searches in ticket summary and description fields.

Examples:
  buddy jira search "payment issue"
  buddy jira search mybuddy
  buddy jira search "API" "error"`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Check if JIRA client is initialized
			if jira.Jira == nil {
				fmt.Printf("Error: JIRA client not initialized. Please ensure JIRA is properly configured.\n")
				os.Exit(1)
			}

			// Get JIRA configuration
			jiraConfig := jira.GetJiraConfig(appCtx.Environment)

			// Validate JIRA username is configured
			if jiraConfig.Auth.Username == "" {
				fmt.Printf("Error: JIRA_USERNAME not configured. Please set it in your .env.%s file\n", appCtx.Environment)
				os.Exit(1)
			}

			// Join all args with spaces for search term
			searchTerm := strings.Join(args, " ")

			// Create context with timeout
			ctx := cmd.Context()

			// Search tickets
			issues, err := jira.Jira.SearchIssues(ctx, searchTerm)
			if err != nil {
				fmt.Printf("Error searching JIRA issues: %v\n", err)
				os.Exit(1)
			}

			if len(issues) == 0 {
				fmt.Printf("No tickets found matching '%s'\n", searchTerm)
				return
			}

			// Build base URL
			baseURL := os.Getenv("JIRA_BASE_URL")
			if baseURL == "" && jiraConfig.Domain != "" {
				baseURL = jiraConfig.Domain
			}

			// Display results as hyperlinks
			for _, issue := range issues {
				ticketURL := fmt.Sprintf("%s/browse/%s", baseURL, issue.Key)
				fmt.Println(ui.CreateHyperlink(ticketURL, issue.Key))
			}
		},
	}
	return cmd
}
