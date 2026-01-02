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
	return r.port.QueryByE2EID(inputID)
}

func (r *rppAdapterPopulator) GetAdapterType() string {
	return "RPP"
}

func (r *rppAdapterPopulator) QueryByAccountsAndTimestamp(sourceAccountID, destinationAccountID, timestamp string) (*domain.RPPAdapterInfo, error) {
	return r.port.QueryByAccountsAndTimestamp(sourceAccountID, destinationAccountID, timestamp)
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
	// Fast adapter requires created_at, will be handled in strategy
	return nil, nil
}

func (f *fastAdapterPopulator) GetAdapterType() string {
	return "Fast"
}

// FastAdapterPortWrapper wraps FastAdapterPort to satisfy the interface
type FastAdapterPortWrapper struct {
	port ports.FastAdapterPort
}

// NewFastAdapterPortWrapper creates a new wrapper
func NewFastAdapterPortWrapper(port ports.FastAdapterPort) FastAdapterPort {
	return &FastAdapterPortWrapper{port: port}
}

func (w *FastAdapterPortWrapper) QueryByInstructionID(instructionID, createdAt string) (*domain.FastAdapterInfo, error) {
	return w.port.QueryByInstructionID(instructionID, createdAt)
}
