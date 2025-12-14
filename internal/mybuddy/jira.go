package mybuddy

import (
	"fmt"
	"os"
	"strings"

	"buddy/clients"
	"buddy/internal/app"
	"buddy/internal/ui"

	"github.com/spf13/cobra"
)

func NewJiraCmd(appCtx *app.Context) *cobra.Command {
	jiraCmd := &cobra.Command{
		Use:   "jira",
		Short: "JIRA ticket operations",
		Long:  `Manage JIRA tickets - list, view details, and more`,
	}

	jiraCmd.AddCommand(NewJiraListCmd(appCtx))
	jiraCmd.AddCommand(NewJiraSearchCmd(appCtx))

	return jiraCmd
}

func NewJiraListCmd(appCtx *app.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIRA tickets",
		Long: `List JIRA tickets assigned to the current user that are not completed or closed.

This command fetches tickets assigned to you from JIRA that are currently in progress.
Only tickets with status NOT IN (COMPLETED, CLOSED) will be shown.`,
		Run: func(cmd *cobra.Command, args []string) {
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

			// Create context with timeout
			ctx := cmd.Context()

			// Get assigned issues for current user
			emails := []string{"currentUser()"}
			issues, err := clients.Jira.GetAssignedIssues(ctx, jiraConfig.Project, emails)
			if err != nil {
				fmt.Printf("Error fetching JIRA issues: %v\n", err)
				os.Exit(1)
			}

			// Map clients.JiraTicket to ui.JiraIssue
			uiIssues := make([]ui.JiraIssue, len(issues))
			for i, issue := range issues {
				uiIssues[i] = ui.JiraIssue{
					Key:         issue.Key,
					Summary:     issue.Summary,
					Status:      issue.Status,
					Priority:    issue.Priority,
					Assignee:    issue.Assignee,
					IssueType:   issue.IssueType,
					CreatedAt:   issue.CreatedAt,
					DueAt:       issue.DueAt,
					Description: issue.Description,
					Attachments: issue.Attachments,
				}
			}

			// Build BaseBrowseURL from JIRA_BASE_URL or jiraConfig.Domain
			baseURL := os.Getenv("JIRA_BASE_URL")
			if baseURL == "" && jiraConfig.Domain != "" {
				baseURL = jiraConfig.Domain
			}

			// Create picker config with mybuddy settings
			config := ui.JiraPickerConfig{
				ProjectKey:        jiraConfig.Project,
				BaseBrowseURL:     baseURL + "/browse",
				ShowAttachments:   true, // mybuddy shows attachments
				MaxDescriptionLen: 0,    // No limit - show full description
				HyperlinksMode:    ui.HyperlinksAuto,
				JiraClient:        clients.Jira,
			}

			// Run picker
			if err := ui.RunJiraPicker(uiIssues, config); err != nil {
				fmt.Printf("Picker error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func NewJiraSearchCmd(appCtx *app.Context) *cobra.Command {
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

			// Join all args with spaces for search term
			searchTerm := strings.Join(args, " ")

			// Create context with timeout
			ctx := cmd.Context()

			// Search tickets
			issues, err := clients.Jira.SearchIssues(ctx, searchTerm)
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
