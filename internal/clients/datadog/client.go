package datadog

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	datadogapi "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"buddy/internal/config"
	"buddy/internal/logging"
)

type DatadogClient struct {
	config    DatadogConfig
	apiClient *datadogapi.APIClient
	logsAPI   *datadogV2.LogsApi
	authCtx   context.Context
	logger    *logging.Logger
}

type DatadogConfig struct {
	BaseURL string
	APIKey  string
	AppKey  string
	Timeout int
}

func NewDatadogClient(env string) *DatadogClient {
	logger := logging.NewDefaultLogger("datadog")

	apiKey := config.Get("DD_API_KEY", "")
	appKey := config.Get("DD_APPLICATION_KEY", "")
	baseURL := config.Get("DATADOG_BASE_URL", "https://api.datadoghq.com")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	apiCfg := datadogapi.NewConfiguration()
	apiCfg.HTTPClient = httpClient
	apiCfg.Servers = datadogapi.ServerConfigurations{{URL: baseURL}}
	apiCfg.OperationServers = map[string]datadogapi.ServerConfigurations{
		"LogsApi.ListLogs":      {{URL: baseURL}},
		"LogsApi.AggregateLogs": {{URL: baseURL}},
		"LogsApi.SubmitLog":     {{URL: baseURL}},
	}

	apiClient := datadogapi.NewAPIClient(apiCfg)

	authCtx := datadogapi.NewDefaultContext(context.Background())
	authCtx = context.WithValue(authCtx, datadogapi.ContextAPIKeys, map[string]datadogapi.APIKey{
		"apiKeyAuth": {Key: apiKey},
		"appKeyAuth": {Key: appKey},
	})

	return &DatadogClient{
		config: DatadogConfig{
			BaseURL: baseURL,
			APIKey:  apiKey,
			AppKey:  appKey,
			Timeout: 30,
		},
		apiClient: apiClient,
		logsAPI:   datadogV2.NewLogsApi(apiClient),
		authCtx:   authCtx,
		logger:    logger,
	}
}

func (c *DatadogClient) SearchLogs(params LogSearchParams) (*LogSearchResponse, error) {
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

	resp, httpResp, err := c.logsAPI.ListLogs(c.authCtx, *datadogV2.NewListLogsOptionalParameters().WithBody(*req))
	if httpResp != nil && httpResp.Body != nil {
		defer func() { _ = httpResp.Body.Close() }()
	}
	if err != nil {
		return nil, err
	}

	out := &LogSearchResponse{
		Data:  make([]LogEvent, 0, len(resp.GetData())),
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
		event := LogEvent{
			ID:   item.GetId(),
			Type: string(item.GetType()),
		}
		if attrs, ok := item.GetAttributesOk(); ok && attrs != nil {
			if m := decodeToMap(attrs); m != nil {
				event.Attributes = m
			}
		}
		out.Data = append(out.Data, event)
	}

	return out, nil
}

func (c *DatadogClient) AggregateLogs(request datadogV2.LogsAggregateRequest) (*datadogV2.LogsAggregateResponse, *http.Response, error) {
	resp, httpResp, err := c.logsAPI.AggregateLogs(c.authCtx, request)
	if httpResp != nil && httpResp.Body != nil {
		defer func() { _ = httpResp.Body.Close() }()
	}
	if err != nil {
		return nil, httpResp, err
	}
	return &resp, httpResp, nil
}

func (c *DatadogClient) SubmitLogs(body []datadogV2.HTTPLogItem, opts *datadogV2.SubmitLogOptionalParameters) (any, *http.Response, error) {
	var (
		resp     any
		httpResp *http.Response
		err      error
	)
	if opts != nil {
		resp, httpResp, err = c.logsAPI.SubmitLog(c.authCtx, body, *opts)
	} else {
		resp, httpResp, err = c.logsAPI.SubmitLog(c.authCtx, body)
	}
	if httpResp != nil && httpResp.Body != nil {
		defer func() { _ = httpResp.Body.Close() }()
	}
	return resp, httpResp, err
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
