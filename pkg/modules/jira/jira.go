package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	JQL        string `json:"jql"`
	StartAt    int    `json:"startAt"`
	MaxResults int    `json:"maxResults"`
	Fields     []string `json:"fields,omitempty"`
}

// jiraSearchResp represents a Jira search response
type jiraSearchResp struct {
	StartAt    int        `json:"startAt"`
	MaxResults int        `json:"maxResults"`
	Total      int        `json:"total"`
	Issues     []jiraIssue `json:"issues"`
}

// jiraIssue represents a Jira issue
type jiraIssue struct {
	Key    string               `json:"key"`
	ID     string               `json:"id"`
	Fields jiraIssueFields      `json:"fields"`
}

// jiraIssueFields represents Jira issue fields
type jiraIssueFields struct {
	Summary     string          `json:"summary"`
	Description interface{}    `json:"description"`
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
	DisplayName string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Name        string `json:"name"`
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
			"type": "codeBlock",
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
	return j.SearchTickets(jql)
}

// SearchTickets searches for tickets using JQL
func (j *jiraModule) SearchTickets(query string) ([]ports.JiraTicket, error) {
	url := fmt.Sprintf("%s/rest/api/3/search", j.baseURL)

	searchReq := jiraSearchReq{
		JQL:        query,
		StartAt:    0,
		MaxResults: 50,
		Fields:     []string{"summary", "status", "priority", "assignee", "reporter", "created", "updated", "labels", "components", "project", "issuetype"},
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
func (j *jiraModule) CreateTicket(ticket *ports.JiraTicket) (string, string, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue", j.baseURL)

	fields := map[string]interface{}{
		"project": map[string]interface{}{"key": "PAY"},
		"summary": ticket.Summary,
		"issuetype": map[string]interface{}{"name": "Task"},
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

	body := jiraIssueReq{Fields: fields}
	b, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal ticket request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", "", fmt.Errorf("failed to create ticket request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to create ticket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("failed to create ticket: %s - %s", resp.Status, string(body))
	}

	var out jiraIssueResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", fmt.Errorf("failed to decode create response: %w", err)
	}

	ticketURL := fmt.Sprintf("%s/browse/%s", j.baseURL, out.Key)
	return out.Key, ticketURL, nil
}

// UpdateTicket updates an existing Jira ticket
func (j *jiraModule) UpdateTicket(ticketKey string, updates map[string]interface{}) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", j.baseURL, ticketKey)

	body := map[string]interface{}{
		"fields": updates,
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
	adfComment := j.adf(comment)
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

// CreateSHIPRM creates a SHIPRM ticket
func (j *jiraModule) CreateSHIPRM(request *ports.SHIPRMRequest) (string, string, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue", j.baseURL)

	// Convert services to Jira format
	services := make([]map[string]interface{}, 0, len(request.Services))
	for _, service := range request.Services {
		services = append(services, map[string]interface{}{
			"workspaceId": "246c1bd0-99bf-42c1-b124-a30c70c816d1",
			"id":          fmt.Sprintf("246c1bd0-99bf-42c1-b124-a30c70c816d1:%s", request.ChangeType),
			"objectId":    request.ChangeType,
		})
	}

	fields := map[string]interface{}{
		"project":           map[string]interface{}{"key": "SHIPRM"},
		"summary":           request.Title,
		"description":       j.adf(request.Description),
		"customfield_11290": services, // Services
		"customfield_10925": j.adf("NA"), // Impact Analysis
		"customfield_11181": services, // Affected Services
		"customfield_11183": request.ReviewDate.Format("2006-01-02"), // Review Date
		"customfield_11187": j.adfCode(request.CURL), // Implementation Steps
		"customfield_10042": j.adf("NA"), // Test Plan
		"customfield_11188": j.adf(request.Description), // Change Description
		"customfield_11189": j.adf(func() string { if request.Impact.RequiresMaintenance { return "Yes" } else { return "No" } }()),
		"customfield_11190": j.adf(func() string { if request.Impact.AffectsCustomers { return "Yes" } else { return "No" } }()),
		"customfield_11191": j.adf(func() string { if request.Impact.DowntimeRequired { return "Yes" } else { return "No" } }()),
		"customfield_11186": j.adf("Standard production change"), // Change Type
		"customfield_11192": j.adf(request.Validation), // Validation Steps
		"issuetype":         map[string]interface{}{"id": "10005"}, // System Change
	}

	body := jiraIssueReq{Fields: fields}
	b, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal SHIPRM request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", "", fmt.Errorf("failed to create SHIPRM request: %w", err)
	}

	resp, err := j.executeWithRetry(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to create SHIPRM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("failed to create SHIPRM: %s - %s", resp.Status, string(body))
	}

	var out jiraIssueResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", fmt.Errorf("failed to decode SHIPRM response: %w", err)
	}

	ticketURL := fmt.Sprintf("%s/browse/%s", j.baseURL, out.Key)
	return out.Key, ticketURL, nil
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
		Key:      issue.Key,
		Summary:  issue.Fields.Summary,
		Status:   issue.Fields.Status.Name,
		Created:  issue.Fields.Created,
		Updated:  issue.Fields.Updated,
		URL:      fmt.Sprintf("%s/browse/%s", j.baseURL, issue.Key),
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