package strategies

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/service/builders"
	"log/slog"
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
		// Payment-engine stores RPP PartnerTxID in transaction_id column, not EndToEndID in external_id
		if result.RPPAdapter.PartnerTxID != "" {
			peInfo, err := s.pePopulator.QueryByTransactionID(
				result.RPPAdapter.PartnerTxID,
			)
			if err == nil && peInfo != nil {
				result.PaymentEngine = peInfo
			}
		} else if result.RPPAdapter.EndToEndID != "" && result.RPPAdapter.CreatedAt != "" {
			// Fallback: use EndToEndID to search when PartnerTxID is empty
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
	if result.PaymentEngine == nil {
		return nil
	}

	if env == "my" && result.RPPAdapter == nil && s.adapterPopulator != nil {
		if result.PaymentEngine.Transfers.ExternalID != "" {
			adapterData, err := s.adapterPopulator.QueryByInputID(
				result.PaymentEngine.Transfers.ExternalID,
			)
			if err == nil && adapterData != nil {
				if rppInfo, ok := adapterData.(*domain.RPPAdapterInfo); ok {
					result.RPPAdapter = rppInfo
					return nil
				}
			}
		}

		if rppPort, ok := s.adapterPopulator.(*rppAdapterPopulator); ok {
			params := domain.RPPQueryParams{
				PartnerTxID:          result.PaymentEngine.Transfers.TransactionID,
				SourceAccountID:      result.PaymentEngine.Transfers.SourceAccountID,
				DestinationAccountID: result.PaymentEngine.Transfers.DestinationAccountID,
				Amount:               result.PaymentEngine.Transfers.Amount,
				Timestamp:            result.PaymentEngine.Transfers.CreatedAt,
			}
			rppInfo, err := rppPort.port.Query(params)
			if err == nil && rppInfo != nil {
				result.RPPAdapter = rppInfo
			}
		}
	}

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

// populatePaymentCoreFromRPP populates PaymentCore directly from RPP adapter data
// This bypasses the need for PaymentEngine data and queries PaymentCore using RPP PartnerTxID
func (s *BasePopulationStrategy) populatePaymentCoreFromRPP(result *domain.TransactionResult) error {
	if s.pcPopulator == nil || result.RPPAdapter == nil {
		return nil
	}

	if result.RPPAdapter.PartnerTxID == "" {
		slog.Debug("Skipping PaymentCore population from RPP: PartnerTxID is empty",
			"inputID", result.InputID)
		return nil
	}

	if result.RPPAdapter.CreatedAt == "" {
		slog.Debug("Skipping PaymentCore population from RPP: CreatedAt is empty",
			"inputID", result.InputID)
		return nil
	}

	transactionID := result.RPPAdapter.PartnerTxID
	createdAt := result.RPPAdapter.CreatedAt

	slog.Info("Populating PaymentCore from RPP",
		"transactionID", transactionID,
		"createdAt", createdAt,
		"inputID", result.InputID)

	// Initialize PaymentCore if nil
	if result.PaymentCore == nil {
		result.PaymentCore = &domain.PaymentCoreInfo{}
	}

	// Query internal transactions (AUTH, CAPTURE)
	internalTxs, err := s.pcPopulator.QueryInternal(transactionID, createdAt)
	if err != nil {
		slog.Warn("Failed to query PaymentCore internal transactions from RPP",
			"error", err,
			"transactionID", transactionID)
	} else if len(internalTxs) > 0 {
		for _, internalTx := range internalTxs {
			switch internalTx.TxType {
			case "AUTH":
				result.PaymentCore.InternalAuth = internalTx
			case "CAPTURE":
				result.PaymentCore.InternalCapture = internalTx
			}
		}
		slog.Info("PaymentCore internal transactions populated from RPP",
			"authFound", result.PaymentCore.InternalAuth.TxID != "",
			"captureFound", result.PaymentCore.InternalCapture.TxID != "")
	}

	// Query external transactions (TRANSFER)
	externalTxs, err := s.pcPopulator.QueryExternal(transactionID, createdAt)
	if err != nil {
		slog.Warn("Failed to query PaymentCore external transactions from RPP",
			"error", err,
			"transactionID", transactionID)
	} else if len(externalTxs) > 0 {
		for _, externalTx := range externalTxs {
			if externalTx.TxType == "TRANSFER" {
				result.PaymentCore.ExternalTransfer = externalTx
				break // Only need first TRANSFER
			}
		}
		slog.Info("PaymentCore external transactions populated from RPP",
			"transferFound", result.PaymentCore.ExternalTransfer.RefID != "")
	}

	return nil
}

// populatePartnerpayFromRPP populates PartnerpayEngine from RPP workflow data
// Uses workflow RunID to query PartnerpayEngine charges
func (s *BasePopulationStrategy) populatePartnerpayFromRPP(result *domain.TransactionResult) error {
	if s.partnerpayPopulator == nil || result.RPPAdapter == nil {
		return nil
	}

	if len(result.RPPAdapter.Workflow) == 0 {
		slog.Debug("Skipping PartnerpayEngine population from RPP: no workflows found",
			"inputID", result.InputID)
		return nil
	}

	slog.Info("Populating PartnerpayEngine from RPP workflows",
		"workflowCount", len(result.RPPAdapter.Workflow),
		"inputID", result.InputID)

	// Query charges for each workflow run_id
	for _, wf := range result.RPPAdapter.Workflow {
		if wf.RunID == "" {
			continue
		}

		chargeInfo, err := s.partnerpayPopulator.QueryCharge(wf.RunID)
		if err != nil {
			slog.Debug("Failed to query PartnerpayEngine charge for workflow",
				"error", err,
				"runID", wf.RunID)
			continue
		}

		if chargeInfo != nil {
			result.PartnerpayEngine = chargeInfo
			slog.Info("PartnerpayEngine populated from RPP workflow",
				"runID", wf.RunID,
				"chargeFound", chargeInfo.Charge.TransactionID != "")
			break // Found charge, no need to continue
		}
	}

	if result.PartnerpayEngine == nil {
		slog.Info("PartnerpayEngine not found from RPP workflows",
			"inputID", result.InputID)
	}

	return nil
}
