package service

import (
	"buddy/internal/clients/doorman"
	"fmt"
)

// DoormanClient implements the ports.ClientPort interface
type DoormanClient struct {
	client doorman.DoormanInterface
}

// NewDoormanClient creates a new DoormanClient
// It ensures the global doorman client is initialized before creating a wrapper
func NewDoormanClient() *DoormanClient {
	// Ensure the global doorman client is initialized
	if doorman.Doorman == nil {
		// Initialize with Malaysia environment as default if not already done
		_ = doorman.NewDoormanClient("my") // If initialization fails, it will be caught when QueryPaymentEngine is called
	}

	return &DoormanClient{
		client: doorman.Doorman,
	}
}

func (d *DoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is nil, was it properly initialized?")
	}
	return d.client.QueryPaymentEngine(query)
}

func (d *DoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is nil, was it properly initialized?")
	}
	return d.client.QueryPaymentCore(query)
}

func (d *DoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is nil, was it properly initialized?")
	}
	return d.client.QueryRppAdapter(query)
}

func (d *DoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is nil, was it properly initialized?")
	}
	return d.client.QueryFastAdapter(query)
}

func (d *DoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is nil, was it properly initialized?")
	}
	return d.client.QueryPartnerpayEngine(query)
}

func (d *DoormanClient) ExecuteQuery(cluster, service, database, query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is nil, was it properly initialized?")
	}
	return d.client.ExecuteQuery(cluster, service, database, query)
}
