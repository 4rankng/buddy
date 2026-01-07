package jira

import (
	"context"
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
			RunJiraView(cmd.Context(), appCtx, args[0], true)
		},
	}
	return cmd
}

// RunJiraView fetches and displays details for a single JIRA ticket
func RunJiraView(ctx context.Context, appCtx *common.Context, issueKey string, showAttachments bool) {
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

	// Get ticket details
	ticket, err := jira.Jira.GetIssueDetails(ctx, issueKey)
	if err != nil {
		fmt.Printf("Error fetching JIRA ticket %s: %v\n", issueKey, err)
		os.Exit(1)
	}

	if ticket == nil {
		fmt.Printf("Error: Ticket %s not found\n", issueKey)
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

	// Create picker config
	config := ui.JiraPickerConfig{
		ProjectKey:        jiraConfig.Project,
		BaseBrowseURL:     baseURL + "/browse",
		ShowAttachments:   showAttachments,
		MaxDescriptionLen: 0, // No limit - show full description
		HyperlinksMode:    ui.HyperlinksAuto,
		JiraClient:        jira.Jira,
	}

	// Print ticket details
	ui.PrintDetails(uiIssue, config, baseURL+"/browse")
}

// NewJiraSearchCmd creates a generic JIRA search command that can be used by both mybuddy and sgbuddy
func NewJiraSearchCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [search terms...]",
		Short: "Search your unresolved JIRA tickets",
		Long: `Search through your unresolved JIRA tickets (not Closed or Completed).
Searches in ticket summary and description fields.

Supports both text search and raw JQL queries:
- Text search: buddy jira search "payment issue"
- JQL query: buddy jira search "project = TS"

Examples:
  buddy jira search "payment issue"
  buddy jira search mybuddy
  buddy jira search "API" "error"
  buddy jira search "project = TS"
  buddy jira search "reporter = currentUser()"`,
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

			// Detect if this is a JQL query (contains field operators like =, !=, IN, etc.)
			isJQL := looksLikeJQL(searchTerm)

			var issues []jira.JiraTicket
			var err error

			if isJQL {
				// Execute raw JQL query
				issues, err = jira.Jira.ExecuteJQL(ctx, searchTerm)
			} else {
				// Search tickets in summary/description
				issues, err = jira.Jira.SearchIssues(ctx, searchTerm)
			}

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

// looksLikeJQL heuristically determines if the search term is a JQL query
func looksLikeJQL(term string) bool {
	term = strings.TrimSpace(term)

	// JQL field operators that indicate this is likely a JQL query
	jqlOperators := []string{" = ", " != ", " ~ ", " !~ ", " > ", " < ", " >= ", " <= ", " IN ", " NOT IN ", " WAS ", " WAS IN ", " WAS NOT IN ", " CHANGED "}

	for _, op := range jqlOperators {
		if strings.Contains(term, op) {
			return true
		}
	}

	return false
}
