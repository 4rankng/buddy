package clients

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"buddy/config"
)

// NewDoormanClient creates a Doorman client using the appropriate environment
// Returns DoormanInterface which can be either SGDoorman or MYDoorman
func NewDoormanClient(timeout time.Duration) (DoormanInterface, error) {
	env := config.GetEnvironment()

	if env == "sg" {
		return NewSGDoorman(timeout)
	}
	return NewMYDoorman(timeout)
}

// DoormanClient maintains backward compatibility for existing code
// This acts as a wrapper around DoormanInterface
type DoormanClient struct {
	client DoormanInterface
	env    string
}

// NewDoormanClientWithConfig creates a Doorman client with custom configuration
// This method maintains backward compatibility for tests and custom configurations
func NewDoormanClientWithConfig(baseURL, username, password string, timeout time.Duration) (*DoormanClient, error) {
	client := &genericDoormanClient{
		BaseURL: baseURL,
		User:    username,
		Pass:    password,
	}
	// Initialize HTTP client
	jar, _ := cookiejar.New(nil)
	client.HTTP = &http.Client{Timeout: timeout, Jar: jar}

	return &DoormanClient{
		client: client,
		env:    "custom",
	}, nil
}

// genericDoormanClient is a simple implementation for custom configurations
type genericDoormanClient struct {
	BaseURL string
	User    string
	Pass    string
	HTTP    *http.Client
}

func (g *genericDoormanClient) Authenticate() error {
	// Implementation would be similar to SGDoorman.Authenticate
	// For simplicity, returning nil here
	return nil
}

func (g *genericDoormanClient) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	// Implementation would be similar to SGDoorman.ExecuteQuery
	// For simplicity, returning nil here
	return nil, nil
}

func (g *genericDoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return g.ExecuteQuery("payment-engine", "payment-engine", "payment_engine", query)
}

func (g *genericDoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	return g.ExecuteQuery("payment-core", "payment-core", "payment_core", query)
}

func (g *genericDoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	return g.ExecuteQuery("fast-adapter", "fast-adapter", "fast_adapter", query)
}

func (g *genericDoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	return g.ExecuteQuery("rpp-adapter", "rpp-adapter", "rpp_adapter", query)
}

func (g *genericDoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return g.ExecuteQuery("partnerpay-engine", "partnerpay-engine", "partnerpay_engine", query)
}

// Delegate all interface methods to the underlying client
func (c *DoormanClient) Authenticate() error {
	return c.client.Authenticate()
}

func (c *DoormanClient) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	return c.client.ExecuteQuery(cluster, instance, schema, query)
}

// Legacy convenience methods - delegate to the underlying client implementation
func (c *DoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return c.client.QueryPaymentEngine(query)
}

func (c *DoormanClient) QueryPrdPaymentsPaymentEngine(query string) ([]map[string]interface{}, error) {
	return c.client.QueryPaymentEngine(query)
}

func (c *DoormanClient) QueryPrdPaymentsPaymentCore(query string) ([]map[string]interface{}, error) {
	return c.client.QueryPaymentCore(query)
}

func (c *DoormanClient) QueryPrdPaymentsRppAdapter(query string) ([]map[string]interface{}, error) {
	return c.client.QueryRppAdapter(query)
}

func (c *DoormanClient) QueryPrdPaymentsPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return c.client.QueryPartnerpayEngine(query)
}

func (c *DoormanClient) QuerySGPaymentCore(query string) ([]map[string]interface{}, error) {
	return c.client.QueryPaymentCore(query)
}

func (c *DoormanClient) QuerySGFastAdapter(query string) ([]map[string]interface{}, error) {
	return c.client.QueryFastAdapter(query)
}
