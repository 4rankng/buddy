package payment

// Module is the payment service module
type Module struct{}

// New creates a new payment service module
func New() *Module {
	return &Module{}
}