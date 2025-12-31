package service

import "time"

// Time windows for data population queries
const (
	// PEExternalIDWindow is the time window for querying PaymentEngine by external ID
	PEExternalIDWindow = 30 * time.Minute

	// PCRegularTxnWindow is the time window for querying PaymentCore regular transactions
	PCRegularTxnWindow = 1 * time.Hour

	// RPPProcessRegistryWindow is the time window for querying RPP process registry workflows
	RPPProcessRegistryWindow = 5 * time.Minute

	// FastInstructionIDWindow is the time window for querying Fast adapter by instruction ID
	FastInstructionIDWindow = 1 * time.Hour
)

// PopulationError represents an error during data population
type PopulationError struct {
	Component string // Component that failed (e.g., "PaymentEngine", "PaymentCore")
	Message   string // Error message
	Err       error  // Underlying error
}

// Error returns the error message
func (e *PopulationError) Error() string {
	if e.Err != nil {
		return e.Component + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Component + ": " + e.Message
}

// Unwrap returns the underlying error for errors.Is/As
func (e *PopulationError) Unwrap() error {
	return e.Err
}

// NewPopulationError creates a new PopulationError
func NewPopulationError(component, message string, err error) *PopulationError {
	return &PopulationError{
		Component: component,
		Message:   message,
		Err:       err,
	}
}
