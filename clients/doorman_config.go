package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"buddy/config"
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

	// Account ID for API requests
	AccountID string

	// Database cluster/instance names
	PaymentEngine    string
	PaymentCore      string
	FastAdapter      string
	RppAdapter       string
	PartnerpayEngine string
}

// DoormanClient singleton with configuration
type DoormanClient struct {
	config        DoormanConfig
	httpClient    *http.Client
	mu            sync.RWMutex
	authenticated bool
}

var (
	singleton *DoormanClient
	once      sync.Once
)

// Ensure DoormanClient implements DoormanInterface
var _ DoormanInterface = (*DoormanClient)(nil)

// Environment-specific configurations
var configs = map[string]DoormanConfig{
	"sg": {
		Host: "https://doorman.sgbank.pr",
		Auth: AuthInfo{
			Username: config.Get("DOORMAN_USERNAME", ""),
			Password: config.Get("DOORMAN_PASSWORD", ""),
		},
		AccountID:        "748118206017", // Singapore environment account ID
		PaymentEngine:    "sg-prd-m-payment-engine",
		PaymentCore:      "sg-prd-m-payment-core",
		FastAdapter:      "sg-prd-m-fast-adapter",
		RppAdapter:       "", // Not available in SG
		PartnerpayEngine: "", // Not available in SG
	},
	"my": {
		Host: "https://doorman.infra.prd.g-bank.app",
		Auth: AuthInfo{
			Username: config.Get("DOORMAN_USERNAME", ""),
			Password: config.Get("DOORMAN_PASSWORD", ""),
		},
		AccountID:        "559634300081", // Malaysia environment account ID
		PaymentEngine:    "prd-payments-payment-engine-rds-mysql",
		PaymentCore:      "prd-payments-payment-core-rds-mysql",
		FastAdapter:      "", // Not available in MY
		RppAdapter:       "prd-payments-rpp-adapter-rds-mysql",
		PartnerpayEngine: "prd-payments-partnerpay-engine-rds-mysql",
	},
}

// GetDoormanClient returns singleton DoormanClient instance
func GetDoormanClient(env string) (*DoormanClient, error) {
	var err error

	once.Do(func() {
		cfg, exists := configs[env]
		if !exists {
			cfg = configs["my"] // Default to Malaysia
		}

		singleton = &DoormanClient{
			config: cfg,
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	})

	return singleton, err
}

// GetConfig returns configuration for client
func (c *DoormanClient) GetConfig() DoormanConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// ResetSingleton resets the singleton instance (useful for testing)
func ResetSingleton() {
	once = sync.Once{}
	singleton = nil
}

// Authenticate performs authentication with doorman service
func (c *DoormanClient) Authenticate() error {
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

	req, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Set up cookie jar if not already set
	if c.httpClient.Jar == nil {
		c.httpClient.Jar, _ = cookiejar.New(nil)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error if close fails
		}
	}()

	if resp.StatusCode >= 300 {
		return errors.New("doorman auth failed: " + resp.Status)
	}

	c.authenticated = true
	return nil
}

// ExecuteQuery executes a query against the specified database cluster
func (c *DoormanClient) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	cfg := c.GetConfig()
	queryURL, _ := url.JoinPath(cfg.Host, "/api/rds/query/execute")

	queryReq := struct {
		AccountID    string `json:"accountID"`
		ClusterName  string `json:"clusterName"`
		InstanceName string `json:"instanceName"`
		Schema       string `json:"schema"`
		Query        string `json:"query"`
	}{
		AccountID:    cfg.AccountID, // Use configurable account ID
		ClusterName:  cluster,
		InstanceName: instance,
		Schema:       schema,
		Query:        query,
	}

	reqBody, _ := json.Marshal(queryReq)

	req, _ := http.NewRequest(http.MethodPost, queryURL, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error if close fails
		}
	}()

	if resp.StatusCode >= 300 {
		return nil, errors.New("doorman query failed: " + resp.Status)
	}

	var response struct {
		Code   int `json:"code"`
		Result struct {
			Headers []string        `json:"headers"`
			Types   []string        `json:"types"`
			Rows    [][]interface{} `json:"rows"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Convert the tabular response to a map-based format
	var data []map[string]interface{}
	for _, row := range response.Result.Rows {
		rowMap := make(map[string]interface{})
		for i, value := range row {
			if i < len(response.Result.Headers) {
				rowMap[response.Result.Headers[i]] = value
			}
		}
		data = append(data, rowMap)
	}

	return data, nil
}

// QueryPaymentEngine queries the payment engine database
func (c *DoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	return c.ExecuteQuery(cfg.PaymentEngine, cfg.PaymentEngine, "prod_payment_engine_db01", query)
}

// QueryPaymentCore queries the payment core database
func (c *DoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	return c.ExecuteQuery(cfg.PaymentCore, cfg.PaymentCore, "prod_payment_core_db01", query)
}

// QueryFastAdapter queries the fast adapter database (Singapore only)
func (c *DoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	if cfg.FastAdapter == "" {
		return nil, errors.New("fast adapter is not available in this environment")
	}
	return c.ExecuteQuery(cfg.FastAdapter, cfg.FastAdapter, "prod_fast_adapter_db01", query)
}

// QueryRppAdapter queries the rpp adapter database (Malaysia only)
func (c *DoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	if cfg.RppAdapter == "" {
		return nil, errors.New("rpp adapter is not available in this environment")
	}
	return c.ExecuteQuery(cfg.RppAdapter, cfg.RppAdapter, "prd-payments-rpp-adapter-rds-mysql", query)
}

// QueryPartnerpayEngine queries the partnerpay engine database (Malaysia only)
func (c *DoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	if cfg.PartnerpayEngine == "" {
		return nil, errors.New("partnerpay engine is not available in this environment")
	}
	return c.ExecuteQuery(cfg.PartnerpayEngine, cfg.PartnerpayEngine, "partnerpay_engine", query)
}
