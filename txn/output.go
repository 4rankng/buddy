package txn

import (
	"fmt"
	"io"
	"os"
)

// WriteBatchResults writes transaction results to an output file in the new format
func WriteBatchResults(results []TransactionResult, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close output file %s: %v\n", outputPath, err)
		}
	}()

	for idx, result := range results {
		writeResult(file, result, idx+1)
	}

	return nil
}

// PrintTransactionStatus prints transaction information in the new format
func PrintTransactionStatus(transactionID string) {
	result := QueryTransactionStatus(transactionID)
	writeResult(os.Stdout, *result, 1)
}

func writeResult(w io.Writer, result TransactionResult, index int) {
	if index <= 0 {
		index = 1
	}

	if result.Error != "" && result.TransferStatus != "NOT_FOUND" {
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nError: %s\n\n", index, result.TransactionID, result.Error); err != nil {
			fmt.Printf("Warning: failed to write error result: %v\n", err)
		}
		return
	}

	if result.TransferStatus == "NOT_FOUND" {
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nStatus: NOT_FOUND\n\n", index, result.TransactionID); err != nil {
			fmt.Printf("Warning: failed to write not found result: %v\n", err)
		}
		return
	}

	if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\n", index, result.TransactionID); err != nil {
		fmt.Printf("Warning: failed to write transaction ID: %v\n", err)
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
		if result.RPPStatus != "" {
			if _, err := fmt.Fprintf(w, "credit_transfer.status: %s\n", result.RPPStatus); err != nil {
				fmt.Printf("Warning: failed to write rpp status: %v\n", err)
			}
		}
		if result.PartnerTxID != "" {
			if _, err := fmt.Fprintf(w, "partner_tx_id: %s\n", result.PartnerTxID); err != nil {
				fmt.Printf("Warning: failed to write partner tx id: %v\n", err)
			}
		}
		if result.RPPWorkflow.Type != "" {
			line := fmt.Sprintf("state=%s attempt=%d", result.RPPWorkflow.GetFormattedState(), result.RPPWorkflow.Attempt)
			if _, err := fmt.Fprintf(w, "workflow_%s: %s run_id=%s\n", result.RPPWorkflow.Type, line, result.RPPWorkflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow: %v\n", err)
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	// Only show payment-engine section if we have data from payment-engine
	if result.TransferStatus != "" || result.PaymentEngineWorkflow.RunID != "" {
		if _, err := fmt.Fprintln(w, "[payment-engine]"); err != nil {
			fmt.Printf("Warning: failed to write payment-engine header: %v\n", err)
		}
		if result.TransferStatus != "" {
			if _, err := fmt.Fprintf(w, "status: %s\n", result.TransferStatus); err != nil {
				fmt.Printf("Warning: failed to write transfer status: %v\n", err)
			}
		}

		if result.CreatedAt != "" {
			if _, err := fmt.Fprintf(w, "created_at: %s\n", result.CreatedAt); err != nil {
				fmt.Printf("Warning: failed to write created at: %v\n", err)
			}
		}

		if result.PaymentEngineWorkflow.RunID != "" {
			line := fmt.Sprintf("state=%s attempt=%d", result.PaymentEngineWorkflow.GetFormattedState(), result.PaymentEngineWorkflow.Attempt)
			if _, err := fmt.Fprintf(w, "workflow_transfer_payment: %s run_id=%s\n", line, result.PaymentEngineWorkflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write payment engine workflow: %v\n", err)
			}
		}

		if _, err := fmt.Fprintln(w, "[payment-core]"); err != nil {
			fmt.Printf("Warning: failed to write payment-core header: %v\n", err)
		}

		if result.InternalTxStatus != "" {
			if _, err := fmt.Fprintf(w, "internal_transaction: %s\n", result.InternalTxStatus); err != nil {
				fmt.Printf("Warning: failed to write internal tx status: %v\n", err)
			}
		}

		if result.ExternalTxStatus != "" {
			if _, err := fmt.Fprintf(w, "external_transaction: %s\n", result.ExternalTxStatus); err != nil {
				fmt.Printf("Warning: failed to write external tx status: %v\n", err)
			}
		}

		for _, workflow := range result.PaymentCoreWorkflows {
			line := fmt.Sprintf("state=%s attempt=%d", FormatWorkflowState(workflow.Type, workflow.State), workflow.Attempt)
			if _, err := fmt.Fprintf(w, "payment_core_workflow_%s: %s run_id=%s\n", workflow.Type, line, workflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write payment core workflow: %v\n", err)
			}
		}
	}

	// Display RPP information if available (but don't duplicate if we already showed it for E2E IDs)
	if (result.RPPWorkflow.RunID != "" || result.ReqBizMsgID != "" || result.PartnerTxID != "") && !IsRppE2EID(result.TransactionID) {
		if _, err := fmt.Fprintln(w, "[rpp-adapter]"); err != nil {
			fmt.Printf("Warning: failed to write rpp-adapter header: %v\n", err)
		}
		if result.ReqBizMsgID != "" {
			if _, err := fmt.Fprintf(w, "req_biz_msg_id: %s\n", result.ReqBizMsgID); err != nil {
				fmt.Printf("Warning: failed to write req biz msg id: %v\n", err)
			}
		}
		if result.PartnerTxID != "" {
			if _, err := fmt.Fprintf(w, "partner_tx_id: %s\n", result.PartnerTxID); err != nil {
				fmt.Printf("Warning: failed to write partner tx id: %v\n", err)
			}
		}
		if result.RPPWorkflow.RunID != "" {
			line := fmt.Sprintf("state=%s attempt=%d", result.RPPWorkflow.GetFormattedState(), result.RPPWorkflow.Attempt)
			if _, err := fmt.Fprintf(w, "workflow_%s: %s run_id=%s\n", result.RPPWorkflow.Type, line, result.RPPWorkflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow: %v\n", err)
			}
		}
	}

	// Add classification section
	if _, err := fmt.Fprintln(w, "[Classification]"); err != nil {
		fmt.Printf("Warning: failed to write classification header: %v\n", err)
	}

	// Use the already identified case type for this transaction
	if result.SOPCase != SOPCaseNone {
		if _, err := fmt.Fprintf(w, "%s\n", result.SOPCase); err != nil {
			fmt.Printf("Warning: failed to write case type: %v\n", err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		fmt.Printf("Warning: failed to write final newline: %v\n", err)
	}
}
