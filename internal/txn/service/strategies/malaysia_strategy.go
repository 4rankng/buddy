package strategies

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"fmt"
	"log/slog"
)

// MalaysiaPopulationStrategy implements the population strategy for Malaysia
type MalaysiaPopulationStrategy struct {
	*BasePopulationStrategy
}

// NewMalaysiaStrategy creates a new Malaysia population strategy
func NewMalaysiaStrategy(
	pePopulator PaymentEnginePopulator,
	pcPopulator PaymentCorePopulator,
	adapterPopulator AdapterPopulator,
	partnerpayPopulator PartnerpayPopulator,
	sopRepo *adapters.SOPRepository,
) *MalaysiaPopulationStrategy {
	base := NewBaseStrategy("my", pePopulator, pcPopulator, adapterPopulator, partnerpayPopulator, sopRepo)
	return &MalaysiaPopulationStrategy{BasePopulationStrategy: base}
}

// GetEnvironment returns the environment
func (s *MalaysiaPopulationStrategy) GetEnvironment() string {
	return s.env
}

// Populate orchestrates the entire data population process for Malaysia
func (s *MalaysiaPopulationStrategy) Populate(input string) (*domain.TransactionResult, error) {
	result := s.newResultBuilder().
		SetInputID(input).
		Build()

	if domain.IsRppE2EID(input) && s.adapterPopulator != nil {
		adapterData, err := s.adapterPopulator.QueryByInputID(input)
		if err == nil && adapterData != nil {
			if rppInfo, ok := adapterData.(*domain.RPPAdapterInfo); ok {
				result.RPPAdapter = rppInfo

				// Auto-fill related systems from RPP data
				slog.Info("RPP data populated, auto-filling related systems",
					"partnerTxID", rppInfo.PartnerTxID,
					"endToEndID", rppInfo.EndToEndID,
					"inputID", input)

				// Populate PaymentEngine from RPP (uses existing logic)
				_ = s.populatePaymentEngineFromAdapters(result, "my")

				// Populate PaymentCore directly from RPP (new logic)
				_ = s.populatePaymentCoreFromRPP(result)

				// Populate PartnerpayEngine from RPP workflows (new logic)
				_ = s.populatePartnerpayFromRPP(result)

				slog.Info("RPP auto-fill complete",
					"paymentEngineFound", result.PaymentEngine != nil,
					"paymentCoreFound", result.PaymentCore != nil,
					"partnerpayEngineFound", result.PartnerpayEngine != nil)
			}
		}
	} else {
		peInfo, err := s.pePopulator.QueryByTransactionID(input)
		if err != nil {
			result.Error = fmt.Sprintf("failed to query payment engine: %v", err)
			return result, nil
		}
		result.PaymentEngine = peInfo
	}

	// Fallback: if PaymentEngine still nil, try adapters
	if result.PaymentEngine == nil {
		if err := s.populatePaymentEngineFromAdapters(result, "my"); err != nil {
			result.Error = err.Error()
			return result, nil
		}
	}

	// Fallback: populate PaymentCore from PaymentEngine if not yet populated
	if result.PaymentCore == nil && result.PaymentEngine != nil {
		_ = s.populatePaymentCore(result)
	}

	// Fallback: populate adapters from PaymentEngine if not yet populated
	if result.PaymentEngine != nil {
		_ = s.populateAdaptersFromPaymentEngine(result, "my")
	}

	s.identifyCase(result)

	return result, nil
}
