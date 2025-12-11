package txn

import (
	"fmt"
	"os"
	"strings"
)

// SQLStatements contains the deploy and rollback SQL statements separated by database
type SQLStatements struct {
	PCDeployStatements    []string
	PCRollbackStatements  []string
	PEDeployStatements    []string
	PERollbackStatements  []string
	RPPDeployStatements   []string
	RPPRollbackStatements []string
}

// GenerateSQLStatements generates SQL statements for all supported cases.
func GenerateSQLStatements(results []TransactionResult) SQLStatements {
	statements := SQLStatements{}

	// Slice to collect run_ids for bulk processing of Case 1
	var pcExtPayment200_11RunIDs []string

	fmt.Println("\n--- Generating SQL Statements ---")

	// Iterate using index to allow modification of the result struct (if needed)
	for i := range results {
		// identifySOPCase now takes a pointer and might trigger prompts
		caseType := identifySOPCase(&results[i])
		results[i].RPPInfo = string(caseType) // Store identified case for reference

		switch caseType {
		case SOPCasePcExternalPaymentFlow200_11:
			// Collect run_id for bulk generation instead of generating individual statements
			runID := getPcExtPayment200_11RunID(results[i])
			if runID != "" {
				pcExtPayment200_11RunIDs = append(pcExtPayment200_11RunIDs, runID)
			}
		case SOPCasePcExternalPaymentFlow201_0RPP210:
			s := FixPcExtPayment201_0RPP210(results[i])
			statements.RPPDeployStatements = append(statements.RPPDeployStatements, s.RPPDeployStatements...)
			statements.RPPRollbackStatements = append(statements.RPPRollbackStatements, s.RPPRollbackStatements...)
		case SOPCasePcExternalPaymentFlow201_0RPP900:
			s := FixPcExtPayment201_0RPP900(results[i])
			statements.RPPDeployStatements = append(statements.RPPDeployStatements, s.RPPDeployStatements...)
			statements.RPPRollbackStatements = append(statements.RPPRollbackStatements, s.RPPRollbackStatements...)
		case SOPCasePeTransferPayment210_0:
			s := FixPeTransferPayment210_0(results[i])
			statements.PEDeployStatements = append(statements.PEDeployStatements, s.PEDeployStatements...)
			statements.PERollbackStatements = append(statements.PERollbackStatements, s.PERollbackStatements...)
		}
	}

	// Generate bulk statements for Case 1 if any found
	if len(pcExtPayment200_11RunIDs) > 0 {
		s := FixPcExtPayment200_11_Bulk(pcExtPayment200_11RunIDs)
		appendStatements(&statements, s)
	}

	return statements
}

// appendStatements is a helper to merge results into the main struct
func appendStatements(main *SQLStatements, new SQLStatements) {
	main.PCDeployStatements = append(main.PCDeployStatements, new.PCDeployStatements...)
	main.PCRollbackStatements = append(main.PCRollbackStatements, new.PCRollbackStatements...)
	main.PEDeployStatements = append(main.PEDeployStatements, new.PEDeployStatements...)
	main.PERollbackStatements = append(main.PERollbackStatements, new.PERollbackStatements...)
	main.RPPDeployStatements = append(main.RPPDeployStatements, new.RPPDeployStatements...)
	main.RPPRollbackStatements = append(main.RPPRollbackStatements, new.RPPRollbackStatements...)
}

// getPcExtPayment200_11RunID extracts the relevant run_id for Case 1
func getPcExtPayment200_11RunID(result TransactionResult) string {
	for _, w := range result.PaymentCoreWorkflows {
		if w.Type == "external_payment_flow" && w.State == "200" && w.Attempt == 11 {
			return w.RunID
		}
	}
	return ""
}

// FixPcExtPayment200_11_Bulk generates SQL statements for multiple transactions using IN clause
func FixPcExtPayment200_11_Bulk(runIDs []string) SQLStatements {
	var pcDeployStatements []string
	var pcRollbackStatements []string

	if len(runIDs) == 0 {
		return SQLStatements{}
	}

	// Create comma-separated list of quoted run_ids
	quotedRunIDs := make([]string, len(runIDs))
	for i, id := range runIDs {
		quotedRunIDs[i] = fmt.Sprintf("'%s'", id)
	}
	inClause := strings.Join(quotedRunIDs, ", ")

	// Generate deploy statement
	deploySQL := fmt.Sprintf(`-- pc_external_payment_flow_200_11 (Bulk Update for %d items)
UPDATE workflow_execution SET state = 202, attempt = 1,
  `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.StreamResp', JSON_OBJECT('TxID', '', 'Status', 'FAILED', 'ErrorCode', 'ADAPTER_ERROR', 'ExternalID', '', 'ErrorMessage', 'Reject from adapter'), '$.State', 202)
WHERE run_id IN (%s)
AND state = 200 AND attempt = 11;`, len(runIDs), inClause)
	pcDeployStatements = append(pcDeployStatements, deploySQL)

	// Generate rollback statement
	rollbackSQL := fmt.Sprintf(`UPDATE workflow_execution SET state = 200, attempt = 11, `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.State', 200)
WHERE run_id IN (%s);`, inClause)
	pcRollbackStatements = append(pcRollbackStatements, rollbackSQL)

	return SQLStatements{
		PCDeployStatements:   pcDeployStatements,
		PCRollbackStatements: pcRollbackStatements,
	}
}

// FixPcExtPayment201_0RPP210 generates SQL for pc_external_payment_flow_201_0_RPP_210 (state 210)
func FixPcExtPayment201_0RPP210(result TransactionResult) SQLStatements {
	var rppDeployStatements []string
	var rppRollbackStatements []string

	// Generate deploy statement
	deploySQL := fmt.Sprintf(`-- RPP 210, PE 220, PC 201. No response from RPP. Move to 222 to resume. ACSP
UPDATE workflow_execution SET state = 222, attempt = 1,
	 `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.State', 222)
WHERE run_id = '%s' AND state = 210;`, result.RPPWorkflow.RunID)
	rppDeployStatements = append(rppDeployStatements, deploySQL)

	// Generate rollback statement
	rollbackSQL := fmt.Sprintf(`UPDATE workflow_execution SET state = 201, attempt = 0,
	 `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.State', 201)
WHERE run_id = '%s';`, result.RPPWorkflow.RunID)
	rppRollbackStatements = append(rppRollbackStatements, rollbackSQL)

	return SQLStatements{
		RPPDeployStatements:   rppDeployStatements,
		RPPRollbackStatements: rppRollbackStatements,
	}
}

// FixPcExtPayment201_0RPP900 generates SQL statements for pc_external_payment_flow_201_0_RPP_900
func FixPcExtPayment201_0RPP900(result TransactionResult) SQLStatements {
	var rppDeployStatements []string
	var rppRollbackStatements []string

	// Ensure we have a valid run ID from the RPP data
	if result.RPPWorkflow.RunID != "" {
		// Generate deploy statement
		deploySQL := fmt.Sprintf(`-- RPP 900, PE 220, PC 201. Republish from RPP to resume. ACSP
UPDATE workflow_execution
SET state = 301, attempt = 1, `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.State', 301)
WHERE run_id = '%s' AND state = 900;`, result.RPPWorkflow.RunID)
		rppDeployStatements = append(rppDeployStatements, deploySQL)

		// Generate rollback statement
		rollbackSQL := fmt.Sprintf(`UPDATE workflow_execution SET state = 900, attempt = 0, `+"`data`"+` = JSON_SET(`+"`data`"+`, '$.State', 900)
WHERE run_id = '%s';`, result.RPPWorkflow.RunID)
		rppRollbackStatements = append(rppRollbackStatements, rollbackSQL)
	}

	return SQLStatements{
		RPPDeployStatements:   rppDeployStatements,
		RPPRollbackStatements: rppRollbackStatements,
	}
}

// FixPeTransferPayment210_0 generates SQL statements for transactions with
// workflow_transfer_payment: stAuthProcessing (210) attempt=0
func FixPeTransferPayment210_0(result TransactionResult) SQLStatements {
	var peDeployStatements []string
	var peRollbackStatements []string

	// Generate deploy statement
	deploySQL := fmt.Sprintf(`-- Reject PE stuck 210. Reject transactions since it hasn't reached Paynet yet
UPDATE workflow_execution
SET  state = 221, attempt = 1, `+"`data`"+` = JSON_SET(
	 `+"`data`"+`, '$.StreamMessage',
	 JSON_OBJECT(
	         'Status', 'FAILED',
	         'ErrorCode', 'ADAPTER_ERROR',
	         'ErrorMessage', 'Manual Rejected'),
	 '$.State', 221)
WHERE run_id = '%s' AND workflow_id = 'workflow_transfer_payment' AND state = 210;`, result.PaymentEngineWorkflow.RunID)
	peDeployStatements = append(peDeployStatements, deploySQL)

	// Generate rollback statement
	rollbackSQL := fmt.Sprintf(`UPDATE workflow_execution
SET  state = 210, attempt = 0, `+"`data`"+` = JSON_SET(
	 `+"`data`"+`, '$.StreamMessage', null,
	   '$.State', 210)
WHERE run_id = '%s' AND workflow_id = 'workflow_transfer_payment';`, result.PaymentEngineWorkflow.RunID)
	peRollbackStatements = append(peRollbackStatements, rollbackSQL)

	return SQLStatements{
		PEDeployStatements:   peDeployStatements,
		PERollbackStatements: peRollbackStatements,
	}
}

// WriteSQLFiles writes the SQL statements to database-specific Deploy.sql and Rollback.sql files
func WriteSQLFiles(statements SQLStatements, basePath string) error {
	// Write PC files
	if len(statements.PCDeployStatements) > 0 {
		deployPath := basePath + "_PC_Deploy.sql"
		if err := writeSQLFile(deployPath, statements.PCDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PCRollbackStatements) > 0 {
		rollbackPath := basePath + "_PC_Rollback.sql"
		if err := writeSQLFile(rollbackPath, statements.PCRollbackStatements); err != nil {
			return err
		}
	}

	// Write PE files
	if len(statements.PEDeployStatements) > 0 {
		deployPath := basePath + "_PE_Deploy.sql"
		if err := writeSQLFile(deployPath, statements.PEDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PERollbackStatements) > 0 {
		rollbackPath := basePath + "_PE_Rollback.sql"
		if err := writeSQLFile(rollbackPath, statements.PERollbackStatements); err != nil {
			return err
		}
	}

	// Write RPP files
	if len(statements.RPPDeployStatements) > 0 {
		deployPath := basePath + "_RPP_Deploy.sql"
		if err := writeSQLFile(deployPath, statements.RPPDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.RPPRollbackStatements) > 0 {
		rollbackPath := basePath + "_RPP_Rollback.sql"
		if err := writeSQLFile(rollbackPath, statements.RPPRollbackStatements); err != nil {
			return err
		}
	}

	return nil
}

// writeSQLFile writes SQL statements to a file
func writeSQLFile(filePath string, statements []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer file.Close()

	for _, stmt := range statements {
		fmt.Fprintln(file, stmt)
		fmt.Fprintln(file)
	}
	fmt.Printf("SQL statements written to %s\n", filePath)
	return nil
}
