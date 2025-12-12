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

type DoormanClient struct {
	BaseURL string
	User    string
	Pass    string
	HTTP    *http.Client
}

type doormanLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type queryPayload struct {
	ClusterName  string `json:"clusterName"`
	InstanceName string `json:"instanceName"`
	Schema       string `json:"schema"`
	Query        string `json:"query"`
}

type queryResponse struct {
	Code   int `json:"code"`
	Result struct {
		Headers []string        `json:"headers"`
		Rows    [][]interface{} `json:"rows"`
	} `json:"result"`
}

func NewDoormanClient(timeout time.Duration) (*DoormanClient, error) {
	base := config.Get("DOORMAN_BASE_URL", "https://doorman.infra.prd.g-bank.app")
	user := config.Get("DOORMAN_USERNAME", "")
	pass := config.Get("DOORMAN_PASSWORD", "")
	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Timeout: timeout, Jar: jar}
	return &DoormanClient{BaseURL: base, User: user, Pass: pass, HTTP: cli}, nil
}

func NewDoormanClientWithConfig(baseURL, username, password string, timeout time.Duration) (*DoormanClient, error) {
	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Timeout: timeout, Jar: jar}
	return &DoormanClient{BaseURL: baseURL, User: username, Pass: password, HTTP: cli}, nil
}

func (c *DoormanClient) Authenticate() error {
	loginURL, _ := url.JoinPath(c.BaseURL, "/api/login/ldap/signin")
	b, _ := json.Marshal(doormanLoginReq{Username: c.User, Password: c.Pass})
	req, _ := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(b))
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

func (c *DoormanClient) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}
	qURL, _ := url.JoinPath(c.BaseURL, "/api/rds/query/execute")
	payload := queryPayload{ClusterName: cluster, InstanceName: instance, Schema: schema, Query: query}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, qURL, bytes.NewReader(b))
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
			// best-effort read; body may be small
			bb, _ := io.ReadAll(resp.Body)
			bodyBytes = bb
		}
		output.LogEvent("doorman_query_error", map[string]any{
			"status":  resp.Status,
			"cluster": cluster,
			"schema":  schema,
			"payload": string(b),
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

// Convenience wrappers matching Python service names
func (c *DoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("sg-prd-m-payment-engine", "sg-prd-m-payment-engine", "prod_payment_engine_db01", query)
}

func (c *DoormanClient) QueryPrdPaymentsPaymentEngine(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-payment-engine-rds-mysql", "prd-payments-payment-engine-rds-mysql", "payment_engine", query)
}

func (c *DoormanClient) QueryPrdPaymentsPaymentCore(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-payment-core-rds-mysql", "prd-payments-payment-core-rds-mysql", "payment_core", query)
}

func (c *DoormanClient) QueryPrdPaymentsRppAdapter(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
}

func (c *DoormanClient) QueryPrdPaymentsPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("prd-payments-partnerpay-engine-rds-mysql", "prd-payments-partnerpay-engine-rds-mysql", "partnerpay_engine", query)
}

func (c *DoormanClient) QuerySGPaymentCore(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("sg-prd-m-payment-core", "sg-prd-m-payment-core", "payment_core", query)
}

func (c *DoormanClient) QuerySGFastAdapter(query string) ([]map[string]interface{}, error) {
	return c.ExecuteQuery("sg-prd-m-fast-adapter", "sg-prd-m-fast-adapter", "fast_adapter", query)
}
