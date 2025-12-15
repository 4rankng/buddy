package doorman

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"buddy/internal/config"
)

// AuthInfo holds authentication information
type AuthInfo struct {
	Username string
	Password string
}

// DBInfo holds database connection information
type DBInfo struct {
	ClusterName  string
	InstanceName string
	Schema       string
}

// DoormanConfig holds environment-specific configuration
type DoormanConfig struct {
	Host string
	Auth AuthInfo

	// Account ID for API requests
	AccountID string

	// Database cluster/instance names
	PaymentEngine    DBInfo
	PaymentCore      DBInfo
	FastAdapter      DBInfo
	RppAdapter       DBInfo
	PartnerpayEngine DBInfo
}

// DoormanClient singleton with configuration
type DoormanClient struct {
	config        DoormanConfig
	httpClient    *http.Client
	mu            sync.RWMutex
	authenticated bool
}

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
		AccountID: "748118206017", // Singapore environment account ID
		PaymentEngine: DBInfo{
			ClusterName:  "sg-prd-m-payment-engine",
			InstanceName: "sg-prd-m-payment-engine",
			Schema:       "prod_payment_engine_db01",
		},
		PaymentCore: DBInfo{
			ClusterName:  "sg-prd-m-payment-core",
			InstanceName: "sg-prd-m-payment-core",
			Schema:       "prod_payment_core_db01",
		},
		FastAdapter: DBInfo{
			ClusterName:  "sg-prd-m-fast-adapter",
			InstanceName: "sg-prd-m-fast-adapter",
			Schema:       "prod_fast_adapter_db01",
		},
		RppAdapter: DBInfo{}, // Not available in SG
		PartnerpayEngine: DBInfo{
			ClusterName:  "sg-prd-m-partnerpay-engine",
			InstanceName: "sg-prd-m-partnerpay-engine",
			Schema:       "prod_partnerpay_engine_db01",
		},
	},
	"my": {
		Host: "https://doorman.infra.prd.g-bank.app",
		Auth: AuthInfo{
			Username: config.Get("DOORMAN_USERNAME", ""),
			Password: config.Get("DOORMAN_PASSWORD", ""),
		},
		AccountID: "559634300081", // Malaysia environment account ID
		PaymentEngine: DBInfo{
			ClusterName:  "prd-payments-payment-engine-rds-mysql",
			InstanceName: "prd-payments-payment-engine-rds-mysql",
			Schema:       "payment_engine",
		},
		PaymentCore: DBInfo{
			ClusterName:  "prd-payments-payment-core-rds-mysql",
			InstanceName: "prd-payments-payment-core-rds-mysql",
			Schema:       "payment_core",
		},
		FastAdapter: DBInfo{}, // Not available in MY
		RppAdapter: DBInfo{
			ClusterName:  "prd-payments-rpp-adapter-rds-mysql",
			InstanceName: "prd-payments-rpp-adapter-rds-mysql",
			Schema:       "rpp_adapter",
		},
		PartnerpayEngine: DBInfo{
			ClusterName:  "prd-payments-partnerpay-engine-rds-mysql",
			InstanceName: "prd-payments-partnerpay-engine-rds-mysql",
			Schema:       "partnerpay_engine",
		},
	},
}

var Doorman DoormanInterface

// NewDoormanClient initializes the global Doorman client with the specified environment
func NewDoormanClient(env string) DoormanInterface {
	if Doorman != nil {
		return Doorman
	}
	fmt.Println("Initialize Doorman client for ", env)
	cfg, exists := configs[env]
	if !exists {
		panic(fmt.Sprintf("country %s is not supported", env))
	}

	Doorman = &DoormanClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return Doorman
}

// GetDoormanClient returns the initialized DoormanClient instance
// Deprecated: Use clients.Doorman directly after initialization
func GetDoormanClient(env string) (DoormanInterface, error) {
	if Doorman == nil {
		panic("Doorman client is not initialized")
	}
	return Doorman, nil
}

// GetConfig returns configuration for client
func (c *DoormanClient) GetConfig() DoormanConfig {
	return c.config
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
	fmt.Println("LOGIN URL: ", loginURL)
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
		_ = resp.Body.Close()
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
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 300 {
		// Read the response body to get more details about the error
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.New("doorman query failed: " + resp.Status + " (failed to read error response)")
		}
		return nil, errors.New("doorman query failed: " + resp.Status + " - " + string(body))
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
	dbInfo := cfg.PaymentEngine
	return c.ExecuteQuery(dbInfo.ClusterName, dbInfo.InstanceName, dbInfo.Schema, query)
}

// QueryPaymentCore queries the payment core database
func (c *DoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	dbInfo := cfg.PaymentCore
	return c.ExecuteQuery(dbInfo.ClusterName, dbInfo.InstanceName, dbInfo.Schema, query)
}

// QueryFastAdapter queries the fast adapter database (Singapore only)
func (c *DoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	dbInfo := cfg.FastAdapter
	if dbInfo.ClusterName == "" {
		return nil, errors.New("fast adapter is not available in this environment")
	}
	return c.ExecuteQuery(dbInfo.ClusterName, dbInfo.InstanceName, dbInfo.Schema, query)
}

// QueryRppAdapter queries the rpp adapter database (Malaysia only)
func (c *DoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	dbInfo := cfg.RppAdapter
	if dbInfo.ClusterName == "" {
		return nil, errors.New("rpp adapter is not available in this environment")
	}
	return c.ExecuteQuery(dbInfo.ClusterName, dbInfo.InstanceName, dbInfo.Schema, query)
}

// QueryPartnerpayEngine queries the partnerpay engine database (Malaysia only)
func (c *DoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	cfg := c.GetConfig()
	dbInfo := cfg.PartnerpayEngine
	if dbInfo.ClusterName == "" {
		return nil, errors.New("partnerpay engine is not available in this environment")
	}
	return c.ExecuteQuery(dbInfo.ClusterName, dbInfo.InstanceName, dbInfo.Schema, query)
}
