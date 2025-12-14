package adapters

import (
	"buddy/internal/txn/domain"
	"bytes"
	"fmt"
	"io"
	"os"
)

// WriteBatchResults writes transaction results to an output file in the new format
func WriteBatchResults(results []domain.TransactionResult, outputPath string) error {
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
// DEPRECATED: Use service.PrintTransactionStatusWithEnv instead
func PrintTransactionStatus(transactionID string) {
	// This function is no longer available in adapters package
	// Please use service.PrintTransactionStatusWithEnv(transactionID, "my") instead
	fmt.Printf("Error: PrintTransactionStatus moved to service package. Use service.PrintTransactionStatusWithEnv instead.\n")
}

// WriteResult writes a single transaction result to the specified writer
func WriteResult(w io.Writer, result domain.TransactionResult, index int) {
	writeResult(w, result, index)
}

// Helper function to display PaymentEngine section
func displayPaymentEngineSection(w io.Writer, pe domain.PaymentEngineInfo) error {
	// Check if we have any payment-engine data
	hasData := pe.Transfers.Status != "" || pe.Transfers.Type != "" || pe.Transfers.TxnSubtype != "" ||
		pe.Transfers.TxnDomain != "" || pe.Transfers.ReferenceID != "" || pe.Transfers.ExternalID != "" ||
		pe.Transfers.CreatedAt != "" || pe.Workflow.RunID != ""

	if !hasData {
		return nil
	}

	// Write section header
	if _, err := fmt.Fprintln(w, "[payment-engine]"); err != nil {
		return err
	}

	// Write transfer fields if not empty
	if pe.Transfers.Type != "" {
		if _, err := fmt.Fprintf(w, "type: %s\n", pe.Transfers.Type); err != nil {
			fmt.Printf("Warning: failed to write transfer type: %v\n", err)
		}
	}
	if pe.Transfers.TxnSubtype != "" {
		if _, err := fmt.Fprintf(w, "subtype: %s\n", pe.Transfers.TxnSubtype); err != nil {
			fmt.Printf("Warning: failed to write transfer subtype: %v\n", err)
		}
	}
	if pe.Transfers.TxnDomain != "" {
		if _, err := fmt.Fprintf(w, "domain: %s\n", pe.Transfers.TxnDomain); err != nil {
			fmt.Printf("Warning: failed to write transfer domain: %v\n", err)
		}
	}
	if pe.Transfers.Status != "" {
		if _, err := fmt.Fprintf(w, "status: %s\n", pe.Transfers.Status); err != nil {
			fmt.Printf("Warning: failed to write transfer status: %v\n", err)
		}
	}
	if pe.Transfers.CreatedAt != "" {
		if _, err := fmt.Fprintf(w, "created_at: %s\n", pe.Transfers.CreatedAt); err != nil {
			fmt.Printf("Warning: failed to write created at: %v\n", err)
		}
	}
	if pe.Transfers.ReferenceID != "" {
		if _, err := fmt.Fprintf(w, "reference_id: %s\n", pe.Transfers.ReferenceID); err != nil {
			fmt.Printf("Warning: failed to write reference id: %v\n", err)
		}
	}
	if pe.Transfers.ExternalID != "" {
		if _, err := fmt.Fprintf(w, "external_id: %s\n", pe.Transfers.ExternalID); err != nil {
			fmt.Printf("Warning: failed to write external id: %v\n", err)
		}
	}

	// Write workflow if exists
	if pe.Workflow.RunID != "" {
		if _, err := fmt.Fprintf(w, "workflow_transfer_payment:\n"); err != nil {
			fmt.Printf("Warning: failed to write workflow header: %v\n", err)
		}
		if _, err := fmt.Fprintf(w, "   state=%s attempt=%d\n", pe.Workflow.GetFormattedState(), pe.Workflow.Attempt); err != nil {
			fmt.Printf("Warning: failed to write workflow state: %v\n", err)
		}
		if _, err := fmt.Fprintf(w, "   run_id=%s\n", pe.Workflow.RunID); err != nil {
			fmt.Printf("Warning: failed to write workflow run id: %v\n", err)
		}
	}

	return nil
}

// Helper function to display PaymentCore section
func displayPaymentCoreSection(w io.Writer, pc domain.PaymentCoreInfo) error {
	// Check if we have any payment-core data
	hasData := len(pc.InternalTxns) > 0 || len(pc.ExternalTxns) > 0 || len(pc.Workflow) > 0

	// Always show payment-core section if payment-engine section was shown
	if _, err := fmt.Fprintln(w, "[payment-core]"); err != nil {
		return err
	}

	if hasData {
		// Display internal transactions
		for _, tx := range pc.InternalTxns {
			if _, err := fmt.Fprintf(w, "internal_transaction:\n"); err != nil {
				fmt.Printf("Warning: failed to write internal transaction header: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   tx_id=%s\n", tx.TxID); err != nil {
				fmt.Printf("Warning: failed to write internal tx id: %v\n", err)
			}
			if tx.GroupID != "" {
				if _, err := fmt.Fprintf(w, "   group_id=%s\n", tx.GroupID); err != nil {
					fmt.Printf("Warning: failed to write internal tx group id: %v\n", err)
				}
			}
			if tx.TxType != "" && tx.TxStatus != "" {
				if _, err := fmt.Fprintf(w, "   type=%s status=%s\n", tx.TxType, tx.TxStatus); err != nil {
					fmt.Printf("Warning: failed to write internal tx type and status: %v\n", err)
				}
			}
		}

		// Display external transactions
		for _, tx := range pc.ExternalTxns {
			if _, err := fmt.Fprintf(w, "external_transaction:\n"); err != nil {
				fmt.Printf("Warning: failed to write external transaction header: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   ref_id=%s\n", tx.RefID); err != nil {
				fmt.Printf("Warning: failed to write external tx ref id: %v\n", err)
			}
			if tx.GroupID != "" {
				if _, err := fmt.Fprintf(w, "   group_id=%s\n", tx.GroupID); err != nil {
					fmt.Printf("Warning: failed to write external tx group id: %v\n", err)
				}
			}
			if tx.TxType != "" && tx.TxStatus != "" {
				if _, err := fmt.Fprintf(w, "   type=%s status=%s\n", tx.TxType, tx.TxStatus); err != nil {
					fmt.Printf("Warning: failed to write external tx type and status: %v\n", err)
				}
			}
		}

		// Display workflows
		for _, workflow := range pc.Workflow {
			if _, err := fmt.Fprintf(w, "%s:\n", workflow.WorkflowID); err != nil {
				fmt.Printf("Warning: failed to write payment core workflow header: %v\n", err)
			}
			line := fmt.Sprintf("   state=%s attempt=%d", domain.FormatWorkflowState(workflow.WorkflowID, workflow.State), workflow.Attempt)
			if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
				fmt.Printf("Warning: failed to write payment core workflow state: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   run_id=%s\n", workflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write payment core workflow run id: %v\n", err)
			}
		}
	} else {
		// Show NOT_FOUND when no transactions exist
		if _, err := fmt.Fprintln(w, "NOT_FOUND"); err != nil {
			fmt.Printf("Warning: failed to write NOT_FOUND: %v\n", err)
		}
	}

	return nil
}

// Helper function to display FastAdapter section
func displayFastAdapterSection(w io.Writer, fa domain.FastAdapterInfo) error {
	// Check if we have any fast-adapter data
	if fa.InstructionID == "" {
		return nil
	}

	// Write section header
	if _, err := fmt.Fprintln(w, "[fast-adapter]"); err != nil {
		return err
	}

	// Write fields if not empty
	if _, err := fmt.Fprintf(w, "FAST ID: %s\n", fa.InstructionID); err != nil {
		fmt.Printf("Warning: failed to write FAST ID: %v\n", err)
	}

	if fa.Type != "" {
		if _, err := fmt.Fprintf(w, "type: %s\n", fa.Type); err != nil {
			fmt.Printf("Warning: failed to write fast adapter type: %v\n", err)
		}
	}

	// Always show status field, even if empty
	statusStr := fa.Status
	if fa.StatusCode > 0 {
		if statusStr != "" {
			statusStr = fmt.Sprintf("%s (%d)", statusStr, fa.StatusCode)
		} else {
			statusStr = fmt.Sprintf("(%d)", fa.StatusCode)
		}
	}
	if _, err := fmt.Fprintf(w, "status: %s\n", statusStr); err != nil {
		fmt.Printf("Warning: failed to write fast adapter status: %v\n", err)
	}

	if fa.CancelReasonCode != "" {
		if _, err := fmt.Fprintf(w, "cancel_reason_code: %s\n", fa.CancelReasonCode); err != nil {
			fmt.Printf("Warning: failed to write cancel reason code: %v\n", err)
		}
	}

	if fa.RejectReasonCode != "" {
		if _, err := fmt.Fprintf(w, "reject_reason_code: %s\n", fa.RejectReasonCode); err != nil {
			fmt.Printf("Warning: failed to write reject reason code: %v\n", err)
		}
	}

	if fa.CreatedAt != "" {
		if _, err := fmt.Fprintf(w, "created_at: %s\n", fa.CreatedAt); err != nil {
			fmt.Printf("Warning: failed to write fast adapter created at: %v\n", err)
		}
	}

	return nil
}

// Helper function to display RPPAdapter section
func displayRPPAdapterSection(w io.Writer, ra domain.RPPAdapterInfo, isE2EID bool, transactionID string) error {
	// For E2E ID mode, we always show the section
	if isE2EID {
		if _, err := fmt.Fprintln(w, "[rpp-adapter]"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "e2e_id: %s\n", transactionID); err != nil {
			fmt.Printf("Warning: failed to write e2e_id: %v\n", err)
		}
		if ra.Status != "" {
			if _, err := fmt.Fprintf(w, "credit_transfer.status: %s\n", ra.Status); err != nil {
				fmt.Printf("Warning: failed to write rpp status: %v\n", err)
			}
		}
		if ra.PartnerTxID != "" {
			if _, err := fmt.Fprintf(w, "partner_tx_id: %s\n", ra.PartnerTxID); err != nil {
				fmt.Printf("Warning: failed to write partner tx id: %v\n", err)
			}
		}
		if ra.Workflow.RunID != "" {
			if _, err := fmt.Fprintf(w, "%s:\n", ra.Workflow.WorkflowID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow header: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   state=%s attempt=%d\n", ra.Workflow.GetFormattedState(), ra.Workflow.Attempt); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow state: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   run_id=%s\n", ra.Workflow.RunID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow run id: %v\n", err)
			}
		}
		return nil
	}

	// For non-E2E IDs, check if we have data
	hasData := ra.Workflow.RunID != "" || ra.ReqBizMsgID != "" || ra.PartnerTxID != "" || ra.Info != ""
	if !hasData {
		return nil
	}

	// Write section header
	if _, err := fmt.Fprintln(w, "[rpp-adapter]"); err != nil {
		return err
	}

	// Write fields if not empty
	if ra.ReqBizMsgID != "" {
		if _, err := fmt.Fprintf(w, "req_biz_msg_id: %s\n", ra.ReqBizMsgID); err != nil {
			fmt.Printf("Warning: failed to write req biz msg id: %v\n", err)
		}
	}
	if ra.PartnerTxID != "" {
		if _, err := fmt.Fprintf(w, "partner_tx_id: %s\n", ra.PartnerTxID); err != nil {
			fmt.Printf("Warning: failed to write partner tx id: %v\n", err)
		}
	}
	if ra.Workflow.RunID != "" {
		if _, err := fmt.Fprintf(w, "%s:\n", ra.Workflow.WorkflowID); err != nil {
			fmt.Printf("Warning: failed to write rpp workflow header: %v\n", err)
		}
		if _, err := fmt.Fprintf(w, "   state=%s attempt=%d\n", ra.Workflow.GetFormattedState(), ra.Workflow.Attempt); err != nil {
			fmt.Printf("Warning: failed to write rpp workflow state: %v\n", err)
		}
		if _, err := fmt.Fprintf(w, "   run_id=%s\n", ra.Workflow.RunID); err != nil {
			fmt.Printf("Warning: failed to write rpp workflow run id: %v\n", err)
		}
	}
	if ra.Info != "" {
		if _, err := fmt.Fprintf(w, "info: %s\n", ra.Info); err != nil {
			fmt.Printf("Warning: failed to write rpp info: %v\n", err)
		}
	}

	return nil
}

// Helper function to display Classification section
func displayClassificationSection(w io.Writer, result domain.TransactionResult) error {
	// Always write section header
	if _, err := fmt.Fprintln(w, "[Classification]"); err != nil {
		return err
	}

	// Show NOT_FOUND for empty case types
	caseType := result.CaseType
	if caseType == "" || caseType == domain.CaseNone {
		caseType = "NOT_FOUND"
	}

	if _, err := fmt.Fprintf(w, "%s\n", caseType); err != nil {
		fmt.Printf("Warning: failed to write case type: %v\n", err)
	}

	return nil
}

func writeResult(w io.Writer, result domain.TransactionResult, index int) {
	if index <= 0 {
		index = 1
	}

	// For RPP E2E IDs with errors, still show the RPP adapter section with NOT FOUND status
	if domain.IsRppE2EID(result.TransactionID) && result.Error != "" {
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

	// Check if there's an error (but NOT_FOUND is not an error for display purposes)
	if result.Error != "" {
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nError: %s\n\n", index, result.TransactionID, result.Error); err != nil {
			fmt.Printf("Warning: failed to write error result: %v\n", err)
		}
		return
	}

	// Use e2e_id for RPP E2E IDs, transaction_id for others

	idLabel := "transaction_id"
	if domain.IsRppE2EID(result.TransactionID) {
		idLabel = "e2e_id"
	}
	if _, err := fmt.Fprintf(w, "### [%d] %s: %s\n", index, idLabel, result.TransactionID); err != nil {
		fmt.Printf("Warning: failed to write ID: %v\n", err)
	}

	// If this is an E2E ID lookup, show RPP adapter info first
	if domain.IsRppE2EID(result.TransactionID) {
		if err := displayRPPAdapterSection(w, result.RPPAdapter, true, result.TransactionID); err != nil {
			fmt.Printf("Warning: failed to display RPP adapter section: %v\n", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	// Display sections based on data availability
	if err := displayPaymentEngineSection(w, result.PaymentEngine); err == nil {
		// payment-engine was shown, now show payment-core
		if err := displayPaymentCoreSection(w, result.PaymentCore); err != nil {
			fmt.Printf("Warning: failed to display payment core section: %v\n", err)
		}
	}

	if err := displayFastAdapterSection(w, result.FastAdapter); err != nil {
		fmt.Printf("Warning: failed to display fast adapter section: %v\n", err)
	}

	// Display RPP adapter for non-E2E IDs
	if !domain.IsRppE2EID(result.TransactionID) {
		if err := displayRPPAdapterSection(w, result.RPPAdapter, false, result.TransactionID); err != nil {
			fmt.Printf("Warning: failed to display RPP adapter section: %v\n", err)
		}
	}

	if err := displayClassificationSection(w, result); err != nil {
		fmt.Printf("Warning: failed to display classification section: %v\n", err)
	}

	if _, err := fmt.Fprintln(w); err != nil {
		fmt.Printf("Warning: failed to write final newline: %v\n", err)
	}
}

// WriteEcoTransactionResult writes a partnerpay-engine transaction result in the required format
func WriteEcoTransactionResult(w io.Writer, result domain.TransactionResult, index int) {
	WriteEcoTransactionInfo(w, result.PartnerpayEngine, result.TransactionID, index)
}

// WriteEcoTransactionInfo writes a partnerpay-engine transaction info in the required format
func WriteEcoTransactionInfo(w io.Writer, info domain.PartnerpayEngineInfo, transactionID string, index int) {
	if index <= 0 {
		index = 1
	}

	// Check if this is a NOT_FOUND error
	if info.Transfers.Status == domain.NotFoundStatus {
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nError: transaction not found\n\n", index, transactionID); err != nil {
			fmt.Printf("Warning: failed to write error result: %v\n", err)
		}
		return
	}

	// Write the transaction ID header
	if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\n", index, transactionID); err != nil {
		fmt.Printf("Warning: failed to write transaction ID: %v\n", err)
	}

	// Write the partnerpay-engine section
	if _, err := fmt.Fprintln(w, "[partnerpay-engine]"); err != nil {
		fmt.Printf("Warning: failed to write partnerpay-engine header: %v\n", err)
	}

	// Write the charge status
	if info.Transfers.Status != "" {
		if _, err := fmt.Fprintf(w, "charge.status: %s", info.Transfers.Status); err != nil {
			fmt.Printf("Warning: failed to write charge status: %v\n", err)
		}

		// If there's a status reason, append it
		if info.Transfers.StatusReason != "" {
			if _, err := fmt.Fprintf(w, " %s", info.Transfers.StatusReason); err != nil {
				fmt.Printf("Warning: failed to write status reason: %v\n", err)
			}
		}

		// If there's a status reason description, append it
		if info.Transfers.StatusReasonDescription != "" {
			if _, err := fmt.Fprintf(w, " %s", info.Transfers.StatusReasonDescription); err != nil {
				fmt.Printf("Warning: failed to write status reason description: %v\n", err)
			}
		}

		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline after charge status: %v\n", err)
		}
	}

	// Write the workflow_charge information if available
	if info.Workflow.RunID != "" {
		line := fmt.Sprintf("workflow_charge: %s Attempt=%d run_id=%s",
			info.Workflow.GetFormattedState(),
			info.Workflow.Attempt,
			info.Workflow.RunID)

		if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
			fmt.Printf("Warning: failed to write workflow_charge: %v\n", err)
		}
	}

	// Add a final newline
	if _, err := fmt.Fprintln(w); err != nil {
		fmt.Printf("Warning: failed to write final newline: %v\n", err)
	}
}
