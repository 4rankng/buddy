package ports

import "time"

// DatadogPort defines the interface for generic Datadog logs and metrics operations
type DatadogPort interface {
	// Log operations
	SearchLogs(params *LogSearchParams) (*LogSearchResponse, error)
	SubmitLogs(logs []LogEvent) (*LogSubmissionResponse, error)

	// Generic operations
	GetAvailableIndexes() ([]string, error)
	TestConnection() error

	// Health check
	HealthCheck() error
}

// LogSearchParams contains parameters for log search
type LogSearchParams struct {
	Query    string            `json:"query"`
	From     string            `json:"from"`
	To       string            `json:"to"`
	Timezone string            `json:"timezone"`
	Sort     string            `json:"sort"`
	Limit    int               `json:"limit"`
	Cursor   string            `json:"cursor"`
	Indexes  []string          `json:"indexes"`
	Tags     map[string]string `json:"tags,omitempty"`
}

// LogSearchResponse contains the response from log search
type LogSearchResponse struct {
	Data   []LogEvent        `json:"data"`
	Links  map[string]string `json:"links,omitempty"`
	Meta   map[string]any    `json:"meta,omitempty"`
	Errors []map[string]any  `json:"errors,omitempty"`
}

// LogEvent represents a log event
type LogEvent struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	Message    string                 `json:"message"`
	Host       string                 `json:"host"`
	Service    string                 `json:"service"`
	Environment string               `json:"environment"`
	Tags       []string               `json:"tags"`
	Attributes map[string]any         `json:"attributes"`
}

// LogSubmissionResponse contains the response from log submission
type LogSubmissionResponse struct {
	Submitted int      `json:"submitted"`
	Errors    []string `json:"errors,omitempty"`
}

