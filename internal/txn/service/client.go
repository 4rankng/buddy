package service

import (
	"buddy/internal/clients/doorman"
)

// DoormanClient implements the ports.ClientPort interface
type DoormanClient struct {
	client doorman.DoormanInterface
}

// NewDoormanClient creates a new DoormanClient
func NewDoormanClient() *DoormanClient {
	return &DoormanClient{
		client: doorman.Doorman,
	}
}

func (d *DoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return d.client.QueryPaymentEngine(query)
}

func (d *DoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	return d.client.QueryPaymentCore(query)
}

func (d *DoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	return d.client.QueryRppAdapter(query)
}

func (d *DoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	return d.client.QueryFastAdapter(query)
}

func (d *DoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return d.client.QueryPartnerpayEngine(query)
}

func (d *DoormanClient) ExecuteQuery(cluster, service, database, query string) ([]map[string]interface{}, error) {
	return d.client.ExecuteQuery(cluster, service, database, query)
}
