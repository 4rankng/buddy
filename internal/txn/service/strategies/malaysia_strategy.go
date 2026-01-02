package strategies

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"fmt"
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

	if result.PaymentEngine == nil {
		if err := s.populatePaymentEngineFromAdapters(result, "my"); err != nil {
			result.Error = err.Error()
			return result, nil
		}
	}

	if result.PaymentCore == nil && result.PaymentEngine != nil {
		_ = s.populatePaymentCore(result)
	}

	if result.PaymentEngine != nil {
		_ = s.populateAdaptersFromPaymentEngine(result, "my")
	}

	s.identifyCase(result)

	return result, nil
}
