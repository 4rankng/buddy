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

	// Step 1: Determine input type and populate primary data
	if domain.IsRppE2EID(input) && s.adapterPopulator != nil {
		// E2E ID: Query RPP adapter first
		adapterData, err := s.adapterPopulator.QueryByInputID(input)
		if err == nil && adapterData != nil {
			if rppInfo, ok := adapterData.(*domain.RPPAdapterInfo); ok {
				result.RPPAdapter = rppInfo
			}
		}
	} else {
		// Transaction ID: Query PaymentEngine directly
		peInfo, err := s.pePopulator.QueryByTransactionID(input)
		if err != nil {
			result.Error = fmt.Sprintf("failed to query payment engine: %v", err)
			return result, nil
		}
		result.PaymentEngine = peInfo
	}

	// Step 2: Ensure PaymentEngine is populated
	if result.PaymentEngine == nil {
		if err := s.populatePaymentEngineFromAdapters(result, "my"); err != nil {
			result.Error = err.Error()
			return result, nil
		}
	}

	// Step 3: Ensure PaymentCore is populated
	if result.PaymentCore == nil && result.PaymentEngine != nil {
		_ = s.populatePaymentCore(result)
	}

	// Step 4: Ensure adapters are populated from PaymentEngine
	if result.PaymentEngine != nil {
		_ = s.populateAdaptersFromPaymentEngine(result, "my")
	}

	// Step 5: Identify SOP case
	s.identifyCase(result)

	return result, nil
}
