package ports

import "buddy/internal/txn/domain"

// Port interfaces for data sources

// PaymentEnginePort defines the interface for payment engine queries
type PaymentEnginePort interface {
	QueryTransfer(transactionID string) (map[string]interface{}, error)
	QueryWorkflow(referenceID string) (map[string]interface{}, error)
	QueryTransferByExternalID(externalID, createdAt string) (map[string]interface{}, error)
}

// PaymentCorePort defines the interface for payment core queries
type PaymentCorePort interface {
	QueryInternalTransactions(transactionID string, createdAt string) ([]map[string]interface{}, error)
	QueryExternalTransactions(transactionID string, createdAt string) ([]map[string]interface{}, error)
	QueryWorkflows(runIDs []string) ([]map[string]interface{}, error)
	QueryPaymentCore(query string) ([]map[string]interface{}, error)
}

// RPPAdapterPort defines the interface for RPP adapter queries
type RPPAdapterPort interface {
	QueryByE2EID(externalID string) (*domain.RPPAdapterInfo, error)
}

// FastAdapterPort defines the interface for fast adapter queries
type FastAdapterPort interface {
	QueryByInstructionID(instructionID, createdAt string) (*domain.FastAdapterInfo, error)
}

// PartnerpayEnginePort defines the interface for partnerpay engine queries
type PartnerpayEnginePort interface {
	QueryCharge(transactionID string) (domain.PartnerpayEngineInfo, error)
}

// ClientPort defines the interface for database client queries
type ClientPort interface {
	QueryPaymentEngine(query string) ([]map[string]interface{}, error)
	QueryPaymentCore(query string) ([]map[string]interface{}, error)
	QueryRppAdapter(query string) ([]map[string]interface{}, error)
	QueryFastAdapter(query string) ([]map[string]interface{}, error)
	QueryPartnerpayEngine(query string) ([]map[string]interface{}, error)
	ExecuteQuery(cluster, service, database, query string) ([]map[string]interface{}, error)
}

// SOPRepositoryPort defines the interface for SOP repository operations
type SOPRepositoryPort interface {
	IdentifyCase(result *domain.TransactionResult, env string)
}
