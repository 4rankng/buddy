package mybuddy

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
		status    string
		limit     int
		assignee  string
		priority  string
		issueType string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIRA tickets",
		Long: `List JIRA tickets with various filtering options.

Examples:
  mybuddy jira list                           # List open tickets assigned to current user
  mybuddy jira list --status=all              # List all tickets
  mybuddy jira list --status=done             # List completed tickets
  mybuddy jira list --assignee="currentUser()" # Explicitly use currentUser()
  mybuddy jira list --assignee=user@example.com # List tickets for specific user
  mybuddy jira list --priority=high           # List high priority tickets
  mybuddy jira list --type=bug                # List bug tickets only
  mybuddy jira list --limit=10                # List maximum 10 tickets`,
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

			// Default to currentUser() if no assignee specified
			if assignee == "" {
				assignee = "currentUser()"
			}

			// Get assigned issues
			emails := []string{assignee}
			issues, err := clients.Jira.GetAssignedIssues(ctx, jiraConfig.Project, emails)
			if err != nil {
				fmt.Printf("Error fetching JIRA issues: %v\n", err)
				os.Exit(1)
			}

			// Apply filters
			issues = filterIssues(issues, status, priority, issueType)

			// Apply limit if specified
			if limit > 0 && len(issues) > limit {
				issues = issues[:limit]
			}

			// Display results
			printJiraIssues(issues, jiraConfig.Project, status, assignee, priority, issueType)
		},
	}

	cmd.Flags().StringVar(&status, "status", "open", "Filter by status (open, in-progress, done, all)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of issues to display (0 = no limit)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee (default: currentUser())")
	cmd.Flags().StringVar(&priority, "priority", "", "Filter by priority (highest, high, medium, low, lowest)")
	cmd.Flags().StringVar(&issueType, "type", "", "Filter by issue type (bug, story, task, etc.)")

	return cmd
}

// filterIssues applies filters to the list of issues
func filterIssues(issues []clients.JiraTicket, status, priority, issueType string) []clients.JiraTicket {
	var filteredIssues []clients.JiraTicket

	for _, issue := range issues {
		// Status filter
		if status != "all" {
			issueStatus := strings.ToLower(issue.Status)
			switch status {
			case "open":
				if issueStatus != "to do" && issueStatus != "open" {
					continue
				}
			case "in-progress":
				if issueStatus != "in progress" && issueStatus != "progress" {
					continue
				}
			case "done":
				if issueStatus != "done" && issueStatus != "closed" && issueStatus != "resolved" && issueStatus != "completed" {
					continue
				}
			}
		}

		// Priority filter
		if priority != "" {
			issuePriority := strings.ToLower(issue.Priority)
			if issuePriority != strings.ToLower(priority) {
				continue
			}
		}

		// Issue type filter
		if issueType != "" {
			issueTypeValue := strings.ToLower(issue.IssueType)
			if issueTypeValue != strings.ToLower(issueType) {
				continue
			}
		}

		filteredIssues = append(filteredIssues, issue)
	}

	return filteredIssues
}

func printJiraIssues(issues []clients.JiraTicket, project, statusFilter, assigneeFilter, priorityFilter, issueTypeFilter string) {
	fmt.Printf("[jira]\n")
	fmt.Printf("project: %s\n", project)
	
	// Show active filters
	var filters []string
	if statusFilter != "" && statusFilter != "open" {
		filters = append(filters, fmt.Sprintf("status=%s", statusFilter))
	}
	if assigneeFilter != "" && assigneeFilter != "currentUser()" {
		filters = append(filters, fmt.Sprintf("assignee=%s", assigneeFilter))
	}
	if priorityFilter != "" {
		filters = append(filters, fmt.Sprintf("priority=%s", priorityFilter))
	}
	if issueTypeFilter != "" {
		filters = append(filters, fmt.Sprintf("type=%s", issueTypeFilter))
	}
	
	if len(filters) > 0 {
		fmt.Printf("filters: %s\n", strings.Join(filters, ", "))
	}
	
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
		
		// Show attachment count if any
		if len(issue.Attachments) > 0 {
			fmt.Printf("attachments: %d\n", len(issue.Attachments))
		}
		
		fmt.Println()
	}
}