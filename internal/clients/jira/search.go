package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"buddy/internal/errors"
)

// GetAssignedIssues fetches issues assigned to specified emails or currentUser()
func (c *JiraClient) GetAssignedIssues(ctx context.Context, projectKey string, emails []string) ([]JiraTicket, error) {
	jql := c.buildAssignedIssuesJQL(projectKey, emails)
	return c.executeSearch(ctx, jql)
}

// SearchIssues searches for issues matching the search term in summary or description
func (c *JiraClient) SearchIssues(ctx context.Context, searchTerm string) ([]JiraTicket, error) {
	jql := c.buildSearchJQL(searchTerm)
	return c.executeSearch(ctx, jql)
}

// ExecuteJQL executes a raw JQL query and returns matching tickets
func (c *JiraClient) ExecuteJQL(ctx context.Context, jql string) ([]JiraTicket, error) {
	return c.executeSearch(ctx, jql)
}

// GetIssueDetails fetches full details for a specific issue
func (c *JiraClient) GetIssueDetails(ctx context.Context, issueKey string) (*JiraTicket, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s", c.config.Domain, issueKey)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to create JIRA request")
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "JIRA request failed")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var issue jiraIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to decode JIRA response")
	}

	return c.convertIssueResponse(&issue)
}

// buildAssignedIssuesJQL builds JQL query for assigned issues
func (c *JiraClient) buildAssignedIssuesJQL(projectKey string, emails []string) string {
	if len(emails) == 1 && emails[0] == "currentUser()" {
		return fmt.Sprintf(
			"project = %s AND assignee = currentUser() AND status NOT IN (Completed, Closed) ORDER BY created DESC",
			projectKey,
		)
	}

	emailList := make([]string, len(emails))
	for i, email := range emails {
		emailList[i] = fmt.Sprintf(`"%s"`, email)
	}

	return fmt.Sprintf(
		"assignee IN (%s) AND project = %s AND status NOT IN (Done, Resolved, Closed, Completed) ORDER BY created ASC",
		strings.Join(emailList, ", "),
		projectKey,
	)
}

// buildSearchJQL builds JQL query for text search
func (c *JiraClient) buildSearchJQL(searchTerm string) string {
	// Trim whitespace from search term to avoid exact matching issues
	trimmedTerm := strings.TrimSpace(searchTerm)
	// Escape search term to prevent JQL injection
	escapedTerm := c.escapeJQLTerm(trimmedTerm)

	return fmt.Sprintf(
		`project = %s AND (summary ~ "%s" OR description ~ "%s") ORDER BY created DESC`,
		c.config.Project,
		escapedTerm,
		escapedTerm,
	)
}

// escapeJQLTerm escapes special characters in JQL search terms
func (c *JiraClient) escapeJQLTerm(term string) string {
	term = strings.ReplaceAll(term, "\\", "\\\\")
	term = strings.ReplaceAll(term, "\"", "\\\"")
	term = strings.ReplaceAll(term, "\n", "\\n")
	term = strings.ReplaceAll(term, "\r", "\\r")
	return term
}

// executeSearch executes a JQL search and returns tickets
func (c *JiraClient) executeSearch(ctx context.Context, jql string) ([]JiraTicket, error) {
	apiURL, err := url.Parse(c.config.Domain + "/rest/api/3/search/jql")
	if err != nil {
		return nil, errors.Validation("invalid JIRA domain")
	}

	requestPayload := map[string]any{
		"jql":        jql,
		"fields":     []string{"assignee", "summary", "issuetype", "key", "priority", "status", "created", "duedate", "customfield_10060", "description", "attachment"},
		"maxResults": c.config.MaxItems,
	}

	reqBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to marshal search payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to create JIRA request")
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "JIRA search request failed")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var searchResponse struct {
		Issues []jiraIssueResponse `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to decode JIRA search response")
	}

	tickets := make([]JiraTicket, len(searchResponse.Issues))
	for i, issue := range searchResponse.Issues {
		ticket, err := c.convertIssueResponse(&issue)
		if err != nil {
			return nil, err
		}
		tickets[i] = *ticket
	}

	return tickets, nil
}
