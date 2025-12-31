package service

import (
	"buddy/internal/clients/doorman"
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	svcAdapters "buddy/internal/txn/service/adapters"
	"buddy/internal/txn/service/population"
	"sync"
)

// AdapterSet contains all adapters needed for transaction queries
type AdapterSet struct {
	PaymentEngine    ports.PaymentEnginePort
	PaymentCore      ports.PaymentCorePort
	RPPAdapter       ports.RPPAdapterPort
	FastAdapter      ports.FastAdapterPort
	PartnerpayEngine ports.PartnerpayEnginePort
}

// TransactionQueryService orchestrates transaction queries across multiple data sources
type TransactionQueryService struct {
	strategy    population.PopulationStrategy
	ecoStrategy population.PopulationStrategy
	adapters    AdapterSet
	sopRepo     *adapters.SOPRepository
	env         string
}

var (
	txnSvc *TransactionQueryService
	once   sync.Once
)

// NewTransactionQueryService creates a new transaction query service singleton
func NewTransactionQueryService(env string) *TransactionQueryService {
	once.Do(func() {
		txnSvc = createTransactionService(env)
	})
	return txnSvc
}

// GetTransactionQueryService returns the singleton instance
func GetTransactionQueryService() *TransactionQueryService {
	if txnSvc == nil {
		panic("TransactionQueryService not initialized. Call NewTransactionQueryService(env) first.")
	}
	return txnSvc
}

// createTransactionService creates a new transaction query service for the given environment
func createTransactionService(env string) *TransactionQueryService {
	var adapterSet AdapterSet

	switch env {
	case "my":
		adapterSet = createMalaysiaAdapters()
	case "sg":
		adapterSet = createSingaporeAdapters()
	default:
		panic("unsupported environment: " + env)
	}

	// Create strategies using the factory
	populationAdapters := population.AdapterSet{
		PaymentEngine:    adapterSet.PaymentEngine,
		PaymentCore:      adapterSet.PaymentCore,
		RPPAdapter:       adapterSet.RPPAdapter,
		FastAdapter:      adapterSet.FastAdapter,
		PartnerpayEngine: adapterSet.PartnerpayEngine,
	}

	strategy := population.NewPopulationStrategy(env, populationAdapters, adapters.SOPRepo)
	ecoStrategy := population.NewEcoStrategy(env, populationAdapters, adapters.SOPRepo)

	return &TransactionQueryService{
		adapters:    adapterSet,
		sopRepo:     adapters.SOPRepo,
		env:         env,
		strategy:    strategy,
		ecoStrategy: ecoStrategy,
	}
}

// createMalaysiaAdapters creates adapters for Malaysia environment
func createMalaysiaAdapters() AdapterSet {
	client := doorman.Doorman
	return AdapterSet{
		PaymentEngine:    svcAdapters.NewPaymentEngineAdapter(client),
		PaymentCore:      svcAdapters.NewPaymentCoreAdapter(client),
		RPPAdapter:       svcAdapters.NewRPPAdapter(client),
		FastAdapter:      svcAdapters.NewFastAdapter(client),
		PartnerpayEngine: svcAdapters.NewPartnerpayEngineAdapter(client),
	}
}

// createSingaporeAdapters creates adapters for Singapore environment
func createSingaporeAdapters() AdapterSet {
	client := doorman.Doorman
	return AdapterSet{
		PaymentEngine:    svcAdapters.NewPaymentEngineAdapter(client),
		PaymentCore:      svcAdapters.NewPaymentCoreAdapter(client),
		RPPAdapter:       nil, // Singapore doesn't use RPP
		FastAdapter:      svcAdapters.NewFastAdapter(client),
		PartnerpayEngine: svcAdapters.NewPartnerpayEngineAdapter(client),
	}
}

// QueryTransaction retrieves complete transaction information by ID
func (s *TransactionQueryService) QueryTransaction(transactionID string) *domain.TransactionResult {
	return s.QueryTransactionWithEnv(transactionID, s.env)
}

// QueryTransactionWithEnv retrieves complete transaction information by ID with specified environment
func (s *TransactionQueryService) QueryTransactionWithEnv(inputID string, env string) *domain.TransactionResult {
	result, err := s.strategy.Populate(inputID)
	if err != nil {
		return &domain.TransactionResult{
			InputID: inputID,
			Error:   err.Error(),
		}
	}
	return result
}

// QueryEcoTransactionWithEnv retrieves ecological transaction information by run_id
func (s *TransactionQueryService) QueryEcoTransactionWithEnv(runID string, env string) *domain.TransactionResult {
	result, err := s.ecoStrategy.Populate(runID)
	if err != nil {
		return &domain.TransactionResult{
			InputID: runID,
			Error:   err.Error(),
		}
	}
	return result
}

// QueryPartnerpayEngine queries the partnerpay-engine database for a transaction by run_id
func (s *TransactionQueryService) QueryPartnerpayEngine(runID string) (domain.PartnerpayEngineInfo, error) {
	return s.adapters.PartnerpayEngine.QueryCharge(runID)
}

// GetAdapters returns the adapter set for this service
func (s *TransactionQueryService) GetAdapters() AdapterSet {
	return s.adapters
}
