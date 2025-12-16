package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"buddy/internal/clients/doorman"
	"buddy/internal/txn/utils"
	internalutils "buddy/internal/utils"
)

// formatTimestampWithTwoDecimals formats an RFC3339 timestamp to have exactly 2 decimal places
func formatTimestampWithTwoDecimals(timestamp string) string {
	// Parse the timestamp
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// If parsing fails, return the original timestamp
		return timestamp
	}

	// Format with nanoseconds, then truncate to 2 decimal places
	nanoStr := fmt.Sprintf("%09d", t.Nanosecond())
	// Take first 2 digits for hundredths of a second
	hundredths := nanoStr[:2]

	// Build the new timestamp with exactly 2 decimal places
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%sZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
		hundredths)
}

// EcoTxnPublisher handles SQL generation for ecosystem transactions
type EcoTxnPublisher struct {
	client           doorman.DoormanInterface
	DeploySQL        strings.Builder
	RollbackSQL      strings.Builder
	OriginalWorkflow *WorkflowExecution // Stores the original workflow execution record for rollback
	OriginalCharge   *ChargeRecord      // Stores the original charge record for rollback
}

// NewEcoTxnPublisher creates a new publisher instance
func NewEcoTxnPublisher() *EcoTxnPublisher {
	return &EcoTxnPublisher{
		client: doorman.Doorman,
	}
}

// ProcessEcoTxnPublish processes a single transaction for publishing
func ProcessEcoTxnPublish(transactionID, env string) error {
	publisher := NewEcoTxnPublisher()

	// Add headers for single transaction
	publisher.DeploySQL.WriteString("-- PPE_Deploy.sql\n")
	publisher.DeploySQL.WriteString(fmt.Sprintf("-- Generated on: %s\n\n", time.Now().Format(time.RFC3339)))
	publisher.RollbackSQL.WriteString("-- PPE_Rollback.sql\n")
	publisher.RollbackSQL.WriteString(fmt.Sprintf("-- Generated on: %s\n\n", time.Now().Format(time.RFC3339)))

	result := publisher.processSingleTransaction(transactionID)
	if result.Success {
		// Write output files for single transaction
		if err := WriteEcoTxnSQLFiles(publisher.DeploySQL.String(), publisher.RollbackSQL.String(), transactionID); err != nil {
			return fmt.Errorf("failed to write SQL files: %v", err)
		}
		return nil
	}
	return result.Error
}

// ProcessEcoTxnPublishBatch processes multiple transactions from a file
func ProcessEcoTxnPublishBatch(filePath, env string) {
	// Read transaction IDs from file
	transactionIDs, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	publisher := NewEcoTxnPublisher()

	// Add headers for batch processing
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

	// Store original charge record for rollback
	p.OriginalCharge = chargeRecord

	// Step 2: Query sg-prd-m-payment-core internal_transaction table
	txID, err := p.queryInternalTransaction(transactionID, chargeRecord.CreatedAt)
	if err != nil {
		result.Error = fmt.Errorf("failed to query internal transaction: %w", err)
		return result
	}

	// Step 3: Query original workflow execution record and ValueTimestamp in one query
	originalWorkflow, valueTimestamp, err := p.queryOriginalWorkflow(txID)
	if err != nil {
		result.Error = fmt.Errorf("failed to query original workflow execution: %w", err)
		return result
	}
	p.OriginalWorkflow = originalWorkflow

	// Step 4: Generate UPDATE SQL for charge table
	p.generateChargeUpdateSQL(transactionID, valueTimestamp)

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

// queryOriginalWorkflow queries the original workflow execution record and extracts ValueTimestamp
func (p *EcoTxnPublisher) queryOriginalWorkflow(txID string) (*WorkflowExecution, string, error) {
	query := fmt.Sprintf(`
		SELECT run_id, state, attempt, data,
		       JSON_EXTRACT(data, '$.NotifyParams.ValueTimestamp') as ValuedAt
		FROM workflow_execution
		WHERE run_id = '%s'`,
		strings.ReplaceAll(txID, "'", "''"))

	rows, err := p.client.QueryPaymentCore(query)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute query: %w", err)
	}

	if len(rows) == 0 {
		return nil, "", fmt.Errorf("no workflow execution found for run_id: %s", txID)
	}

	row := rows[0]
	workflow := &WorkflowExecution{
		RunID:   toString(row["run_id"]),
		State:   toInt(row["state"]),
		Attempt: toInt(row["attempt"]),
		Data:    toString(row["data"]),
	}

	valueTimestamp := toString(row["ValuedAt"])
	if valueTimestamp == "" {
		return nil, "", fmt.Errorf("no ValueTimestamp found for run_id: %s", txID)
	}

	// Remove quotes if present and format
	valueTimestamp = strings.Trim(valueTimestamp, `"`)
	valueTimestamp = formatTimestampWithTwoDecimals(valueTimestamp)

	return workflow, valueTimestamp, nil
}

// generateChargeUpdateSQL generates UPDATE SQL for charge table
func (p *EcoTxnPublisher) generateChargeUpdateSQL(transactionID, valueAt string) {
	// Format timestamp with exactly 2 decimal places
	formattedValueAt := formatTimestampWithTwoDecimals(valueAt)

	// Generate deploy SQL - update charge to COMPLETED status
	p.DeploySQL.WriteString(fmt.Sprintf(`-- Update charge table for transaction_id: %s
UPDATE charge
SET valued_at = '%s', updated_at = '%s'
WHERE transaction_id = '%s';

`,
		transactionID,
		strings.ReplaceAll(formattedValueAt, "'", "''"),
		strings.ReplaceAll(p.OriginalCharge.UpdatedAt, "'", "''"),
		strings.ReplaceAll(transactionID, "'", "''")))

	// Generate rollback SQL - restore original status and valued_at
	if p.OriginalCharge != nil {
		// For rollback, set valued_at to default timestamp as per GrabTxn.md
		rollbackValuedAt := "0000-00-00T00:00:00.00Z"
		if p.OriginalCharge.ValuedAt != "" && p.OriginalCharge.ValuedAt != "0000-00-00 00:00:00" {
			rollbackValuedAt = p.OriginalCharge.ValuedAt
		}

		p.RollbackSQL.WriteString(fmt.Sprintf(`-- Rollback charge table for transaction_id: %s
UPDATE charge
SET status = '%s', valued_at = '%s', updated_at = '%s'
WHERE transaction_id = '%s';

`,
			transactionID,
			strings.ReplaceAll(p.OriginalCharge.Status, "'", "''"),
			strings.ReplaceAll(rollbackValuedAt, "'", "''"),
			strings.ReplaceAll(p.OriginalCharge.UpdatedAt, "'", "''"),
			strings.ReplaceAll(transactionID, "'", "''")))
	}
}

// generateWorkflowUpdateSQL generates UPDATE SQL for workflow_execution table
func (p *EcoTxnPublisher) generateWorkflowUpdateSQL(transactionID string, chargeRecord *ChargeRecord, valueAt string) {
	// Format timestamp with exactly 2 decimal places
	formattedValueAt := formatTimestampWithTwoDecimals(valueAt)

	// Extract original ChargeStorage from the stored workflow record
	var originalChargeStorageJSON string
	if p.OriginalWorkflow != nil {
		// Parse the data JSON to extract ChargeStorage
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(p.OriginalWorkflow.Data), &data); err == nil {
			if chargeStorage, exists := data["ChargeStorage"]; exists {
				if csBytes, err := json.Marshal(chargeStorage); err == nil {
					originalChargeStorageJSON = string(csBytes)
				}
			}
		}
	}

	// Create updated ChargeStorage with the new ValuedAt
	chargeStorage := ChargeStorage{
		ID:                      chargeRecord.ID,
		Amount:                  chargeRecord.Amount,
		Status:                  chargeRecord.Status,
		Remarks:                 chargeRecord.Remarks,
		TxnType:                 chargeRecord.TxnType,
		Currency:                chargeRecord.Currency,
		Metadata:                stringPtr(chargeRecord.Metadata),
		ValuedAt:                formattedValueAt,
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

	// Build the JSON_OBJECT string
	chargeStorageSQL := buildChargeStorageJSONObject(chargeStorage)

	// Generate UPDATE statement for workflow_execution
	originalState := 0
	if p.OriginalWorkflow != nil {
		originalState = p.OriginalWorkflow.State
	}

	p.DeploySQL.WriteString(fmt.Sprintf(`-- Update workflow_execution for transaction_id: %s
UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', %s)
WHERE
    run_id = '%s'
    and state = %d;

`,
		transactionID,
		chargeStorageSQL,
		strings.ReplaceAll(transactionID, "'", "''"),
		originalState))

	// Generate corresponding rollback using ORIGINAL values
	if p.OriginalWorkflow != nil {
		// Get original state and attempt
		originalState := p.OriginalWorkflow.State
		originalAttempt := p.OriginalWorkflow.Attempt

		if originalChargeStorageJSON != "" {
			// Convert the JSON string to MySQL JSON_OBJECT format using mysqljson package
			originalChargeStorageSQL, err := internalutils.ToMySQLJSONObjectExpr(originalChargeStorageJSON)
			if err != nil {
				fmt.Printf("Warning: Could not convert original ChargeStorage to JSON_OBJECT: %v\n", err)
				// Fallback: use raw JSON (though this might not be ideal)
				originalChargeStorageSQL = fmt.Sprintf("'%s'", escapeSQL(originalChargeStorageJSON))
			}

			p.RollbackSQL.WriteString(fmt.Sprintf(`-- Rollback workflow_execution for transaction_id: %s
UPDATE workflow_execution
SET
	state = %d,
	attempt = %d,
    data = JSON_SET(data,
		'$.State', %d,
		'$.ChargeStorage', %s)
WHERE
    run_id = '%s';

`,
				transactionID,
				originalState,
				originalAttempt,
				originalState,
				originalChargeStorageSQL,
				strings.ReplaceAll(transactionID, "'", "''")))
		} else {
			// Fallback: restore only state and attempt if original ChargeStorage not available
			p.RollbackSQL.WriteString(fmt.Sprintf(`-- Rollback workflow_execution for transaction_id: %s
UPDATE workflow_execution
SET
	state = %d,
	attempt = %d
WHERE
    run_id = '%s';

`,
				transactionID,
				originalState,
				originalAttempt,
				strings.ReplaceAll(transactionID, "'", "''")))
		}
	}
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

	// Try to use the mysqljson package first
	if expr, err := internalutils.ToMySQLJSONObjectExpr(jsonStr); err == nil {
		return expr
	}

	// If it's not a JSON object (e.g., it's a primitive value), fall back to simple escaping
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If it's not valid JSON, return it as a string
		return fmt.Sprintf("'%s'", escapeSQL(jsonStr))
	}

	// For non-object values, we need to handle them differently
	switch v := data.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeSQL(v))
	case float64:
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
		// For other types, try to convert back to JSON and use as string
		if jsonBytes, err := json.Marshal(data); err == nil {
			return fmt.Sprintf("'%s'", escapeSQL(string(jsonBytes)))
		}
		return "NULL"
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
	// Write PPE_Deploy.sql
	deployPath := "PPE_Deploy.sql"
	if err := os.WriteFile(deployPath, []byte(deploySQL), 0644); err != nil {
		return fmt.Errorf("failed to write PPE_Deploy.sql: %v", err)
	}
	fmt.Printf("Deploy statements written to %s\n", deployPath)

	// Write PPE_Rollback.sql
	rollbackPath := "PPE_Rollback.sql"
	if err := os.WriteFile(rollbackPath, []byte(rollbackSQL), 0644); err != nil {
		return fmt.Errorf("failed to write PPE_Rollback.sql: %v", err)
	}
	fmt.Printf("Rollback statements written to %s\n", rollbackPath)

	return nil
}
