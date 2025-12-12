package clients

import (
	"time"
)

// JiraInterface defines the contract for JIRA operations
// This interface provides a clean abstraction for JIRA API interactions
type JiraInterface interface {
	// Core ticket operations
	GetAssignedIssues(projectKey string, emails []string) ([]JiraTicket, error)
	GetIssueDetails(issueKey string) (*JiraTicket, error)

	// Attachment operations
	GetAttachmentContent(attachmentURL string) ([]byte, error)
	ParseCSVAttachment(content string) ([]CSVRow, error)

	// Ticket lifecycle operations
	CloseTicket(issueKey string, reasonType string) error
}

// JiraTicket represents a JIRA issue/ticket
type JiraTicket struct {
	ID          string       `json:"id"`
	Key         string       `json:"key"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Assignee    string       `json:"assignee"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
	CreatedAt   time.Time    `json:"created_at"`
	DueAt       *time.Time   `json:"due_at,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	IssueType   string       `json:"issue_type"`
}

// Attachment represents a JIRA attachment
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
	URL      string `json:"url"`
	Content  string `json:"content"`
}

// CSVRow represents a parsed CSV row from JIRA attachments
type CSVRow struct {
	TransactionDate *string `json:"transaction_date,omitempty"`
	BatchID         *string `json:"batch_id,omitempty"`
	EndToEndID      *string `json:"end_to_end_id,omitempty"`
	TransactionID   *string `json:"transaction_id,omitempty"`
	ReqBizMsgID     *string `json:"req_biz_msg_id,omitempty"`
	InternalStatus  *string `json:"internal_status,omitempty"`
	PaynetStatus    *string `json:"paynet_status,omitempty"`
}

// JiraError represents a JIRA API error
type JiraError struct {
	StatusCode int
	Message    string
	Details    map[string]interface{}
}

func (e *JiraError) Error() string {
	return e.Message
}

// JiraAuthInfo holds JIRA authentication information
type JiraAuthInfo struct {
	Domain   string
	Username string
	APIKey   string
}

// Ensure JiraClient implements JiraInterface
var _ JiraInterface = (*JiraClient)(nil)
