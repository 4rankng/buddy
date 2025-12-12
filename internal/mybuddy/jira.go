package mybuddy

import (
	"fmt"
	"os"

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

			// Display results
			printJiraIssues(issues, jiraConfig.Project)
		},
	}

	return cmd
}

func printJiraIssues(issues []clients.JiraTicket, project string) {
	fmt.Printf("[jira]\n")
	fmt.Printf("project: %s\n", project)
	fmt.Printf("total: %d\n\n", len(issues))

	if len(issues) == 0 {
		fmt.Printf("No issues found.\n")
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

		// Show attachment count if any
		if len(issue.Attachments) > 0 {
			fmt.Printf("attachments: %d\n", len(issue.Attachments))
		}

		fmt.Println()
	}
}
