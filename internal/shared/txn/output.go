package txn

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// WriteBatchResults writes transaction results to an output file in the new format
func WriteBatchResults(results []TransactionResult, outputPath string) error {
	var buffer bytes.Buffer
	for idx, result := range results {
		writeResult(&buffer, result, idx+1)
	}

	if err := os.WriteFile(outputPath, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

// PrintTransactionStatus prints transaction information in the new format
func PrintTransactionStatus(transactionID string) {
	PrintTransactionStatusWithEnv(transactionID, "my")
}

// PrintTransactionStatusWithEnv prints transaction information in the new format with specified environment
func PrintTransactionStatusWithEnv(transactionID string, env string) {
	result := QueryTransactionStatusWithEnv(transactionID, env)
	writeResult(os.Stdout, *result, 1)
}

// WriteResult writes a single transaction result to the specified writer
func WriteResult(w io.Writer, result TransactionResult, index int) {
	writeResult(w, result, index)
}

func writeResult(w io.Writer, result TransactionResult, index int) {
	if index <= 0 {
		index = 1
	}

	peStatus := result.PaymentEngine.Transfers.Status
	ppeStatus := result.PartnerpayEngine.Transfers.Status

	// For RPP E2E IDs with errors, still show the RPP adapter section with NOT FOUND status
	if IsRppE2EID(result.TransactionID) && result.Error != "" {
		if _, err := fmt.Fprintf(w, "### [%d] e2e_id: %s\n", index, result.TransactionID); err != nil {
			fmt.Printf("Warning: failed to write e2e_id: %v\n", err)
		}
		if _, err := fmt.Fprintln(w, "[rpp-adapter]"); err != nil {
			fmt.Printf("Warning: failed to write rpp-adapter header: %v\n", err)
		}
		if _, err := fmt.Fprintln(w, "NOT FOUND"); err != nil {
			fmt.Printf("Warning: failed to write NOT FOUND: %v\n", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
		return
	}

	notFound := peStatus == NotFoundStatus || ppeStatus == NotFoundStatus

	if result.Error != "" && !notFound {
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nError: %s\n\n", index, result.TransactionID, result.Error); err != nil {
			fmt.Printf("Warning: failed to write error result: %v\n", err)
		}
		return
	}

	if notFound {
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nStatus: %s\n\n", index, result.TransactionID, NotFoundStatus); err != nil {
			fmt.Printf("Warning: failed to write not found result: %v\n", err)
		}
		return
	}

	// Use e2e_id for RPP E2E IDs, transaction_id for others
	idLabel := "transaction_id"
	if IsRppE2EID(result.TransactionID) {
		idLabel = "e2e_id"
	}
	if _, err := fmt.Fprintf(w, "### [%d] %s: %s\n", index, idLabel, result.TransactionID); err != nil {
		fmt.Printf("Warning: failed to write ID: %v\n", err)
	}

	// If this is an E2E ID lookup (determined by checking if it matches E2E ID pattern),
	// show RPP adapter info first
	if IsRppE2EID(result.TransactionID) {
		if _, err := fmt.Fprintln(w, "[rpp-adapter]"); err != nil {
			fmt.Printf("Warning: failed to write rpp-adapter header: %v\n", err)
		}
		if _, err := fmt.Fprintf(w, "e2e_id: %s\n", result.TransactionID); err != nil {
			fmt.Printf("Warning: failed to write e2e_id: %v\n", err)
		}
		if result.RPPAdapter.Status != "" {
			if _, err := fmt.Fprintf(w, "credit_transfer.status: %s\n", result.RPPAdapter.Status); err != nil {
				fmt.Printf("Warning: failed to write rpp status: %v\n", err)
			}
		}
		if result.RPPAdapter.PartnerTxID != "" {
			if _, err := fmt.Fprintf(w, "partner_tx_id: %s\n", result.RPPAdapter.PartnerTxID); err != nil {
				fmt.Printf("Warning: failed to write partner tx id: %v\n", err)
			}
		}
		if result.RPPAdapter.Workflow.WorkflowID != "" {
			line := fmt.Sprintf("state=%s attempt=%d", result.RPPAdapter.Workflow.GetFormattedState(), result.RPPAdapter.Workflow.Attempt)
			if _, err := fmt.Fprintf(w, "workflow_%s: %s run_id=%s\n", result.RPPAdapter.Workflow.WorkflowID, line, result.RPPAdapter.Workflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow: %v\n", err)
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	// Only show payment-engine section if we have data from payment-engine
	if peStatus != "" || result.PaymentEngine.Workflow.RunID != "" {
		if _, err := fmt.Fprintln(w, "[payment-engine]"); err != nil {
			fmt.Printf("Warning: failed to write payment-engine header: %v\n", err)
		}
		if peStatus != "" {
			if _, err := fmt.Fprintf(w, "status: %s\n", peStatus); err != nil {
				fmt.Printf("Warning: failed to write transfer status: %v\n", err)
			}
		}

		if result.PaymentEngine.Transfers.CreatedAt != "" {
			if _, err := fmt.Fprintf(w, "created_at: %s\n", result.PaymentEngine.Transfers.CreatedAt); err != nil {
				fmt.Printf("Warning: failed to write created at: %v\n", err)
			}
		}

		if result.PaymentEngine.Workflow.RunID != "" {
			line := fmt.Sprintf("state=%s attempt=%d", result.PaymentEngine.Workflow.GetFormattedState(), result.PaymentEngine.Workflow.Attempt)
			if _, err := fmt.Fprintf(w, "workflow_transfer_payment: %s run_id=%s\n", line, result.PaymentEngine.Workflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write payment engine workflow: %v\n", err)
			}
		}

		if _, err := fmt.Fprintln(w, "[payment-core]"); err != nil {
			fmt.Printf("Warning: failed to write payment-core header: %v\n", err)
		}

		if len(result.PaymentCore.InternalTxns) > 0 {
			var statuses []string
			for _, tx := range result.PaymentCore.InternalTxns {
				var parts []string
				if tx.TxType != "" {
					parts = append(parts, tx.TxType)
				}
				if tx.TxStatus != "" {
					parts = append(parts, tx.TxStatus)
				}
				if len(parts) > 0 {
					statuses = append(statuses, strings.Join(parts, " "))
				}
			}
			if len(statuses) > 0 {
				if _, err := fmt.Fprintf(w, "internal_transaction: %s\n", strings.Join(statuses, " , ")); err != nil {
					fmt.Printf("Warning: failed to write internal tx status: %v\n", err)
				}
			}
		}

		if len(result.PaymentCore.ExternalTxns) > 0 {
			var statuses []string
			for _, tx := range result.PaymentCore.ExternalTxns {
				var parts []string
				if tx.TxType != "" {
					parts = append(parts, tx.TxType)
				}
				if tx.TxStatus != "" {
					parts = append(parts, tx.TxStatus)
				}
				if len(parts) > 0 {
					statuses = append(statuses, strings.Join(parts, " "))
				}
			}
			if len(statuses) > 0 {
				if _, err := fmt.Fprintf(w, "external_transaction: %s\n", strings.Join(statuses, " , ")); err != nil {
					fmt.Printf("Warning: failed to write external tx status: %v\n", err)
				}
			}
		}

		for _, workflow := range result.PaymentCore.Workflow {
			line := fmt.Sprintf("state=%s attempt=%d", FormatWorkflowState(workflow.WorkflowID, workflow.State), workflow.Attempt)
			if _, err := fmt.Fprintf(w, "%s: %s run_id=%s\n", workflow.WorkflowID, line, workflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write payment core workflow: %v\n", err)
			}
		}
	}

	// Display RPP information if available (but don't duplicate if we already showed it for E2E IDs)
	if (result.RPPAdapter.Workflow.RunID != "" || result.RPPAdapter.ReqBizMsgID != "" || result.RPPAdapter.PartnerTxID != "") && !IsRppE2EID(result.TransactionID) {
		if _, err := fmt.Fprintln(w, "[rpp-adapter]"); err != nil {
			fmt.Printf("Warning: failed to write rpp-adapter header: %v\n", err)
		}
		if result.RPPAdapter.ReqBizMsgID != "" {
			if _, err := fmt.Fprintf(w, "req_biz_msg_id: %s\n", result.RPPAdapter.ReqBizMsgID); err != nil {
				fmt.Printf("Warning: failed to write req biz msg id: %v\n", err)
			}
		}
		if result.RPPAdapter.PartnerTxID != "" {
			if _, err := fmt.Fprintf(w, "partner_tx_id: %s\n", result.RPPAdapter.PartnerTxID); err != nil {
				fmt.Printf("Warning: failed to write partner tx id: %v\n", err)
			}
		}
		if result.RPPAdapter.Workflow.RunID != "" {
			line := fmt.Sprintf("state=%s attempt=%d", result.RPPAdapter.Workflow.GetFormattedState(), result.RPPAdapter.Workflow.Attempt)
			if _, err := fmt.Fprintf(w, "workflow_%s: %s run_id=%s\n", result.RPPAdapter.Workflow.WorkflowID, line, result.RPPAdapter.Workflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow: %v\n", err)
			}
		}
	}

	// Add classification section - always show it for valid transactions
	// Only show for transactions that have actual data (not NOT_FOUND or error)
	if result.Error == "" && !notFound {
		if _, err := fmt.Fprintln(w, "[Classification]"); err != nil {
			fmt.Printf("Warning: failed to write classification header: %v\n", err)
		}

		// Use the already identified case type for this transaction
		if result.CaseType != SOPCaseNone {
			if _, err := fmt.Fprintf(w, "%s\n", result.CaseType); err != nil {
				fmt.Printf("Warning: failed to write case type: %v\n", err)
			}
		} else {
			// Debug: print that no case was identified
			// This helps identify why some transactions don't show classifications
			if _, err := fmt.Fprintf(w, "no_case_matched\n"); err != nil {
				fmt.Printf("Warning: failed to write no case matched: %v\n", err)
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		fmt.Printf("Warning: failed to write final newline: %v\n", err)
	}
}

// WriteEcoTransactionResult writes a partnerpay-engine transaction result in the required format
func WriteEcoTransactionResult(w io.Writer, result TransactionResult, index int) {
	if index <= 0 {
		index = 1
	}

	// Check if this is a NOT_FOUND error
	if result.PartnerpayEngine.Transfers.Status == NotFoundStatus {
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nError: %s\n\n", index, result.TransactionID, result.Error); err != nil {
			fmt.Printf("Warning: failed to write error result: %v\n", err)
		}
		return
	}

	// Write the transaction ID header
	if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\n", index, result.TransactionID); err != nil {
		fmt.Printf("Warning: failed to write transaction ID: %v\n", err)
	}

	// Write the partnerpay-engine section
	if _, err := fmt.Fprintln(w, "[partnerpay-engine]"); err != nil {
		fmt.Printf("Warning: failed to write partnerpay-engine header: %v\n", err)
	}

	// Write the charge status
	if result.PartnerpayEngine.Transfers.Status != "" {
		if _, err := fmt.Fprintf(w, "charge.status: %s", result.PartnerpayEngine.Transfers.Status); err != nil {
			fmt.Printf("Warning: failed to write charge status: %v\n", err)
		}

		// If there's an error message (status_reason), append it
		if result.Error != "" {
			if _, err := fmt.Fprintf(w, " %s", result.Error); err != nil {
				fmt.Printf("Warning: failed to write error message: %v\n", err)
			}
		}

		// If there's a status reason description, append it
		if result.PartnerpayEngine.Transfers.StatusReasonDescription != "" {
			if _, err := fmt.Fprintf(w, " %s", result.PartnerpayEngine.Transfers.StatusReasonDescription); err != nil {
				fmt.Printf("Warning: failed to write status reason description: %v\n", err)
			}
		}

		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline after charge status: %v\n", err)
		}
	}

	// Write the workflow_charge information if available
	if result.PartnerpayEngine.Workflow.RunID != "" {
		line := fmt.Sprintf("workflow_charge: %s Attempt=%d run_id=%s",
			result.PartnerpayEngine.Workflow.GetFormattedState(),
			result.PartnerpayEngine.Workflow.Attempt,
			result.PartnerpayEngine.Workflow.RunID)

		if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
			fmt.Printf("Warning: failed to write workflow_charge: %v\n", err)
		}
	}

	// Add a final newline
	if _, err := fmt.Fprintln(w); err != nil {
		fmt.Printf("Warning: failed to write final newline: %v\n", err)
	}
}
