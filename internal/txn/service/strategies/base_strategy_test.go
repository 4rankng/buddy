package strategies

import (
	"buddy/internal/txn/domain"
	"errors"
	"testing"
)

// MockPaymentEnginePopulator is a mock implementation of PaymentEnginePopulator
type MockPaymentEnginePopulator struct {
	QueryByTransactionIDFunc func(transactionID string) (*domain.PaymentEngineInfo, error)
	QueryByExternalIDFunc    func(externalID, createdAt string) (*domain.PaymentEngineInfo, error)
	QueryWorkflowFunc        func(referenceID string) (*domain.WorkflowInfo, error)
}

func (m *MockPaymentEnginePopulator) QueryByTransactionID(transactionID string) (*domain.PaymentEngineInfo, error) {
	if m.QueryByTransactionIDFunc != nil {
		return m.QueryByTransactionIDFunc(transactionID)
	}
	return nil, nil
}

func (m *MockPaymentEnginePopulator) QueryByExternalID(externalID, createdAt string) (*domain.PaymentEngineInfo, error) {
	if m.QueryByExternalIDFunc != nil {
		return m.QueryByExternalIDFunc(externalID, createdAt)
	}
	return nil, nil
}

func (m *MockPaymentEnginePopulator) QueryWorkflow(referenceID string) (*domain.WorkflowInfo, error) {
	if m.QueryWorkflowFunc != nil {
		return m.QueryWorkflowFunc(referenceID)
	}
	return nil, nil
}

func TestPopulatePaymentEngineFromAdapters_RPP_PartnerTxID(t *testing.T) {
	// Setup
	mockPE := &MockPaymentEnginePopulator{}
	strategy := NewBaseStrategy("my", mockPE, nil, nil, nil, nil)

	// Case 1: RPP Adapter has PartnerTxID, and we find PE data using it
	t.Run("FoundByPartnerTxID", func(t *testing.T) {
		partnerTxID := "partner_123"

		// Input result with RPP adapter data but no PE data yet
		result := &domain.TransactionResult{
			InputID: "input_123",
			RPPAdapter: &domain.RPPAdapterInfo{
				PartnerTxID: partnerTxID,
				// EndToEndID might be missing or lookup failed
				EndToEndID: "e2e_123",
			},
		}

		// Mock PE behavior
		// 1. QueryByExternalID fails or returns nil (simulating current failure)
		mockPE.QueryByExternalIDFunc = func(externalID, createdAt string) (*domain.PaymentEngineInfo, error) {
			return nil, nil // Not found by External ID
		}

		// 2. QueryByTransactionID returns success (this is what we want to verify is called)
		expectedPE := &domain.PaymentEngineInfo{
			Transfers: domain.PETransfersInfo{
				TransactionID: partnerTxID,
				Status:        "SUCCESS",
			},
		}
		mockPE.QueryByTransactionIDFunc = func(transactionID string) (*domain.PaymentEngineInfo, error) {
			t.Logf("QueryByTransactionID called with %s", transactionID)
			if transactionID == partnerTxID {
				return expectedPE, nil
			}
			return nil, errors.New("unexpected transaction ID")
		}

		// Execute
		err := strategy.populatePaymentEngineFromAdapters(result, "my")
		t.Logf("Result PE: %v", result.PaymentEngine)

		// Verify
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.PaymentEngine == nil {
			t.Error("PaymentEngine should have been populated")
		} else if result.PaymentEngine.Transfers.TransactionID != partnerTxID {
			t.Errorf("expected PE TransactionID %s, got %s", partnerTxID, result.PaymentEngine.Transfers.TransactionID)
		}
	})

	// Case 2: Already found by ExternalID, shouldn't overwrite/query again unnecessarily
	// (Optional: depending on implementation preference, but good for stability)
	// For now, let's focus on the feature: adding the fallback.
}
