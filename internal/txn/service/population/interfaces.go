package population

import "buddy/internal/txn/domain"

// PaymentEnginePopulator handles PaymentEngine data population
type PaymentEnginePopulator interface {
	// QueryByTransactionID fetches transfer and workflow by transaction ID
	QueryByTransactionID(transactionID string) (*domain.PaymentEngineInfo, error)

	// QueryByExternalID fetches transfer by external ID within time window
	QueryByExternalID(externalID, createdAt string) (*domain.PaymentEngineInfo, error)

	// QueryWorkflow fetches workflow by reference ID
	QueryWorkflow(referenceID string) (*domain.WorkflowInfo, error)
}

// PaymentCorePopulator handles PaymentCore data population
type PaymentCorePopulator interface {
	// QueryInternal fetches internal transactions (AUTH, CAPTURE)
	QueryInternal(transactionID, createdAt string) ([]domain.PCInternalInfo, error)

	// QueryExternal fetches external transactions (TRANSFER)
	QueryExternal(transactionID, createdAt string) ([]domain.PCExternalInfo, error)

	// QueryWorkflow fetches workflow by run ID
	QueryWorkflow(runID string) domain.WorkflowInfo
}

// AdapterPopulator handles RPP/Fast adapter data population
type AdapterPopulator interface {
	// QueryByInputID fetches adapter data by input ID
	QueryByInputID(inputID string) (interface{}, error)

	// GetAdapterType returns the adapter type ("RPP" or "Fast")
	GetAdapterType() string
}

// PartnerpayPopulator handles PartnerpayEngine data population
type PartnerpayPopulator interface {
	// QueryCharge fetches charge information by run ID
	QueryCharge(runID string) (*domain.PartnerpayEngineInfo, error)
}

// PopulationStrategy defines the strategy for populating transaction data
type PopulationStrategy interface {
	// Populate orchestrates the entire data population process
	Populate(input string) (*domain.TransactionResult, error)

	// GetEnvironment returns the environment ("my" or "sg")
	GetEnvironment() string
}
