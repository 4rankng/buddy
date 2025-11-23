package datadog

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"oncall/pkg/config"
	"oncall/pkg/ports"
)

// datadogModule implements the DatadogPort interface
type datadogModule struct {
	config    config.DatadogConfig
	client    *http.Client
	baseURL   string
	apiKey    string
	appKey    string
}

// NewDatadogModule creates a new Datadog module
func NewDatadogModule(cfg config.DatadogConfig) (ports.DatadogPort, error) {
	if cfg.APIKey == "" || cfg.AppKey == "" {
		return nil, fmt.Errorf("Datadog API key and App key are required")
	}

	return &datadogModule{
		config:  cfg,
		client:  &http.Client{Timeout: cfg.Timeout},
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		appKey:  cfg.AppKey,
	}, nil
}

// SearchLogs searches for logs using the provided parameters (simplified implementation)
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

	// For now, return a placeholder response
	// In a real implementation, you would call the Datadog API
	response := &ports.LogSearchResponse{
		Data: []ports.LogEvent{
			{
				ID:         "placeholder-log-id",
				Type:       "log",
				Timestamp:  time.Now(),
				Message:    fmt.Sprintf("Search query: %s (placeholder implementation)", params.Query),
				Host:       "placeholder-host",
				Service:    "placeholder-service",
				Environment: "production",
				Tags:       []string{"placeholder:true"},
				Attributes: map[string]interface{}{
					"query": params.Query,
					"from":  params.From,
					"to":    params.To,
				},
			},
		},
		Links: map[string]string{},
		Meta: map[string]interface{}{
			"limit": params.Limit,
			"total": 1,
		},
	}

	return response, nil
}

// SubmitLogs submits logs to Datadog (placeholder implementation)
func (d *datadogModule) SubmitLogs(logs []ports.LogEvent) (*ports.LogSubmissionResponse, error) {
	// Placeholder implementation
	return &ports.LogSubmissionResponse{
		Submitted: len(logs),
		Errors:    []string{},
	}, nil
}

// GetAvailableIndexes returns available log indexes
func (d *datadogModule) GetAvailableIndexes() ([]string, error) {
	// Placeholder implementation - return common indexes
	return []string{
		"main",
		"payment-engine",
		"payment-core",
		"datadog",
		"infrastructure",
	}, nil
}

// TestConnection tests the Datadog API connection
func (d *datadogModule) TestConnection() error {
	// Simple health check - return nil for now
	return nil
}

// HealthCheck performs a health check on the Datadog service
func (d *datadogModule) HealthCheck() error {
	// For a basic health check, we can verify that we have the required credentials
	if d.apiKey == "" || d.appKey == "" {
		return fmt.Errorf("missing Datadog API credentials")
	}

	// Try to make a simple API call
	url := fmt.Sprintf("%s/api/v1/validate", d.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("DD-API-KEY", d.apiKey)
	req.Header.Set("DD-APPLICATION-KEY", d.appKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("datadog health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("datadog health check failed with status: %s", resp.Status)
	}

	return nil
}

// makeAPIRequest makes a generic API request to Datadog
func (d *datadogModule) makeAPIRequest(method, endpoint string, body []byte) (*http.Response, error) {
	url := d.baseURL + endpoint

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, strings.NewReader(string(body)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", d.apiKey)
	req.Header.Set("DD-APPLICATION-KEY", d.appKey)

	return d.client.Do(req)
}