package txn

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"buddy/clients"
)

// ChargeRecord represents a record from the charge table
type ChargeRecord struct {
	ID                      int     `json:"id"`
	TransactionID           string  `json:"transaction_id"`
	Amount                  float64 `json:"amount"`
	Status                  string  `json:"status"`
	Currency                string  `json:"currency"`
	PartnerID               string  `json:"partner_id"`
	CustomerID              string  `json:"customer_id"`
	ExternalID              string  `json:"external_id"`
	ReferenceID             string  `json:"reference_id"`
	TxnDomain               string  `json:"txn_domain"`
	TxnType                 string  `json:"txn_type"`
	TxnSubtype              string  `json:"txn_subtype"`
	Remarks                 string  `json:"remarks"`
	Metadata                string  `json:"metadata"`
	Properties              string  `json:"properties"`
	ValuedAt                string  `json:"valued_at"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
	BillingToken            string  `json:"billing_token"`
	StatusReason            string  `json:"status_reason"`
	CaptureMethod           string  `json:"capture_method"`
	SourceAccount           string  `json:"source_account"`
	DestinationAccount      string  `json:"destination_account"`
	CapturedAmount          float64 `json:"captured_amount"`
	TransactionPayLoad      string  `json:"transaction_payload"`
	StatusReasonDescription string  `json:"status_reason_description"`
}

// ChargeStorage represents the ChargeStorage object for workflow_execution
type ChargeStorage struct {
	ID                      int     `json:"ID"`
	Amount                  float64 `json:"Amount"`
	Status                  string  `json:"Status"`
	Remarks                 string  `json:"Remarks"`
	TxnType                 string  `json:"TxnType"`
	Currency                string  `json:"Currency"`
	Metadata                *string `json:"Metadata"`
	ValuedAt                string  `json:"ValuedAt"`
	CreatedAt               string  `json:"CreatedAt"`
	PartnerID               string  `json:"PartnerID"`
	TxnDomain               string  `json:"TxnDomain"`
	UpdatedAt               string  `json:"UpdatedAt"`
	CustomerID              string  `json:"CustomerID"`
	ExternalID              string  `json:"ExternalID"`
	Properties              *string `json:"Properties"`
	TxnSubtype              string  `json:"TxnSubtype"`
	ReferenceID             string  `json:"ReferenceID"`
	BillingToken            string  `json:"BillingToken"`
	StatusReason            string  `json:"StatusReason"`
	CaptureMethod           string  `json:"CaptureMethod"`
	SourceAccount           *string `json:"SourceAccount"`
	TransactionID           string  `json:"TransactionID"`
	CapturedAmount          float64 `json:"CapturedAmount"`
	DestinationAccount      *string `json:"DestinationAccount"`
	TransactionPayLoad      *string `json:"TransactionPayLoad"`
	StatusReasonDescription string  `json:"StatusReasonDescription"`
}

// ProcessResult holds the result of processing a single transaction
type ProcessResult struct {
	TransactionID string
	Success       bool
	ValueAt       string
	Error         error
}

// EcoTxnPublisher handles SQL generation for ecosystem transactions
type EcoTxnPublisher struct {
	client      clients.DoormanInterface
	DeploySQL   strings.Builder
	RollbackSQL strings.Builder
}

// NewEcoTxnPublisher creates a new publisher instance
func NewEcoTxnPublisher() *EcoTxnPublisher {
	return &EcoTxnPublisher{
		client: clients.Doorman,
	}
}

// ProcessEcoTxnPublish processes a single transaction for publishing
func ProcessEcoTxnPublish(transactionID, env string) error {
	publisher := NewEcoTxnPublisher()
	result := publisher.processSingleTransaction(transactionID)
	if result.Success {
		return nil
	}
	return result.Error
}

// ProcessEcoTxnPublishBatch processes multiple transactions from a file
func ProcessEcoTxnPublishBatch(filePath, env string) {
	// Read transaction IDs from file
	transactionIDs, err := ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	publisher := NewEcoTxnPublisher()

	// Add headers
	publisher.DeploySQL.WriteString("-- PPE_Deploy.sql\n")
	publisher.DeploySQL.WriteString(fmt.Sprintf("-- Generated on: %s\n\n", time.Now().Format(time.RFC3339)))
	publisher.RollbackSQL.WriteString("-- PPE_Rollback.sql\n")
	publisher.RollbackSQL.WriteString(fmt.Sprintf("-- Generated on: %s\n\n", time.Now().Format(time.RFC3339)))

	processedCount := 0
	failedCount := 0

	// Process each transaction
	for _, txID := range transactionIDs {
		result := publisher.processSingleTransaction(txID)
		if result.Success {
			processedCount++
		} else {
			failedCount++
			fmt.Printf("Failed to process transaction ID: %s, Error: %v\n", txID, result.Error)
		}
	}

	fmt.Printf("Processing completed. Success: %d, Failed: %d\n", processedCount, failedCount)

	// Write output files
	if err := WriteEcoTxnSQLFiles(publisher.DeploySQL.String(), publisher.RollbackSQL.String(), filePath); err != nil {
		fmt.Printf("Error writing SQL files: %v\n", err)
		return
	}
}

// processSingleTransaction processes a single transaction ID
func (p *EcoTxnPublisher) processSingleTransaction(transactionID string) ProcessResult {
	result := ProcessResult{
		TransactionID: transactionID,
		Success:       false,
	}

	// Step 1: Query charge table for the transaction ID
	chargeRecord, err := p.queryChargeRecord(transactionID)
	if err != nil {
		result.Error = fmt.Errorf("failed to query charge record: %w", err)
		return result
	}

	// Step 2: Query sg-prd-m-payment-core internal_transaction table
	txID, err := p.queryInternalTransaction(transactionID, chargeRecord.CreatedAt)
	if err != nil {
		result.Error = fmt.Errorf("failed to query internal transaction: %w", err)
		return result
	}

	// Step 3: Query workflow_execution table for ValueTimestamp
	valueTimestamp, err := p.queryWorkflowExecution(txID)
	if err != nil {
		result.Error = fmt.Errorf("failed to query workflow execution: %w", err)
		return result
	}

	// Step 4: Generate UPDATE SQL for charge table
	p.generateChargeUpdateSQL(transactionID, chargeRecord.Status, valueTimestamp)

	// Step 5: Generate UPDATE SQL for workflow_execution table
	p.generateWorkflowUpdateSQL(transactionID, chargeRecord, valueTimestamp)

	result.Success = true
	result.ValueAt = valueTimestamp
	return result
}

// queryChargeRecord queries the charge table for a single transaction ID
func (p *EcoTxnPublisher) queryChargeRecord(transactionID string) (*ChargeRecord, error) {
	query := fmt.Sprintf("SELECT * FROM charge WHERE transaction_id = '%s'",
		strings.ReplaceAll(transactionID, "'", "''"))

	rows, err := p.client.QueryPartnerpayEngine(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no charge record found for transaction_id: %s", transactionID)
	}

	if len(rows) > 1 {
		return nil, fmt.Errorf("multiple charge records found for transaction_id: %s", transactionID)
	}

	row := rows[0]
	record := &ChargeRecord{
		ID:                      toInt(row["id"]),
		TransactionID:           toString(row["transaction_id"]),
		Amount:                  toFloat64(row["amount"]),
		Status:                  toString(row["status"]),
		Currency:                toString(row["currency"]),
		PartnerID:               toString(row["partner_id"]),
		CustomerID:              toString(row["customer_id"]),
		ExternalID:              toString(row["external_id"]),
		ReferenceID:             toString(row["reference_id"]),
		TxnDomain:               toString(row["txn_domain"]),
		TxnType:                 toString(row["txn_type"]),
		TxnSubtype:              toString(row["txn_subtype"]),
		Remarks:                 toString(row["remarks"]),
		Metadata:                toString(row["metadata"]),
		Properties:              toString(row["properties"]),
		ValuedAt:                toString(row["valued_at"]),
		CreatedAt:               toString(row["created_at"]),
		UpdatedAt:               toString(row["updated_at"]),
		BillingToken:            toString(row["billing_token"]),
		StatusReason:            toString(row["status_reason"]),
		CaptureMethod:           toString(row["capture_method"]),
		SourceAccount:           toString(row["source_account"]),
		DestinationAccount:      toString(row["destination_account"]),
		CapturedAmount:          toFloat64(row["captured_amount"]),
		TransactionPayLoad:      toString(row["transaction_payload"]),
		StatusReasonDescription: toString(row["status_reason_description"]),
	}

	return record, nil
}

// queryInternalTransaction queries sg-prd-m-payment-core internal_transaction table
func (p *EcoTxnPublisher) queryInternalTransaction(transactionID, chargeCreatedAt string) (string, error) {
	// Parse the charge created_at timestamp
	chargeTime, err := time.Parse("2006-01-02 15:04:05", chargeCreatedAt)
	if err != nil {
		// Try alternative format
		chargeTime, err = time.Parse(time.RFC3339, chargeCreatedAt)
		if err != nil {
			return "", fmt.Errorf("failed to parse charge created_at timestamp: %w", err)
		}
	}

	// Calculate 1-hour range
	startTime := chargeTime.Add(-1 * time.Hour)
	endTime := chargeTime.Add(1 * time.Hour)

	query := fmt.Sprintf(`
		SELECT tx_id
		FROM internal_transaction
		WHERE group_id = '%s'
		AND tx_type = 'CAPTURE'
		AND status = 'SUCCESS'
		AND created_at >= '%s'
		AND created_at <= '%s'
		LIMIT 1`,
		strings.ReplaceAll(transactionID, "'", "''"),
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	rows, err := p.client.QueryPaymentCore(query)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("no internal transaction found for group_id: %s", transactionID)
	}

	txID := toString(rows[0]["tx_id"])
	if txID == "" {
		return "", fmt.Errorf("empty tx_id found for group_id: %s", transactionID)
	}

	return txID, nil
}

// queryWorkflowExecution queries workflow_execution table for ValueTimestamp
func (p *EcoTxnPublisher) queryWorkflowExecution(txID string) (string, error) {
	query := fmt.Sprintf(`
		SELECT JSON_EXTRACT(data, '$.NotifyParams.ValueTimestamp') as ValuedAt
		FROM workflow_execution
		WHERE run_id = '%s'`,
		strings.ReplaceAll(txID, "'", "''"))

	rows, err := p.client.QueryPaymentCore(query)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("no workflow execution found for run_id: %s", txID)
	}

	valueTimestamp := toString(rows[0]["ValuedAt"])
	if valueTimestamp == "" {
		return "", fmt.Errorf("no ValueTimestamp found for run_id: %s", txID)
	}

	// Remove quotes if present
	valueTimestamp = strings.Trim(valueTimestamp, `"`)

	return valueTimestamp, nil
}

// generateChargeUpdateSQL generates UPDATE SQL for charge table
func (p *EcoTxnPublisher) generateChargeUpdateSQL(transactionID, originalStatus, valueAt string) {
	p.DeploySQL.WriteString(fmt.Sprintf(`-- Update charge table for transaction_id: %s
UPDATE charge
SET status = 'COMPLETED', valued_at = '%s'
WHERE transaction_id = '%s';

`,
		transactionID,
		strings.ReplaceAll(valueAt, "'", "''"),
		strings.ReplaceAll(transactionID, "'", "''")))

	// Generate corresponding rollback
	p.RollbackSQL.WriteString(fmt.Sprintf(`-- Rollback charge table for transaction_id: %s
UPDATE charge
SET status = '%s', valued_at = NULL
WHERE transaction_id = '%s';

`,
		transactionID,
		strings.ReplaceAll(originalStatus, "'", "''"),
		strings.ReplaceAll(transactionID, "'", "''")))
}

// generateWorkflowUpdateSQL generates UPDATE SQL for workflow_execution table
func (p *EcoTxnPublisher) generateWorkflowUpdateSQL(transactionID string, chargeRecord *ChargeRecord, valueAt string) {
	// Create updated ChargeStorage with the new ValuedAt
	chargeStorage := ChargeStorage{
		ID:                      chargeRecord.ID,
		Amount:                  chargeRecord.Amount,
		Status:                  "COMPLETED",
		Remarks:                 chargeRecord.Remarks,
		TxnType:                 chargeRecord.TxnType,
		Currency:                chargeRecord.Currency,
		Metadata:                stringPtr(chargeRecord.Metadata),
		ValuedAt:                valueAt,
		CreatedAt:               chargeRecord.CreatedAt,
		PartnerID:               chargeRecord.PartnerID,
		TxnDomain:               chargeRecord.TxnDomain,
		UpdatedAt:               time.Now().Format("2006-01-02 15:04:05"),
		CustomerID:              chargeRecord.CustomerID,
		ExternalID:              chargeRecord.ExternalID,
		Properties:              stringPtr(chargeRecord.Properties),
		TxnSubtype:              chargeRecord.TxnSubtype,
		ReferenceID:             chargeRecord.ReferenceID,
		BillingToken:            chargeRecord.BillingToken,
		StatusReason:            chargeRecord.StatusReason,
		CaptureMethod:           chargeRecord.CaptureMethod,
		SourceAccount:           stringPtr(chargeRecord.SourceAccount),
		TransactionID:           chargeRecord.TransactionID,
		CapturedAmount:          chargeRecord.CapturedAmount,
		DestinationAccount:      stringPtr(chargeRecord.DestinationAccount),
		TransactionPayLoad:      stringPtr(chargeRecord.TransactionPayLoad),
		StatusReasonDescription: chargeRecord.StatusReasonDescription,
	}

	// Build the JSON_OBJECT string
	chargeStorageSQL := buildChargeStorageJSONObject(chargeStorage)

	// Generate UPDATE statement for workflow_execution
	p.DeploySQL.WriteString(fmt.Sprintf(`-- Update workflow_execution for transaction_id: %s
UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', %s)
WHERE
    run_id = '%s';

`,
		transactionID,
		chargeStorageSQL,
		strings.ReplaceAll(transactionID, "'", "''")))

	// Generate corresponding rollback (restore original ChargeStorage)
	originalChargeStorage := ChargeStorage{
		ID:                      chargeRecord.ID,
		Amount:                  chargeRecord.Amount,
		Status:                  chargeRecord.Status,
		Remarks:                 chargeRecord.Remarks,
		TxnType:                 chargeRecord.TxnType,
		Currency:                chargeRecord.Currency,
		Metadata:                stringPtr(chargeRecord.Metadata),
		ValuedAt:                chargeRecord.ValuedAt,
		CreatedAt:               chargeRecord.CreatedAt,
		PartnerID:               chargeRecord.PartnerID,
		TxnDomain:               chargeRecord.TxnDomain,
		UpdatedAt:               chargeRecord.UpdatedAt,
		CustomerID:              chargeRecord.CustomerID,
		ExternalID:              chargeRecord.ExternalID,
		Properties:              stringPtr(chargeRecord.Properties),
		TxnSubtype:              chargeRecord.TxnSubtype,
		ReferenceID:             chargeRecord.ReferenceID,
		BillingToken:            chargeRecord.BillingToken,
		StatusReason:            chargeRecord.StatusReason,
		CaptureMethod:           chargeRecord.CaptureMethod,
		SourceAccount:           stringPtr(chargeRecord.SourceAccount),
		TransactionID:           chargeRecord.TransactionID,
		CapturedAmount:          chargeRecord.CapturedAmount,
		DestinationAccount:      stringPtr(chargeRecord.DestinationAccount),
		TransactionPayLoad:      stringPtr(chargeRecord.TransactionPayLoad),
		StatusReasonDescription: chargeRecord.StatusReasonDescription,
	}

	originalChargeStorageSQL := buildChargeStorageJSONObject(originalChargeStorage)

	p.RollbackSQL.WriteString(fmt.Sprintf(`-- Rollback workflow_execution for transaction_id: %s
UPDATE workflow_execution
SET
    data = JSON_SET(data,
            '$.ChargeStorage', %s)
WHERE
    run_id = '%s';

`,
		transactionID,
		originalChargeStorageSQL,
		strings.ReplaceAll(transactionID, "'", "''")))
}

// buildChargeStorageJSONObject constructs a MySQL JSON_OBJECT string for the ChargeStorage struct
func buildChargeStorageJSONObject(cs ChargeStorage) string {
	var sb strings.Builder
	sb.WriteString("JSON_OBJECT(\n")

	// Helper to write field with indentation
	writeField := func(key string, val string, isLast bool) {
		sb.WriteString(fmt.Sprintf("\t'%s', %s", key, val))
		if !isLast {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	// Fields in the correct order matching GrabTxn.md
	fields := []struct {
		Key string
		Val string
	}{
		{"ID", fmt.Sprintf("%d", cs.ID)},
		{"Amount", fmt.Sprintf("%d", int(cs.Amount))}, // Amount as integer
		{"Status", fmt.Sprintf("'%s'", escapeSQL(cs.Status))},
		{"Remarks", fmt.Sprintf("'%s'", escapeSQL(cs.Remarks))},
		{"TxnType", fmt.Sprintf("'%s'", escapeSQL(cs.TxnType))},
		{"Currency", fmt.Sprintf("'%s'", escapeSQL(cs.Currency))},
		{"Metadata", parseAndBuildNestedJSON(cs.Metadata)},
		{"ValuedAt", fmt.Sprintf("'%s'", escapeSQL(cs.ValuedAt))},
		{"CreatedAt", fmt.Sprintf("'%s'", escapeSQL(cs.CreatedAt))},
		{"PartnerID", fmt.Sprintf("'%s'", escapeSQL(cs.PartnerID))},
		{"TxnDomain", fmt.Sprintf("'%s'", escapeSQL(cs.TxnDomain))},
		{"UpdatedAt", fmt.Sprintf("'%s'", escapeSQL(cs.UpdatedAt))},
		{"CustomerID", fmt.Sprintf("'%s'", escapeSQL(cs.CustomerID))},
		{"ExternalID", fmt.Sprintf("'%s'", escapeSQL(cs.ExternalID))},
		{"Properties", parseAndBuildNestedJSON(cs.Properties)},
		{"TxnSubtype", fmt.Sprintf("'%s'", escapeSQL(cs.TxnSubtype))},
		{"ReferenceID", fmt.Sprintf("'%s'", escapeSQL(cs.ReferenceID))},
		{"BillingToken", fmt.Sprintf("'%s'", escapeSQL(cs.BillingToken))},
		{"StatusReason", fmt.Sprintf("'%s'", escapeSQL(cs.StatusReason))},
		{"CaptureMethod", fmt.Sprintf("'%s'", escapeSQL(cs.CaptureMethod))},
		{"SourceAccount", parseAndBuildNestedJSON(cs.SourceAccount)},
		{"TransactionID", fmt.Sprintf("'%s'", escapeSQL(cs.TransactionID))},
		{"CapturedAmount", fmt.Sprintf("%d", int(cs.CapturedAmount))}, // CapturedAmount as integer
		{"DestinationAccount", parseAndBuildNestedJSON(cs.DestinationAccount)},
		{"TransactionPayLoad", parseAndBuildNestedJSON(cs.TransactionPayLoad)},
		{"StatusReasonDescription", fmt.Sprintf("'%s'", escapeSQL(cs.StatusReasonDescription))},
	}

	for i, f := range fields {
		writeField(f.Key, f.Val, i == len(fields)-1)
	}

	sb.WriteString(")")
	return sb.String()
}

// parseAndBuildNestedJSON takes a pointer to a JSON string and returns a SQL JSON_OBJECT/JSON_ARRAY construction
func parseAndBuildNestedJSON(jsonStrPtr *string) string {
	if jsonStrPtr == nil || *jsonStrPtr == "" {
		return "NULL"
	}
	jsonStr := *jsonStrPtr

	// If it's just "null", return NULL
	if jsonStr == "null" {
		return "NULL"
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		// If it's not valid JSON, return it as a string
		return fmt.Sprintf("'%s'", escapeSQL(jsonStr))
	}

	return buildRecursiveJSONObject(data)
}

// buildRecursiveJSONObject recursively converts Go types to MySQL JSON construction strings
func buildRecursiveJSONObject(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return "JSON_OBJECT()"
		}
		var sb strings.Builder
		sb.WriteString("JSON_OBJECT(")

		// Sort keys for deterministic output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			valStr := buildRecursiveJSONObject(v[k])
			sb.WriteString(fmt.Sprintf("'%s', %s", escapeSQL(k), valStr))
			if i < len(keys)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(")")
		return sb.String()

	case []interface{}:
		if len(v) == 0 {
			return "JSON_ARRAY()"
		}
		var sb strings.Builder
		sb.WriteString("JSON_ARRAY(")
		for i, item := range v {
			sb.WriteString(buildRecursiveJSONObject(item))
			if i < len(v)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(")")
		return sb.String()

	case string:
		return fmt.Sprintf("'%s'", escapeSQL(v))

	case float64:
		// JSON numbers are float64 in Go. Check if it's actually an integer to print cleanly
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%f", v)

	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"

	case nil:
		return "NULL"

	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// escapeSQL escapes single quotes for SQL
func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// stringPtr returns a pointer to the string if not empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Helper functions for type conversion
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if f, ok := v.(float32); ok {
		return float64(f)
	}
	return 0
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	if i, ok := v.(int); ok {
		return i
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if s, ok := v.(string); ok {
		// Try to parse string as int
		var i int
		if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
			return i
		}
	}
	return 0
}

// WriteEcoTxnSQLFiles writes the generated SQL to files
func WriteEcoTxnSQLFiles(deploySQL, rollbackSQL, basePath string) error {
	// Write Deploy.sql
	deployPath := strings.TrimSuffix(basePath, ".txt") + "_Deploy.sql"
	if err := os.WriteFile(deployPath, []byte(deploySQL), 0644); err != nil {
		return fmt.Errorf("failed to write Deploy.sql: %v", err)
	}
	fmt.Printf("Deploy statements written to %s\n", deployPath)

	// Write Rollback.sql
	rollbackPath := strings.TrimSuffix(basePath, ".txt") + "_Rollback.sql"
	if err := os.WriteFile(rollbackPath, []byte(rollbackSQL), 0644); err != nil {
		return fmt.Errorf("failed to write Rollback.sql: %v", err)
	}
	fmt.Printf("Rollback statements written to %s\n", rollbackPath)

	return nil
}
