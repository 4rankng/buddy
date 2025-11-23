package ports

import "time"

// JiraPort defines the interface for Jira ticket management and operations
type JiraPort interface {
	// Ticket operations
	GetTicket(ticketKey string) (*JiraTicket, error)
	GetAssignedTickets(team string) ([]JiraTicket, error)
	SearchTickets(query string) ([]JiraTicket, error)
	CreateTicket(ticket *JiraTicket) (string, string, error) // Returns ticket ID and URL
	UpdateTicket(ticketKey string, updates map[string]interface{}) error

	// Comment operations
	AddComment(ticketKey string, comment string) error
	AddCommentWithADF(ticketKey string, comment map[string]interface{}) error

	// SHIPRM operations
	CreateSHIPRM(request *SHIPRMRequest) (string, string, error) // Returns ticket ID and URL

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
	ID        string                 `json:"id"`
	Author    string                 `json:"author"`
	Body      string                 `json:"body"`
	Created   time.Time              `json:"created"`
	Updated   time.Time              `json:"updated"`
	IsPublic  bool                   `json:"is_public"`
	ADF       map[string]interface{} `json:"adf,omitempty"`
}

// SHIPRMRequest represents a SHIPRM (System High Impact Production Risk Management) request
type SHIPRMRequest struct {
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	ChangeType   string                 `json:"change_type"`
	CURL         string                 `json:"curl"`
	Services     []SHIPRMService        `json:"services"`
	Impact       SHIPRMImpact           `json:"impact"`
	Dependencies []string               `json:"dependencies"`
	BackoutPlan  string                 `json:"backout_plan"`
	Validation   string                 `json:"validation"`
	ReviewDate   time.Time              `json:"review_date"`
	Additional   map[string]interface{} `json:"additional,omitempty"`
}

// SHIPRMService represents a service affected by a SHIPRM change
type SHIPRMService struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Component   string `json:"component"`
	Team        string `json:"team"`
}

// SHIPRMImpact represents the impact assessment for a SHIPRM change
type SHIPRMImpact struct {
	Description        string `json:"description"`
	AffectsCustomers   bool   `json:"affects_customers"`
	RequiresMaintenance bool  `json:"requires_maintenance"`
	DowntimeRequired   bool   `json:"downtime_required"`
	RiskLevel         string `json:"risk_level"`
}

// ADFDocument represents an Atlassian Document Format document
type ADFDocument struct {
	Version int                    `json:"version"`
	Type    string                 `json:"type"`
	Content []map[string]interface{} `json:"content"`
}