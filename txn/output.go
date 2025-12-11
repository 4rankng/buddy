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
	defer file.Close()

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
		fmt.Fprintf(w, "### [%d] transaction_id: %s\nError: %s\n\n", index, result.TransactionID, result.Error)
		return
	}

	if result.TransferStatus == "NOT_FOUND" {
		fmt.Fprintf(w, "### [%d] transaction_id: %s\nStatus: NOT_FOUND\n\n", index, result.TransactionID)
		return
	}

	fmt.Fprintf(w, "### [%d] transaction_id: %s\n", index, result.TransactionID)
	fmt.Fprintln(w, "[payment-engine]")
	fmt.Fprintf(w, "status: %s\n", result.TransferStatus)

	if result.CreatedAt != "" {
		fmt.Fprintf(w, "created_at: %s\n", result.CreatedAt)
	}

	if result.PaymentEngineWorkflow.RunID != "" {
		line := fmt.Sprintf("state=%s attempt=%d", result.PaymentEngineWorkflow.State, result.PaymentEngineWorkflow.Attempt)
		fmt.Fprintf(w, "workflow_transfer_payment: %s run_id=%s\n", line, result.PaymentEngineWorkflow.RunID)
	}

	fmt.Fprintln(w, "[payment-core]")

	if result.InternalTxStatus != "" {
		fmt.Fprintf(w, "internal_transaction: %s\n", result.InternalTxStatus)
	}

	if result.ExternalTxStatus != "" {
		fmt.Fprintf(w, "external_transaction: %s\n", result.ExternalTxStatus)
	}

	for _, workflow := range result.PaymentCoreWorkflows {
		line := fmt.Sprintf("state=%s attempt=%d", workflow.State, workflow.Attempt)
		fmt.Fprintf(w, "payment_core_workflow_%s: %s run_id=%s\n", workflow.Type, line, workflow.RunID)
	}

	fmt.Fprintln(w)
}
