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
	if domain.IsRppE2EID(result.InputID) && result.Error != "" {
		if _, err := fmt.Fprintf(w, "### [%d] e2e_id: %s\n", index, result.InputID); err != nil {
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
		if _, err := fmt.Fprintf(w, "### [%d] transaction_id: %s\nError: %s\n\n", index, result.InputID, result.Error); err != nil {
			fmt.Printf("Warning: failed to write error result: %v\n", err)
		}
		return
	}

	// Use e2e_id for RPP E2E IDs, transaction_id for others

	idLabel := "transaction_id"
	if domain.IsRppE2EID(result.InputID) {
		idLabel = "e2e_id"
	}
	if _, err := fmt.Fprintf(w, "### [%d] %s: %s\n", index, idLabel, result.InputID); err != nil {
		fmt.Printf("Warning: failed to write ID: %v\n", err)
	}

	// Display sections in fix order: PaymentEngine -> PaymentCore -> FastAdapter/RPPAdapter -> PartnerpayEngine

	// 1. Payment Engine
	if result.PaymentEngine != nil {
		if err := displayPaymentEngineSection(w, *result.PaymentEngine); err != nil {
			fmt.Printf("Warning: failed to display payment engine section: %v\n", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	// 2. Payment Core
	if result.PaymentCore != nil {
		if err := displayPaymentCoreSection(w, *result.PaymentCore); err != nil {
			fmt.Printf("Warning: failed to display payment core section: %v\n", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	// 3. Fast Adapter (SG) or RPP Adapter (MY)
	if result.FastAdapter != nil {
		if err := displayFastAdapterSection(w, *result.FastAdapter); err != nil {
			fmt.Printf("Warning: failed to display fast adapter section: %v\n", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	if result.RPPAdapter != nil {
		if err := displayRPPAdapterSection(w, *result.RPPAdapter, domain.IsRppE2EID(result.InputID), result.InputID); err != nil {
			fmt.Printf("Warning: failed to display RPP adapter section: %v\n", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	// 4. Partnerpay Engine
	if result.PartnerpayEngine != nil {
		// Create a minimal TransactionResult with just the partnerpay-engine info
		peResult := domain.TransactionResult{
			InputID:          result.InputID,
			PartnerpayEngine: result.PartnerpayEngine,
		}
		WriteEcoTransactionInfo(w, peResult, result.InputID, index)
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	if err := displayClassificationSection(w, result); err != nil {
		fmt.Printf("Warning: failed to display classification section: %v\n", err)
	}

	if _, err := fmt.Fprintln(w); err != nil {
		fmt.Printf("Warning: failed to write final newline: %v\n", err)
	}
}
