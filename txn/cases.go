package txn

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"buddy/clients"
)

// findPaymentCoreWorkflow finds a specific workflow type in the PaymentCoreWorkflows slice
func findPaymentCoreWorkflow(workflows []WorkflowInfo, workflowType string) *WorkflowInfo {
	for i := range workflows {
		if workflows[i].Type == workflowType {
			return &workflows[i]
		}
	}
	return nil
}

// matchSOPCasePcExternalPaymentFlow200_11 checks if a transaction matches Case 1 criteria
func matchSOPCasePcExternalPaymentFlow200_11(result TransactionResult) bool {
	externalWorkflow := findPaymentCoreWorkflow(result.PaymentCoreWorkflows, "external_payment_flow")
	if externalWorkflow == nil {
		return false
	}

	return externalWorkflow.RunID != "" &&
		externalWorkflow.State == "200" &&
		externalWorkflow.Attempt == 11
}

// matchSOPCasePeTransferPayment210_0 checks if a transaction matches Case 4 criteria
func matchSOPCasePeTransferPayment210_0(result TransactionResult) bool {
	return result.PaymentEngineWorkflow.RunID != "" &&
		result.PaymentEngineWorkflow.State == "210" &&
		result.PaymentEngineWorkflow.Attempt == 0
}

// matchSOPCaseRppCashoutReject101_19 checks if a transaction matches Case 5 criteria
func matchSOPCaseRppCashoutReject101_19(result TransactionResult) bool {
	// Check if RPP workflow matches criteria
	return result.RPPWorkflow.RunID != "" &&
		result.RPPWorkflow.State == "101" &&
		result.RPPWorkflow.Attempt == 19 &&
		result.RPPWorkflow.Type == "wf_ct_cashout"
}

// matchSOPCaseRppQrPaymentReject210_0 checks if a transaction matches the new QR payment reject criteria
func matchSOPCaseRppQrPaymentReject210_0(result TransactionResult) bool {
	// Check if RPP workflow matches criteria for QR payment reject
	return result.RPPWorkflow.RunID != "" &&
		result.RPPWorkflow.State == "210" &&
		result.RPPWorkflow.Attempt == 0 &&
		result.RPPWorkflow.Type == "wf_ct_qr_payment"
}

// MatchSOPCaseRppNoResponseResume checks if a transaction matches the RPP no response resume criteria
func MatchSOPCaseRppNoResponseResume(result TransactionResult) bool {
	// Check if RPP workflow matches criteria for resume (timeout scenario)
	return result.RPPWorkflow.RunID != "" &&
		result.RPPWorkflow.State == "210" &&
		result.RPPWorkflow.Attempt == 0 &&
		(result.RPPWorkflow.Type == "wf_ct_cashout" || result.RPPWorkflow.Type == "wf_ct_qr_payment")
}

// isRPPStuckCandidate checks if the transaction matches the ambiguous state for RPP cases
// (Workflow Transfer 220 && External Payment 201/0)
func isRPPStuckCandidate(result TransactionResult) bool {
	externalWorkflow := findPaymentCoreWorkflow(result.PaymentCoreWorkflows, "external_payment_flow")
	if externalWorkflow == nil {
		return false
	}

	return result.PaymentEngineWorkflow.RunID != "" &&
		result.PaymentEngineWorkflow.State == "220" &&
		externalWorkflow.RunID != "" &&
		externalWorkflow.State == "201" &&
		externalWorkflow.Attempt == 0 &&
		result.PaymentEngineWorkflow.Attempt == 0
}

// identifySOPCase determines which SOP case a transaction matches.
// It accepts a pointer to TransactionResult to allow enrichment of data (prompts/lookups).
func identifySOPCase(result *TransactionResult) SOPCase {
	// Check if we've already identified the case
	if result.SOPCase != SOPCaseNone {
		return result.SOPCase
	}

	// 1. Check for distinct static cases
	if matchSOPCasePcExternalPaymentFlow200_11(*result) {
		result.SOPCase = SOPCasePcExternalPaymentFlow200_11
		return result.SOPCase
	}
	if matchSOPCasePeTransferPayment210_0(*result) {
		result.SOPCase = SOPCasePeTransferPayment210_0
		return result.SOPCase
	}
	if matchSOPCaseRppCashoutReject101_19(*result) {
		result.SOPCase = SOPCaseRppCashoutReject101_19
		return result.SOPCase
	}
	if matchSOPCaseRppQrPaymentReject210_0(*result) {
		result.SOPCase = SOPCaseRppQrPaymentReject210_0
		return result.SOPCase
	}
	if MatchSOPCaseRppNoResponseResume(*result) {
		result.SOPCase = SOPCaseRppNoResponseResume
		return result.SOPCase
	}

	// 2. Check for RPP Stuck Candidate (Ambiguous Case 2 vs Case 3)
	if isRPPStuckCandidate(*result) {
		fmt.Printf("\n--- Analyzing Potential RPP Stuck Case for TxID: %s ---\n", result.TransactionID)

		// Step A: Prompt for req_biz_msg_id if missing
		if result.ReqBizMsgID == "" {
			var err error
			result.ReqBizMsgID, err = promptForInput(fmt.Sprintf("Enter req_biz_msg_id for %s", result.TransactionID))
			if err != nil {
				fmt.Printf("Skipping RPP check due to input error: %v\n", err)
				result.SOPCase = SOPCaseNone
				return result.SOPCase
			}
		}

		// Step B: Resolve RPP Status
		// Priority 1: If we have ReqBizMsgID but no PartnerTxID/RPPStatus, try the nested query
		if result.RPPStatus == "" && result.ReqBizMsgID != "" {
			fmt.Printf("Attempting to resolve RPP status via req_biz_msg_id: %s\n", result.ReqBizMsgID)
			rppStatus, runID, attempt := checkRPPAdapterStatusByReqBizMsgID(result.ReqBizMsgID)

			if rppStatus != "NOT_FOUND" && rppStatus != "ERROR" {
				result.RPPStatus = rppStatus
				result.PartnerTxID = runID // The run_id found IS the partner_tx_id
				result.RPPWorkflow.RunID = runID
				result.RPPWorkflow.Attempt = attempt
				fmt.Printf("RPP Found -> Status: %s, RunID: %s, Attempt: %d\n", rppStatus, runID, attempt)
			}
		}

		// Priority 2: If we still don't have status but have a PartnerTxID (manual entry?), try direct lookup
		if result.RPPStatus == "" && result.PartnerTxID != "" {
			rppStatus, rppRunID := checkRPPAdapterStatus(result.PartnerTxID)
			result.RPPStatus = rppStatus
			result.RPPWorkflow.RunID = rppRunID
			fmt.Printf("RPP Adapter Status (via PartnerID): %s, RunID: %s\n", result.RPPStatus, result.RPPWorkflow.RunID)
		}

		// Step C: Classify based on RPP Status
		// If RPP workflow is in state 900 (Completed), we match Case 3.
		// Otherwise (state 210, NOT_FOUND, ERROR, etc.), we match Case 2 to resume/retry.
		if result.RPPStatus == "900" {
			result.SOPCase = SOPCasePcExternalPaymentFlow201_0RPP900
		} else {
			// Even if NOT_FOUND or ERROR, we generally assume we need to retry/resume (Case 2)
			// unless we want to abort. Standard SOP implies checking if it's NOT success, treat as stuck/retry.
			result.SOPCase = SOPCasePcExternalPaymentFlow201_0RPP210
		}
		return result.SOPCase
	}

	result.SOPCase = SOPCaseNone
	return result.SOPCase
}

// IdentifySOPCases identifies SOP cases for all transactions without generating SQL
func IdentifySOPCases(results []TransactionResult) {
	for i := range results {
		if results[i].SOPCase == SOPCaseNone {
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
	factory := clients.NewDoormanClientFactory("my")
	client, err := factory.CreateClient(30 * time.Second)
	if err != nil {
		fmt.Printf("Error creating doorman client: %v\n", err)
		return "ERROR", "", 0
	}

	// Nested query logic as requested
	query := fmt.Sprintf("SELECT run_id, attempt, state FROM workflow_execution WHERE run_id = (SELECT partner_tx_id FROM credit_transfer WHERE req_biz_msg_id = '%s')", reqBizMsgID)

	results, err := client.ExecuteQuery("prd-payments-rpp-adapter-rds-mysql", "prd-payments-rpp-adapter-rds-mysql", "rpp_adapter", query)
	if err != nil {
		fmt.Printf("Error querying RPP Adapter with nested query: %v\n", err)
		return "ERROR", "", 0
	}

	if len(results) == 0 {
		fmt.Println("No RPP Adapter workflow found for this req_biz_msg_id.")
		return "NOT_FOUND", "", 0
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
	fmt.Printf("Checking RPP Adapter for partner_tx_id: %s...\n", partnerTxID)

	factory := clients.NewDoormanClientFactory("my")
	client, err := factory.CreateClient(30 * time.Second)
	if err != nil {
		fmt.Printf("Error creating doorman client: %v\n", err)
		return "ERROR", ""
	}

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
		return "NOT_FOUND", ""
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
