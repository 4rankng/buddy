package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"oncall/pkg/config"
	"oncall/pkg/ports"
)

// jiraModule implements the JiraPort interface
type jiraModule struct {
	config  config.JiraConfig
	client  *http.Client
	baseURL string
}

// NewJiraModule creates a new Jira module
func NewJiraModule(cfg config.JiraConfig) (ports.JiraPort, error) {
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	return &jiraModule{
		config:  cfg,
		client:  client,
		baseURL: cfg.BaseURL,
	}, nil
}

// jiraIssueReq represents a Jira issue creation request
type jiraIssueReq struct {
	Fields map[string]any `json:"fields"`
}

// jiraIssueResp represents a Jira issue creation response
type jiraIssueResp struct {
	Key string `json:"key"`
	ID  string `json:"id"`
}

// jiraSearchReq represents a Jira search request
type jiraSearchReq struct {
	JQL        string   `json:"jql"`
	StartAt    int      `json:"startAt"`
	MaxResults int      `json:"maxResults"`
	Fields     []string `json:"fields,omitempty"`
}

// jiraSearchResp represents a Jira search response
type jiraSearchResp struct {
	StartAt    int         `json:"startAt"`
	MaxResults int         `json:"maxResults"`
	Total      int         `json:"total"`
	Issues     []jiraIssue `json:"issues"`
}

// jiraIssue represents a Jira issue
type jiraIssue struct {
	Key    string          `json:"key"`
	ID     string          `json:"id"`
	Fields jiraIssueFields `json:"fields"`
}

// jiraIssueFields represents Jira issue fields
type jiraIssueFields struct {
	Summary     string          `json:"summary"`
	Description interface{}     `json:"description"`
	Status      jiraStatus      `json:"status"`
	Priority    jiraPriority    `json:"priority"`
	Assignee    *jiraUser       `json:"assignee"`
	Reporter    jiraUser        `json:"reporter"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
	Labels      []string        `json:"labels"`
	Components  []jiraComponent `json:"components"`
	Project     jiraProject     `json:"project"`
	IssueType   jiraIssueType   `json:"issuetype"`
}

// jiraStatus represents a Jira status
type jiraStatus struct {
	Name string `json:"name"`
}

// jiraPriority represents a Jira priority
type jiraPriority struct {
	Name string `json:"name"`
}

// jiraUser represents a Jira user
type jiraUser struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Name         string `json:"name"`
}

// jiraComponent represents a Jira component
type jiraComponent struct {
	Name string `json:"name"`
}

// jiraProject represents a Jira project
type jiraProject struct {
	Key string `json:"key"`
}

// jiraIssueType represents a Jira issue type
type jiraIssueType struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// executeWithRetry executes an HTTP request with retry logic
func (j *jiraModule) executeWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= j.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(j.config.RetryDelay)
		}

		req.SetBasicAuth(j.config.Email, j.config.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := j.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// adf creates an Atlassian Document Format document
func (j *jiraModule) adf(text string) ports.ADFDocument {
	return ports.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []map[string]interface{}{{
			"type": "paragraph",
			"content": []map[string]interface{}{{
				"type": "text",
				"text": text,
			}},
		}},
	}
}

// adfCode creates an ADF code block
func (j *jiraModule) adfCode(text string) ports.ADFDocument {
	return ports.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []map[string]interface{}{{
			"type":  "codeBlock",
			"attrs": map[string]interface{}{"language": "bash"},
			"content": []map[string]interface{}{{
				"type": "text",
				"text": text,
			}},
		}},
	}
}

// GetTicket retrieves a specific Jira ticket
func (j *jiraModule) GetTicket(ticketKey string) (*ports.JiraTicket, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", j.baseURL, ticketKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get ticket %s: %s - %s", ticketKey, resp.Status, string(body))
	}

	var issue jiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return j.convertToJiraTicket(issue), nil
}

// GetAssignedTickets retrieves tickets assigned to the specified team/user
func (j *jiraModule) GetAssignedTickets(team string) ([]ports.JiraTicket, error) {
	jql := fmt.Sprintf(`project = "PAY" AND assignee = "%s" AND status not in (Done, Closed, Resolved) ORDER BY created DESC`, team)
	options := &ports.SearchOptions{
		MaxResults: 50,
		Fields:     []string{"summary", "status", "priority", "assignee", "reporter", "created", "updated"},
	}
	return j.SearchTickets(jql, options)
}

// GetTicketsByFilter retrieves tickets using a filter ID
func (j *jiraModule) GetTicketsByFilter(filterID string, options *ports.SearchOptions) ([]ports.JiraTicket, error) {
	jql := fmt.Sprintf("filter=%s", filterID)
	return j.SearchTickets(jql, options)
}

// SearchTickets searches for tickets using JQL
func (j *jiraModule) SearchTickets(query string, options *ports.SearchOptions) ([]ports.JiraTicket, error) {
	url := fmt.Sprintf("%s/rest/api/3/search", j.baseURL)

	maxResults := 50
	if options != nil && options.MaxResults > 0 {
		maxResults = options.MaxResults
	}

	fields := []string{"summary", "status", "priority", "assignee", "reporter", "created", "updated", "labels", "components", "project", "issuetype"}
	if options != nil && len(options.Fields) > 0 {
		fields = options.Fields
	}

	searchReq := jiraSearchReq{
		JQL:        query,
		StartAt:    0,
		MaxResults: maxResults,
		Fields:     fields,
	}

	b, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s - %s", resp.Status, string(body))
	}

	var searchResp jiraSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	tickets := make([]ports.JiraTicket, 0, len(searchResp.Issues))
	for _, issue := range searchResp.Issues {
		ticket := j.convertToJiraTicket(issue)
		tickets = append(tickets, *ticket)
	}

	return tickets, nil
}

// CreateTicket creates a new Jira ticket
func (j *jiraModule) CreateTicket(ticket *ports.CreateTicketRequest) (*ports.JiraTicket, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue", j.baseURL)

	fields := map[string]interface{}{
		"project":   map[string]interface{}{"key": ticket.Project},
		"summary":   ticket.Summary,
		"issuetype": map[string]interface{}{"name": ticket.IssueType},
	}

	// Add description if provided
	if ticket.Description != "" {
		fields["description"] = j.adf(ticket.Description)
	}

	// Add priority if provided
	if ticket.Priority != "" {
		fields["priority"] = map[string]interface{}{"name": ticket.Priority}
	}

	// Add assignee if provided
	if ticket.Assignee != "" {
		fields["assignee"] = map[string]interface{}{"name": ticket.Assignee}
	}

	// Add labels if provided
	if len(ticket.Labels) > 0 {
		fields["labels"] = ticket.Labels
	}

	// Add components if provided
	if len(ticket.Components) > 0 {
		components := make([]map[string]interface{}, 0, len(ticket.Components))
		for _, component := range ticket.Components {
			components = append(components, map[string]interface{}{"name": component})
		}
		fields["components"] = components
	}

	// Add additional fields if provided
	if len(ticket.Fields) > 0 {
		for k, v := range ticket.Fields {
			fields[k] = v
		}
	}

	body := jiraIssueReq{Fields: fields}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ticket request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create ticket: %s - %s", resp.Status, string(body))
	}

	var out jiraIssueResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode create response: %w", err)
	}

	// Get the created ticket to return full details
	return j.GetTicket(out.Key)
}

// CreateShiprmTicket creates a new SHIPRM ticket
func (j *jiraModule) CreateShiprmTicket(request *ports.CreateShiprmTicketRequest) (*ports.JiraTicket, error) {
	req := &ports.CreateTicketRequest{
		Project:     "SHIPRM",
		IssueType:   "Task", // Assuming 'Task' is the issue type, can be adjusted
		Summary:     request.Summary,
		Description: request.Description,
		Assignee:    request.Assignee,
	}
	return j.CreateTicket(req)
}

// UpdateTicket updates an existing Jira ticket
func (j *jiraModule) UpdateTicket(ticketKey string, updates *ports.UpdateTicketRequest) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", j.baseURL, ticketKey)

	fields := map[string]interface{}{}

	if updates.Summary != "" {
		fields["summary"] = updates.Summary
	}
	if updates.Description != "" {
		fields["description"] = j.adf(updates.Description)
	}
	if updates.Status != "" {
		fields["status"] = map[string]interface{}{"name": updates.Status}
	}
	if updates.Priority != "" {
		fields["priority"] = map[string]interface{}{"name": updates.Priority}
	}
	if updates.Assignee != "" {
		fields["assignee"] = map[string]interface{}{"name": updates.Assignee}
	}
	if len(updates.Labels) > 0 {
		fields["labels"] = updates.Labels
	}
	if len(updates.Components) > 0 {
		components := make([]map[string]interface{}, 0, len(updates.Components))
		for _, component := range updates.Components {
			components = append(components, map[string]interface{}{"name": component})
		}
		fields["components"] = components
	}

	if len(updates.Fields) > 0 {
		for k, v := range updates.Fields {
			fields[k] = v
		}
	}

	body := map[string]interface{}{
		"fields": fields,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create update request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update ticket %s: %s - %s", ticketKey, resp.Status, string(body))
	}

	return nil
}

// AddComment adds a comment to a ticket
func (j *jiraModule) AddComment(ticketKey string, comment string) error {
	adfComment := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []map[string]interface{}{{
			"type": "paragraph",
			"content": []map[string]interface{}{{
				"type": "text",
				"text": comment,
			}},
		}},
	}
	return j.AddCommentWithADF(ticketKey, adfComment)
}

// AddCommentWithADF adds a comment to a ticket with ADF formatting
func (j *jiraModule) AddCommentWithADF(ticketKey string, comment map[string]interface{}) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", j.baseURL, ticketKey)

	body := map[string]interface{}{
		"body": comment,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal comment request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create comment request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add comment to %s: %s - %s", ticketKey, resp.Status, string(body))
	}

	return nil
}

// GetProjects returns a list of available Jira projects
func (j *jiraModule) GetProjects() ([]ports.JiraProject, error) {
	url := fmt.Sprintf("%s/rest/api/3/project", j.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create projects request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get projects: %s - %s", resp.Status, string(body))
	}

	var projects []ports.JiraProject
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode projects response: %w", err)
	}

	return projects, nil
}

// GetIssueTypes returns issue types for a specific project
func (j *jiraModule) GetIssueTypes(projectKey string) ([]ports.JiraIssueType, error) {
	url := fmt.Sprintf("%s/rest/api/3/project/%s/issuetypes", j.baseURL, projectKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue types request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue types: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get issue types: %s - %s", resp.Status, string(body))
	}

	var issueTypes []ports.JiraIssueType
	if err := json.NewDecoder(resp.Body).Decode(&issueTypes); err != nil {
		return nil, fmt.Errorf("failed to decode issue types response: %w", err)
	}

	return issueTypes, nil
}

// HealthCheck performs a health check on the Jira service
func (j *jiraModule) HealthCheck() error {
	url := fmt.Sprintf("%s/rest/api/3/myself", j.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return fmt.Errorf("Jira health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Jira health check failed with status: %s", resp.Status)
	}

	return nil
}

// convertToJiraTicket converts a jiraIssue to a ports.JiraTicket
func (j *jiraModule) convertToJiraTicket(issue jiraIssue) *ports.JiraTicket {
	ticket := &ports.JiraTicket{
		Key:     issue.Key,
		Summary: issue.Fields.Summary,
		Status:  issue.Fields.Status.Name,
		Created: issue.Fields.Created,
		Updated: issue.Fields.Updated,
		URL:     fmt.Sprintf("%s/browse/%s", j.baseURL, issue.Key),
	}

	// Handle description (could be string or ADF)
	if desc, ok := issue.Fields.Description.(string); ok {
		ticket.Description = desc
	} else if descMap, ok := issue.Fields.Description.(map[string]interface{}); ok {
		// Extract text from ADF if needed
		if content, ok := descMap["content"].([]interface{}); ok && len(content) > 0 {
			if paragraph, ok := content[0].(map[string]interface{}); ok {
				if paraContent, ok := paragraph["content"].([]interface{}); ok && len(paraContent) > 0 {
					if textElement, ok := paraContent[0].(map[string]interface{}); ok {
						if text, ok := textElement["text"].(string); ok {
							ticket.Description = text
						}
					}
				}
			}
		}
	}

	if issue.Fields.Priority.Name != "" {
		ticket.Priority = issue.Fields.Priority.Name
	}

	if issue.Fields.Assignee != nil {
		ticket.Assignee = issue.Fields.Assignee.DisplayName
	}

	ticket.Reporter = issue.Fields.Reporter.DisplayName

	ticket.Labels = issue.Fields.Labels

	for _, component := range issue.Fields.Components {
		ticket.Components = append(ticket.Components, component.Name)
	}

	return ticket
}
