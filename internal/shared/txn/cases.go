package txn

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"buddy/clients"
)

// findPaymentCoreWorkflow finds a specific workflow type in the PaymentCoreWorkflows slice
func findPaymentCoreWorkflow(workflows []WorkflowInfo, workflowType string) *WorkflowInfo {
	for i := range workflows {
		if workflows[i].WorkflowID == workflowType {
			return &workflows[i]
		}
	}
	return nil
}

// matchSOPCasePcExternalPaymentFlow200_11 checks if a transaction matches Case 1 criteria
func matchSOPCasePcExternalPaymentFlow200_11(result TransactionResult) bool {
	externalWorkflow := findPaymentCoreWorkflow(result.PaymentCore.Workflow, "external_payment_flow")
	if externalWorkflow == nil {
		return false
	}

	return externalWorkflow.RunID != "" &&
		externalWorkflow.State == "200" &&
		externalWorkflow.Attempt == 11
}

// matchSOPCasePeTransferPayment210_0 checks if a transaction matches Case 4 criteria
func matchSOPCasePeTransferPayment210_0(result TransactionResult) bool {
	return result.PaymentEngine.Workflow.RunID != "" &&
		result.PaymentEngine.Workflow.State == "210" &&
		result.PaymentEngine.Workflow.Attempt == 0
}

// matchSOPCaseRppCashoutReject101_19 checks if a transaction matches Case 5 criteria
func matchSOPCaseRppCashoutReject101_19(result TransactionResult) bool {
	// Check if RPP workflow matches criteria
	return result.RPPAdapter.Workflow.RunID != "" &&
		result.RPPAdapter.Workflow.State == "101" &&
		result.RPPAdapter.Workflow.Attempt == 19 &&
		result.RPPAdapter.Workflow.WorkflowID == "wf_ct_cashout"
}

// matchSOPCaseRppQrPaymentReject210_0 checks if a transaction matches the new QR payment reject criteria
func matchSOPCaseRppQrPaymentReject210_0(result TransactionResult) bool {
	// Check if RPP workflow matches criteria for QR payment reject
	return result.RPPAdapter.Workflow.RunID != "" &&
		result.RPPAdapter.Workflow.State == "210" &&
		result.RPPAdapter.Workflow.Attempt == 0 &&
		result.RPPAdapter.Workflow.WorkflowID == "wf_ct_qr_payment"
}

// MatchSOPCaseRppNoResponseResume checks if a transaction matches the RPP no response resume criteria
func MatchSOPCaseRppNoResponseResume(result TransactionResult) bool {
	// Check if RPP workflow matches criteria for resume (timeout scenario)
	return result.RPPAdapter.Workflow.RunID != "" &&
		result.RPPAdapter.Workflow.State == "210" &&
		result.RPPAdapter.Workflow.Attempt == 0 &&
		(result.RPPAdapter.Workflow.WorkflowID == "wf_ct_cashout" || result.RPPAdapter.Workflow.WorkflowID == "wf_ct_qr_payment")
}

// isRPPStuckCandidate checks if the transaction matches the ambiguous state for RPP cases
// (Workflow Transfer 220 && External Payment 201/0)
func isRPPStuckCandidate(result TransactionResult) bool {
	externalWorkflow := findPaymentCoreWorkflow(result.PaymentCore.Workflow, "external_payment_flow")
	if externalWorkflow == nil {
		return false
	}

	return result.PaymentEngine.Workflow.RunID != "" &&
		result.PaymentEngine.Workflow.State == "220" &&
		externalWorkflow.RunID != "" &&
		externalWorkflow.State == "201" &&
		externalWorkflow.Attempt == 0 &&
		result.PaymentEngine.Workflow.Attempt == 0
}

// identifySOPCase determines which SOP case a transaction matches.
// It accepts a pointer to TransactionResult to allow enrichment of data (prompts/lookups).
func identifySOPCase(result *TransactionResult) SOPCase {
	// Check if we've already identified the case
	if result.CaseType != SOPCaseNone {
		return result.CaseType
	}

	// 1. Check for distinct static cases
	if matchSOPCasePcExternalPaymentFlow200_11(*result) {
		result.CaseType = SOPCasePcExternalPaymentFlow200_11
		return result.CaseType
	}
	if matchSOPCasePeTransferPayment210_0(*result) {
		result.CaseType = SOPCasePeTransferPayment210_0
		return result.CaseType
	}
	if matchSOPCaseRppCashoutReject101_19(*result) {
		result.CaseType = SOPCaseRppCashoutReject101_19
		return result.CaseType
	}
	if matchSOPCaseRppQrPaymentReject210_0(*result) {
		result.CaseType = SOPCaseRppQrPaymentReject210_0
		return result.CaseType
	}
	if MatchSOPCaseRppNoResponseResume(*result) {
		result.CaseType = SOPCaseRppNoResponseResume
		return result.CaseType
	}

	// 2. Check for RPP Stuck Candidate (Ambiguous Case 2 vs Case 3)
	if isRPPStuckCandidate(*result) {
		fmt.Printf("\n--- Analyzing Potential RPP Stuck Case for TxID: %s ---\n", result.TransactionID)

		// Step A: Prompt for req_biz_msg_id if missing
		if result.RPPAdapter.ReqBizMsgID == "" {
			var err error
			result.RPPAdapter.ReqBizMsgID, err = promptForInput(fmt.Sprintf("Enter req_biz_msg_id for %s", result.TransactionID))
			if err != nil {
				fmt.Printf("Skipping RPP check due to input error: %v\n", err)
				result.CaseType = SOPCaseNone
				return result.CaseType
			}
		}

		// Step B: Resolve RPP Status
		// Priority 1: If we have ReqBizMsgID but no PartnerTxID/RPPStatus, try the nested query
		if result.RPPAdapter.Status == "" && result.RPPAdapter.ReqBizMsgID != "" {
			fmt.Printf("Attempting to resolve RPP status via req_biz_msg_id: %s\n", result.RPPAdapter.ReqBizMsgID)
			rppStatus, runID, attempt := checkRPPAdapterStatusByReqBizMsgID(result.RPPAdapter.ReqBizMsgID)

			if rppStatus != NotFoundStatus && rppStatus != "ERROR" {
				result.RPPAdapter.Status = rppStatus
				result.RPPAdapter.PartnerTxID = runID // The run_id found IS the partner_tx_id
				result.RPPAdapter.Workflow.RunID = runID
				result.RPPAdapter.Workflow.Attempt = attempt
				fmt.Printf("RPP Found -> Status: %s, RunID: %s, Attempt: %d\n", rppStatus, runID, attempt)
			}
		}

		// Priority 2: If we still don't have status but have a PartnerTxID (manual entry?), try direct lookup
		if result.RPPAdapter.Status == "" && result.RPPAdapter.PartnerTxID != "" {
			rppStatus, rppRunID := checkRPPAdapterStatus(result.RPPAdapter.PartnerTxID)
			result.RPPAdapter.Status = rppStatus
			result.RPPAdapter.Workflow.RunID = rppRunID
			fmt.Printf("RPP Adapter Status (via PartnerID): %s, RunID: %s\n", result.RPPAdapter.Status, result.RPPAdapter.Workflow.RunID)
		}

		// Step C: Classify based on RPP Status
		// If RPP workflow is in state 900 (Completed), we match Case 3.
		// Otherwise (state 210, NOT_FOUND, ERROR, etc.), we match Case 2 to resume/retry.
		if result.RPPAdapter.Status == "900" {
			result.CaseType = SOPCasePcExternalPaymentFlow201_0RPP900
		} else {
			// Even if NOT_FOUND or ERROR, we generally assume we need to retry/resume (Case 2)
			// unless we want to abort. Standard SOP implies checking if it's NOT success, treat as stuck/retry.
			result.CaseType = SOPCasePcExternalPaymentFlow201_0RPP210
		}
		return result.CaseType
	}

	result.CaseType = SOPCaseNone
	return result.CaseType
}

// IdentifySOPCases identifies SOP cases for all transactions without generating SQL
func IdentifySOPCases(results []TransactionResult) {
	for i := range results {
		if results[i].CaseType == SOPCaseNone {
			identifySOPCase(&results[i])
		}
	}
}

// promptForInput reads a line from stdin
func promptForInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// checkRPPAdapterStatusByReqBizMsgID performs the nested query to find run_id/partner_tx_id via req_biz_msg_id
func checkRPPAdapterStatusByReqBizMsgID(reqBizMsgID string) (string, string, int) {
	return checkRPPAdapterStatusByReqBizMsgIDWithEnv(reqBizMsgID, "my")
}

// checkRPPAdapterStatusByReqBizMsgIDWithEnv performs the nested query to find run_id/partner_tx_id via req_biz_msg_id with environment
func checkRPPAdapterStatusByReqBizMsgIDWithEnv(reqBizMsgID string, env string) (string, string, int) {
	client := clients.Doorman

	// Nested query logic as requested
	query := fmt.Sprintf("SELECT run_id, attempt, state FROM workflow_execution WHERE run_id = (SELECT partner_tx_id FROM credit_transfer WHERE req_biz_msg_id = '%s')", reqBizMsgID)

	results, err := client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil {
		fmt.Printf("Error querying RPP Adapter with nested query: %v\n", err)
		return "ERROR", "", 0
	}

	if len(results) == 0 {
		fmt.Println("No RPP Adapter workflow found for this req_biz_msg_id.")
		return NotFoundStatus, "", 0
	}

	// Parse results
	var stateStr string
	var runIDStr string
	var attemptInt int

	if stateVal, ok := results[0]["state"]; ok {
		stateStr = fmt.Sprintf("%v", stateVal)
	}

	if runIDVal, ok := results[0]["run_id"]; ok {
		runIDStr = fmt.Sprintf("%v", runIDVal)
	}

	if attemptVal, ok := results[0]["attempt"]; ok {
		if attemptFloat, ok := attemptVal.(float64); ok {
			attemptInt = int(attemptFloat)
		}
	}

	return stateStr, runIDStr, attemptInt
}

// checkRPPAdapterStatus queries the RPP Adapter database for the real workflow state and run_id using PartnerTxID
func checkRPPAdapterStatus(partnerTxID string) (string, string) {
	return checkRPPAdapterStatusWithEnv(partnerTxID, "my")
}

// checkRPPAdapterStatusWithEnv queries the RPP Adapter database for the real workflow state and run_id using PartnerTxID with environment
func checkRPPAdapterStatusWithEnv(partnerTxID string, env string) (string, string) {
	fmt.Printf("Checking RPP Adapter for partner_tx_id: %s...\n", partnerTxID)

	client := clients.Doorman

	// Query the workflow_execution table in the RPP Adapter database
	query := fmt.Sprintf("SELECT state, run_id, workflow_id, attempt FROM workflow_execution WHERE run_id = '%s'", partnerTxID)

	// ExecuteQuery using the specific RPP Adapter connection details
	results, err := client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil {
		fmt.Printf("Error querying RPP Adapter: %v\n", err)
		return "ERROR", ""
	}

	if len(results) == 0 {
		fmt.Println("No RPP Adapter workflow found.")
		return NotFoundStatus, ""
	}

	// Parse and return the state and run_id
	if stateVal, ok := results[0]["state"]; ok {
		state := fmt.Sprintf("%v", stateVal)
		runID := ""
		if runIDVal, ok := results[0]["run_id"]; ok {
			runID = fmt.Sprintf("%v", runIDVal)
		}
		return state, runID
	}

	return "UNKNOWN", ""
}
