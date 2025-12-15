package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// JiraClient implements the JiraInterface
type JiraClient struct {
	config     JiraConfig
	httpClient *http.Client
}

// NewJiraClient creates a new JIRA client instance
func NewJiraClient(env string) *JiraClient {
	fmt.Println("Initialize Jira client for ", env)
	cfg := GetJiraConfig(env)
	return &JiraClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// GetAssignedIssues fetches issues assigned to specified emails or currentUser()
func (c *JiraClient) GetAssignedIssues(ctx context.Context, projectKey string, emails []string) ([]JiraTicket, error) {
	// Build JQL query
	var jql string
	if len(emails) == 1 && emails[0] == "currentUser()" {
		// Use currentUser() function when explicitly requested
		jql = fmt.Sprintf(
			"project = %s AND assignee = currentUser() AND status NOT IN (Completed, Closed) ORDER BY created DESC",
			projectKey,
		)
	} else {
		// Build email list for specific users
		emailList := make([]string, len(emails))
		for i, email := range emails {
			emailList[i] = fmt.Sprintf(`"%s"`, email)
		}

		jql = fmt.Sprintf(
			"assignee IN (%s) AND project = %s AND status NOT IN (Done, Resolved, Closed, Completed) ORDER BY created ASC",
			strings.Join(emailList, ", "),
			projectKey,
		)
	}

	// Build request URL with correct endpoint
	apiURL, err := url.Parse(c.config.Domain + "/rest/api/3/search/jql")
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("invalid JIRA domain: %v", err),
		}
	}

	requestPayload := map[string]interface{}{
		"jql":        jql,
		"fields":     []string{"assignee", "summary", "issuetype", "key", "priority", "status", "created", "duedate", "customfield_10060", "description", "attachment"},
		"maxResults": c.config.MaxItems,
	}

	reqBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to marshal search payload: %v", err),
		}
	}

	// Create and execute request with context
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Content-Type", "application/json")

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	// Parse response
	var searchResponse struct {
		Issues []jiraIssueResponse `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to decode response: %v", err),
		}
	}

	// Convert to JiraTicket objects
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

// GetIssueDetails fetches full details for a specific issue
func (c *JiraClient) GetIssueDetails(ctx context.Context, issueKey string) (*JiraTicket, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s", c.config.Domain, issueKey)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to create request: %v", err),
		}
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var issue jiraIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to decode response: %v", err),
		}
	}

	return c.convertIssueResponse(&issue)
}

// SearchIssues searches for issues matching the search term in summary or description
func (c *JiraClient) SearchIssues(ctx context.Context, searchTerm string) ([]JiraTicket, error) {
	// Escape search term to prevent JQL injection
	escapedTerm := strings.ReplaceAll(searchTerm, "\"", "\\\"")
	escapedTerm = strings.ReplaceAll(escapedTerm, "\\", "\\\\")
	escapedTerm = strings.ReplaceAll(escapedTerm, "\n", "\\n")
	escapedTerm = strings.ReplaceAll(escapedTerm, "\r", "\\r")

	// Build JQL query for searching in summary and description
	jql := fmt.Sprintf(
		`project = %s AND assignee = currentUser() AND status NOT IN (Completed, Closed) AND (summary ~ "%s" OR description ~ "%s") ORDER BY created DESC`,
		c.config.Project,
		escapedTerm,
		escapedTerm,
	)

	// Build request URL with correct endpoint
	apiURL, err := url.Parse(c.config.Domain + "/rest/api/3/search/jql")
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("invalid JIRA domain: %v", err),
		}
	}

	requestPayload := map[string]interface{}{
		"jql":        jql,
		"fields":     []string{"assignee", "summary", "issuetype", "key", "priority", "status", "created", "duedate", "customfield_10060", "description", "attachment"},
		"maxResults": c.config.MaxItems,
	}

	reqBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to marshal search payload: %v", err),
		}
	}

	// Create and execute request with context
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Content-Type", "application/json")

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	// Parse response
	var searchResponse struct {
		Issues []jiraIssueResponse `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to decode response: %v", err),
		}
	}

	// Convert to JiraTicket objects
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

// GetAttachmentContent downloads attachment content
func (c *JiraClient) GetAttachmentContent(ctx context.Context, attachmentURL string) ([]byte, error) {
	// Handle redirects with context
	client := &http.Client{
		Timeout: c.httpClient.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Copy authorization header for redirects
			c.setAuthHeaders(req)
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", attachmentURL, nil)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to create request: %v", err),
		}
	}

	c.setAuthHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to download attachment: %v", err),
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to read attachment content: %v", err),
		}
	}

	return content, nil
}

// DownloadAttachment downloads an attachment and saves it to the specified path
func (c *JiraClient) DownloadAttachment(ctx context.Context, attachment Attachment, savePath string) error {
	// Get attachment content
	content, err := c.GetAttachmentContent(ctx, attachment.URL)
	if err != nil {
		return fmt.Errorf("failed to download attachment: %w", err)
	}

	// Handle file naming if savePath is a directory
	if strings.HasSuffix(savePath, "/") || savePath == "." || savePath == "" {
		filename := attachment.Filename
		fullPath := filepath.Join(savePath, filename)

		// If file already exists, add suffix
		if _, err := os.Stat(fullPath); err == nil {
			ext := filepath.Ext(filename)
			base := strings.TrimSuffix(filename, ext)
			i := 1
			for {
				newFilename := fmt.Sprintf("%s_%d%s", base, i, ext)
				newPath := filepath.Join(savePath, newFilename)
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					fullPath = newPath
					break
				}
				i++
			}
		}
		savePath = fullPath
	}

	// Write file
	if err := os.WriteFile(savePath, content, 0644); err != nil {
		return fmt.Errorf("failed to save attachment: %w", err)
	}

	return nil
}

// ParseCSVAttachment parses CSV content from JIRA attachments
func (c *JiraClient) ParseCSVAttachment(content string) ([]CSVRow, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to parse CSV: %v", err),
		}
	}

	if len(records) == 0 {
		return nil, nil
	}

	// Find header row
	headerRowIndex := -1
	fieldMappings := c.getFieldMappings()

	for i, row := range records {
		if c.findHeaderRow(row, fieldMappings) {
			headerRowIndex = i
			break
		}
	}

	if headerRowIndex < 0 {
		return nil, &JiraError{
			StatusCode: 0,
			Message:    "header row not found in CSV",
		}
	}

	// Map columns
	headerRow := records[headerRowIndex]
	c.mapColumnIndices(headerRow, fieldMappings)

	// Process data rows
	var rows []CSVRow
	for _, row := range records[headerRowIndex+1:] {
		if c.isEmptyRow(row) || c.isSummaryRow(row) {
			continue
		}

		csvRow := c.processCSVRow(row, fieldMappings)
		if csvRow != nil {
			rows = append(rows, *csvRow)
		}
	}

	return rows, nil
}

// CloseTicket closes a JIRA ticket by transitioning it
func (c *JiraClient) CloseTicket(ctx context.Context, issueKey string, reasonType string) error {
	// First get available transitions
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", c.config.Domain, issueKey)

	// Get transitions
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to create request: %v", err),
		}
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("request failed: %v", err),
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return c.handleAPIError(resp)
	}

	var transitions struct {
		Transitions []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"transitions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&transitions); err != nil {
		return &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to decode transitions: %v", err),
		}
	}

	// Find the "Done" or "Close" transition
	var targetTransitionID string
	for _, t := range transitions.Transitions {
		if strings.Contains(strings.ToLower(t.Name), "done") ||
			strings.Contains(strings.ToLower(t.Name), "close") {
			targetTransitionID = t.ID
			break
		}
	}

	if targetTransitionID == "" {
		return &JiraError{
			StatusCode: 0,
			Message:    "no close transition found",
		}
	}

	// Execute transition
	transitionReq := struct {
		Transition struct {
			ID string `json:"id"`
		} `json:"transition"`
	}{}
	transitionReq.Transition.ID = targetTransitionID

	reqBody, err := json.Marshal(transitionReq)
	if err != nil {
		return &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to marshal transition request: %v", err),
		}
	}

	req, err = http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to create transition request: %v", err),
		}
	}

	c.setAuthHeaders(req)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return &JiraError{
			StatusCode: 0,
			Message:    fmt.Sprintf("transition failed: %v", err),
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleAPIError(resp)
	}

	return nil
}

// Helper methods

type jiraIssueResponse struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary     interface{} `json:"summary"`
		Description interface{} `json:"description"`
		Assignee    *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority *struct {
			Name string `json:"name"`
		} `json:"priority"`
		Created          string      `json:"created"`
		DueDate          string      `json:"duedate"`
		CustomField10060 interface{} `json:"customfield_10060"`
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

func (c *JiraClient) setAuthHeaders(req *http.Request) {
	// Validate credentials before setting headers
	if c.config.Auth.Username == "" || c.config.Auth.APIKey == "" {
		// This should never happen if config is validated during initialization
		return
	}

	// Create Basic Auth header
	auth := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", c.config.Auth.Username, c.config.Auth.APIKey)),
	)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))

	// Set standard headers
	req.Header.Set("Accept", "application/json")

	// Only set Content-Type for requests with a body
	if req.Method != "GET" && req.Method != "HEAD" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add user agent for better debugging
	req.Header.Set("User-Agent", "buddy-jira-client/1.0")
}

func (c *JiraClient) convertIssueResponse(issue *jiraIssueResponse) (*JiraTicket, error) {
	ticket := &JiraTicket{
		ID:        issue.ID,
		Key:       issue.Key,
		IssueType: issue.Fields.IssueType.Name,
		CreatedAt: c.parseTime(issue.Fields.Created),
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

	// Status
	ticket.Status = issue.Fields.Status.Name

	// Priority
	if issue.Fields.Priority != nil {
		ticket.Priority = issue.Fields.Priority.Name
	}

	// Due date
	if issue.Fields.DueDate != "" {
		if dueDate := c.parseTime(issue.Fields.DueDate); !dueDate.IsZero() {
			ticket.DueAt = &dueDate
		}
	} else if issue.Fields.CustomField10060 != nil {
		if customFieldStr, ok := issue.Fields.CustomField10060.(string); ok && customFieldStr != "" {
			if dueDate := c.parseTime(customFieldStr); !dueDate.IsZero() {
				ticket.DueAt = &dueDate
			}
		}
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

func (c *JiraClient) extractTextFromADF(description interface{}) string {
	if description == nil {
		return ""
	}

	if str, ok := description.(string); ok {
		return str
	}

	// Handle ADF format
	if desc, ok := description.(map[string]interface{}); ok {
		return c.extractADFContent(desc)
	}

	return ""
}

func (c *JiraClient) extractADFContent(node map[string]interface{}) string {
	var text strings.Builder

	if nodeType, ok := node["type"].(string); ok && nodeType == "text" {
		if txt, ok := node["text"].(string); ok {
			text.WriteString(txt)
		}
	}

	if content, ok := node["content"].([]interface{}); ok {
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				text.WriteString(c.extractADFContent(childMap))
			}
		}
	}

	if nodeType, ok := node["type"].(string); ok && nodeType == "paragraph" {
		text.WriteString("\n\n")
	}

	return text.String()
}

func (c *JiraClient) parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}

	// Try different time formats
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

func (c *JiraClient) handleAPIError(resp *http.Response) *JiraError {
	body, _ := io.ReadAll(resp.Body)

	errResp := &JiraError{
		StatusCode: resp.StatusCode,
		Message:    fmt.Sprintf("JIRA API error: %d", resp.StatusCode),
		Details:    make(map[string]interface{}),
	}

	if len(body) > 0 {
		var apiError struct {
			ErrorMessages []string          `json:"errorMessages"`
			Errors        map[string]string `json:"errors"`
		}

		if json.Unmarshal(body, &apiError) == nil {
			if len(apiError.ErrorMessages) > 0 {
				errResp.Message = apiError.ErrorMessages[0]
			}
			for k, v := range apiError.Errors {
				errResp.Details[k] = v
			}
		}
	}

	return errResp
}

type csvFieldMapping struct {
	Index  int
	Fields []string
}

func (c *JiraClient) getFieldMappings() map[string]*csvFieldMapping {
	return map[string]*csvFieldMapping{
		"transaction_date": {
			Fields: []string{"date"},
		},
		"batch_id": {
			Fields: []string{"batch id", "partner_tx_id"},
		},
		"end_to_end_id": {
			Fields: []string{"tar02 bmid", "original_bizmsgid"},
		},
		"transaction_id": {
			Fields: []string{"transaction id"},
		},
		"req_biz_msg_id": {
			Fields: []string{"req_biz_msg_id"},
		},
		"internal_status": {
			Fields: []string{"dbmy status"},
		},
		"paynet_status": {
			Fields: []string{"column_status", "tar02 sts", "rpp_status"},
		},
	}
}

func (c *JiraClient) findHeaderRow(row []string, mappings map[string]*csvFieldMapping) bool {
	if len(row) == 0 {
		return false
	}

	lowerRow := make([]string, len(row))
	for i, cell := range row {
		lowerRow[i] = strings.ToLower(strings.TrimSpace(cell))
	}

	matches := 0
	for _, mapping := range mappings {
		for _, field := range mapping.Fields {
			for _, cell := range lowerRow {
				if cell == field {
					matches++
					break
				}
			}
		}
	}

	// Consider it a header if we match at least 3 fields
	return matches >= 3
}

func (c *JiraClient) mapColumnIndices(headerRow []string, mappings map[string]*csvFieldMapping) {
	for i, header := range headerRow {
		lowerHeader := strings.ToLower(strings.TrimSpace(header))

		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				if lowerHeader == field {
					mapping.Index = i
					break
				}
			}
		}
	}
}

func (c *JiraClient) isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func (c *JiraClient) isSummaryRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(row[0])), "total")
}

func (c *JiraClient) processCSVRow(row []string, mappings map[string]*csvFieldMapping) *CSVRow {
	csvRow := &CSVRow{}
	hasData := false

	for fieldName, mapping := range mappings {
		if mapping.Index >= 0 && mapping.Index < len(row) {
			value := strings.TrimSpace(row[mapping.Index])
			if value != "" && value != "-" {
				ptr := &value
				switch fieldName {
				case "transaction_date":
					csvRow.TransactionDate = ptr
				case "batch_id":
					csvRow.BatchID = ptr
				case "end_to_end_id":
					csvRow.EndToEndID = ptr
				case "transaction_id":
					csvRow.TransactionID = ptr
				case "req_biz_msg_id":
					csvRow.ReqBizMsgID = ptr
				case "internal_status":
					csvRow.InternalStatus = ptr
				case "paynet_status":
					csvRow.PaynetStatus = ptr
				}
				hasData = true
			}
		}
	}

	if !hasData {
		return nil
	}

	return csvRow
}
