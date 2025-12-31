package population

import (
	"buddy/internal/txn/adapters"
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"buddy/internal/txn/service/strategies"
)

// AdapterSet contains all adapters needed for transaction queries
type AdapterSet struct {
	PaymentEngine    ports.PaymentEnginePort
	PaymentCore      ports.PaymentCorePort
	RPPAdapter       ports.RPPAdapterPort
	FastAdapter      ports.FastAdapterPort
	PartnerpayEngine ports.PartnerpayEnginePort
}

// NewPopulationStrategy creates a population strategy based on environment
func NewPopulationStrategy(
	env string,
	adapters AdapterSet,
	sopRepo *adapters.SOPRepository,
) PopulationStrategy {
	// Create populators - these implement interfaces in strategies package
	pePopulator := NewPEPopulator(adapters.PaymentEngine)
	pcPopulator := NewPCPopulator(adapters.PaymentCore)
	partnerpayPopulator := NewPartnerpayPopulator(adapters.PartnerpayEngine)

	var adapterPopulator interface{}
	var fastAdapterPort interface{}

	if env == "my" && adapters.RPPAdapter != nil {
		adapterPopulator = strategies.NewRPPAdapterPopulator(adapters.RPPAdapter)
	} else if env == "sg" && adapters.FastAdapter != nil {
		adapterPopulator = strategies.NewFastAdapterPopulator(adapters.FastAdapter)
		fastAdapterPort = strategies.NewFastAdapterPortWrapper(adapters.FastAdapter)
	}

	// Create strategy based on environment
	var strat interface{}
	switch env {
	case "my":
		strat = strategies.NewMalaysiaStrategy(
			pePopulator,
			pcPopulator,
			adapterPopulator.(strategies.AdapterPopulator),
			partnerpayPopulator,
			sopRepo,
		)
	case "sg":
		strat = strategies.NewSingaporeStrategy(
			pePopulator,
			pcPopulator,
			adapterPopulator.(strategies.AdapterPopulator),
			partnerpayPopulator,
			sopRepo,
			fastAdapterPort.(strategies.FastAdapterPort),
		)
	default:
		panic("unsupported environment: " + env)
	}

	// Wrap the strategy to satisfy PopulationStrategy interface
	return &strategyWrapper{strat: strat}
}

// NewEcoStrategy creates an eco transaction population strategy
func NewEcoStrategy(
	env string,
	adapters AdapterSet,
	sopRepo *adapters.SOPRepository,
) PopulationStrategy {
	pcPopulator := NewPCPopulator(adapters.PaymentCore)
	partnerpayPopulator := NewPartnerpayPopulator(adapters.PartnerpayEngine)

	ecoStrat := strategies.NewEcoStrategy(
		env,
		pcPopulator,
		partnerpayPopulator,
		sopRepo,
	)

	return &strategyWrapper{strat: ecoStrat}
}

// strategyWrapper wraps the strategy implementation to satisfy PopulationStrategy interface
type strategyWrapper struct {
	strat interface{}
}

func (w *strategyWrapper) Populate(input string) (*domain.TransactionResult, error) {
	type populateInterface interface {
		Populate(input string) (*domain.TransactionResult, error)
	}
	return w.strat.(populateInterface).Populate(input)
}

func (w *strategyWrapper) GetEnvironment() string {
	type envInterface interface {
		GetEnvironment() string
	}
	return w.strat.(envInterface).GetEnvironment()
}
