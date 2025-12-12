package sgbuddy

import (
	"fmt"
	"os"
	"strings"

	"buddy/clients"
	"buddy/internal/app"

	"github.com/spf13/cobra"
)

func NewJiraCmd(appCtx *app.Context) *cobra.Command {
	jiraCmd := &cobra.Command{
		Use:   "jira",
		Short: "JIRA ticket operations",
		Long:  `Manage JIRA tickets - list, view details, and more`,
	}

	jiraCmd.AddCommand(NewJiraListCmd(appCtx))

	return jiraCmd
}

func NewJiraListCmd(appCtx *app.Context) *cobra.Command {
	var (
		status string
		limit  int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIRA tickets assigned to you",
		Long: `List JIRA tickets assigned to the current user (based on JIRA_USERNAME configuration).

Examples:
  sgbuddy jira list                    # List open tickets (default)
  sgbuddy jira list --status=all       # List all tickets
  sgbuddy jira list --status=done      # List completed tickets
  sgbuddy jira list --limit=10         # List maximum 10 tickets`,
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

			// Get assigned issues using currentUser() for more efficient query
			emails := []string{"currentUser()"}
			issues, err := clients.Jira.GetAssignedIssues(ctx, jiraConfig.Project, emails)
			if err != nil {
				fmt.Printf("Error fetching JIRA issues: %v\n", err)
				os.Exit(1)
			}

			// Filter by status if specified
			if status != "all" {
				var filteredIssues []clients.JiraTicket
				for _, issue := range issues {
					issueStatus := strings.ToLower(issue.Status)
					switch status {
					case "open":
						if issueStatus == "to do" || issueStatus == "open" {
							filteredIssues = append(filteredIssues, issue)
						}
					case "in-progress":
						if issueStatus == "in progress" || issueStatus == "progress" {
							filteredIssues = append(filteredIssues, issue)
						}
					case "done":
						if issueStatus == "done" || issueStatus == "closed" || issueStatus == "resolved" {
							filteredIssues = append(filteredIssues, issue)
						}
					}
				}
				issues = filteredIssues
			}

			// Apply limit if specified
			if limit > 0 && len(issues) > limit {
				issues = issues[:limit]
			}

			// Display results
			printJiraIssues(issues, jiraConfig.Project, status)
		},
	}

	cmd.Flags().StringVar(&status, "status", "open", "Filter by status (open, in-progress, done, all)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of issues to display (0 = no limit)")

	return cmd
}

func printJiraIssues(issues []clients.JiraTicket, project, statusFilter string) {
	fmt.Printf("[jira]\n")
	fmt.Printf("project: %s\n", project)
	fmt.Printf("status_filter: %s\n", statusFilter)
	fmt.Printf("total: %d\n\n", len(issues))

	if len(issues) == 0 {
		fmt.Printf("No issues found matching your criteria.\n")
		return
	}

	for i, issue := range issues {
		fmt.Printf("[issue %d]\n", i+1)
		fmt.Printf("key: %s\n", issue.Key)
		fmt.Printf("summary: %s\n", issue.Summary)
		fmt.Printf("status: %s\n", issue.Status)
		if issue.Priority != "" {
			fmt.Printf("priority: %s\n", issue.Priority)
		}
		if issue.Assignee != "" {
			fmt.Printf("assignee: %s\n", issue.Assignee)
		}
		if !issue.CreatedAt.IsZero() {
			fmt.Printf("created: %s\n", issue.CreatedAt.Format("2006-01-02"))
		}
		if issue.DueAt != nil && !issue.DueAt.IsZero() {
			fmt.Printf("due: %s\n", issue.DueAt.Format("2006-01-02"))
		}
		fmt.Printf("type: %s\n", issue.IssueType)
		fmt.Println()
	}
}
