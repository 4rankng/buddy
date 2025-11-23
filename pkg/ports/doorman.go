package ports

// DoormanPort defines the interface for SQL query execution and database operations
type DoormanPort interface {
	// ExecuteQuery executes a read-only SQL query on a specific cluster/instance/schema
	ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error)

	// Health check
	HealthCheck() error

	// Database information
	GetAvailableClusters() ([]DatabaseCluster, error)
}

// DatabaseCluster represents an available database cluster
type DatabaseCluster struct {
	Name        string `json:"name"`
	Instance    string `json:"instance"`
	Schema      string `json:"schema"`
	Description string `json:"description"`
	Environment string `json:"environment"`
}

// DMLRequest represents a DML (Data Manipulation Language) request
type DMLRequest struct {
	ClusterName  string `json:"cluster_name"`
	InstanceName string `json:"instance_name"`
	Schema       string `json:"schema"`
	Query        string `json:"query"`
	Description  string `json:"description"`
	RequestedBy  string `json:"requested_by"`
}