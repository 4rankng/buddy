package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"buddy/config"
	"buddy/output"
)

// SGDoorman implements DoormanInterface for Singapore environment
type SGDoorman struct {
	BaseURL string
	User    string
	Pass    string
	HTTP    *http.Client
}

// NewSGDoorman creates a new Singapore doorman client
func NewSGDoorman(timeout time.Duration) (*SGDoorman, error) {
	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Timeout: timeout, Jar: jar}

	return &SGDoorman{
		BaseURL: "https://doorman.sgbank.pr",
		User:    config.Get("DOORMAN_USERNAME", ""),
		Pass:    config.Get("DOORMAN_PASSWORD", ""),
		HTTP:    cli,
	}, nil
}

// Authenticate performs authentication with the doorman service
func (c *SGDoorman) Authenticate() error {
	loginURL, _ := url.JoinPath(c.BaseURL, "/api/login/ldap/signin")
	loginReq := doormanLoginReq{Username: c.User, Password: c.Pass}
	reqBody, _ := json.Marshal(loginReq)

	// Debug logging
	output.LogEvent("doorman_auth_attempt", map[string]any{
		"url":      loginURL,
		"username": c.User,
		"request":  string(reqBody),
	})

	req, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
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
	return nil
}

// ExecuteQuery executes a query against the specified database cluster
func (c *SGDoorman) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}
	qURL, _ := url.JoinPath(c.BaseURL, "/api/rds/query/execute")
	payload := queryPayload{ClusterName: cluster, InstanceName: instance, Schema: schema, Query: query}
	reqBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, qURL, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		output.LogEvent("doorman_http_error", map[string]any{"error": err.Error(), "cluster": cluster, "schema": schema})
		return nil, err
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
		output.LogEvent("doorman_query_error", map[string]any{
			"status":  resp.Status,
			"cluster": cluster,
			"schema":  schema,
			"payload": string(reqBody),
			"body":    string(bodyBytes),
		})
		return nil, errors.New("doorman query failed: " + resp.Status)
	}
	var qr queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		output.LogEvent("doorman_decode_error", map[string]any{"error": err.Error(), "cluster": cluster, "schema": schema})
		return nil, err
	}
	if qr.Code != 200 {
		output.LogEvent("doorman_query_non200", map[string]any{"code": qr.Code, "cluster": cluster, "schema": schema})
		return nil, errors.New("doorman query returned non-200 code")
	}
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

// QueryPaymentEngine queries the Singapore payment-engine database
func (c *SGDoorman) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("sg-prd-m-payment-engine", "sg-prd-m-payment-engine", "prod_payment_engine_db01", query)
}

// QueryPaymentCore queries the Singapore payment-core database
func (c *SGDoorman) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("sg-prd-m-payment-core", "sg-prd-m-payment-core", "prod_payment_core_db01", query)
}

// QueryFastAdapter queries the Singapore fast-adapter database
func (c *SGDoorman) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("sg-prd-m-fast-adapter", "sg-prd-m-fast-adapter", "prod_fast_adapter_db01", query)
}

// QueryRppAdapter is not available in Singapore environment
func (c *SGDoorman) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	return nil, errors.New("RPP adapter is not available in Singapore environment")
}

// QueryPartnerpayEngine is not available in Singapore environment
func (c *SGDoorman) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return nil, errors.New("Partnerpay engine is not available in Singapore environment")
}
