package adapters

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"buddy/internal/clients/doorman"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDoormanClient is a mock implementation of doorman.DoormanInterface
type MockDoormanClient struct {
	responses map[string][]map[string]interface{}
	errors    map[string]error
}

func NewMockDoormanClient() *MockDoormanClient {
	return &MockDoormanClient{
		responses: make(map[string][]map[string]interface{}),
		errors:    make(map[string]error),
	}
}

func (m *MockDoormanClient) SetResponse(query string, response []map[string]interface{}) {
	m.responses[query] = response
}

func (m *MockDoormanClient) SetError(query string, err error) {
	m.errors[query] = err
}

func (m *MockDoormanClient) Authenticate() error {
	return nil
}

func (m *MockDoormanClient) ExecuteQuery(cluster, instance, schema, query string) ([]map[string]interface{}, error) {
	if err, exists := m.errors[query]; exists {
		return nil, err
	}
	if resp, exists := m.responses[query]; exists {
		return resp, nil
	}
	return []map[string]interface{}{}, nil
}

func (m *MockDoormanClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	if err, exists := m.errors[query]; exists {
		return nil, err
	}
	if resp, exists := m.responses[query]; exists {
		return resp, nil
	}
	return []map[string]interface{}{}, nil
}

func (m *MockDoormanClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	if err, exists := m.errors[query]; exists {
		return nil, err
	}
	if resp, exists := m.responses[query]; exists {
		return resp, nil
	}
	// Try to match partial queries (contains check)
	for q, r := range m.responses {
		if strings.Contains(query, q) {
			return r, nil
		}
	}
	// Return mock data for any query containing group_id
	if strings.Contains(query, "group_id = 'test-timestamp-123'") {
		return []map[string]interface{}{
			{"tx_id": "tx-timestamp-123"},
		}, nil
	}
	if strings.Contains(query, "group_id = '7eba1b67c9174d21bb66bb089ebd6fd3'") {
		return []map[string]interface{}{
			{"tx_id": "tx-internal-123"},
		}, nil
	}
	// Return mock data for workflow queries
	if strings.Contains(query, "run_id = 'tx-timestamp-123'") {
		return []map[string]interface{}{
			{
				"run_id":   "tx-timestamp-123",
				"state":    200,
				"attempt":  0,
				"data":     json.RawMessage(`{"State": 200, "ValueTimestamp": "2025-10-24T15:30:01Z", "ChargeStorage": {"ID": 123, "Status": "PENDING"}}`),
				"ValuedAt": "2025-10-24T15:30:01Z",
			},
		}, nil
	}
	if strings.Contains(query, "run_id = 'tx-internal-123'") {
		return []map[string]interface{}{
			{
				"run_id":   "tx-internal-123",
				"state":    300,
				"attempt":  0,
				"data":     json.RawMessage(`{"State": 300, "ValueTimestamp": "2025-10-24T15:30:01Z", "ChargeStorage": {"ID": 786874, "Status": "PENDING"}}`),
				"ValuedAt": "2025-10-24T15:30:01Z",
			},
		}, nil
	}
	return []map[string]interface{}{}, nil
}

func (m *MockDoormanClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	if err, exists := m.errors[query]; exists {
		return nil, err
	}
	if resp, exists := m.responses[query]; exists {
		return resp, nil
	}
	return []map[string]interface{}{}, nil
}

func (m *MockDoormanClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	if err, exists := m.errors[query]; exists {
		return nil, err
	}
	if resp, exists := m.responses[query]; exists {
		return resp, nil
	}
	return []map[string]interface{}{}, nil
}

func (m *MockDoormanClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	if err, exists := m.errors[query]; exists {
		return nil, err
	}
	if resp, exists := m.responses[query]; exists {
		return resp, nil
	}
	// Try to match partial queries (contains check)
	for q, r := range m.responses {
		if strings.Contains(query, q) {
			return r, nil
		}
	}
	return []map[string]interface{}{}, nil
}

func TestEcoTxn_ChargeUpdateSQLTimestampPreservation(t *testing.T) {
	// Test that both deploy and rollback SQL preserve the updated_at field

	// Create a mock client
	originalClient := doorman.Doorman
	defer func() {
		doorman.Doorman = originalClient
	}()

	mockClient := NewMockDoormanClient()
	doorman.Doorman = mockClient

	// Charge record with specific timestamps
	chargeResponse := []map[string]interface{}{
		{
			"id":             123,
			"transaction_id": "test-timestamp-123",
			"amount":         100.0,
			"status":         "PENDING",
			"currency":       "SGD",
			"created_at":     "2023-12-15T10:00:00.00Z",
			"updated_at":     "2023-12-16T07:06:07.00Z", // Specific timestamp to preserve
			"valued_at":      "0000-00-00T00:00:00.00Z",
			// Include other required fields with empty values
			"partner_id":                "",
			"customer_id":               "",
			"external_id":               "",
			"reference_id":              "",
			"txn_domain":                "",
			"txn_type":                  "",
			"txn_subtype":               "",
			"remarks":                   "",
			"metadata":                  "",
			"properties":                "",
			"billing_token":             "",
			"status_reason":             "",
			"capture_method":            "",
			"source_account":            "",
			"captured_amount":           0.0,
			"destination_account":       "",
			"transaction_payload":       "",
			"status_reason_description": "",
		},
	}

	mockClient.SetResponse("SELECT * FROM charge WHERE transaction_id = 'test-timestamp-123'", chargeResponse)

	// Process the transaction using the public API
	err := ProcessEcoTxnPublish("test-timestamp-123", "test")
	require.NoError(t, err)

	// Read the generated SQL files
	deploySQL, err := os.ReadFile("PPE_Deploy.sql")
	require.NoError(t, err)
	rollbackSQL, err := os.ReadFile("PPE_Rollback.sql")
	require.NoError(t, err)

	// Clean up generated files
	_ = os.Remove("PPE_Deploy.sql")
	_ = os.Remove("PPE_Rollback.sql")

	// Verify deploy SQL includes updated_at preservation
	assert.Contains(t, string(deploySQL), "updated_at = '2023-12-16T07:06:07.00Z'",
		"Deploy SQL should preserve original updated_at timestamp")
	assert.Contains(t, string(deploySQL), "SET valued_at = '2025-10-24T15:30:01.00Z', updated_at = '2023-12-16T07:06:07.00Z'",
		"Deploy SQL should include valued_at and updated_at fields in correct order")

	// Verify rollback SQL includes updated_at preservation
	assert.Contains(t, string(rollbackSQL), "updated_at = '2023-12-16T07:06:07.00Z'",
		"Rollback SQL should preserve original updated_at timestamp")
	assert.Contains(t, string(rollbackSQL), "SET status = 'PENDING', valued_at = '0000-00-00T00:00:00.00Z', updated_at = '2023-12-16T07:06:07.00Z'",
		"Rollback SQL should include all three fields in correct order")
}

func TestEcoTxn_FullIntegrationTest(t *testing.T) {
	// Create a mock client
	originalClient := doorman.Doorman
	defer func() {
		doorman.Doorman = originalClient
	}()

	mockClient := NewMockDoormanClient()
	doorman.Doorman = mockClient

	// Set up mock responses
	chargeResponse := []map[string]interface{}{
		{
			"id":                        786874,
			"transaction_id":            "7eba1b67c9174d21bb66bb089ebd6fd3",
			"amount":                    200.0,
			"status":                    "PENDING",
			"currency":                  "SGD",
			"partner_id":                "28a61200-7d6d-4419-947a-173ff81cf7db",
			"customer_id":               "0948a27e-bd0d-4aff-8071-0bf3fc9469fd",
			"external_id":               "be618a3ad75648a79ae8be1ee4fa0d43",
			"reference_id":              "be618a3ad75648a79ae8be1ee4fa0d43",
			"txn_domain":                "DEPOSITS",
			"txn_type":                  "SPEND_MONEY",
			"txn_subtype":               "GRAB",
			"remarks":                   "",
			"metadata":                  `{"featureCode": "A-8NCF3OMGWQPOD9", "service": "Goblin"}`,
			"properties":                `{"AuthorisationID": "abc654828dc047bca0898989a574a41a", "NotificationFlags": {"Email": 1, "Push": 1, "Sms": 0}, "VerdictID": 4032259}`,
			"valued_at":                 "0000-00-00 00:00:00",
			"created_at":                "2025-12-10T02:02:51.00Z",
			"updated_at":                "2025-12-16T07:06:07.00Z",
			"billing_token":             "1e0b4d582c204a55af42f6ceb84a0d73",
			"status_reason":             "",
			"capture_method":            "AUTOMATIC",
			"source_account":            `{"DisplayName": "", "Number": "8880261519"}`,
			"captured_amount":           200.0,
			"destination_account":       `{"DisplayName": "", "Number": "209421001"}`,
			"transaction_payload":       `{"ActivityID": "A-8NCF3OMGWQPOD9", "ActivityType": "DEFAULT"}`,
			"status_reason_description": "",
		},
	}

	mockClient.SetResponse("SELECT * FROM charge WHERE transaction_id = '7eba1b67c9174d21bb66bb089ebd6fd3'", chargeResponse)

	// Process the transaction using the public API
	err := ProcessEcoTxnPublish("7eba1b67c9174d21bb66bb089ebd6fd3", "test")
	require.NoError(t, err)

	// Read the generated SQL files
	deploySQL, err := os.ReadFile("PPE_Deploy.sql")
	require.NoError(t, err)
	rollbackSQL, err := os.ReadFile("PPE_Rollback.sql")
	require.NoError(t, err)

	// Clean up generated files
	_ = os.Remove("PPE_Deploy.sql")
	_ = os.Remove("PPE_Rollback.sql")

	// Verify deploy SQL contains expected elements
	deployStr := string(deploySQL)
	assert.Contains(t, deployStr, "UPDATE charge")
	assert.Contains(t, deployStr, "valued_at = '2025-10-24T15:30:01.00Z'")  // This should be from payment-core ValueTimestamp
	assert.Contains(t, deployStr, "updated_at = '2025-12-16T07:06:07.00Z'") // This should be preserved from charge
	assert.Contains(t, deployStr, "UPDATE workflow_execution")
	assert.Contains(t, deployStr, "state = 800")
	assert.Contains(t, deployStr, "'UpdatedAt', '2025-12-16T07:06:07.00Z'") // Preserved from charge record

	// Verify rollback SQL contains expected elements
	rollbackStr := string(rollbackSQL)
	assert.Contains(t, rollbackStr, "UPDATE charge")
	assert.Contains(t, rollbackStr, "valued_at = '0000-00-00T00:00:00.00Z'")
	assert.Contains(t, rollbackStr, "updated_at = '2025-12-16T07:06:07.00Z'") // Preserved from original
	assert.Contains(t, rollbackStr, "state = 300")
	assert.Contains(t, rollbackStr, "attempt = 0")
}
