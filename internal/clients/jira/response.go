package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"buddy/internal/errors"
)

// jiraIssueResponse represents the JIRA API issue response structure
type jiraIssueResponse struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary     any `json:"summary"`
		Description any `json:"description"`
		Assignee    *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority *struct {
			Name string `json:"name"`
		} `json:"priority"`
		Created          string `json:"created"`
		DueDate          string `json:"duedate"`
		CustomField10060 any    `json:"customfield_10060"`
		IssueType        struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Attachment []struct {
			ID       string `json:"id"`
			Filename string `json:"filename"`
			MimeType string `json:"mimeType"`
			Content  string `json:"content"`
		} `json:"attachment"`
	} `json:"fields"`
}

// convertIssueResponse converts JIRA API response to JiraTicket
func (c *JiraClient) convertIssueResponse(issue *jiraIssueResponse) (*JiraTicket, error) {
	ticket := &JiraTicket{
		ID:        issue.ID,
		Key:       issue.Key,
		IssueType: issue.Fields.IssueType.Name,
		CreatedAt: c.parseTime(issue.Fields.Created),
		Status:    issue.Fields.Status.Name,
	}

	// Summary
	if summary, ok := issue.Fields.Summary.(string); ok {
		ticket.Summary = summary
	}

	// Description
	ticket.Description = c.extractTextFromADF(issue.Fields.Description)

	// Assignee
	if issue.Fields.Assignee != nil {
		ticket.Assignee = issue.Fields.Assignee.DisplayName
	}

	// Priority
	if issue.Fields.Priority != nil {
		ticket.Priority = issue.Fields.Priority.Name
	}

	// Due date
	if dueDate := c.parseDueDate(issue.Fields.DueDate, issue.Fields.CustomField10060); !dueDate.IsZero() {
		ticket.DueAt = &dueDate
	}

	// Attachments
	for _, att := range issue.Fields.Attachment {
		ticket.Attachments = append(ticket.Attachments, Attachment{
			ID:       att.ID,
			Filename: att.Filename,
			MimeType: att.MimeType,
			URL:      att.Content,
			Content:  att.Content,
		})
	}

	return ticket, nil
}

// parseDueDate parses due date from multiple possible fields
func (c *JiraClient) parseDueDate(dueDate string, customField any) time.Time {
	if dueDate != "" {
		if parsed := c.parseTime(dueDate); !parsed.IsZero() {
			return parsed
		}
	}

	if customField != nil {
		if customFieldStr, ok := customField.(string); ok && customFieldStr != "" {
			if parsed := c.parseTime(customFieldStr); !parsed.IsZero() {
				return parsed
			}
		}
	}

	return time.Time{}
}

// extractTextFromADF extracts text content from JIRA's ADF format
func (c *JiraClient) extractTextFromADF(description any) string {
	if description == nil {
		return ""
	}

	if str, ok := description.(string); ok {
		return str
	}

	// Handle ADF format
	if desc, ok := description.(map[string]any); ok {
		return c.extractADFContent(desc)
	}

	return ""
}

// extractADFContent recursively extracts text from ADF nodes
func (c *JiraClient) extractADFContent(node map[string]any) string {
	var text strings.Builder

	if nodeType, ok := node["type"].(string); ok && nodeType == "text" {
		if txt, ok := node["text"].(string); ok {
			text.WriteString(txt)
		}
	}

	if content, ok := node["content"].([]any); ok {
		for _, child := range content {
			if childMap, ok := child.(map[string]any); ok {
				text.WriteString(c.extractADFContent(childMap))
			}
		}
	}

	if nodeType, ok := node["type"].(string); ok && nodeType == "paragraph" {
		text.WriteString("\n\n")
	}

	return text.String()
}

// parseTime parses time strings in various formats
func (c *JiraClient) parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}

	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000+0000",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	return time.Time{}
}

// handleAPIError converts HTTP error responses to BuddyError
func (c *JiraClient) handleAPIError(resp *http.Response) error {
	body, _ := readResponseBody(resp)

	buddyErr := errors.External("JIRA", fmt.Errorf("API error: %d", resp.StatusCode)).
		WithContext("status_code", resp.StatusCode)

	if len(body) > 0 {
		var apiError struct {
			ErrorMessages []string          `json:"errorMessages"`
			Errors        map[string]string `json:"errors"`
		}

		if json.Unmarshal(body, &apiError) == nil {
			if len(apiError.ErrorMessages) > 0 {
				buddyErr.Message = apiError.ErrorMessages[0]
			}
			for k, v := range apiError.Errors {
				_ = buddyErr.WithContext(k, v)
			}
		}
	}

	return buddyErr
}

// readResponseBody safely reads HTTP response body
func readResponseBody(resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}
