package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"io"
)

// displayPaymentEngineSection displays PaymentEngine section
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
		// Use the actual workflow ID from the data
		if _, err := fmt.Fprintf(w, "%s:\n", pe.Workflow.WorkflowID); err != nil {
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

// displayPaymentCoreSection displays PaymentCore section
func displayPaymentCoreSection(w io.Writer, pc domain.PaymentCoreInfo) error {
	// Check if we have any payment-core data
	hasData := pc.InternalCapture.TxID != "" || pc.InternalAuth.TxID != "" || pc.ExternalTransfer.RefID != ""

	// Always show payment-core section if payment-engine section was shown
	if _, err := fmt.Fprintln(w, "[payment-core]"); err != nil {
		return err
	}

	if hasData {
		// Display Internal CAPTURE transaction
		if pc.InternalCapture.TxID != "" {
			// Show status on the same line as transaction type
			if _, err := fmt.Fprintf(w, "internal_capture: %s\n", pc.InternalCapture.TxStatus); err != nil {
				fmt.Printf("Warning: failed to write internal transaction header: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   tx_id=%s\n", pc.InternalCapture.TxID); err != nil {
				fmt.Printf("Warning: failed to write internal tx id: %v\n", err)
			}
			if pc.InternalCapture.GroupID != "" {
				if _, err := fmt.Fprintf(w, "   group_id=%s\n", pc.InternalCapture.GroupID); err != nil {
					fmt.Printf("Warning: failed to write internal tx group id: %v\n", err)
				}
			}
			if pc.InternalCapture.TxType != "" {
				if _, err := fmt.Fprintf(w, "   type=%s\n", pc.InternalCapture.TxType); err != nil {
					fmt.Printf("Warning: failed to write internal tx type: %v\n", err)
				}
			}
			// Always display error fields
			if _, err := fmt.Fprintf(w, "   error_code='%s' error_msg='%s'\n", pc.InternalCapture.ErrorCode, pc.InternalCapture.ErrorMsg); err != nil {
				fmt.Printf("Warning: failed to write internal tx error fields: %v\n", err)
			}
			// Display workflow for this transaction
			if pc.InternalCapture.Workflow.WorkflowID != "" {
				if _, err := fmt.Fprintf(w, "   workflow:\n"); err != nil {
					fmt.Printf("Warning: failed to write workflow header: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "      workflow_id=%s\n", pc.InternalCapture.Workflow.WorkflowID); err != nil {
					fmt.Printf("Warning: failed to write workflow id: %v\n", err)
				}
				line := fmt.Sprintf("      state=%s attempt=%d", domain.FormatWorkflowState(pc.InternalCapture.Workflow.WorkflowID, pc.InternalCapture.Workflow.State), pc.InternalCapture.Workflow.Attempt)
				if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
					fmt.Printf("Warning: failed to write workflow state: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "      run_id=%s\n", pc.InternalCapture.Workflow.RunID); err != nil {
					fmt.Printf("Warning: failed to write workflow run id: %v\n", err)
				}
			}
		}

		// Display Internal AUTH transaction
		if pc.InternalAuth.TxID != "" {
			// Show status on the same line as transaction type
			if _, err := fmt.Fprintf(w, "internal_auth: %s\n", pc.InternalAuth.TxStatus); err != nil {
				fmt.Printf("Warning: failed to write internal transaction header: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   tx_id=%s\n", pc.InternalAuth.TxID); err != nil {
				fmt.Printf("Warning: failed to write internal tx id: %v\n", err)
			}
			if pc.InternalAuth.GroupID != "" {
				if _, err := fmt.Fprintf(w, "   group_id=%s\n", pc.InternalAuth.GroupID); err != nil {
					fmt.Printf("Warning: failed to write internal tx group id: %v\n", err)
				}
			}
			if pc.InternalAuth.TxType != "" {
				if _, err := fmt.Fprintf(w, "   type=%s\n", pc.InternalAuth.TxType); err != nil {
					fmt.Printf("Warning: failed to write internal tx type: %v\n", err)
				}
			}
			// Always display error fields
			if _, err := fmt.Fprintf(w, "   error_code='%s' error_msg='%s'\n", pc.InternalAuth.ErrorCode, pc.InternalAuth.ErrorMsg); err != nil {
				fmt.Printf("Warning: failed to write internal tx error fields: %v\n", err)
			}
			// Display workflow for this transaction
			if pc.InternalAuth.Workflow.WorkflowID != "" {
				if _, err := fmt.Fprintf(w, "   workflow:\n"); err != nil {
					fmt.Printf("Warning: failed to write workflow header: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "      workflow_id=%s\n", pc.InternalAuth.Workflow.WorkflowID); err != nil {
					fmt.Printf("Warning: failed to write workflow id: %v\n", err)
				}
				line := fmt.Sprintf("      state=%s attempt=%d", domain.FormatWorkflowState(pc.InternalAuth.Workflow.WorkflowID, pc.InternalAuth.Workflow.State), pc.InternalAuth.Workflow.Attempt)
				if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
					fmt.Printf("Warning: failed to write workflow state: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "      run_id=%s\n", pc.InternalAuth.Workflow.RunID); err != nil {
					fmt.Printf("Warning: failed to write workflow run id: %v\n", err)
				}
			}
		}

		// Display External Transfer transaction (TRANSFER)
		if pc.ExternalTransfer.RefID != "" {
			if _, err := fmt.Fprintf(w, "external_transaction:\n"); err != nil {
				fmt.Printf("Warning: failed to write external transaction header: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   ref_id=%s\n", pc.ExternalTransfer.RefID); err != nil {
				fmt.Printf("Warning: failed to write external tx ref id: %v\n", err)
			}
			if pc.ExternalTransfer.GroupID != "" {
				if _, err := fmt.Fprintf(w, "   group_id=%s\n", pc.ExternalTransfer.GroupID); err != nil {
					fmt.Printf("Warning: failed to write external tx group id: %v\n", err)
				}
			}
			if pc.ExternalTransfer.TxType != "" && pc.ExternalTransfer.TxStatus != "" {
				if _, err := fmt.Fprintf(w, "   type=%s status=%s\n", pc.ExternalTransfer.TxType, pc.ExternalTransfer.TxStatus); err != nil {
					fmt.Printf("Warning: failed to write external tx type and status: %v\n", err)
				}
			}
			// Display workflow for this transaction
			if pc.ExternalTransfer.Workflow.WorkflowID != "" {
				if _, err := fmt.Fprintf(w, "   workflow:\n"); err != nil {
					fmt.Printf("Warning: failed to write workflow header: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "      workflow_id=%s\n", pc.ExternalTransfer.Workflow.WorkflowID); err != nil {
					fmt.Printf("Warning: failed to write workflow id: %v\n", err)
				}
				line := fmt.Sprintf("      state=%s attempt=%d", domain.FormatWorkflowState(pc.ExternalTransfer.Workflow.WorkflowID, pc.ExternalTransfer.Workflow.State), pc.ExternalTransfer.Workflow.Attempt)
				if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
					fmt.Printf("Warning: failed to write workflow state: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "      run_id=%s\n", pc.ExternalTransfer.Workflow.RunID); err != nil {
					fmt.Printf("Warning: failed to write workflow run id: %v\n", err)
				}
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

// WriteEcoTransactionResult writes a partnerpay-engine transaction result in the required format
func WriteEcoTransactionResult(w io.Writer, result domain.TransactionResult, index int) {
	if result.PartnerpayEngine != nil {
		WriteEcoTransactionInfo(w, result, result.InputID, index)
	}
}

// WriteEcoTransactionInfo writes a partnerpay-engine transaction info in the required format
func WriteEcoTransactionInfo(w io.Writer, result domain.TransactionResult, transactionID string, index int) {
	if index <= 0 {
		index = 1
	}

	// Check if this is a NOT_FOUND error
	if result.PartnerpayEngine != nil && result.PartnerpayEngine.Charge.Status == domain.NotFoundStatus {
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
	if result.PartnerpayEngine != nil {
		if _, err := fmt.Fprintln(w, "[partnerpay-engine]"); err != nil {
			fmt.Printf("Warning: failed to write partnerpay-engine header: %v\n", err)
		}

		// Write the charge status
		if result.PartnerpayEngine.Charge.Status != "" {
			if _, err := fmt.Fprintf(w, "charge.status: %s", result.PartnerpayEngine.Charge.Status); err != nil {
				fmt.Printf("Warning: failed to write charge status: %v\n", err)
			}

			// If there's a status reason, append it
			if result.PartnerpayEngine.Charge.StatusReason != "" {
				if _, err := fmt.Fprintf(w, " %s", result.PartnerpayEngine.Charge.StatusReason); err != nil {
					fmt.Printf("Warning: failed to write status reason: %v\n", err)
				}
			}

			// If there's a status reason description, append it
			if result.PartnerpayEngine.Charge.StatusReasonDescription != "" {
				if _, err := fmt.Fprintf(w, " %s", result.PartnerpayEngine.Charge.StatusReasonDescription); err != nil {
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

		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline after partnerpay-engine section: %v\n", err)
		}
	}

	// Write payment-core section if available
	if result.PaymentCore != nil {
		if err := displayPaymentCoreSection(w, *result.PaymentCore); err != nil {
			fmt.Printf("Warning: failed to display payment core section: %v\n", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	// Write classification section
	if err := displayClassificationSection(w, result); err != nil {
		fmt.Printf("Warning: failed to display classification section: %v\n", err)
	}

	if _, err := fmt.Fprintln(w); err != nil {
		fmt.Printf("Warning: failed to write final newline: %v\n", err)
	}
}
