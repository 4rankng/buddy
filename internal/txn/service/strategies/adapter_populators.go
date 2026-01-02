package strategies

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
)

// RPP Adapter Populator (Malaysia)
type rppAdapterPopulator struct {
	port ports.RPPAdapterPort
}

// NewRPPAdapterPopulator creates a new RPP adapter populator
func NewRPPAdapterPopulator(port ports.RPPAdapterPort) AdapterPopulator {
	return &rppAdapterPopulator{port: port}
}

func (r *rppAdapterPopulator) QueryByInputID(inputID string) (interface{}, error) {
	return r.port.Query(domain.RPPQueryParams{EndToEndID: inputID})
}

func (r *rppAdapterPopulator) GetAdapterType() string {
	return "RPP"
}

// Fast Adapter Populator (Singapore)
type fastAdapterPopulator struct {
	port ports.FastAdapterPort
}

// NewFastAdapterPopulator creates a new Fast adapter populator
func NewFastAdapterPopulator(port ports.FastAdapterPort) AdapterPopulator {
	return &fastAdapterPopulator{port: port}
}

func (f *fastAdapterPopulator) QueryByInputID(inputID string) (interface{}, error) {
	return nil, nil
}

func (f *fastAdapterPopulator) GetAdapterType() string {
	return "Fast"
}

// FastAdapterPort wrapper for strategies
type FastAdapterPort interface {
	Query(params domain.FastQueryParams) (*domain.FastAdapterInfo, error)
}

// FastAdapterPortWrapper wraps ports.FastAdapterPort
type FastAdapterPortWrapper struct {
	port ports.FastAdapterPort
}

// NewFastAdapterPortWrapper creates a new wrapper
func NewFastAdapterPortWrapper(port ports.FastAdapterPort) FastAdapterPort {
	return &FastAdapterPortWrapper{port: port}
}

func (w *FastAdapterPortWrapper) Query(params domain.FastQueryParams) (*domain.FastAdapterInfo, error) {
	return w.port.Query(params)
}
