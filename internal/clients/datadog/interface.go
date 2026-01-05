package datadog

import (
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

type LogSearchParams struct {
	Query   string
	From    string
	To      string
	Sort    string
	Limit   int
	Cursor  string
	Indexes []string
}

type LogEvent struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

type LogSearchResponse struct {
	Data  []LogEvent        `json:"data"`
	Links map[string]string `json:"links,omitempty"`
	Meta  map[string]any    `json:"meta,omitempty"`
}

type DatadogInterface interface {
	SearchLogs(params LogSearchParams) (*LogSearchResponse, error)
	AggregateLogs(request datadogV2.LogsAggregateRequest) (*datadogV2.LogsAggregateResponse, *http.Response, error)
	SubmitLogs(body []datadogV2.HTTPLogItem, opts *datadogV2.SubmitLogOptionalParameters) (any, *http.Response, error)
}
