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

// MYDoorman implements DoormanInterface for Malaysia environment
type MYDoorman struct {
	BaseURL string
	User    string
	Pass    string
	HTTP    *http.Client
}

// NewMYDoorman creates a new Malaysia doorman client
func NewMYDoorman(timeout time.Duration) (*MYDoorman, error) {
	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Timeout: timeout, Jar: jar}

	return &MYDoorman{
		BaseURL: "https://doorman.infra.prd.g-bank.app",
		User:    config.Get("DOORMAN_USERNAME", ""),
		Pass:    config.Get("DOORMAN_PASSWORD", ""),
		HTTP:    cli,
	}, nil
}

// Authenticate performs authentication with the doorman service
func (c *MYDoorman) Authenticate() error {
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
func (c *MYDoorman) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
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

// QueryPaymentEngine queries the Malaysia payment-engine database
func (c *MYDoorman) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-payment-engine-rds-mysql", "prd-payments-payment-engine-rds-mysql", "payment_engine", query)
}

// QueryPaymentCore queries the Malaysia payment-core database
func (c *MYDoorman) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-payment-core-rds-mysql", "prd-payments-payment-core-rds-mysql", "payment_core", query)
}

// QueryRppAdapter queries the Malaysia rpp-adapter database
func (c *MYDoorman) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
}

// QueryPartnerpayEngine queries the Malaysia partnerpay-engine database
func (c *MYDoorman) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-partnerpay-engine-rds-mysql", "prd-payments-partnerpay-engine-rds-mysql", "partnerpay_engine", query)
}

// QueryFastAdapter is not available in Malaysia environment
func (c *MYDoorman) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	return nil, errors.New("Fast adapter is not available in Malaysia environment")
}
