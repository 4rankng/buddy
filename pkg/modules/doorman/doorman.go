package doorman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// NewDoormanModule creates a new Doorman module
func NewDoormanModule(cfg config.DoormanConfig) (ports.DoormanPort, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: cfg.Timeout,
		Jar:     jar,
	}

	return &doormanModule{
		config:  cfg,
		client:  client,
		baseURL: cfg.BaseURL,
	}, nil
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

// Authenticate performs authentication with Doorman
func (d *doormanModule) authenticate() error {
	loginURL, _ := url.JoinPath(d.baseURL, "/api/login/ldap/signin")

	payload := doormanLoginReq{
		Username: d.config.User,
		Password: d.config.Password,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.executeWithRetry(req)
	if err != nil {
		return fmt.Errorf("doorman authentication failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("doorman auth failed: %s - %s", resp.Status, string(bodyBytes))
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
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Validate query to prevent SQL injection
	if err := d.validateQuery(query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	qURL, _ := url.JoinPath(d.baseURL, "/api/rds/query/execute")

	payload := queryPayload{
		ClusterName:  cluster,
		InstanceName: instance,
		Schema:       schema,
		Query:        strings.TrimSpace(query),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, qURL, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.executeWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("doorman query failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var qr queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	if qr.Code != 200 {
		return nil, fmt.Errorf("doorman query returned non-200 code: %d", qr.Code)
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

// validateQuery performs basic SQL query validation
func (d *doormanModule) validateQuery(query string) error {
	query = strings.ToUpper(strings.TrimSpace(query))

	// List of potentially dangerous SQL statements
	dangerousStatements := []string{
		"DROP", "DELETE", "UPDATE", "INSERT", "CREATE", "ALTER",
		"TRUNCATE", "EXEC", "EXECUTE", "GRANT", "REVOKE",
	}

	for _, stmt := range dangerousStatements {
		if strings.HasPrefix(query, stmt) {
			return fmt.Errorf("potentially dangerous SQL statement detected: %s", stmt)
		}
	}

	return nil
}

// QueryPaymentEngine executes queries on the Payment Engine cluster
func (d *doormanModule) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return d.ExecuteQuery("sg-prd-m-payment-engine", "sg-prd-m-payment-engine", "prod_payment_engine_db01", query)
}

// QueryPaymentCore executes queries on the Payment Core cluster
func (d *doormanModule) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	return d.ExecuteQuery("sg-prd-m-payment-core", "sg-prd-m-payment-core", "prod_payment_core_db01", query)
}

// QueryPartnerPayEngine executes queries on the Partner Pay Engine cluster
func (d *doormanModule) QueryPartnerPayEngine(query string) ([]map[string]interface{}, error) {
	return d.ExecuteQuery("sg-prd-m-partnerpay-engine", "sg-prd-m-partnerpay-engine", "prod_partnerpay_engine_db01", query)
}

// QueryPairingService executes queries on the Pairing Service cluster
func (d *doormanModule) QueryPairingService(query string) ([]map[string]interface{}, error) {
	return d.ExecuteQuery("sg-prd-m-pairing-service", "sg-prd-m-pairing-service", "prod_pairing_service_db01", query)
}

// QueryTransactionLimit executes queries on the Transaction Limit cluster
func (d *doormanModule) QueryTransactionLimit(query string) ([]map[string]interface{}, error) {
	return d.ExecuteQuery("sg-prd-m-transaction-limit", "sg-prd-m-transaction-limit", "prod_transaction_limit_db01", query)
}

// HealthCheck performs a health check on the Doorman service
func (d *doormanModule) HealthCheck() error {
	// Try to authenticate as a health check
	if err := d.authenticate(); err != nil {
		return fmt.Errorf("doorman health check failed: %w", err)
	}
	return nil
}

// GetTransactionStatus queries the status of a specific transaction
func (d *doormanModule) GetTransactionStatus(transactionID string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM transactions WHERE transaction_id = '%s' LIMIT 1", transactionID)
	return d.QueryPaymentCore(query)
}

// GetStuckTransactions retrieves transactions that are stuck in a particular state
func (d *doormanModule) GetStuckTransactions(state string, hours int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT transaction_id, status, created_at, updated_at, amount
		FROM transactions
		WHERE status = '%s'
		AND updated_at < NOW() - INTERVAL '%d hours'
		ORDER BY created_at DESC
		LIMIT 100
	`, state, hours)

	return d.QueryPaymentCore(query)
}

// FixStuckTransaction applies a fix to a stuck transaction
func (d *doormanModule) FixStuckTransaction(transactionID, fixType string) ([]map[string]interface{}, error) {
	var query string

	switch fixType {
	case "mark_failed":
		query = fmt.Sprintf("UPDATE transactions SET status = 'FAILED', updated_at = NOW() WHERE transaction_id = '%s'", transactionID)
	case "retry":
		query = fmt.Sprintf("UPDATE transactions SET status = 'PENDING', retry_count = retry_count + 1, updated_at = NOW() WHERE transaction_id = '%s'", transactionID)
	case "cancel":
		query = fmt.Sprintf("UPDATE transactions SET status = 'CANCELLED', updated_at = NOW() WHERE transaction_id = '%s'", transactionID)
	default:
		return nil, fmt.Errorf("unknown fix type: %s", fixType)
	}

	return d.QueryPaymentCore(query)
}