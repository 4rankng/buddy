package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteDMLFiles writes deploy and rollback DML files for the transaction
func WriteDMLFiles(result domain.TransactionResult, outputDir string) error {
	if result.PaymentCore == nil {
		return fmt.Errorf("no payment-core data to write DML files")
	}

	// Create output directory if it doesn't exist
	if outputDir == "" {
		outputDir = "output"
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate filenames with timestamp
	timestamp := time.Now().Format("20060102_150405")
	deployFile := filepath.Join(outputDir, fmt.Sprintf("deploy_%s_%s.sql", result.InputID, timestamp))
	rollbackFile := filepath.Join(outputDir, fmt.Sprintf("rollback_%s_%s.sql", result.InputID, timestamp))

	// Write deploy file
	if err := writeDeployFile(deployFile, result); err != nil {
		return fmt.Errorf("failed to write deploy file: %v", err)
	}

	// Write rollback file
	if err := writeRollbackFile(rollbackFile, result); err != nil {
		return fmt.Errorf("failed to write rollback file: %v", err)
	}

	fmt.Printf("DML files written:\n")
	fmt.Printf("  Deploy:   %s\n", deployFile)
	fmt.Printf("  Rollback: %s\n", rollbackFile)

	return nil
}

// writeDeployFile writes the deploy DML file
func writeDeployFile(filename string, result domain.TransactionResult) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Write header
	if _, err := fmt.Fprintf(file, "-- Deploy DML for transaction %s\n", result.InputID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "-- Generated at: %s\n\n", time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	// Write payment-core internal transactions
	if result.PaymentCore.InternalCapture.TxID != "" {
		if _, err := fmt.Fprintf(file, "-- Internal Capture (AUTH) Transaction\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE internal_transaction SET status = 'COMPLETED' WHERE tx_id = '%s' AND group_id = '%s';\n\n",
			result.PaymentCore.InternalCapture.TxID, result.PaymentCore.InternalCapture.GroupID); err != nil {
			return err
		}
	}

	if result.PaymentCore.InternalAuth.TxID != "" {
		if _, err := fmt.Fprintf(file, "-- Internal Auth (CAPTURE) Transaction\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE internal_transaction SET status = 'COMPLETED' WHERE tx_id = '%s' AND group_id = '%s';\n\n",
			result.PaymentCore.InternalAuth.TxID, result.PaymentCore.InternalAuth.GroupID); err != nil {
			return err
		}
	}

	// Write payment-core external transactions
	if result.PaymentCore.ExternalTransfer.RefID != "" {
		if _, err := fmt.Fprintf(file, "-- External Transfer Transaction\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE external_transaction SET status = 'COMPLETED' WHERE ref_id = '%s' AND group_id = '%s';\n\n",
			result.PaymentCore.ExternalTransfer.RefID, result.PaymentCore.ExternalTransfer.GroupID); err != nil {
			return err
		}
	}

	// Write partnerpay-engine update
	if result.PartnerpayEngine != nil && result.PartnerpayEngine.Transfers.TransactionID != "" {
		if _, err := fmt.Fprintf(file, "-- Partnerpay-engine Charge Update\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE charge SET status = 'COMPLETED' WHERE transaction_id = '%s';\n\n",
			result.PartnerpayEngine.Transfers.TransactionID); err != nil {
			return err
		}
	}

	return nil
}

// writeRollbackFile writes the rollback DML file
func writeRollbackFile(filename string, result domain.TransactionResult) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Write header
	if _, err := fmt.Fprintf(file, "-- Rollback DML for transaction %s\n", result.InputID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "-- Generated at: %s\n\n", time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	// Write payment-core internal transactions rollback
	if result.PaymentCore.InternalCapture.TxID != "" {
		if _, err := fmt.Fprintf(file, "-- Internal Capture (AUTH) Transaction Rollback\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE internal_transaction SET status = '%s' WHERE tx_id = '%s' AND group_id = '%s';\n\n",
			result.PaymentCore.InternalCapture.TxStatus, result.PaymentCore.InternalCapture.TxID, result.PaymentCore.InternalCapture.GroupID); err != nil {
			return err
		}
	}

	if result.PaymentCore.InternalAuth.TxID != "" {
		if _, err := fmt.Fprintf(file, "-- Internal Auth (CAPTURE) Transaction Rollback\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE internal_transaction SET status = '%s' WHERE tx_id = '%s' AND group_id = '%s';\n\n",
			result.PaymentCore.InternalAuth.TxStatus, result.PaymentCore.InternalAuth.TxID, result.PaymentCore.InternalAuth.GroupID); err != nil {
			return err
		}
	}

	// Write payment-core external transactions rollback
	if result.PaymentCore.ExternalTransfer.RefID != "" {
		if _, err := fmt.Fprintf(file, "-- External Transfer Transaction Rollback\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE external_transaction SET status = '%s' WHERE ref_id = '%s' AND group_id = '%s';\n\n",
			result.PaymentCore.ExternalTransfer.TxStatus, result.PaymentCore.ExternalTransfer.RefID, result.PaymentCore.ExternalTransfer.GroupID); err != nil {
			return err
		}
	}

	// Write partnerpay-engine rollback
	if result.PartnerpayEngine != nil && result.PartnerpayEngine.Transfers.TransactionID != "" {
		if _, err := fmt.Fprintf(file, "-- Partnerpay-engine Charge Rollback\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(file, "UPDATE charge SET status = '%s' WHERE transaction_id = '%s';\n\n",
			result.PartnerpayEngine.Transfers.Status, result.PartnerpayEngine.Transfers.TransactionID); err != nil {
			return err
		}
	}

	return nil
}
