package service

import (
	"buddy/internal/clients/doorman"
	"fmt"
)

// DoormanClient implements the ports.ClientPort interface
type DoormanClient struct {
	client doorman.DoormanInterface
}

func (d *DoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is not initialized - check environment configuration")
	}
	return d.client.QueryPaymentEngine(query)
}

func (d *DoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is not initialized - check environment configuration")
	}
	return d.client.QueryPaymentCore(query)
}

func (d *DoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is not initialized - check environment configuration")
	}
	return d.client.QueryRppAdapter(query)
}

func (d *DoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is not initialized - check environment configuration")
	}
	return d.client.QueryFastAdapter(query)
}

func (d *DoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is not initialized - check environment configuration")
	}
	return d.client.QueryPartnerpayEngine(query)
}

func (d *DoormanClient) ExecuteQuery(cluster, service, database, query string) ([]map[string]interface{}, error) {
	if d.client == nil {
		return nil, fmt.Errorf("doorman client is not initialized - check environment configuration")
	}
	return d.client.ExecuteQuery(cluster, service, database, query)
}
