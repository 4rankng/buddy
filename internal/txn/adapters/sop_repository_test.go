package adapters

import (
	"buddy/internal/txn/domain"
	"testing"
)

func TestPeCaptureProcessingPcCaptureFailedRppSuccess(t *testing.T) {
	// Test case for the specific transaction that should match CasePeCaptureProcessingPcCaptureFailedRppSuccess
	// Based on the provided transaction data:
	// E2E ID: 20251212GXSPMYKL040OQR49782779
	// Payment Engine: state=stCaptureProcessing(230) attempt=0
	// Payment Core Internal Capture: state=stFailed(500) attempt=0, type=CAPTURE, status=FAILED
	// RPP Adapter: state=stSuccess(900) attempt=0, status=PROCESSING

	result := &domain.TransactionResult{
		InputID: "20251212GXSPMYKL040OQR49782779",
		PaymentEngine: &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				Type:        "PAYMENT",
				TxnSubtype:  "RPP_NETWORK",
				TxnDomain:   "DEPOSITS",
				Status:      "FAILED",
				ExternalID:  "20251212GXSPMYKL040OQR49782779",
				ReferenceID: "33FABDFC-C067-45B1-AF50-288D63E508EE",
				CreatedAt:   "2025-12-11T16:10:12.381401Z",
			},
			Workflow: domain.WorkflowInfo{
				WorkflowID: "workflow_transfer_payment",
				State:      "230",
				Attempt:    0,
				RunID:      "33FABDFC-C067-45B1-AF50-288D63E508EE",
			},
		},
		PaymentCore: &domain.PaymentCoreInfo{
			InternalCapture: domain.PCInternalInfo{
				TxID:      "3b7cd88629444d7fa08aa573d92dfe8c",
				GroupID:   "8da1ee9e11a44cf389b7604aea39cfb4",
				TxType:    "CAPTURE",
				TxStatus:  "FAILED",
				CreatedAt: "2025-12-11T16:10:12.381401Z",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "500",
					Attempt:    0,
					RunID:      "3b7cd88629444d7fa08aa573d92dfe8c",
				},
			},
			InternalAuth: domain.PCInternalInfo{
				TxID:      "bf599efb7a37477897e51d7df818ca68",
				GroupID:   "8da1ee9e11a44cf389b7604aea39cfb4",
				TxType:    "AUTH",
				TxStatus:  "SUCCESS",
				CreatedAt: "2025-12-11T16:10:12.381401Z",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "internal_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "bf599efb7a37477897e51d7df818ca68",
				},
			},
			ExternalTransfer: domain.PCExternalInfo{
				RefID:     "7b4ac93e5d1f493fb9c5def1f6fc4f39",
				GroupID:   "8da1ee9e11a44cf389b7604aea39cfb4",
				TxType:    "TRANSFER",
				TxStatus:  "SUCCESS",
				CreatedAt: "2025-12-11T16:10:12.381401Z",
				Workflow: domain.WorkflowInfo{
					WorkflowID: "external_payment_flow",
					State:      "900",
					Attempt:    0,
					RunID:      "7b4ac93e5d1f493fb9c5def1f6fc4f39",
				},
			},
		},
		RPPAdapter: &domain.RPPAdapterInfo{
			ReqBizMsgID: "20251212GXSPMYKL040OQR49782779",
			PartnerTxID: "8da1ee9e11a44cf389b7604aea39cfb4",
			EndToEndID:  "20251212GXSPMYKL040OQR49782779",
			Status:      "PROCESSING",
			CreatedAt:   "2025-12-11T16:10:12.381401Z",
			Workflow: domain.WorkflowInfo{
				WorkflowID: "wf_ct_qr_payment",
				State:      "900",
				Attempt:    0,
				RunID:      "8da1ee9e11a44cf389b7604aea39cfb4",
			},
			Info: "RPP Status: PROCESSING",
		},
		CaseType: domain.CaseNone,
	}

	sopRepo := NewSOPRepository()
	caseType := sopRepo.IdentifyCase(result, "my")

	if caseType != domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess {
		t.Errorf("Expected case type %s, got %s", domain.CasePeCaptureProcessingPcCaptureFailedRppSuccess, caseType)
	}
}
