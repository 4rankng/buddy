package ports

import "time"

// DoormanPort defines the interface for SQL query execution and database operations
type DoormanPort interface {
	// ExecuteQuery executes a SQL query on a specific cluster/instance/schema
	ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error)

	// Payment Engine specific query methods
	QueryPaymentEngine(query string) ([]map[string]interface{}, error)
	QueryPaymentCore(query string) ([]map[string]interface{}, error)
	QueryPartnerPayEngine(query string) ([]map[string]interface{}, error)
	QueryPairingService(query string) ([]map[string]interface{}, error)
	QueryTransactionLimit(query string) ([]map[string]interface{}, error)

	// Health check
	HealthCheck() error
}

// QueryResult represents the result of a SQL query
type QueryResult struct {
	Headers []string                 `json:"headers"`
	Rows    [][]interface{}          `json:"rows"`
	Meta    map[string]interface{}   `json:"meta,omitempty"`
}

// ClusterInfo contains information about available database clusters
type ClusterInfo struct {
	Name        string `json:"name"`
	Instance    string `json:"instance"`
	Schema      string `json:"schema"`
	Description string `json:"description"`
}

// DMLRequest represents a DML (Data Manipulation Language) request
type DMLRequest struct {
	ClusterName  string `json:"cluster_name"`
	InstanceName string `json:"instance_name"`
	Schema       string `json:"schema"`
	Query        string `json:"query"`
	Description  string `json:"description"`
}