package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	datadogapi "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"oncall/pkg/config"
	"oncall/pkg/ports"
)

// datadogModule implements the DatadogPort interface
type datadogModule struct {
	config    config.DatadogConfig
	apiClient *datadogapi.APIClient
	logsAPI   *datadogV2.LogsApi
	authCtx   context.Context
}

// NewDatadogModule creates a new Datadog module
func NewDatadogModule(cfg config.DatadogConfig) (ports.DatadogPort, error) {
	if cfg.APIKey == "" || cfg.AppKey == "" {
		return nil, fmt.Errorf("Datadog API key and App key are required")
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	svc := &datadogModule{
		config:    cfg,
		apiClient: datadogapi.NewAPIClient(configuration(cfg.BaseURL, httpClient)),
	}

	svc.logsAPI = datadogV2.NewLogsApi(svc.apiClient)
	svc.authCtx = newAuthContext(cfg.APIKey, cfg.AppKey)

	return svc, nil
}

// SearchLogs searches for logs using the provided parameters
func (d *datadogModule) SearchLogs(params *ports.LogSearchParams) (*ports.LogSearchResponse, error) {
	if params.Query == "" {
		params.Query = "*"
	}
	if params.From == "" {
		params.From = "now-15m"
	}
	if params.To == "" {
		params.To = "now"
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	req := datadogV2.NewLogsListRequest()
	filter := datadogV2.NewLogsQueryFilter()
	filter.SetQuery(params.Query)
	filter.SetFrom(params.From)
	filter.SetTo(params.To)

	if len(params.Indexes) > 0 {
		filter.SetIndexes(params.Indexes)
	}

	if len(params.Tags) > 0 {
		tags := make([]string, 0, len(params.Tags))
		for k, v := range params.Tags {
			tags = append(tags, fmt.Sprintf("%s:%s", k, v))
		}
		filter.SetIndexes(tags)
	}

	req.SetFilter(*filter)

	page := datadogV2.NewLogsListRequestPage()
	limit := int32(params.Limit)
	page.SetLimit(limit)

	if params.Cursor != "" {
		page.SetCursor(params.Cursor)
	}

	req.SetPage(*page)

	normalizedSort := normalizeSort(params.Sort)
	if normalizedSort != "" {
		if sortVal, err := datadogV2.NewLogsSortFromValue(normalizedSort); err == nil {
			req.SetSort(*sortVal)
		}
	}

	resp, httpResp, err := d.logsAPI.ListLogs(d.authCtx, *datadogV2.NewListLogsOptionalParameters().WithBody(*req))
	closeBody(httpResp)

	if err != nil {
		status, code := httpStatus(httpResp)
		return nil, fmt.Errorf("datadog search error: %v (status: %s, code: %d)", err, status, code)
	}

	out := &ports.LogSearchResponse{
		Data:  make([]ports.LogEvent, 0, len(resp.GetData())),
		Links: map[string]string{},
		Meta:  map[string]any{},
	}

	if links, ok := resp.GetLinksOk(); ok && links != nil {
		if next, ok := links.GetNextOk(); ok && next != nil {
			out.Links["next"] = *next
		}
	}

	if meta, ok := resp.GetMetaOk(); ok && meta != nil {
		if m := decodeToMap(meta); m != nil {
			out.Meta = m
		}
	}

	for _, item := range resp.GetData() {
		event := ports.LogEvent{
			ID:   item.GetId(),
			Type: string(item.GetType()),
		}

		if attrs, ok := item.GetAttributesOk(); ok && attrs != nil {
			if m := decodeToMap(attrs); m != nil {
				event.Attributes = m

				// Extract common fields
				if timestamp, ok := m["timestamp"].(string); ok {
					if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
						event.Timestamp = t
					}
				}

				if message, ok := m["message"].(string); ok {
					event.Message = message
				}

				if host, ok := m["host"].(string); ok {
					event.Host = host
				}

				if service, ok := m["service"].(string); ok {
					event.Service = service
				}

				if env, ok := m["env"].(string); ok {
					event.Environment = env
				}

				// Extract tags
				if tags, ok := m["tags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagStr, ok := tag.(string); ok {
							event.Tags = append(event.Tags, tagStr)
						}
					}
				}
			}
		}

		out.Data = append(out.Data, event)
	}

	return out, nil
}

// SubmitLogs submits logs to Datadog
func (d *datadogModule) SubmitLogs(logs []ports.LogEvent) (*ports.LogSubmissionResponse, error) {
	httpLogItems := make([]datadogV2.HTTPLogItem, 0, len(logs))

	for _, log := range logs {
		item := datadogV2.NewHTTPLogItem()
		item.SetId(log.ID)
		item.SetMessage(log.Message)
		item.SetHost(log.Host)
		item.SetService(log.Service)

		// Add tags
		if len(log.Tags) > 0 {
			item.SetTags(log.Tags)
		}

		// Add attributes
		attributes := make(map[string]interface{})
		for k, v := range log.Attributes {
			attributes[k] = v
		}
		if len(attributes) > 0 {
			item.SetAttributes(attributes)
		}

		httpLogItems = append(httpLogItems, *item)
	}

	resp, httpResp, err := d.logsAPI.SubmitLog(d.authCtx, httpLogItems)
	closeBody(httpResp)

	if err != nil {
		status, code := httpStatus(httpResp)
		return &ports.LogSubmissionResponse{
			Submitted: 0,
			Errors:    []string{fmt.Sprintf("submission error: %v (status: %s, code: %d)", err, status, code)},
		}, err
	}

	submittedCount := len(logs)
	if resp != nil {
		// Extract submission count from response if available
		if m := decodeToMap(resp); m != nil {
			if count, ok := m["count"].(float64); ok {
				submittedCount = int(count)
			}
		}
	}

	return &ports.LogSubmissionResponse{
		Submitted: submittedCount,
		Errors:    []string{},
	}, nil
}

// AggregateLogs performs log aggregation
func (d *datadogModule) AggregateLogs(request *ports.LogAggregationRequest) (*ports.LogAggregationResponse, error) {
	ddRequest := datadogV2.NewLogsAggregateRequest()

	filter := datadogV2.NewLogsAggregateRequestFilter()
	filter.SetQuery(request.Query)
	filter.SetFrom(request.From)
	filter.SetTo(request.To)
	ddRequest.SetFilter(*filter)

	// Convert aggregations
	computes := make([]datadogV2.LogsAggregateFunction, 0, len(request.Aggregations))
	for _, agg := range request.Aggregations {
		compute := datadogV2.NewLogsAggregateFunction()
		compute.SetType(agg.Type)

		if agg.Field != "" {
			compute.SetField(agg.Field)
		}

		if agg.As != "" {
			compute.SetAs(agg.As)
		}

		computes = append(computes, *compute)
	}

	ddRequest.SetCompute(computes)

	resp, httpResp, err := d.logsAPI.AggregateLogs(d.authCtx, *ddRequest)
	closeBody(httpResp)

	if err != nil {
		status, code := httpStatus(httpResp)
		return nil, fmt.Errorf("datadog aggregation error: %v (status: %s, code: %d)", err, status, code)
	}

	out := &ports.LogAggregationResponse{
		Buckets: make([]ports.LogBucket, 0, len(resp.GetBuckets())),
		Links:   map[string]string{},
		Meta:    map[string]any{},
	}

	for _, bucket := range resp.GetBuckets() {
		logBucket := ports.LogBucket{
			By: map[string]any{},
		}

		if by, ok := bucket.GetByOk(); ok && by != nil {
			logBucket.By = decodeToMap(by)
		}

		computers := make([]ports.LogCompute, 0, len(bucket.GetCompute()))
		for _, compute := range bucket.GetCompute() {
			logCompute := ports.LogCompute{
				Type: string(compute.GetType()),
			}

			if val, ok := compute.GetValueOk(); ok && val != nil {
				logCompute.Value = *val
			}

			computers = append(computers, logCompute)
		}

		logBucket.Computers = computers
		out.Buckets = append(out.Buckets, logBucket)
	}

	return out, nil
}

// GetMetricQuery retrieves metrics using a query string
func (d *datadogModule) GetMetricQuery(query string, from, to time.Time) (*ports.MetricResponse, error) {
	// This would require implementing the MetricsApi from Datadog
	// For now, return a placeholder implementation
	return &ports.MetricResponse{
		Query:      query,
		From:       from,
		To:         to,
		Series:     []ports.MetricSeries{},
		Resolution: 60,
	}, fmt.Errorf("metric query not yet implemented")
}

// GetMetricsByTags retrieves metrics by tags
func (d *datadogModule) GetMetricsByTags(tags map[string]string, from, to time.Time) ([]ports.MetricPoint, error) {
	// This would require implementing the MetricsApi from Datadog
	// For now, return a placeholder implementation
	return []ports.MetricPoint{}, fmt.Errorf("metrics by tags not yet implemented")
}

// CreateMonitor creates a new monitor
func (d *datadogModule) CreateMonitor(monitor *ports.Monitor) (*ports.MonitorResponse, error) {
	// This would require implementing the MonitorsApi from Datadog
	// For now, return a placeholder implementation
	return &ports.MonitorResponse{
		Monitor: monitor,
		Errors:  []string{"monitor creation not yet implemented"},
	}, fmt.Errorf("monitor creation not yet implemented")
}

// GetMonitor retrieves a specific monitor
func (d *datadogModule) GetMonitor(monitorID int) (*ports.Monitor, error) {
	// This would require implementing the MonitorsApi from Datadog
	// For now, return a placeholder implementation
	return nil, fmt.Errorf("get monitor not yet implemented")
}

// ListMonitors lists monitors with optional tag filtering
func (d *datadogModule) ListMonitors(tags map[string]string) ([]ports.Monitor, error) {
	// This would require implementing the MonitorsApi from Datadog
	// For now, return a placeholder implementation
	return []ports.Monitor{}, fmt.Errorf("list monitors not yet implemented")
}

// HealthCheck performs a health check on the Datadog service
func (d *datadogModule) HealthCheck() error {
	// Try a simple log search as a health check
	params := &ports.LogSearchParams{
		Query: "*",
		From:  "now-5m",
		To:    "now",
		Limit: 1,
	}

	_, err := d.SearchLogs(params)
	if err != nil {
		return fmt.Errorf("datadog health check failed: %w", err)
	}

	return nil
}

// Helper functions from the archived app

func configuration(baseURL string, client *http.Client) *datadogapi.Configuration {
	cfg := datadogapi.NewConfiguration()
	cfg.HTTPClient = client
	cfg.Servers = datadogapi.ServerConfigurations{{
		URL: baseURL,
	}}
	cfg.OperationServers = map[string]datadogapi.ServerConfigurations{
		"LogsApi.ListLogs":      {{URL: baseURL}},
		"LogsApi.AggregateLogs": {{URL: baseURL}},
		"LogsApi.SubmitLog":     {{URL: baseURL}},
	}
	return cfg
}

func newAuthContext(apiKey, appKey string) context.Context {
	ctx := datadogapi.NewDefaultContext(context.Background())
	ctx = context.WithValue(ctx, datadogapi.ContextAPIKeys, map[string]datadogapi.APIKey{
		"apiKeyAuth": {Key: apiKey},
		"appKeyAuth": {Key: appKey},
	})
	return ctx
}

func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func httpStatus(resp *http.Response) (string, int) {
	if resp == nil {
		return "", 0
	}
	return resp.Status, resp.StatusCode
}

func decodeToMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case map[string]any:
		return v
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil
	}
	return out
}

func normalizeSort(sort string) string {
	s := strings.TrimSpace(strings.ToLower(sort))
	switch s {
	case "", "-timestamp", "desc", "descending":
		return string(datadogV2.LOGSSORT_TIMESTAMP_DESCENDING)
	case "timestamp", "asc", "ascending":
		return string(datadogV2.LOGSSORT_TIMESTAMP_ASCENDING)
	default:
		return sort
	}
}