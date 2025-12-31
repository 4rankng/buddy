package strategies

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/service/builders"
	"fmt"
)

// EcoPopulationStrategy implements the population strategy for ecological transactions
type EcoPopulationStrategy struct {
	pcPopulator         PaymentCorePopulator
	partnerpayPopulator PartnerpayPopulator
	sopRepo             *adapters.SOPRepository
	env                 string
}

// NewEcoStrategy creates a new eco population strategy
func NewEcoStrategy(
	env string,
	pcPopulator PaymentCorePopulator,
	partnerpayPopulator PartnerpayPopulator,
	sopRepo *adapters.SOPRepository,
) *EcoPopulationStrategy {
	return &EcoPopulationStrategy{
		pcPopulator:         pcPopulator,
		partnerpayPopulator: partnerpayPopulator,
		sopRepo:             sopRepo,
		env:                 env,
	}
}

// GetEnvironment returns the environment
func (s *EcoPopulationStrategy) GetEnvironment() string {
	return s.env
}

// Populate orchestrates the entire data population process for eco transactions
func (s *EcoPopulationStrategy) Populate(input string) (*domain.TransactionResult, error) {
	result := s.newResultBuilder().
		SetInputID(input).
		Build()

	// Step 1: Fill PartnerpayEngine first for ecological transactions
	if s.partnerpayPopulator != nil {
		ppeInfo, err := s.partnerpayPopulator.QueryCharge(input)
		if err == nil && ppeInfo != nil {
			result.PartnerpayEngine = ppeInfo
		} else {
			result.Error = fmt.Sprintf("failed to query partnerpay engine: %v", err)
			return result, nil
		}
	}

	// Step 2: Populate PaymentCore using group_id (InputID)
	if result.PartnerpayEngine != nil && s.pcPopulator != nil {
		s.populatePaymentCoreEco(result)
	}

	// Step 3: Identify SOP case
	s.identifyCase(result)

	return result, nil
}

// populatePaymentCoreEco queries payment-core using partnerpay-engine transaction info
func (s *EcoPopulationStrategy) populatePaymentCoreEco(result *domain.TransactionResult) {
	if result.PartnerpayEngine == nil || s.pcPopulator == nil {
		return
	}

	// Initialize PaymentCore if nil
	if result.PaymentCore == nil {
		result.PaymentCore = &domain.PaymentCoreInfo{}
	}

	// Use the original run_id (InputID) as group_id to query payment-core
	groupID := result.InputID
	createdAt := result.PartnerpayEngine.Charge.CreatedAt

	if groupID == "" || createdAt == "" {
		return
	}

	// Query internal transactions
	internalTxs, err := s.pcPopulator.QueryInternal(groupID, createdAt)
	if err == nil {
		for _, internalTx := range internalTxs {
			switch internalTx.TxType {
			case "AUTH":
				result.PaymentCore.InternalAuth = internalTx
			case "CAPTURE":
				result.PaymentCore.InternalCapture = internalTx
			}
		}
	}

	// Query external transactions
	externalTxs, err := s.pcPopulator.QueryExternal(groupID, createdAt)
	if err == nil {
		for _, externalTx := range externalTxs {
			if externalTx.TxType == "TRANSFER" {
				result.PaymentCore.ExternalTransfer = externalTx
				break // Only need first TRANSFER
			}
		}
	}
}

// identifyCase identifies the SOP case for the transaction
func (s *EcoPopulationStrategy) identifyCase(result *domain.TransactionResult) {
	if s.sopRepo != nil {
		s.sopRepo.IdentifyCase(result, s.env)
	}
}

// newResultBuilder creates a new result builder
func (s *EcoPopulationStrategy) newResultBuilder() *builders.TransactionResultBuilder {
	return builders.NewResultBuilder()
}
