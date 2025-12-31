package population

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"fmt"
)

type partnerpayPopulator struct {
	port ports.PartnerpayEnginePort
}

// NewPartnerpayPopulator creates a new PartnerpayPopulator
func NewPartnerpayPopulator(port ports.PartnerpayEnginePort) PartnerpayPopulator {
	return &partnerpayPopulator{port: port}
}

// QueryCharge fetches charge information by run ID
func (p *partnerpayPopulator) QueryCharge(runID string) (*domain.PartnerpayEngineInfo, error) {
	if p.port == nil {
		return nil, fmt.Errorf("partnerpay engine not available")
	}

	info, err := p.port.QueryCharge(runID)
	if err != nil {
		return nil, err
	}

	return &info, nil
}
