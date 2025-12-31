package strategies

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"fmt"
)

// FastAdapterPort defines the interface for Fast adapter queries
type FastAdapterPort interface {
	QueryByInstructionID(instructionID, createdAt string) (*domain.FastAdapterInfo, error)
}

// SingaporePopulationStrategy implements the population strategy for Singapore
type SingaporePopulationStrategy struct {
	*BasePopulationStrategy
	fastAdapterPort FastAdapterPort
}

// NewSingaporeStrategy creates a new Singapore population strategy
func NewSingaporeStrategy(
	pePopulator PaymentEnginePopulator,
	pcPopulator PaymentCorePopulator,
	adapterPopulator AdapterPopulator,
	partnerpayPopulator PartnerpayPopulator,
	sopRepo *adapters.SOPRepository,
	fastAdapterPort FastAdapterPort,
) *SingaporePopulationStrategy {
	base := NewBaseStrategy("sg", pePopulator, pcPopulator, adapterPopulator, partnerpayPopulator, sopRepo)
	return &SingaporePopulationStrategy{
		BasePopulationStrategy: base,
		fastAdapterPort:        fastAdapterPort,
	}
}

// GetEnvironment returns the environment
func (s *SingaporePopulationStrategy) GetEnvironment() string {
	return s.env
}

// Populate orchestrates the entire data population process for Singapore
func (s *SingaporePopulationStrategy) Populate(input string) (*domain.TransactionResult, error) {
	result := s.newResultBuilder().
		SetInputID(input).
		Build()

	// Step 1: Determine input type and populate primary data
	if domain.IsRppE2EID(input) && s.adapterPopulator != nil {
		// E2E ID: Query Fast adapter first
		if s.fastAdapterPort != nil {
			fastInfo, err := s.fastAdapterPort.QueryByInstructionID(input, "")
			if err == nil && fastInfo != nil {
				result.FastAdapter = fastInfo
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
		if err := s.populatePaymentEngineFromAdapters(result, "sg"); err != nil {
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
		_ = s.populateAdaptersFromPaymentEngine(result, "sg")

		// Special handling for Fast adapter in Singapore
		if result.FastAdapter == nil && s.fastAdapterPort != nil {
			if result.PaymentEngine.Transfers.ExternalID != "" {
				fastInfo, err := s.fastAdapterPort.QueryByInstructionID(
					result.PaymentEngine.Transfers.ExternalID,
					result.PaymentEngine.Transfers.CreatedAt,
				)
				if err == nil && fastInfo != nil {
					result.FastAdapter = fastInfo
				}
			}
		}
	}

	// Step 5: Identify SOP case
	s.identifyCase(result)

	return result, nil
}
