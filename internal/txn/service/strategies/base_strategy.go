package strategies

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/service/builders"
)

// PaymentEnginePopulator handles PaymentEngine data population
type PaymentEnginePopulator interface {
	QueryByTransactionID(transactionID string) (*domain.PaymentEngineInfo, error)
	QueryByExternalID(externalID, createdAt string) (*domain.PaymentEngineInfo, error)
	QueryWorkflow(referenceID string) (*domain.WorkflowInfo, error)
}

// PaymentCorePopulator handles PaymentCore data population
type PaymentCorePopulator interface {
	QueryInternal(transactionID, createdAt string) ([]domain.PCInternalInfo, error)
	QueryExternal(transactionID, createdAt string) ([]domain.PCExternalInfo, error)
	QueryWorkflow(runID string) domain.WorkflowInfo
}

// AdapterPopulator handles RPP/Fast adapter data population
type AdapterPopulator interface {
	QueryByInputID(inputID string) (interface{}, error)
	GetAdapterType() string
}

// PartnerpayPopulator handles PartnerpayEngine data population
type PartnerpayPopulator interface {
	QueryCharge(runID string) (*domain.PartnerpayEngineInfo, error)
}

// BasePopulationStrategy contains shared logic for Malaysia and Singapore strategies
type BasePopulationStrategy struct {
	env                 string
	pePopulator         PaymentEnginePopulator
	pcPopulator         PaymentCorePopulator
	adapterPopulator    AdapterPopulator
	partnerpayPopulator PartnerpayPopulator
	sopRepo             *adapters.SOPRepository
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy(
	env string,
	pePopulator PaymentEnginePopulator,
	pcPopulator PaymentCorePopulator,
	adapterPopulator AdapterPopulator,
	partnerpayPopulator PartnerpayPopulator,
	sopRepo *adapters.SOPRepository,
) *BasePopulationStrategy {
	return &BasePopulationStrategy{
		env:                 env,
		pePopulator:         pePopulator,
		pcPopulator:         pcPopulator,
		adapterPopulator:    adapterPopulator,
		partnerpayPopulator: partnerpayPopulator,
		sopRepo:             sopRepo,
	}
}

// GetEnvironment returns the environment
func (s *BasePopulationStrategy) GetEnvironment() string {
	return s.env
}

// populatePaymentEngineFromAdapters populates PaymentEngine from RPP/Fast adapters
func (s *BasePopulationStrategy) populatePaymentEngineFromAdapters(result *domain.TransactionResult, env string) error {
	if s.pePopulator == nil {
		return nil
	}

	if env == "my" && result.RPPAdapter != nil {
		// Use RPP adapter data
		if result.RPPAdapter.EndToEndID != "" && result.RPPAdapter.CreatedAt != "" {
			peInfo, err := s.pePopulator.QueryByExternalID(
				result.RPPAdapter.EndToEndID,
				result.RPPAdapter.CreatedAt,
			)
			if err == nil && peInfo != nil {
				result.PaymentEngine = peInfo
			}
		}
	} else if env == "sg" && result.FastAdapter != nil {
		// Use Fast adapter data
		if result.FastAdapter.InstructionID != "" && result.FastAdapter.CreatedAt != "" {
			peInfo, err := s.pePopulator.QueryByExternalID(
				result.FastAdapter.InstructionID,
				result.FastAdapter.CreatedAt,
			)
			if err == nil && peInfo != nil {
				result.PaymentEngine = peInfo
			}
		}
	}

	return nil
}

// populateAdaptersFromPaymentEngine populates adapters from PaymentEngine
func (s *BasePopulationStrategy) populateAdaptersFromPaymentEngine(result *domain.TransactionResult, env string) error {
	if result.PaymentEngine == nil || result.PaymentEngine.Transfers.ExternalID == "" {
		return nil
	}

	if env == "my" && result.RPPAdapter == nil && s.adapterPopulator != nil {
		// Query RPP adapter using external_id
		adapterData, err := s.adapterPopulator.QueryByInputID(
			result.PaymentEngine.Transfers.ExternalID,
		)
		if err == nil && adapterData != nil {
			if rppInfo, ok := adapterData.(*domain.RPPAdapterInfo); ok {
				result.RPPAdapter = rppInfo
			}
		}
	}
	// For Singapore, Fast adapter is handled in the concrete strategy implementation

	return nil
}

// populatePaymentCore populates PaymentCore from PaymentEngine
func (s *BasePopulationStrategy) populatePaymentCore(result *domain.TransactionResult) error {
	if s.pcPopulator == nil || result.PaymentEngine == nil {
		return nil
	}

	transactionID := result.PaymentEngine.Transfers.TransactionID
	if transactionID == "" {
		transactionID = result.InputID
	}

	createdAt := result.PaymentEngine.Transfers.CreatedAt
	if createdAt == "" {
		return nil
	}

	// Initialize PaymentCore if nil
	if result.PaymentCore == nil {
		result.PaymentCore = &domain.PaymentCoreInfo{}
	}

	// Query internal transactions
	internalTxs, err := s.pcPopulator.QueryInternal(transactionID, createdAt)
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
	externalTxs, err := s.pcPopulator.QueryExternal(transactionID, createdAt)
	if err == nil {
		for _, externalTx := range externalTxs {
			if externalTx.TxType == "TRANSFER" {
				result.PaymentCore.ExternalTransfer = externalTx
				break // Only need first TRANSFER
			}
		}
	}

	return nil
}

// identifyCase identifies the SOP case for the transaction
func (s *BasePopulationStrategy) identifyCase(result *domain.TransactionResult) {
	if s.sopRepo != nil {
		s.sopRepo.IdentifyCase(result, s.env)
	}
}

// newResultBuilder creates a new result builder
func (s *BasePopulationStrategy) newResultBuilder() *builders.TransactionResultBuilder {
	return builders.NewResultBuilder()
}
