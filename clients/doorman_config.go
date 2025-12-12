package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"buddy/config"
	"buddy/output"
)

// AuthInfo holds authentication information
type AuthInfo struct {
	Username string
	Password string
}

// DoormanConfig holds environment-specific configuration
type DoormanConfig struct {
	Host string
	Auth AuthInfo

	// Database cluster/instance names
	PaymentEngine    string
	PaymentCore      string
	FastAdapter      string
	RppAdapter       string
	PartnerpayEngine string
}

// DoormanClient singleton with configuration
type DoormanClientSingleton struct {
	config        DoormanConfig
	httpClient    *http.Client
	mu            sync.RWMutex
	authenticated bool
}

var (
	singleton *DoormanClientSingleton
	once      sync.Once
)

// Ensure DoormanClientSingleton implements DoormanInterface
var _ DoormanInterface = (*DoormanClientSingleton)(nil)

// Environment-specific configurations
var configs = map[string]DoormanConfig{
	"sg": {
		Host: "https://doorman.sgbank.pr",
		Auth: AuthInfo{
			Username: config.Get("DOORMAN_USERNAME", ""),
			Password: config.Get("DOORMAN_PASSWORD", ""),
		},
		PaymentEngine:    "payment-engine",
		PaymentCore:      "payment-core",
		FastAdapter:      "fast-adapter",
		RppAdapter:       "", // Not available in SG
		PartnerpayEngine: "", // Not available in SG
	},
	"my": {
		Host: "https://doorman.infra.prd.g-bank.app",
		Auth: AuthInfo{
			Username: config.Get("DOORMAN_USERNAME", ""),
			Password: config.Get("DOORMAN_PASSWORD", ""),
		},
		PaymentEngine:    "payment-engine",
		PaymentCore:      "payment-core",
		FastAdapter:      "", // Not available in MY
		RppAdapter:       "rpp-adapter",
		PartnerpayEngine: "partnerpay-engine",
	},
}

// GetDoormanClient returns the singleton DoormanClient instance
func GetDoormanClient(env string) (*DoormanClientSingleton, error) {
	var err error

	once.Do(func() {
		cfg, exists := configs[env]
		if !exists {
			cfg = configs["my"] // Default to Malaysia
		}

		singleton = &DoormanClientSingleton{
			config: cfg,
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	})

	return singleton, err
}

// GetConfig returns the configuration for the client
func (c *DoormanClientSingleton) GetConfig() DoormanConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// ResetSingleton resets the singleton instance (useful for testing)
func ResetSingleton() {
	once = sync.Once{}
	singleton = nil
}

// Authenticate performs authentication with the doorman service
func (c *DoormanClientSingleton) Authenticate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.authenticated {
		return nil
	}

	cfg := c.config
	loginURL, _ := url.JoinPath(cfg.Host, "/api/login/ldap/signin")

	loginReq := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}

	reqBody, _ := json.Marshal(loginReq)

	// Debug logging
	output.LogEvent("doorman_auth_attempt", map[string]any{
		"url":      loginURL,
		"username": cfg.Auth.Username,
		"request":  string(reqBody),
	})

	req, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Set up cookie jar if not already set
	if c.httpClient.Jar == nil {
		c.httpClient.Jar, _ = cookiejar.New(nil)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		output.LogEvent("doorman_auth_http_error", map[string]any{"error": err.Error()})
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			output.LogEvent("doorman_auth_close_error", map[string]any{"error": err.Error()})
		}
	}()

	if resp.StatusCode >= 300 {
		var bodyBytes []byte
		if resp.Body != nil {
			bb, _ := io.ReadAll(resp.Body)
			bodyBytes = bb
		}
		output.LogEvent("doorman_auth_failed", map[string]any{"status": resp.Status, "body": string(bodyBytes)})
		return errors.New("doorman auth failed: " + resp.Status)
	}

	c.authenticated = true
	return nil
}

// ExecuteQuery executes a query against the specified database cluster
func (c *DoormanClientSingleton) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	cfg := c.GetConfig()
	queryURL, _ := url.JoinPath(cfg.Host, "/api/sql/query")

	queryReq := struct {
		Cluster  string `json:"cluster"`
		Instance string `json:"instance"`
		Schema   string `json:"schema"`
		Query    string `json:"query"`
	}{
		Cluster:  cluster,
		Instance: instance,
		Schema:   schema,
		Query:    query,
	}

	reqBody, _ := json.Marshal(queryReq)

	// Debug logging
	output.LogEvent("doorman_query_attempt", map[string]any{
		"url":     queryURL,
		"cluster": cluster,
		"schema":  schema,
		"query":   query,
	})

	req, _ := http.NewRequest(http.MethodPost, queryURL, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		output.LogEvent("doorman_query_http_error", map[string]any{"error": err.Error()})
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			output.LogEvent("doorman_query_close_error", map[string]any{"error": err.Error()})
		}
	}()

	if resp.StatusCode >= 300 {
		var bodyBytes []byte
		if resp.Body != nil {
			bb, _ := io.ReadAll(resp.Body)
			bodyBytes = bb
		}
		output.LogEvent("doorman_query_failed", map[string]any{"status": resp.Status, "body": string(bodyBytes)})
		return nil, errors.New("doorman query failed: " + resp.Status)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// QueryPaymentEngine queries the payment engine database
func (c *DoormanClientSingleton) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	return c.ExecuteQuery(cfg.PaymentEngine, cfg.PaymentEngine, "payment_engine", query)
}

// QueryPaymentCore queries the payment core database
func (c *DoormanClientSingleton) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	return c.ExecuteQuery(cfg.PaymentCore, cfg.PaymentCore, "payment_core", query)
}

// QueryFastAdapter queries the fast adapter database (Singapore only)
func (c *DoormanClientSingleton) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	if cfg.FastAdapter == "" {
		return nil, errors.New("fast adapter is not available in this environment")
	}
	return c.ExecuteQuery(cfg.FastAdapter, cfg.FastAdapter, "fast_adapter", query)
}

// QueryRppAdapter queries the rpp adapter database (Malaysia only)
func (c *DoormanClientSingleton) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	if cfg.RppAdapter == "" {
		return nil, errors.New("rpp adapter is not available in this environment")
	}
	return c.ExecuteQuery(cfg.RppAdapter, cfg.RppAdapter, "rpp_adapter", query)
}

// QueryPartnerpayEngine queries the partnerpay engine database (Malaysia only)
func (c *DoormanClientSingleton) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	if cfg.PartnerpayEngine == "" {
		return nil, errors.New("partnerpay engine is not available in this environment")
	}
	return c.ExecuteQuery(cfg.PartnerpayEngine, cfg.PartnerpayEngine, "partnerpay_engine", query)
}
