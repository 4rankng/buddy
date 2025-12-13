package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"buddy/clients"
	"github.com/manifoldco/promptui"
	"github.com/pkg/browser"
)

// Attachment represents a JIRA attachment for UI display
type Attachment = clients.Attachment

// JiraIssue represents a JIRA ticket for UI display
type JiraIssue struct {
	Key         string
	Summary     string
	Status      string
	Priority    string
	Assignee    string
	IssueType   string
	CreatedAt   time.Time
	DueAt       *time.Time
	Description string
	Attachments []Attachment
}

// HyperlinksMode controls hyperlink behavior
type HyperlinksMode int

const (
	HyperlinksAuto HyperlinksMode = iota
	HyperlinksOn
	HyperlinksOff
)

// JiraPickerConfig configures the interactive picker
type JiraPickerConfig struct {
	ProjectKey        string
	BaseBrowseURL     string // e.g. https://your-jira/browse (will be normalized)
	ShowAttachments   bool
	MaxDescriptionLen int // 0 = no limit
	HyperlinksMode    HyperlinksMode
	JiraClient        clients.JiraInterface // Client for downloading attachments
}

// RunJiraPicker runs an interactive JIRA ticket picker
func RunJiraPicker(issues []JiraIssue, cfg JiraPickerConfig) error {
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	// Normalize BaseBrowseURL
	baseURL := normalizeURL(cfg.BaseBrowseURL)

	for {
		// Show ticket selection
		issue, err := selectTicket(issues, cfg)
		if err != nil {
			if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
				return nil // Normal exit
			}
			return err
		}

		// Show ticket details
		printDetails(*issue, cfg, baseURL)

		// Show action menu
		action, attachmentIndex, err := selectAction(baseURL == "" /* hasBrowser */, cfg.ShowAttachments && len(issue.Attachments) > 0 /* hasAttachments */, issue.Attachments)
		if err != nil {
			if err == promptui.ErrEOF || err == promptui.ErrInterrupt {
				return nil // Normal exit
			}
			return err
		}

		switch action {
		case "Open in browser":
			if baseURL != "" {
				ticketURL := baseURL + "/" + issue.Key
				if err := browser.OpenURL(ticketURL); err != nil {
					fmt.Printf("Error opening browser: %v\n", err)
				}
			}
		case "Download attachment":
			if cfg.JiraClient != nil && cfg.ShowAttachments && len(issue.Attachments) > 0 {
				attachment, err := selectAttachmentToShow(issue.Attachments)
				if err != nil {
					fmt.Printf("Error selecting attachment: %v\n", err)
					continue
				}

				fmt.Printf("Downloading %s...\n", attachment.Filename)
				err = cfg.JiraClient.DownloadAttachment(context.Background(), *attachment, ".")
				if err != nil {
					fmt.Printf("Error downloading attachment: %v\n", err)
				} else {
					fmt.Printf("Successfully downloaded to current directory\n")
				}
			}
		case "download_by_number":
			if cfg.JiraClient != nil && attachmentIndex >= 0 && attachmentIndex < len(issue.Attachments) {
				attachment := issue.Attachments[attachmentIndex]
				fmt.Printf("Downloading %s...\n", attachment.Filename)
				err = cfg.JiraClient.DownloadAttachment(context.Background(), attachment, ".")
				if err != nil {
					fmt.Printf("Error downloading attachment: %v\n", err)
				} else {
					fmt.Printf("Successfully downloaded to current directory\n")
				}
			}
		case "Quit":
			return nil
		}
		// Loop continues for "Back to list" (default)
	}
}

// normalizeURL cleans up the base URL
func normalizeURL(url string) string {
	if url == "" {
		return ""
	}
	// Trim spaces and trailing slash
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	return url
}

// selectTicket shows the ticket selection prompt
func selectTicket(issues []JiraIssue, cfg JiraPickerConfig) (*JiraIssue, error) {
	// Prepare search items
	items := make([]string, len(issues))
	for i, issue := range issues {
		summary := issue.Summary
		if len(summary) > 80 {
			summary = summary[:77] + "..."
		}
		items[i] = fmt.Sprintf("[%d] %s - %s (%s)", i+1, issue.Key, summary, issue.Status)
	}

	prompt := promptui.Select{
		Label:             "Select a JIRA ticket",
		Items:             items,
		Size:              min(12, len(issues)),
		StartInSearchMode: true,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   `{{ "✔" | cyan }} {{ . | cyan }}`,
			Inactive: `  {{ . }}`,
			Selected: `{{ "✔" | green }} {{ . | green }}`,
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &issues[index], nil
}

// printDetails displays ticket details
func printDetails(issue JiraIssue, cfg JiraPickerConfig, baseURL string) {
	fmt.Printf("\n[issue]\n")

	// Print key with URL, using hyperlinks if enabled and supported
	if baseURL != "" {
		ticketURL := baseURL + "/" + issue.Key
		if shouldEnableHyperlinks(cfg.HyperlinksMode) {
			fmt.Printf("key: %s\n", CreateHyperlink(ticketURL, issue.Key))
		} else {
			fmt.Printf("key: %s (%s)\n", issue.Key, ticketURL)
		}
	} else {
		fmt.Printf("key: %s\n", issue.Key)
	}

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

	if cfg.ShowAttachments && len(issue.Attachments) > 0 {
		fmt.Printf("attachments:\n")
		for i, att := range issue.Attachments {
			if shouldEnableHyperlinks(cfg.HyperlinksMode) {
				fmt.Printf("  [%d] %s\n", i+1, CreateHyperlink(att.URL, att.Filename))
			} else {
				fmt.Printf("  [%d] %s (%s)\n", i+1, att.Filename, att.URL)
			}
		}
		fmt.Printf("\nTip: Enter attachment number to download\n")
	}

	// Print description
	if issue.Description != "" {
		desc := issue.Description
		if cfg.MaxDescriptionLen > 0 && len(desc) > cfg.MaxDescriptionLen {
			desc = desc[:cfg.MaxDescriptionLen-3] + "..."
		}
		lines := WordWrap(desc, 80)
		for i, line := range lines {
			if i == 0 {
				fmt.Printf("description: %s\n", line)
			} else {
				fmt.Printf("             %s\n", line)
			}
		}
	}

	fmt.Println()
}

// shouldEnableHyperlinks determines if hyperlinks should be enabled
func shouldEnableHyperlinks(mode HyperlinksMode) bool {
	switch mode {
	case HyperlinksOn:
		return true
	case HyperlinksOff:
		return false
	case HyperlinksAuto:
		// Disable if TERM is dumb or CI is set
		if os.Getenv("TERM") == "dumb" {
			return false
		}
		if os.Getenv("CI") != "" {
			return false
		}
		return true
	default:
		return false
	}
}

// selectAction shows the action selection prompt with support for attachment number input
func selectAction(hasBrowser bool, hasAttachments bool, attachments []Attachment) (string, int, error) {
	actions := []string{
		"Back to list",
		"Quit",
	}

	if hasBrowser {
		actions = append([]string{"Open in browser"}, actions...)
	}

	if hasAttachments {
		actions = append([]string{"Download attachment", "Enter attachment number"}, actions...)
	}

	prompt := promptui.Select{
		Label: "Select action",
		Items: actions,
		Size:  len(actions),
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   `{{ "✔" | cyan }} {{ . | cyan }}`,
			Inactive: `  {{ . }}`,
			Selected: `{{ "✔" | green }} {{ . | green }}`,
		},
	}

	_, action, err := prompt.Run()
	if err != nil {
		return "", 0, err
	}

	// If "Enter attachment number" is selected, prompt for the number
	if action == "Enter attachment number" && len(attachments) > 0 {
		validate := func(input string) error {
			num := 0
			_, err := fmt.Sscanf(input, "%d", &num)
			if err != nil {
				return fmt.Errorf("please enter a valid number")
			}
			if num < 1 || num > len(attachments) {
				return fmt.Errorf("please enter a number between 1 and %d", len(attachments))
			}
			return nil
		}

		promptNum := promptui.Prompt{
			Label:    "Enter attachment number",
			Validate: validate,
		}

		result, err := promptNum.Run()
		if err != nil {
			return "", 0, err
		}

		var attachmentNum int
		_, err = fmt.Sscanf(result, "%d", &attachmentNum)
		if err == nil {
			return "download_by_number", attachmentNum - 1, nil // Convert to 0-based index
		}
	}

	return action, 0, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// selectAttachmentToShow prompts user to select an attachment from the list
func selectAttachmentToShow(attachments []Attachment) (*Attachment, error) {
	if len(attachments) == 0 {
		return nil, fmt.Errorf("no attachments available")
	}

	// If there's only one attachment, return it
	if len(attachments) == 1 {
		return &attachments[0], nil
	}

	// Multiple attachments: let user choose
	items := make([]string, len(attachments))
	for i, att := range attachments {
		items[i] = att.Filename
	}

	prompt := promptui.Select{
		Label: "Select attachment to download",
		Items: items,
		Size:  min(12, len(attachments)),
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   `{{ "✔" | cyan }} {{ . | cyan }}`,
			Inactive: `  {{ . }}`,
			Selected: `{{ "✔" | green }} {{ . | green }}`,
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &attachments[index], nil
}
