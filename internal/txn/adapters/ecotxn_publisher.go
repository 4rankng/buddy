package adapters

import (
	"fmt"
	"os"
	"strings"
	"time"

	"buddy/internal/clients/doorman"
	"buddy/internal/txn/utils"
	internalutils "buddy/internal/utils"
)

// EcoTxnPublisher handles SQL generation for ecosystem transactions
type EcoTxnPublisher struct {
	client      doorman.DoormanInterface
	PPEDeploy   strings.Builder
	PPERollback strings.Builder
	PCDeploy    strings.Builder
	PCRollback  strings.Builder

	// Trackers for batch processing of PC records
	PCRunIDs []string
}

// NewEcoTxnPublisher creates a new publisher instance
func NewEcoTxnPublisher() *EcoTxnPublisher {
	return &EcoTxnPublisher{
		client:   doorman.Doorman,
		PCRunIDs: make([]string, 0),
	}
}

// ProcessEcoTxnPublish processes a single transaction for publishing
func ProcessEcoTxnPublish(transactionID, env string) error {
	publisher := NewEcoTxnPublisher()

	result := publisher.processSingleTransaction(transactionID)
	if result.Success {
		publisher.generatePCBatchSQL()

		if err := WriteEcoTxnSQLFiles(
			publisher.PPEDeploy.String(),
			publisher.PPERollback.String(),
			publisher.PCDeploy.String(),
			publisher.PCRollback.String(),
		); err != nil {
			return fmt.Errorf("failed to write SQL files: %v", err)
		}
		return nil
	}
	return result.Error
}

// ProcessEcoTxnPublishBatch processes multiple transactions from a file
func ProcessEcoTxnPublishBatch(filePath, env string) {
	transactionIDs, err := utils.ReadTransactionIDsFromFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	publisher := NewEcoTxnPublisher()
	processedCount := 0
	failedCount := 0

	for _, txID := range transactionIDs {
		result := publisher.processSingleTransaction(txID)
		if result.Success {
			processedCount++
		} else {
			failedCount++
			fmt.Printf("Failed to process transaction ID: %s, Error: %v\n", txID, result.Error)
		}
	}

	publisher.generatePCBatchSQL()

	fmt.Printf("Processing completed. Success: %d, Failed: %d\n", processedCount, failedCount)

	if err := WriteEcoTxnSQLFiles(
		publisher.PPEDeploy.String(),
		publisher.PPERollback.String(),
		publisher.PCDeploy.String(),
		publisher.PCRollback.String(),
	); err != nil {
		fmt.Printf("Error writing SQL files: %v\n", err)
		return
	}
}

// processSingleTransaction processes a single transaction ID
func (p *EcoTxnPublisher) processSingleTransaction(transactionID string) ProcessResult {
	result := ProcessResult{TransactionID: transactionID, Success: false}

	// 1. Query PPE Charge table
	chargeRecord, err := p.queryChargeRecord(transactionID)
	if err != nil {
		result.Error = fmt.Errorf("failed to query charge record: %w", err)
		return result
	}

	// 2. Query PC internal_transaction table to find the tx_id (run_id) and valued_at
	pcRunID, pcValuedAt, err := p.queryInternalTransaction(transactionID, chargeRecord.CreatedAt)
	if err != nil {
		result.Error = fmt.Errorf("failed to query payment-core internal transaction: %w", err)
		return result
	}
	p.PCRunIDs = append(p.PCRunIDs, pcRunID)

	// 3. Construct ChargeStorage JSON for workflow_execution update
	chargeStorageSQL := buildChargeStorageJSONObject(chargeRecord, pcValuedAt)

	// 4. Generate PPE SQL (Deploy)
	p.PPEDeploy.WriteString(fmt.Sprintf("-- Transaction: %s\n", transactionID))
	p.PPEDeploy.WriteString(fmt.Sprintf(`UPDATE charge
SET 
    valued_at = '%s',
    updated_at = '%s'
WHERE transaction_id = '%s';

UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', %s)
WHERE
    run_id = '%s';

`, pcValuedAt, chargeRecord.UpdatedAt, transactionID, chargeStorageSQL, transactionID))

	// 5. Generate PPE SQL (Rollback)
	p.PPERollback.WriteString(fmt.Sprintf("-- Transaction: %s\n", transactionID))
	p.PPERollback.WriteString(fmt.Sprintf(`UPDATE charge
SET 
    valued_at = '%s',
    updated_at = '%s'
WHERE transaction_id = '%s';

UPDATE workflow_execution
SET
    state = 800,
    attempt = 1,
    data = JSON_SET(data,
            '$.State', 800,
            '$.ChargeStorage', JSON_OBJECT())
WHERE
    run_id = '%s';

`, chargeRecord.ValuedAt, chargeRecord.UpdatedAt, transactionID, transactionID))

	result.Success = true
	return result
}

// generatePCBatchSQL generates the PC Deploy and Rollback files based on collected Run IDs
func (p *EcoTxnPublisher) generatePCBatchSQL() {
	if len(p.PCRunIDs) == 0 {
		return
	}

	quotedIDs := make([]string, len(p.PCRunIDs))
	for i, id := range p.PCRunIDs {
		quotedIDs[i] = fmt.Sprintf("'%s'", id)
	}
	idList := strings.Join(quotedIDs, ", ")

	p.PCDeploy.WriteString(fmt.Sprintf(`UPDATE workflow_execution SET attempt=1, state=902
WHERE workflow_id='internal_payment_flow' 
AND run_id in (%s) 
AND state=900;
`, idList))

	p.PCRollback.WriteString(fmt.Sprintf(`UPDATE workflow_execution SET attempt=0, state=900
WHERE workflow_id='internal_payment_flow' 
AND run_id in (%s);
`, idList))
}

// queryChargeRecord queries the charge table and populates the record
func (p *EcoTxnPublisher) queryChargeRecord(transactionID string) (*ChargeRecord, error) {
	query := fmt.Sprintf("SELECT * FROM charge WHERE transaction_id = '%s'", escapeSQL(transactionID))
	rows, err := p.client.QueryPartnerpayEngine(query)
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("charge record not found: %v", err)
	}
	row := rows[0]

	return &ChargeRecord{
		ID:                      toInt(row["id"]),
		TransactionID:           toString(row["transaction_id"]),
		Amount:                  toFloat64(row["amount"]),
		Status:                  toString(row["status"]),
		Remarks:                 toString(row["remarks"]),
		TxnType:                 toString(row["txn_type"]),
		Currency:                toString(row["currency"]),
		Metadata:                toString(row["metadata"]),
		ValuedAt:                toString(row["valued_at"]),
		CreatedAt:               toString(row["created_at"]),
		PartnerID:               toString(row["partner_id"]),
		TxnDomain:               toString(row["txn_domain"]),
		UpdatedAt:               toString(row["updated_at"]),
		CustomerID:              toString(row["customer_id"]),
		ExternalID:              toString(row["external_id"]),
		Properties:              toString(row["properties"]),
		TxnSubtype:              toString(row["txn_subtype"]),
		ReferenceID:             toString(row["reference_id"]),
		BillingToken:            toString(row["billing_token"]),
		StatusReason:            toString(row["status_reason"]),
		CaptureMethod:           toString(row["capture_method"]),
		SourceAccount:           toString(row["source_account"]),
		CapturedAmount:          toFloat64(row["captured_amount"]),
		DestinationAccount:      toString(row["destination_account"]),
		TransactionPayLoad:      toString(row["transaction_payload"]),
		StatusReasonDescription: toString(row["status_reason_description"]),
	}, nil
}

// queryInternalTransaction finds the tx_id and ValueTimestamp in PC
func (p *EcoTxnPublisher) queryInternalTransaction(transactionID, createdAt string) (string, string, error) {
	chargeTime, _ := time.Parse("2006-01-02 15:04:05", createdAt)
	if chargeTime.IsZero() {
		chargeTime, _ = time.Parse(time.RFC3339, createdAt)
	}

	startTime := chargeTime.Add(-1 * time.Hour).Format("2006-01-02 15:04:05")
	endTime := chargeTime.Add(1 * time.Hour).Format("2006-01-02 15:04:05")

	// Query for tx_id
	queryTx := fmt.Sprintf(`
		SELECT tx_id FROM internal_transaction 
		WHERE group_id = '%s' 
		AND created_at >= '%s' AND created_at <= '%s' 
		LIMIT 1`, escapeSQL(transactionID), startTime, endTime)

	rowsTx, err := p.client.QueryPaymentCore(queryTx)
	if err != nil || len(rowsTx) == 0 {
		return "", "", fmt.Errorf("internal_transaction not found")
	}
	txID := toString(rowsTx[0]["tx_id"])

	// Query for ValueTimestamp in workflow execution
	queryWF := fmt.Sprintf(`
		SELECT JSON_EXTRACT(data, '$.NotifyParams.ValueTimestamp') as ValuedAt 
		FROM workflow_execution WHERE run_id = '%s'`, escapeSQL(txID))

	rowsWF, err := p.client.QueryPaymentCore(queryWF)
	if err != nil || len(rowsWF) == 0 {
		return txID, "", fmt.Errorf("workflow metadata not found")
	}
	valAt := strings.Trim(toString(rowsWF[0]["ValuedAt"]), "\"")

	return txID, valAt, nil
}

// buildChargeStorageJSONObject constructs the MySQL JSON_OBJECT string
func buildChargeStorageJSONObject(cr *ChargeRecord, valuedAt string) string {
	var sb strings.Builder
	sb.WriteString("JSON_OBJECT(\n")

	fields := []struct {
		Key string
		Val string
	}{
		{"ID", fmt.Sprintf("%d", cr.ID)},
		{"Amount", fmt.Sprintf("%d", int(cr.Amount))},
		{"Status", fmt.Sprintf("'%s'", escapeSQL(cr.Status))},
		{"Remarks", fmt.Sprintf("'%s'", escapeSQL(cr.Remarks))},
		{"TxnType", fmt.Sprintf("'%s'", escapeSQL(cr.TxnType))},
		{"Currency", fmt.Sprintf("'%s'", escapeSQL(cr.Currency))},
		{"Metadata", parseToSQLJSON(cr.Metadata)},
		{"ValuedAt", fmt.Sprintf("'%s'", escapeSQL(valuedAt))},
		{"CreatedAt", fmt.Sprintf("'%s'", escapeSQL(cr.CreatedAt))},
		{"PartnerID", fmt.Sprintf("'%s'", escapeSQL(cr.PartnerID))},
		{"TxnDomain", fmt.Sprintf("'%s'", escapeSQL(cr.TxnDomain))},
		{"UpdatedAt", fmt.Sprintf("'%s'", escapeSQL(cr.UpdatedAt))},
		{"CustomerID", fmt.Sprintf("'%s'", escapeSQL(cr.CustomerID))},
		{"ExternalID", fmt.Sprintf("'%s'", escapeSQL(cr.ExternalID))},
		{"Properties", parseToSQLJSON(cr.Properties)},
		{"TxnSubtype", fmt.Sprintf("'%s'", escapeSQL(cr.TxnSubtype))},
		{"ReferenceID", fmt.Sprintf("'%s'", escapeSQL(cr.ReferenceID))},
		{"BillingToken", fmt.Sprintf("'%s'", escapeSQL(cr.BillingToken))},
		{"StatusReason", fmt.Sprintf("'%s'", escapeSQL(cr.StatusReason))},
		{"CaptureMethod", fmt.Sprintf("'%s'", escapeSQL(cr.CaptureMethod))},
		{"SourceAccount", parseToSQLJSON(cr.SourceAccount)},
		{"TransactionID", fmt.Sprintf("'%s'", escapeSQL(cr.TransactionID))},
		{"CapturedAmount", fmt.Sprintf("%d", int(cr.CapturedAmount))},
		{"DestinationAccount", parseToSQLJSON(cr.DestinationAccount)},
		{"TransactionPayLoad", parseToSQLJSON(cr.TransactionPayLoad)},
		{"StatusReasonDescription", fmt.Sprintf("'%s'", escapeSQL(cr.StatusReasonDescription))},
	}

	for i, f := range fields {
		sb.WriteString(fmt.Sprintf("    '%s', %s", f.Key, f.Val))
		if i < len(fields)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(")")
	return sb.String()
}

func parseToSQLJSON(jsonStr string) string {
	if jsonStr == "" || jsonStr == "null" {
		return "NULL"
	}
	if expr, err := internalutils.ToMySQLJSONObjectExpr(jsonStr); err == nil {
		return expr
	}
	return fmt.Sprintf("'%s'", escapeSQL(jsonStr))
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch i := v.(type) {
	case int:
		return i
	case float64:
		return int(i)
	case string:
		var res int
		fmt.Sscanf(i, "%d", &res)
		return res
	}
	return 0
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch f := v.(type) {
	case float64:
		return f
	case float32:
		return float64(f)
	}
	return 0
}

func WriteEcoTxnSQLFiles(ppeD, ppeR, pcD, pcR string) error {
	files := map[string]string{
		"PPE_Deploy.sql":   ppeD,
		"PPE_Rollback.sql": ppeR,
		"PC_Deploy.sql":    pcD,
		"PC_Rollback.sql":  pcR,
	}
	for name, content := range files {
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}
