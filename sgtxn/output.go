package sgtxn

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PrintSGTransactionStatus prints a single transaction result to stdout
func PrintSGTransactionStatus(result SGTransactionResult, index int) {
	WriteSGResult(os.Stdout, result, index)
}

// writef is a helper function that writes formatted text and handles errors
func writef(w io.Writer, format string, args ...interface{}) {
	// For CLI output, write errors are ignored to avoid cluttering the output
	// This is especially true when writing to stdout which is rarely error-prone
	_, _ = fmt.Fprintf(w, format, args...)
}

// WriteSGResult writes a transaction result in the specified format
func WriteSGResult(w io.Writer, result SGTransactionResult, index int) {
	// Header
	writef(w, "### [%d] transaction_id: %s\n", index, result.TransactionID)

	// Payment Engine section
	writef(w, "[payment-engine]\n")
	if result.Error != "" {
		writef(w, "error: %s\n", result.Error)
	} else {
		writef(w, "status: %s\n", result.TransferStatus)
		if result.CreatedAt != "" {
			writef(w, "created_at: %s\n", result.CreatedAt)
		}
		if result.PaymentEngineWorkflow.Type != "" {
			writef(w, "%s: state=%s attempt=%d run_id=%s\n",
				result.PaymentEngineWorkflow.Type,
				result.PaymentEngineWorkflow.State,
				result.PaymentEngineWorkflow.Attempt,
				result.PaymentEngineWorkflow.RunID)
		}
	}

	// Payment Core section
	writef(w, "[payment-core]\n")
	if len(result.InternalTxStatuses) > 0 {
		statuses := make([]string, len(result.InternalTxStatuses))
		for i, status := range result.InternalTxStatuses {
			statuses[i] = mapInternalStatus(status)
		}
		writef(w, "internal_transaction: %s\n", strings.Join(statuses, " | "))
	}
	if len(result.ExternalTxStatuses) > 0 {
		statuses := make([]string, len(result.ExternalTxStatuses))
		for i, status := range result.ExternalTxStatuses {
			statuses[i] = mapExternalStatus(status)
		}
		writef(w, "external_transaction: %s\n", strings.Join(statuses, " | "))
	}
	for _, workflow := range result.PaymentCoreWorkflows {
		writef(w, "%s: state=%s attempt=%d run_id=%s\n",
			workflow.Type,
			workflow.State,
			workflow.Attempt,
			workflow.RunID)
	}

	// Fast Adapter section
	writef(w, "[fast-adapter]\n")
	if result.ExternalID != "" {
		writef(w, "FAST ID: %s\n", result.ExternalID)
	}
	if result.FastAdapterType != "" {
		writef(w, "type: %s\n", result.FastAdapterType)
	}
	if result.FastAdapterStatus != "" {
		writef(w, "transactions.status: %s\n", result.FastAdapterStatus)
	}
	if result.FastAdapterCancelCode != "" {
		writef(w, "cancel_reason_code: %s\n", result.FastAdapterCancelCode)
	}
	if result.FastAdapterRejectCode != "" {
		writef(w, "reject_reason_code: %s\n", result.FastAdapterRejectCode)
	}

	// Classification section
	writef(w, "[Classification]\n")
	writef(w, "no_case_matched\n")

	// Add spacing between transactions
	writef(w, "\n\n")
}

// mapInternalStatus maps internal transaction status
func mapInternalStatus(status string) string {
	switch status {
	case "AUTH_SUCCESS":
		return "AUTH SUCCESS"
	case "CAPTURE_SUCCESS":
		return "CAPTURE SUCCESS"
	case "AUTH_FAILED":
		return "AUTH FAILED"
	case "CAPTURE_FAILED":
		return "CAPTURE FAILED"
	case "AUTH_PROCESSING":
		return "AUTH PROCESSING"
	case "CAPTURE_PROCESSING":
		return "CAPTURE PROCESSING"
	default:
		return status
	}
}

// mapExternalStatus maps external transaction status
func mapExternalStatus(status string) string {
	switch status {
	case "TRANSFER_SUCCESS":
		return "TRANSFER SUCCESS"
	case "TRANSFER_FAILED":
		return "TRANSFER FAILED"
	case "TRANSFER_PROCESSING":
		return "TRANSFER PROCESSING"
	case "TRANSFER_CANCELED":
		return "TRANSFER CANCELED"
	default:
		return status
	}
}
