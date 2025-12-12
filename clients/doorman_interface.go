package clients

// DoormanInterface defines the contract for database query operations
// This interface provides a clean abstraction for environment-specific doorman clients
type DoormanInterface interface {
	// Core authentication and query methods
	Authenticate() error
	ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error)

	// Environment-specific query methods
	// Note: Not all environments support all services - implementations should return clear errors for unsupported services

	// Payment Engine queries - available in both Singapore and Malaysia
	QueryPaymentEngine(query string) ([]map[string]interface{}, error)

	// Payment Core queries - available in both Singapore and Malaysia
	QueryPaymentCore(query string) ([]map[string]interface{}, error)

	// Fast Adapter queries - Singapore only
	QueryFastAdapter(query string) ([]map[string]interface{}, error)

	// RPP Adapter queries - Malaysia only
	QueryRppAdapter(query string) ([]map[string]interface{}, error)

	// Partnerpay Engine queries - Malaysia only
	QueryPartnerpayEngine(query string) ([]map[string]interface{}, error)

}
