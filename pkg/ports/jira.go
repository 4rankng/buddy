package ports

import "time"

// JiraPort defines the interface for generic Jira ticket management and operations
type JiraPort interface {
	// Ticket operations
	GetTicket(ticketKey string) (*JiraTicket, error)
	SearchTickets(jql string, options *SearchOptions) ([]JiraTicket, error)
	GetTicketsByFilter(filterID string, options *SearchOptions) ([]JiraTicket, error)
	CreateTicket(ticket *CreateTicketRequest) (*JiraTicket, error)
	UpdateTicket(ticketKey string, updates *UpdateTicketRequest) error
	CreateShiprmTicket(request *CreateShiprmTicketRequest) (*JiraTicket, error)

	// Comment operations
	AddComment(ticketKey string, comment string) error
	AddCommentWithADF(ticketKey string, comment map[string]interface{}) error

	// Project and issue type operations
	GetProjects() ([]JiraProject, error)
	GetIssueTypes(projectKey string) ([]JiraIssueType, error)

	// Health check
	HealthCheck() error
}

// JiraTicket represents a Jira ticket
type JiraTicket struct {
	Key         string                 `json:"key"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Priority    string                 `json:"priority"`
	Assignee    string                 `json:"assignee"`
	Reporter    string                 `json:"reporter"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Labels      []string               `json:"labels"`
	Components  []string               `json:"components"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	URL         string                 `json:"url"`
}

// JiraComment represents a Jira comment
type JiraComment struct {
	ID       string                 `json:"id"`
	Author   string                 `json:"author"`
	Body     string                 `json:"body"`
	Created  time.Time              `json:"created"`
	Updated  time.Time              `json:"updated"`
	IsPublic bool                   `json:"is_public"`
	ADF      map[string]interface{} `json:"adf,omitempty"`
}

// Generic Jira Types

// CreateTicketRequest represents a generic ticket creation request
type CreateTicketRequest struct {
	Project     string                 `json:"project"`
	IssueType   string                 `json:"issue_type"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority,omitempty"`
	Assignee    string                 `json:"assignee,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Components  []string               `json:"components,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// UpdateTicketRequest represents a generic ticket update request
type UpdateTicketRequest struct {
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Priority    string                 `json:"priority,omitempty"`
	Assignee    string                 `json:"assignee,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Components  []string               `json:"components,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// CreateShiprmTicketRequest represents a request to create a SHIPRM ticket
type CreateShiprmTicketRequest struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Assignee    string `json:"assignee,omitempty"`
}

// SearchOptions represents options for ticket searching
type SearchOptions struct {
	Fields     []string `json:"fields,omitempty"`
	MaxResults int      `json:"max_results,omitempty"`
	StartAt    int      `json:"start_at,omitempty"`
}

// JiraProject represents a Jira project
type JiraProject struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ProjectTypeKey string `json:"project_type_key"`
	Lead           string `json:"lead"`
}

// JiraIssueType represents a Jira issue type
type JiraIssueType struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

// ADFDocument represents an Atlassian Document Format document
type ADFDocument struct {
	Version int                      `json:"version"`
	Type    string                   `json:"type"`
	Content []map[string]interface{} `json:"content"`
}
