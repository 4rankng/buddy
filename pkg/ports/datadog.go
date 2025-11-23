package ports

import "time"

// DatadogPort defines the interface for Datadog logs and metrics operations
type DatadogPort interface {
	// Log operations
	SearchLogs(params *LogSearchParams) (*LogSearchResponse, error)
	SubmitLogs(logs []LogEvent) (*LogSubmissionResponse, error)
	AggregateLogs(request *LogAggregationRequest) (*LogAggregationResponse, error)

	// Metrics operations
	GetMetricQuery(query string, from, to time.Time) (*MetricResponse, error)
	GetMetricsByTags(tags map[string]string, from, to time.Time) ([]MetricPoint, error)

	// Monitor operations
	CreateMonitor(monitor *Monitor) (*MonitorResponse, error)
	GetMonitor(monitorID int) (*Monitor, error)
	ListMonitors(tags map[string]string) ([]Monitor, error)

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

// LogAggregationRequest represents a log aggregation request
type LogAggregationRequest struct {
	Query   string                 `json:"query"`
	From    time.Time              `json:"from"`
	To      time.Time              `json:"to"`
	Aggregations []LogAggregation   `json:"aggregations"`
}

// LogAggregation defines how to aggregate logs
type LogAggregation struct {
	Type   string `json:"type"`   // count, avg, sum, min, max
	Field  string `json:"field"`
	As     string `json:"as"`
}

// LogAggregationResponse contains aggregated log data
type LogAggregationResponse struct {
	Buckets []LogBucket `json:"buckets"`
	Links   map[string]string `json:"links,omitempty"`
	Meta    map[string]any   `json:"meta,omitempty"`
}

// LogBucket represents an aggregation bucket
type LogBucket struct {
	By     map[string]any `json:"by,omitempty"`
	Computers []LogCompute `json:"computers"`
}

// LogCompute represents a computed aggregation value
type LogCompute struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// MetricResponse contains metric query results
type MetricResponse struct {
	Query      string         `json:"query"`
	From       time.Time      `json:"from"`
	To         time.Time      `json:"to"`
	Series     []MetricSeries `json:"series"`
	Resolution int            `json:"resolution"`
}

// MetricSeries represents a time series of metric data
type MetricSeries struct {
	Name       string        `json:"name"`
	Tags       map[string]string `json:"tags"`
	Points     []MetricPoint `json:"points"`
	Host       string        `json:"host"`
	Service    string        `json:"service"`
}

// MetricPoint represents a single metric data point
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// Monitor represents a Datadog monitor
type Monitor struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Query       string                 `json:"query"`
	Message     string                 `json:"message"`
	Tags        []string               `json:"tags"`
	Options     MonitorOptions         `json:"options"`
	State       MonitorState           `json:"state"`
	Created     time.Time              `json:"created"`
	Modified    time.Time              `json:"modified"`
	Multi       bool                   `json:"multi"`
	Priority    int                    `json:"priority"`
	OverallState string               `json:"overall_state"`
}

// MonitorOptions contains configuration options for a monitor
type MonitorOptions struct {
	Thresholds      map[string]float64 `json:"thresholds"`
	NotifyAudit     bool               `json:"notify_audit"`
	RequireFullWindow bool            `json:"require_full_window"`
	NotifyNoData    bool               `json:"notify_no_data"`
	RenegotiateWindow int              `json:"renotify_every"`
	EvaluationDelay int                `json:"evaluation_delay"`
}

// MonitorState represents the current state of a monitor
type MonitorState struct {
	Groupings []MonitorGrouping `json:"groupings"`
}

// MonitorGrouping represents a monitor state grouping
type MonitorGrouping struct {
	Group   string          `json:"group"`
	Status  string          `json:"status"`
	LastUpdated time.Time   `json:"last_updated"`
}

// MonitorResponse contains the response from monitor creation/retrieval
type MonitorResponse struct {
	Monitor *Monitor `json:"monitor"`
	Errors  []string `json:"errors,omitempty"`
}