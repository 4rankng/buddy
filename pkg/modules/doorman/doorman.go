package doorman

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"oncall/pkg/config"
	"oncall/pkg/ports"
)

// doormanModule implements the DoormanPort interface
type doormanModule struct {
	config  config.DoormanConfig
	client  *http.Client
	baseURL string
}

// doormanLoginReq represents the login request payload
type doormanLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// queryPayload represents the SQL query payload
type queryPayload struct {
	ClusterName  string `json:"clusterName"`
	InstanceName string `json:"instanceName"`
	Schema       string `json:"schema"`
	Query        string `json:"query"`
}

// queryResponse represents the SQL query response
type queryResponse struct {
	Code   int `json:"code"`
	Result struct {
		Headers []string        `json:"headers"`
		Rows    [][]interface{} `json:"rows"`
	} `json:"result"`
}

// NewDoormanModule creates a new Doorman module
func NewDoormanModule(cfg config.DoormanConfig) (ports.DoormanPort, error) {
	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Timeout: cfg.Timeout, Jar: jar}
	return &doormanModule{
		config:  cfg,
		client:  cli,
		baseURL: cfg.BaseURL,
	}, nil
}

// authenticate performs authentication with Doorman
func (d *doormanModule) authenticate() error {
	loginURL, _ := url.JoinPath(d.baseURL, "/api/login/ldap/signin")

	payload := doormanLoginReq{
		Username: d.config.User,
		Password: d.config.Password,
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return errors.New("doorman auth failed: " + resp.Status)
	}
	return nil
}

// executeWithRetry executes an HTTP request with retry logic
func (d *doormanModule) executeWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= d.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(d.config.RetryDelay)
		}

		resp, err := d.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// ExecuteQuery executes a SQL query on the specified cluster/instance/schema
func (d *doormanModule) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	if err := d.authenticate(); err != nil {
		return nil, err
	}

	// Validate query to prevent SQL injection
	if err := d.validateQuery(query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	qURL, _ := url.JoinPath(d.baseURL, "/api/rds/query/execute")
	payload := queryPayload{ClusterName: cluster, InstanceName: instance, Schema: schema, Query: strings.TrimSpace(query)}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, qURL, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.executeWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, errors.New("doorman query failed: " + resp.Status)
	}

	var qr queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return nil, err
	}

	if qr.Code != 200 {
		return nil, errors.New("doorman query returned non-200 code")
	}

	// Convert response to map format
	rows := make([]map[string]interface{}, 0, len(qr.Result.Rows))
	for _, r := range qr.Result.Rows {
		row := map[string]interface{}{}
		for i, h := range qr.Result.Headers {
			if i < len(r) {
				row[h] = r[i]
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// CreateDMLTicket creates a DML ticket for database operations
func (d *doormanModule) CreateDMLTicket(request *ports.DMLRequest) (string, error) {
	if err := d.authenticate(); err != nil {
		return "", err
	}

	// Validate query to prevent SQL injection (basic check)
	if err := d.validateQuery(request.Query); err != nil {
		return "", fmt.Errorf("query validation failed: %w", err)
	}

	url, _ := url.JoinPath(d.baseURL, "/api/rds/dml/create_ticket")

	payload := map[string]interface{}{
		"clusterName":       request.ClusterName,
		"instanceName":      request.InstanceName,
		"schema":            request.Schema,
		"originalQuery":     request.Query,
		"rollbackQuery":     "", // Optional, can be added to request if needed
		"toolLabel":         "direct",
		"skipWhereClause":   false,
		"skipRollbackQuery": true,
		"note":              request.Description,
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.executeWithRetry(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", errors.New("doorman ticket creation failed: " + resp.Status)
	}

	var result struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Result) > 0 {
		return result.Result[0].ID, nil
	}

	return "", errors.New("no ticket ID returned")
}


// GetAvailableClusters returns a list of available database clusters
func (d *doormanModule) GetAvailableClusters() ([]ports.DatabaseCluster, error) {
	// Return a list of known clusters from configuration
	clusters := []ports.DatabaseCluster{
		{
			Name:        "sg-prd-m-payment-engine",
			Instance:    "sg-prd-m-payment-engine",
			Schema:      "prod_payment_engine_db01",
			Description: "Payment processing engine database",
			Environment: "production",
		},
		{
			Name:        "sg-prd-m-payment-core",
			Instance:    "sg-prd-m-payment-core",
			Schema:      "prod_payment_core_db01",
			Description: "Payment core services database",
			Environment: "production",
		},
		{
			Name:        "sg-prd-m-partnerpay-engine",
			Instance:    "sg-prd-m-partnerpay-engine",
			Schema:      "prod_partnerpay_engine_db01",
			Description: "Partner payment processing database",
			Environment: "production",
		},
		{
			Name:        "sg-prd-m-pairing-service",
			Instance:    "sg-prd-m-pairing-service",
			Schema:      "prod_pairing_service_db01",
			Description: "Device pairing service database",
			Environment: "production",
		},
		{
			Name:        "sg-prd-m-transaction-limit",
			Instance:    "sg-prd-m-transaction-limit",
			Schema:      "prod_transaction_limit_db01",
			Description: "Transaction limit management database",
			Environment: "production",
		},
	}

	return clusters, nil
}

// validateQuery performs basic SQL query validation
func (d *doormanModule) validateQuery(query string) error {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))

	// List of potentially dangerous SQL statements
	dangerousStatements := []string{
		"DROP", "DELETE", "UPDATE", "INSERT", "CREATE", "ALTER",
		"TRUNCATE", "EXEC", "EXECUTE", "GRANT", "REVOKE",
	}

	for _, stmt := range dangerousStatements {
		if strings.HasPrefix(queryUpper, stmt) {
			return fmt.Errorf("potentially dangerous SQL statement detected: %s", stmt)
		}
	}

	return nil
}


// HealthCheck performs a health check on the Doorman service
func (d *doormanModule) HealthCheck() error {
	// Try to authenticate as a health check
	if err := d.authenticate(); err != nil {
		return fmt.Errorf("doorman health check failed: %w", err)
	}
	return nil
}
