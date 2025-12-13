build command

sgbuddy ecotxn publish abc // single transaction_id
or
sgbuddy ecotxn publish TSE-833.txt // file path contains multile transactions

PPE_Deploy.sql

-- Update charge table for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE charge
SET status = 'COMPLETED', valued_at = '2025-10-24T15:30:01.311411Z'
WHERE transaction_id = '1ed87447b552420790357c2d5abe5509';

-- Update workflow_execution for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT(
	'ID', 2489358,
	'Amount', 1300,  // amount in integer
	'Status', 'COMPLETED',
	'Remarks', '',
	'TxnType', 'SPEND_MONEY',
	'Currency', 'SGD',
	'Metadata', JSON_OBJECT('featureCode', 'A-8HDXTHVGW4VFAV', 'service', 'Transport'),
	'ValuedAt', '2025-10-24T15:30:01.311411Z',
	'CreatedAt', '2025-10-24T14:58:11Z',
	'PartnerID', 'b2da6c1e-b2e4-4162-82b7-ce43ebf8b211',
	'TxnDomain', 'DEPOSITS',
	'UpdatedAt', '2025-12-10 15:40:24',
	'CustomerID', '6f1c366d-8f21-4fbd-ae3e-d52edb32b754',
	'ExternalID', '80adc50d977a4519912781ef034987e8',
	'Properties', JSON_OBJECT('AuthorisationID', '81cdfddd213b48a9be168fea7368c999', 'CancelIdempotencyKey', '', 'CaptureIdempotencyKey', '80c97e874e044d69bf53f47925027cd1', 'NotificationFlags', JSON_OBJECT('Email', 0, 'Push', 0, 'Sms', 0), 'VerdictID', 37536250),
	'TxnSubtype', 'GRAB',
	'ReferenceID', '80adc50d977a4519912781ef034987e8',
	'BillingToken', '3bbee8f00440469295a34c57418cafd7',
	'StatusReason', '',
	'CaptureMethod', 'MANUAL',
	'SourceAccount', JSON_OBJECT('DisplayName', '', 'Number', '8885548902'),
	'TransactionID', '1ed87447b552420790357c2d5abe5509',
	'CapturedAmount', 1300, // in integer
	'DestinationAccount', JSON_OBJECT('DisplayName', '', 'Number', '209421001'),
	'TransactionPayLoad', JSON_OBJECT('ActivityID', 'A-8HDXTHVGW4VFAV', 'ActivityType', 'TRANSPORT'),
	'StatusReasonDescription', ''
))
WHERE
    run_id = '1ed87447b552420790357c2d5abe5509';

PPE_Rollback.sql


-- Rollback charge table for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE charge
SET status = 'COMPLETED', valued_at = NULL
WHERE transaction_id = '1ed87447b552420790357c2d5abe5509';

-- Rollback workflow_execution for transaction_id: 1ed87447b552420790357c2d5abe5509
UPDATE workflow_execution
SET
    data = JSON_SET(data,
            '$.ChargeStorage', JSON_OBJECT(
	'ID', 2489358,
	'Amount', 1300, // amount in integer
	'Status', 'COMPLETED',
	'Remarks', '',
	'TxnType', 'SPEND_MONEY',
	'Currency', 'SGD',
	'Metadata', JSON_OBJECT('featureCode', 'A-8HDXTHVGW4VFAV', 'service', 'Transport'),
	'ValuedAt', '2025-10-24T15:30:01Z',
	'CreatedAt', '2025-10-24T14:58:11Z',
	'PartnerID', 'b2da6c1e-b2e4-4162-82b7-ce43ebf8b211',
	'TxnDomain', 'DEPOSITS',
	'UpdatedAt', '2025-12-10T06:25:06Z',
	'CustomerID', '6f1c366d-8f21-4fbd-ae3e-d52edb32b754',
	'ExternalID', '80adc50d977a4519912781ef034987e8',
	'Properties', JSON_OBJECT('AuthorisationID', '81cdfddd213b48a9be168fea7368c999', 'CancelIdempotencyKey', '', 'CaptureIdempotencyKey', '80c97e874e044d69bf53f47925027cd1', 'NotificationFlags', JSON_OBJECT('Email', 0, 'Push', 0, 'Sms', 0), 'VerdictID', 37536250),
	'TxnSubtype', 'GRAB',
	'ReferenceID', '80adc50d977a4519912781ef034987e8',
	'BillingToken', '3bbee8f00440469295a34c57418cafd7',
	'StatusReason', '',
	'CaptureMethod', 'MANUAL',
	'SourceAccount', JSON_OBJECT('DisplayName', '', 'Number', '8885548902'),
	'TransactionID', '1ed87447b552420790357c2d5abe5509',
	'CapturedAmount', 1300.000000,
	'DestinationAccount', JSON_OBJECT('DisplayName', '', 'Number', '209421001'),
	'TransactionPayLoad', JSON_OBJECT('ActivityID', 'A-8HDXTHVGW4VFAV', 'ActivityType', 'TRANSPORT'),
	'StatusReasonDescription', ''
))
WHERE
    run_id = '1ed87447b552420790357c2d5abe5509';


Here is the code for reference    

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"oncall/oncall-cli/clients"
	"oncall/oncall-cli/config"
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

// TransactionProcessor holds the state for processing a single transaction
type TransactionProcessor struct {
	Client         *clients.DoormanClient
	DeploySQL      strings.Builder
	RollbackSQL    strings.Builder
	ProcessedCount int
	FailedCount    int
}

// ProcessResult holds the result of processing a single transaction
type ProcessResult struct {
	TransactionID string
	Success       bool
	ValueAt       string
	Error         error
}

func main() {
	// Parse command line flags
	dryRun := flag.Bool("dry-run", false, "Only generate SQL statements without executing them")
	outputDir := flag.String("output-dir", ".", "Output directory for SQL files (default: current directory)")
	flag.Parse()

	// Bootstrap configuration
	config.BootstrapViper()
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// List of run_ids to process
	runIDs := []string{
		"1ed87447b552420790357c2d5abe5509",
		"ef2282dcdf00458fa309d7a9442232d6",
		// "455266f8473c4e1595a91c3a4cfb7f9b",
		// "488ecd9edc1a4475ae2e6d45b48b16dd",
		// "7f3a6a0bf3f54b7395ede50c09bc1d9c",
		// "9ebfb54135b74a40a0028a781ff57e97",
		// "12f916a09aef4f0daae4ee4e595d8c86",
		// "53fab3a498ce4f81ad15b43f0e141812",
		// "55a2d73bfbaf45c2bd11f8c97e5f2249",
		// "c4e7cbea9fe44c83b81030523732040b",
		// "8420d992e0474383bcc3b03320e2bfb2",
		// "e9ed49b59e7e40f4a9e6527a8c24d4d6",
	}

	// Create doorman client
	client, err := clients.NewDoormanClient(30 * time.Second)
	if err != nil {
		log.Fatalf("Failed to create doorman client: %v", err)
	}

	// Initialize processor
	processor := &TransactionProcessor{
		Client:      client,
		DeploySQL:   strings.Builder{},
		RollbackSQL: strings.Builder{},
	}

	// Add header to Deploy.sql
	processor.DeploySQL.WriteString("-- Deploy Script\n")
	processor.DeploySQL.WriteString("-- Generated on: ")
	processor.DeploySQL.WriteString(time.Now().Format(time.RFC3339))
	processor.DeploySQL.WriteString("\n\n")
	processor.DeploySQL.WriteString("-- Processing each transaction ID individually\n")
	processor.DeploySQL.WriteString("-- Workflow: Charge -> Internal Transaction -> Workflow Execution -> Charge Update\n\n")

	// Add header to Rollback.sql
	processor.RollbackSQL.WriteString("-- Rollback Script\n")
	processor.RollbackSQL.WriteString("-- Generated on: ")
	processor.RollbackSQL.WriteString(time.Now().Format(time.RFC3339))
	processor.RollbackSQL.WriteString("\n\n")
	processor.RollbackSQL.WriteString("-- Rollback statements for individual transaction updates\n\n")

	// Process each transaction ID individually
	log.Printf("Processing %d transaction IDs individually...", len(runIDs))
	for _, transactionID := range runIDs {
		log.Printf("Processing transaction ID: %s", transactionID)
		result := processor.processSingleTransaction(transactionID)

		if result.Success {
			processor.ProcessedCount++
			log.Printf("Successfully processed transaction ID: %s", transactionID)
		} else {
			processor.FailedCount++
			log.Printf("Failed to process transaction ID: %s, Error: %v", transactionID, result.Error)
		}
	}

	log.Printf("Processing completed. Success: %d, Failed: %d", processor.ProcessedCount, processor.FailedCount)

	// Write Deploy.sql
	deployPath := fmt.Sprintf("%s/Deploy.sql", *outputDir)
	err = os.WriteFile(deployPath, []byte(processor.DeploySQL.String()), 0644)
	if err != nil {
		log.Fatalf("Failed to write Deploy.sql: %v", err)
	}
	log.Printf("Deploy statements written to %s", deployPath)

	// Write Rollback.sql
	rollbackPath := fmt.Sprintf("%s/Rollback.sql", *outputDir)
	err = os.WriteFile(rollbackPath, []byte(processor.RollbackSQL.String()), 0644)
	if err != nil {
		log.Fatalf("Failed to write Rollback.sql: %v", err)
	}
	log.Printf("Rollback statements written to %s", rollbackPath)

	if !*dryRun {
		log.Println("Executing updates...")
		// Execute the updates here if needed
		log.Println("Note: Actual execution not implemented in this script")
	}

	log.Println("Script completed successfully")
}

// Helper functions
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

// processSingleTransaction processes a single transaction ID through the complete workflow
func (tp *TransactionProcessor) processSingleTransaction(transactionID string) ProcessResult {
	result := ProcessResult{
		TransactionID: transactionID,
		Success:       false,
	}

	// Step 1: Query charge table for the transaction ID
	chargeRecord, err := tp.queryChargeRecord(transactionID)
	if err != nil {
		result.Error = fmt.Errorf("failed to query charge record: %w", err)
		return result
	}

	// NOTE: Disabled 'COMPLETED' check to force update generation
	// if chargeRecord.Status == "COMPLETED" {
	// 	result.Success = true
	// 	result.ValueAt = chargeRecord.ValuedAt
	// 	log.Printf("Transaction ID %s already has status COMPLETED, skipping", transactionID)
	// 	return result
	// }

	// Step 2: Query sg-prd-m-payment-core internal_transaction table
	txID, err := tp.queryInternalTransaction(transactionID, chargeRecord.CreatedAt)
	if err != nil {
		result.Error = fmt.Errorf("failed to query internal transaction: %w", err)
		return result
	}

	// Step 3: Query workflow_execution table for ValueTimestamp
	valueTimestamp, err := tp.queryWorkflowExecution(txID)
	if err != nil {
		result.Error = fmt.Errorf("failed to query workflow execution: %w", err)
		return result
	}

	// Step 4: Generate UPDATE SQL for charge table
	tp.generateChargeUpdateSQL(transactionID, chargeRecord.Status, valueTimestamp)

	// Step 5: Generate UPDATE SQL for workflow_execution table
	tp.generateWorkflowUpdateSQL(transactionID, chargeRecord, valueTimestamp)

	result.Success = true
	result.ValueAt = valueTimestamp
	return result
}

// queryChargeRecord queries the charge table for a single transaction ID
func (tp *TransactionProcessor) queryChargeRecord(transactionID string) (*ChargeRecord, error) {
	query := fmt.Sprintf("SELECT * FROM charge WHERE transaction_id = '%s'",
		strings.ReplaceAll(transactionID, "'", "''"))

	log.Printf("Querying charge table for transaction_id: %s", transactionID)
	rows, err := tp.Client.QueryPartnerPayEngine(query)
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
func (tp *TransactionProcessor) queryInternalTransaction(transactionID, chargeCreatedAt string) (string, error) {
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

	log.Printf("Querying internal_transaction for group_id: %s", transactionID)
	rows, err := tp.Client.QueryPaymentCore(query)
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
func (tp *TransactionProcessor) queryWorkflowExecution(txID string) (string, error) {
	query := fmt.Sprintf(`
		SELECT JSON_EXTRACT(data, '$.NotifyParams.ValueTimestamp') as ValuedAt
		FROM workflow_execution
		WHERE run_id = '%s'`,
		strings.ReplaceAll(txID, "'", "''"))

	log.Printf("Querying workflow_execution for run_id: %s", txID)
	rows, err := tp.Client.QueryPaymentCore(query)
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
func (tp *TransactionProcessor) generateChargeUpdateSQL(transactionID, originalStatus, valueAt string) {
	tp.DeploySQL.WriteString(fmt.Sprintf(`-- Update charge table for transaction_id: %s
UPDATE charge
SET status = 'COMPLETED', valued_at = '%s'
WHERE transaction_id = '%s';

`,
		transactionID,
		strings.ReplaceAll(valueAt, "'", "''"),
		strings.ReplaceAll(transactionID, "'", "''")))

	// Generate corresponding rollback
	tp.RollbackSQL.WriteString(fmt.Sprintf(`-- Rollback charge table for transaction_id: %s
UPDATE charge
SET status = '%s', valued_at = NULL
WHERE transaction_id = '%s';

`,
		transactionID,
		strings.ReplaceAll(originalStatus, "'", "''"),
		strings.ReplaceAll(transactionID, "'", "''")))
}

// generateWorkflowUpdateSQL generates UPDATE SQL for workflow_execution table
func (tp *TransactionProcessor) generateWorkflowUpdateSQL(transactionID string, chargeRecord *ChargeRecord, valueAt string) {
	// Create updated ChargeStorage with the new ValuedAt
	chargeStorage := ChargeStorage{
		ID:                      chargeRecord.ID,
		Amount:                  chargeRecord.Amount,
		Status:                  "COMPLETED",
		Remarks:                 chargeRecord.Remarks,
		TxnType:                 chargeRecord.TxnType,
		Currency:                chargeRecord.Currency,
		Metadata:                stringPtr(chargeRecord.Metadata),
		ValuedAt:                valueAt, // Use the ValueTimestamp from workflow_execution
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
	tp.DeploySQL.WriteString(fmt.Sprintf(`-- Update workflow_execution for transaction_id: %s
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

	tp.RollbackSQL.WriteString(fmt.Sprintf(`-- Rollback workflow_execution for transaction_id: %s
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

	// We list fields manually to ensure order and handling of nested objects
	fields := []struct {
		Key string
		Val string
	}{
		{"ID", fmt.Sprintf("%d", cs.ID)},
		{"Amount", fmt.Sprintf("%f", cs.Amount)},
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
		{"CapturedAmount", fmt.Sprintf("%f", cs.CapturedAmount)},
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

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
