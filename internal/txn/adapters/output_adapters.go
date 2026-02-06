package adapters

import (
	"buddy/internal/txn/domain"
	"fmt"
	"io"
)

// displayFastAdapterSection displays FastAdapter section
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

// displayRPPAdapterSection displays RPPAdapter section
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
		// Display all workflows
		for _, wf := range ra.Workflow {
			if wf.RunID != "" {
				if _, err := fmt.Fprintf(w, "%s:\n", wf.WorkflowID); err != nil {
					fmt.Printf("Warning: failed to write rpp workflow header: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "   state=%s attempt=%d\n", wf.GetFormattedState(), wf.Attempt); err != nil {
					fmt.Printf("Warning: failed to write rpp workflow state: %v\n", err)
				}
				if _, err := fmt.Fprintf(w, "   run_id=%s\n", wf.RunID); err != nil {
					fmt.Printf("Warning: failed to write rpp workflow run id: %v\n", err)
				}
			}
		}
		return nil
	}

	// For non-E2E IDs, check if we have data
	hasData := len(ra.Workflow) > 0 || ra.ReqBizMsgID != "" || ra.PartnerTxID != "" || ra.Info != ""
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
	if ra.PartnerMsgID != "" {
		if _, err := fmt.Fprintf(w, "partner_msg_id: %s\n", ra.PartnerMsgID); err != nil {
			fmt.Printf("Warning: failed to write partner msg id: %v\n", err)
		}
	}
	// Display all workflows
	for _, wf := range ra.Workflow {
		if wf.RunID != "" {
			if _, err := fmt.Fprintf(w, "%s:\n", wf.WorkflowID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow header: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   state=%s attempt=%d\n", wf.GetFormattedState(), wf.Attempt); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow state: %v\n", err)
			}
			if _, err := fmt.Fprintf(w, "   run_id=%s\n", wf.RunID); err != nil {
				fmt.Printf("Warning: failed to write rpp workflow run id: %v\n", err)
			}
		}
	}
	if ra.Info != "" {
		if _, err := fmt.Fprintf(w, "info: %s\n", ra.Info); err != nil {
			fmt.Printf("Warning: failed to write rpp info: %v\n", err)
		}
	}

	return nil
}
