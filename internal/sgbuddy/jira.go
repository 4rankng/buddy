package sgbuddy

import (
	"bufio"
	"fmt"
	"os"
	"unicode/utf8"

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

	return jiraCmd
}

func NewJiraListCmd(appCtx *app.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIRA tickets assigned to you",
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

			// Get assigned issues using currentUser() for more efficient query
			emails := []string{"currentUser()"}
			issues, err := clients.Jira.GetAssignedIssues(ctx, jiraConfig.Project, emails)
			if err != nil {
				fmt.Printf("Error fetching JIRA issues: %v\n", err)
				os.Exit(1)
			}

			// Display results
			printJiraIssues(issues, jiraConfig)
		},
	}

	return cmd
}

func printJiraIssues(issues []clients.JiraTicket, jiraConfig clients.JiraConfig) {
	fmt.Printf("[jira]\n")
	fmt.Printf("project: %s\n", jiraConfig.Project)
	fmt.Printf("total: %d\n\n", len(issues))

	if len(issues) == 0 {
		fmt.Printf("No issues found.\n")
		return
	}

	isInteractive := ui.IsInteractive()
	descLength := ui.GetDescriptionLength()

	for i, issue := range issues {
		fmt.Printf("[issue %d]\n", i+1)

		// Display key as clickable hyperlink
		ticketURL := fmt.Sprintf("%s/browse/%s", jiraConfig.Domain, issue.Key)
		fmt.Printf("key: %s\n", ui.CreateHyperlink(ticketURL, issue.Key))

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

		// Handle description display
		if issue.Description != "" {
			if isInteractive {
				// Interactive mode with expand/collapse
				displayInteractiveDescription(issue.Description, descLength)
			} else {
				// Non-interactive mode: show truncated or full based on length
				if utf8.RuneCountInString(issue.Description) > descLength {
					truncated := ui.TruncateText(issue.Description, descLength)
					fmt.Printf("description: %s\n", truncated)
				} else {
					// Show full description if it's short enough
					lines := ui.WordWrap(issue.Description, 80)
					for j, line := range lines {
						if j == 0 {
							fmt.Printf("description: %s\n", line)
						} else {
							fmt.Printf("             %s\n", line)
						}
					}
				}
			}
		}

		fmt.Println()
	}
}

func displayInteractiveDescription(description string, truncateLength int) {
	reader := bufio.NewReader(os.Stdin)
	expanded := false
	truncated := ui.TruncateText(description, truncateLength)

	for {
		if !expanded {
			// Show truncated description with "Read more" option
			if utf8.RuneCountInString(description) > truncateLength {
				fmt.Printf("description: %s [Read more]\n", truncated)
			} else {
				// Description is short enough, show full
				lines := ui.WordWrap(description, 80)
				for j, line := range lines {
					if j == 0 {
						fmt.Printf("description: %s\n", line)
					} else {
						fmt.Printf("             %s\n", line)
					}
				}
				break
			}
		} else {
			// Show full description with "Read less" option
			lines := ui.WordWrap(description, 80)
			for j, line := range lines {
				if j == 0 {
					fmt.Printf("description: %s\n", line)
				} else {
					fmt.Printf("             %s\n", line)
				}
			}
			fmt.Printf("[Read less]\n")
		}

		// Prompt user for action
		fmt.Print("\nPress Enter to continue...")
		_, _ = reader.ReadString('\n')

		// Toggle expansion state
		expanded = !expanded

		// Clear the description area for redisplay
		// Move cursor up to overwrite the description
		if expanded {
			// We're about to show full description, need to clear truncated
			fmt.Printf("\033[1A\033[2K") // Move up and clear line
		} else {
			// We're about to show truncated, need to clear full description
			lines := ui.WordWrap(description, 80)
			for range lines {
				fmt.Printf("\033[1A\033[2K") // Move up and clear each line
			}
			fmt.Printf("\033[1A\033[2K") // Clear the "[Read less]" line
			fmt.Printf("\033[1A\033[2K") // Clear the "Press Enter" line
		}
	}
}
